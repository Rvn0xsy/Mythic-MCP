package server

import (
	"context"
	"fmt"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nbaertsch/Mythic-MCP/pkg/config"
	"github.com/nbaertsch/mythic-sdk-go/pkg/mythic"
)

// Server is the main MCP server wrapping Mythic SDK
type Server struct {
	config       *config.Config
	mcpServer    *mcp.Server
	mythicClient *mythic.Client
}

// NewServer creates a new MCP server with Mythic integration
func NewServer(cfg *config.Config) (*Server, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	// Create Mythic SDK client configuration
	mythicConfig := &mythic.Config{
		ServerURL:     cfg.MythicURL,
		APIToken:      cfg.APIToken,
		Username:      cfg.Username,
		Password:      cfg.Password,
		SSL:           cfg.SSL,
		SkipTLSVerify: cfg.SkipTLSVerify,
		Timeout:       cfg.Timeout,
	}

	// Create Mythic client
	mythicClient, err := mythic.NewClient(mythicConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Mythic client: %w", err)
	}

	// Create MCP server
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "mythic-mcp",
		Version: "1.0.0",
	}, nil)

	server := &Server{
		config:       cfg,
		mcpServer:    mcpServer,
		mythicClient: mythicClient,
	}

	// Register MCP tools
	if err := server.registerTools(); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}

	return server, nil
}

// Run starts the MCP server
func (s *Server) Run(ctx context.Context, transport mcp.Transport) error {
	// Authenticate with Mythic
	log.Println("Authenticating with Mythic...")
	if err := s.mythicClient.Login(ctx); err != nil {
		return fmt.Errorf("failed to authenticate with Mythic: %w", err)
	}
	log.Println("Successfully authenticated with Mythic")

	// Run MCP server with transport
	log.Println("Starting MCP server...")
	return s.mcpServer.Run(ctx, transport)
}

// Close cleans up server resources
func (s *Server) Close() error {
	if s.mythicClient != nil {
		return s.mythicClient.Close()
	}
	return nil
}

// registerTools registers all MCP tools
func (s *Server) registerTools() error {
	// Phase 1: Authentication tools
	s.registerAuthTools()

	// Future phases will add more tool categories:
	// - Operations tools (Phase 2)
	// - Callback tools (Phase 3)
	// - Task tools (Phase 3)
	// - Payload tools (Phase 3)
	// - File tools (Phase 2)
	// - etc.

	return nil
}
