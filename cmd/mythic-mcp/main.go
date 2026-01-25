package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nbaertsch/Mythic-MCP/pkg/config"
	"github.com/nbaertsch/Mythic-MCP/pkg/server"
)

var (
	// Version is set via ldflags during build
	Version = "dev"
)

func main() {
	// Handle version command
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("mythic-mcp version %s\n", Version)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
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

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received shutdown signal, stopping server...")
		cancel()
	}()

	// Create stdio transport for MCP
	transport := &mcp.StdioTransport{}

	// Run server
	log.Println("Starting Mythic MCP Server...")
	if err := srv.Run(ctx, transport); err != nil {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Server stopped")
}
