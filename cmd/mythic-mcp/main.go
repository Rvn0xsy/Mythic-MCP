package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nbaertsch/Mythic-MCP/pkg/config"
	"github.com/nbaertsch/Mythic-MCP/pkg/server"
)

var (
	// Version is set via ldflags during build
	Version = "dev"
)

func main() {
	// Parse flags
	configFile := flag.String("config", "", "Path to config.toml (overrides MCP_CONFIG_FILE and default)")
	flag.Parse()

	// Handle version command
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("mythic-mcp version %s\n", Version)
		os.Exit(0)
	}

	// -config flag takes precedence
	if *configFile != "" {
		os.Setenv("MCP_CONFIG_FILE", *configFile)
	}

	// Load configuration (config.toml if present, then environment variables)
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Auto-detect file vending base URL from HTTP listen address if not set
	if cfg.FileVendingEnabled && cfg.FileVendingBaseURL == "" {
		port := os.Getenv("MCP_HTTP_PORT")
		if port == "" {
			port = "3333"
		}
		httpAddr := os.Getenv("MCP_HTTP_ADDR")
		if httpAddr == "" {
			httpAddr = "0.0.0.0:" + port
		}
		cfg.FileVendingBaseURL = fmt.Sprintf("http://%s", httpAddr)
	}

	// Create server
	srv, err := server.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Close()

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Authenticate with Mythic before accepting connections
	if err := srv.Authenticate(ctx); err != nil {
		log.Fatalf("Mythic authentication failed: %v", err)
	}

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received shutdown signal, stopping server...")
		cancel()
	}()

	// Choose transport based on MCP_TRANSPORT env var (default: "http")
	transport := os.Getenv("MCP_TRANSPORT")
	if transport == "" {
		transport = "http"
	}

	switch transport {
	case "stdio":
		// Legacy: Run over stdio (for local/CLI use)
		log.Println("Starting Mythic MCP Server (stdio transport)...")
		if err := srv.Run(ctx, &mcp.StdioTransport{}); err != nil {
			log.Fatalf("Server error: %v", err)
		}

	case "http":
		// Determine listen address
		addr := os.Getenv("MCP_HTTP_ADDR")
		if addr == "" {
			port := os.Getenv("MCP_HTTP_PORT")
			if port == "" {
				port = "3333"
			}
			addr = "0.0.0.0:" + port
		}

		// Create Streamable HTTP handler (MCP 2025-03-26 spec)
		var mcpHandler http.Handler = mcp.NewStreamableHTTPHandler(
			func(r *http.Request) *mcp.Server { return srv.MCPServer() },
			nil, // default StreamableHTTPOptions
		)

		// Add bearer token auth if MCP_AUTH_TOKEN is configured
		if cfg.AuthToken != "" {
			verifier := auth.TokenVerifier(func(ctx context.Context, token string, req *http.Request) (*auth.TokenInfo, error) {
				if token != cfg.AuthToken {
					return nil, auth.ErrInvalidToken
				}
				return &auth.TokenInfo{
					UserID:    "mcp",
					Scopes:    []string{"mcp"},
					Expiration: time.Now().Add(24 * time.Hour),
				}, nil
			})
			mcpHandler = auth.RequireBearerToken(verifier, nil)(mcpHandler)
			log.Println("Bearer token auth enabled for /mcp endpoint")
		}

		mux := http.NewServeMux()
		mux.Handle("/mcp", mcpHandler)

		// File vending download endpoint
		if fs := srv.FileStore(); fs != nil {
			mux.HandleFunc("/download/", fs.ServeDownload)
			log.Println("File vending enabled — download endpoint: /download/{file_id}?token=...")
		}

		// Health check endpoint (no auth required)
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		})

		httpServer := &http.Server{
			Addr:    addr,
			Handler: mux,
		}

		// Graceful shutdown
		go func() {
			<-ctx.Done()
			log.Println("Shutting down HTTP server...")
			httpServer.Close()
		}()

		// TLS or plain HTTP
		if cfg.TLSCertFile != "" && cfg.TLSKeyFile != "" {
			log.Printf("Starting Mythic MCP Server (HTTPS) on %s", addr)
			log.Printf("  MCP endpoint:    https://%s/mcp", addr)
			log.Printf("  Health check:    https://%s/healthz", addr)
			if err := httpServer.ListenAndServeTLS(cfg.TLSCertFile, cfg.TLSKeyFile); err != nil && err != http.ErrServerClosed {
				log.Fatalf("HTTPS server error: %v", err)
			}
		} else {
			log.Printf("Starting Mythic MCP Server (HTTP) on %s", addr)
			log.Printf("  MCP endpoint:    http://%s/mcp", addr)
			log.Printf("  Health check:    http://%s/healthz", addr)
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("HTTP server error: %v", err)
			}
		}

	default:
		log.Fatalf("Unknown MCP_TRANSPORT=%q (supported: http, stdio)", transport)
	}

	log.Println("Server stopped")
}
