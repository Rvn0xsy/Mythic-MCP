# Mythic MCP Server Architecture

**Author:** Claude Code
**Date:** 2026-01-24
**Purpose:** Design architecture for MCP server wrapping Mythic Go SDK
**Philosophy:** CI-First Development

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Architecture Overview](#architecture-overview)
3. [Component Design](#component-design)
4. [MCP Protocol Integration](#mcp-protocol-integration)
5. [Authentication & Security](#authentication--security)
6. [Error Handling](#error-handling)
7. [Project Structure](#project-structure)

---

## Executive Summary

### Project Goal

Build a production-grade Model Context Protocol (MCP) server that exposes all Mythic C2 Framework operations as MCP tools, enabling AI assistants to interact with Mythic for red team operations.

### Key Principles

1. **Comprehensive Coverage** - Wrap all 204+ Mythic SDK methods as MCP tools
2. **Type Safety** - Leverage Go's strong typing throughout
3. **Production Ready** - >90% test coverage, comprehensive error handling
4. **CI-First** - All functionality validated through integration tests against real Mythic
5. **Zero Data Loss** - All Mythic SDK capabilities preserved, no simplification

### Architecture Layers

```
┌─────────────────────────────────────────┐
│      AI Assistant (Claude Desktop)      │
│                                          │
└────────────────┬────────────────────────┘
                 │ MCP Protocol (JSON-RPC)
                 │
┌────────────────▼────────────────────────┐
│          MCP Server (this project)      │
│  ┌──────────────────────────────────┐   │
│  │  MCP Protocol Handler            │   │
│  │  - Tool registration             │   │
│  │  - Resource management           │   │
│  │  - Prompt templates              │   │
│  └────────────┬─────────────────────┘   │
│               │                          │
│  ┌────────────▼─────────────────────┐   │
│  │  Mythic Tool Wrappers            │   │
│  │  - Input validation              │   │
│  │  - Output formatting             │   │
│  │  - Error translation             │   │
│  └────────────┬─────────────────────┘   │
└───────────────┼──────────────────────────┘
                │
┌───────────────▼──────────────────────────┐
│      Mythic Go SDK (upstream)            │
│  - GraphQL client                        │
│  - WebSocket subscriptions               │
│  - Authentication                        │
│  - Type definitions                      │
└────────────────┬─────────────────────────┘
                 │ GraphQL + WebSocket
                 │
┌────────────────▼─────────────────────────┐
│      Mythic C2 Framework Instance        │
│  - PostgreSQL database                   │
│  - RabbitMQ messaging                    │
│  - Docker containers                     │
└──────────────────────────────────────────┘
```

---

## Architecture Overview

### System Components

**1. MCP Server Core**
- Implements official Go MCP SDK (`github.com/modelcontextprotocol/go-sdk`)
- Registers all Mythic operations as MCP tools
- Manages client lifecycle and state
- Handles JSON-RPC communication

**2. Mythic Client Manager**
- Singleton Mythic SDK client instance
- Connection pooling and lifecycle management
- Authentication state tracking
- Operation context management

**3. Tool Wrappers**
- One wrapper per Mythic SDK method category
- Input validation using MCP tool schemas
- Output formatting for AI consumption
- Error translation to user-friendly messages

**4. Configuration Management**
- Environment-based configuration
- Secure credential handling
- Connection parameters
- Feature flags for development

### Design Decisions

| Decision | Rationale |
|----------|-----------|
| Use official MCP Go SDK | Production support, maintained by Google + Anthropic |
| Thin wrapper pattern | Preserve all Mythic SDK functionality, no logic duplication |
| Stateful server | Maintain authenticated session, reuse connections |
| Comprehensive tool coverage | Enable full Mythic automation from AI |
| Synchronous tools | Match Mythic SDK semantics, WebSocket for async events |

---

## Component Design

### 1. MCP Server (`pkg/server/server.go`)

```go
package server

import (
    "context"
    "github.com/modelcontextprotocol/go-sdk/mcp"
    "github.com/nbaertsch/mythic-sdk-go/pkg/mythic"
)

// Server is the main MCP server wrapping Mythic SDK
type Server struct {
    mcpServer     *mcp.Server
    mythicClient  *mythic.Client
    config        *Config
}

// NewServer creates a new MCP server with Mythic integration
func NewServer(config *Config) (*Server, error) {
    // Initialize Mythic client
    mythicClient, err := mythic.NewClient(&mythic.Config{
        ServerURL: config.MythicURL,
        APIToken:  config.APIToken,
        SSL:       config.SSL,
        SkipTLSVerify: config.SkipTLSVerify,
    })
    if err != nil {
        return nil, err
    }

    // Create MCP server
    mcpServer := mcp.NewServer()

    server := &Server{
        mcpServer:    mcpServer,
        mythicClient: mythicClient,
        config:       config,
    }

    // Register all tools
    server.registerTools()

    // Register resources (optional)
    server.registerResources()

    // Register prompts (optional)
    server.registerPrompts()

    return server, nil
}

// Run starts the MCP server
func (s *Server) Run(ctx context.Context, transport mcp.Transport) error {
    // Authenticate with Mythic
    if err := s.mythicClient.Login(ctx); err != nil {
        return fmt.Errorf("failed to authenticate with Mythic: %w", err)
    }

    // Run MCP server
    return s.mcpServer.Connect(transport)
}
```

### 2. Tool Registration (`pkg/server/tools.go`)

```go
package server

// registerTools registers all Mythic SDK operations as MCP tools
func (s *Server) registerTools() {
    // Authentication tools
    s.registerAuthTools()

    // Operation management
    s.registerOperationTools()

    // Callback operations
    s.registerCallbackTools()

    // Task execution
    s.registerTaskTools()

    // Payload generation
    s.registerPayloadTools()

    // File operations
    s.registerFileTools()

    // ... (all other categories)
}

// Example: Register callback tools
func (s *Server) registerCallbackTools() {
    // Get all callbacks
    s.mcpServer.AddTool("mythic_get_all_callbacks", mcp.Tool{
        Description: "Retrieve all callbacks from Mythic server",
        InputSchema: mcp.ToolInputSchema{
            Type: "object",
            Properties: map[string]interface{}{},
        },
    }, func(ctx context.Context, args map[string]interface{}) (*mcp.ToolResult, error) {
        callbacks, err := s.mythicClient.GetAllCallbacks(ctx)
        if err != nil {
            return nil, fmt.Errorf("failed to get callbacks: %w", err)
        }

        return formatCallbacksResult(callbacks), nil
    })

    // Get callback by ID
    s.mcpServer.AddTool("mythic_get_callback", mcp.Tool{
        Description: "Get a specific callback by ID",
        InputSchema: mcp.ToolInputSchema{
            Type: "object",
            Properties: map[string]interface{}{
                "callback_id": map[string]interface{}{
                    "type":        "integer",
                    "description": "The callback ID to retrieve",
                },
            },
            Required: []string{"callback_id"},
        },
    }, func(ctx context.Context, args map[string]interface{}) (*mcp.ToolResult, error) {
        callbackID := int(args["callback_id"].(float64))

        callback, err := s.mythicClient.GetCallbackByID(ctx, callbackID)
        if err != nil {
            return nil, fmt.Errorf("failed to get callback: %w", err)
        }

        return formatCallbackResult(callback), nil
    })

    // Issue task
    s.mcpServer.AddTool("mythic_issue_task", mcp.Tool{
        Description: "Issue a task to a callback",
        InputSchema: mcp.ToolInputSchema{
            Type: "object",
            Properties: map[string]interface{}{
                "callback_id": map[string]interface{}{
                    "type":        "integer",
                    "description": "The callback ID to task",
                },
                "command": map[string]interface{}{
                    "type":        "string",
                    "description": "The command to execute",
                },
                "params": map[string]interface{}{
                    "type":        "string",
                    "description": "Command parameters (JSON string)",
                },
            },
            Required: []string{"callback_id", "command"},
        },
    }, func(ctx context.Context, args map[string]interface{}) (*mcp.ToolResult, error) {
        callbackID := int(args["callback_id"].(float64))
        command := args["command"].(string)
        params := ""
        if p, ok := args["params"]; ok {
            params = p.(string)
        }

        task, err := s.mythicClient.IssueTask(ctx, &mythic.TaskRequest{
            CallbackDisplayID: callbackID,
            Command:           command,
            Params:            params,
        })
        if err != nil {
            return nil, fmt.Errorf("failed to issue task: %w", err)
        }

        return formatTaskResult(task), nil
    })
}
```

### 3. Output Formatting (`pkg/server/formatters.go`)

```go
package server

import (
    "encoding/json"
    "github.com/modelcontextprotocol/go-sdk/mcp"
    "github.com/nbaertsch/mythic-sdk-go/pkg/mythic"
)

// formatCallbacksResult formats callbacks for MCP response
func formatCallbacksResult(callbacks []mythic.Callback) *mcp.ToolResult {
    data, _ := json.MarshalIndent(callbacks, "", "  ")

    return &mcp.ToolResult{
        Content: []interface{}{
            mcp.TextContent{
                Type: "text",
                Text: string(data),
            },
        },
    }
}

// formatTaskResult formats task for MCP response
func formatTaskResult(task *mythic.Task) *mcp.ToolResult {
    data, _ := json.MarshalIndent(task, "", "  ")

    return &mcp.ToolResult{
        Content: []interface{}{
            mcp.TextContent{
                Type: "text",
                Text: string(data),
            },
        },
        IsError: task.Status == "error",
    }
}
```

---

## MCP Protocol Integration

### Tool Categories

Tools will be organized by Mythic functionality area:

**Authentication (7 tools)**
- `mythic_login` - Authenticate with username/password
- `mythic_logout` - End session
- `mythic_create_api_token` - Generate API token
- `mythic_delete_api_token` - Revoke API token
- `mythic_get_me` - Get current user info
- `mythic_is_authenticated` - Check auth status
- `mythic_refresh_token` - Refresh access token

**Operations (11 tools)**
- `mythic_get_operations` - List all operations
- `mythic_get_operation` - Get operation by ID
- `mythic_create_operation` - Create new operation
- `mythic_update_operation` - Modify operation
- `mythic_set_current_operation` - Switch operation context
- `mythic_get_current_operation` - Get active operation
- ... (continuing for all 204 methods)

**Callbacks (14 tools)**
- `mythic_get_all_callbacks` - List all callbacks
- `mythic_get_active_callbacks` - List active callbacks
- `mythic_get_callback` - Get callback details
- `mythic_update_callback` - Modify callback
- `mythic_delete_callback` - Remove callback
- ... (etc)

**Tasks (12 tools)**
- `mythic_issue_task` - Execute command on callback
- `mythic_get_task` - Get task status
- `mythic_wait_for_task` - Wait for completion
- `mythic_get_task_output` - Retrieve task results
- ... (etc)

**Payloads (12 tools)**
- `mythic_create_payload` - Build new payload
- `mythic_download_payload` - Download payload binary
- `mythic_get_payload` - Get payload info
- ... (etc)

### Resource Exposure (Optional)

MCP resources for browsing Mythic state:

```json
{
  "resources": [
    {
      "uri": "mythic://operations",
      "name": "Mythic Operations",
      "description": "List of all operations",
      "mimeType": "application/json"
    },
    {
      "uri": "mythic://callbacks",
      "name": "Active Callbacks",
      "description": "All active agent callbacks",
      "mimeType": "application/json"
    },
    {
      "uri": "mythic://payloads",
      "name": "Generated Payloads",
      "description": "All built payloads",
      "mimeType": "application/json"
    }
  ]
}
```

### Prompt Templates (Optional)

Pre-built prompts for common workflows:

```json
{
  "prompts": [
    {
      "name": "deploy_agent",
      "description": "Guide through agent deployment workflow",
      "arguments": [
        {
          "name": "target_os",
          "description": "Target operating system (Windows, Linux, macOS)",
          "required": true
        },
        {
          "name": "agent_type",
          "description": "Payload type (apollo, poseidon, etc.)",
          "required": true
        }
      ]
    },
    {
      "name": "execute_task",
      "description": "Execute command on callback",
      "arguments": [
        {
          "name": "callback_id",
          "description": "Callback ID to task",
          "required": true
        },
        {
          "name": "command",
          "description": "Command to execute",
          "required": true
        }
      ]
    }
  ]
}
```

---

## Authentication & Security

### Configuration

```go
type Config struct {
    // Mythic connection
    MythicURL      string `env:"MYTHIC_URL" required:"true"`
    APIToken       string `env:"MYTHIC_API_TOKEN"`
    Username       string `env:"MYTHIC_USERNAME"`
    Password       string `env:"MYTHIC_PASSWORD"`
    SSL            bool   `env:"MYTHIC_SSL" default:"true"`
    SkipTLSVerify  bool   `env:"MYTHIC_SKIP_TLS_VERIFY" default:"false"`

    // Server options
    LogLevel       string `env:"LOG_LEVEL" default:"info"`
    Timeout        time.Duration `env:"TIMEOUT" default:"30s"`
}
```

### Authentication Flow

1. **Server Startup**
   - Load configuration from environment
   - Create Mythic SDK client
   - Authenticate using API token or username/password
   - Verify connection

2. **Session Management**
   - Maintain single authenticated session
   - Automatic token refresh
   - Reconnection on connection loss

3. **Tool Execution**
   - All tools execute in authenticated context
   - No per-tool authentication needed
   - Inherit operation context from client

### Security Considerations

- **Credential Storage** - Environment variables only, never hardcoded
- **TLS Verification** - Enabled by default, skip only for development
- **Token Management** - Automatic refresh, secure storage
- **Input Validation** - All tool inputs validated against schemas
- **Error Messages** - No sensitive data in error responses

---

## Error Handling

### Error Translation

Mythic SDK errors are translated to user-friendly MCP error responses:

```go
func translateError(err error) *mcp.ToolError {
    switch {
    case errors.Is(err, mythic.ErrNotAuthenticated):
        return &mcp.ToolError{
            Code:    mcp.ErrorCodeInvalidRequest,
            Message: "Not authenticated with Mythic server. Please check credentials.",
        }
    case errors.Is(err, mythic.ErrNotFound):
        return &mcp.ToolError{
            Code:    mcp.ErrorCodeInvalidRequest,
            Message: "Requested resource not found",
        }
    case errors.Is(err, mythic.ErrInvalidInput):
        return &mcp.ToolError{
            Code:    mcp.ErrorCodeInvalidParams,
            Message: "Invalid input parameters",
        }
    default:
        return &mcp.ToolError{
            Code:    mcp.ErrorCodeInternalError,
            Message: fmt.Sprintf("Internal error: %v", err),
        }
    }
}
```

### Error Categories

1. **Authentication Errors** - Invalid credentials, expired tokens
2. **Network Errors** - Connection failures, timeouts
3. **Validation Errors** - Invalid tool inputs
4. **Mythic API Errors** - GraphQL errors, operation failures
5. **Internal Errors** - Unexpected server errors

---

## Project Structure

```
mythic-mcp/
├── cmd/
│   └── mythic-mcp/
│       └── main.go                 # Server entry point
├── pkg/
│   ├── server/
│   │   ├── server.go               # MCP server implementation
│   │   ├── tools.go                # Tool registration
│   │   ├── tools_auth.go           # Authentication tools
│   │   ├── tools_operations.go     # Operation tools
│   │   ├── tools_callbacks.go      # Callback tools
│   │   ├── tools_tasks.go          # Task tools
│   │   ├── tools_payloads.go       # Payload tools
│   │   ├── tools_files.go          # File tools
│   │   ├── tools_*.go              # Other category tools
│   │   ├── formatters.go           # Output formatting
│   │   ├── validators.go           # Input validation
│   │   └── errors.go               # Error translation
│   ├── config/
│   │   └── config.go               # Configuration management
│   └── client/
│       └── manager.go              # Mythic client lifecycle
├── tests/
│   ├── integration/
│   │   ├── e2e_helpers.go          # Test infrastructure
│   │   ├── e2e_auth_test.go        # Auth workflow tests
│   │   ├── e2e_operations_test.go  # Operations tests
│   │   ├── e2e_callbacks_test.go   # Callback tests
│   │   ├── e2e_tasks_test.go       # Task tests
│   │   └── e2e_*.go                # Other E2E tests
│   └── unit/
│       ├── formatters_test.go      # Unit tests
│       └── validators_test.go
├── docs/
│   ├── 01-ARCHITECTURE.md          # This document
│   ├── 02-API-MAPPING.md           # Mythic → MCP mapping
│   ├── 03-TEST-STRATEGY.md         # Test design
│   ├── 04-CI-CD-DESIGN.md          # CI/CD pipeline
│   └── 05-IMPLEMENTATION-ROADMAP.md # Implementation phases
├── scripts/
│   ├── setup-mythic.sh             # Dev environment setup
│   └── run-tests.sh                # Test execution
├── .github/
│   └── workflows/
│       ├── test.yml                # CI pipeline
│       └── release.yml             # Release automation
├── docker-compose.test.yml         # Test environment
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## Next Steps

1. **Complete Design Phase**
   - [ ] Finalize API mapping (docs/02-API-MAPPING.md)
   - [ ] Design test strategy (docs/03-TEST-STRATEGY.md)
   - [ ] Plan CI/CD pipeline (docs/04-CI-CD-DESIGN.md)
   - [ ] Create implementation roadmap (docs/05-IMPLEMENTATION-ROADMAP.md)

2. **Implementation Phase 1: Core Infrastructure**
   - [ ] Initialize Go module
   - [ ] Integrate MCP Go SDK
   - [ ] Create basic server structure
   - [ ] Implement configuration management
   - [ ] Add logging framework

3. **Implementation Phase 2: Authentication Tools**
   - [ ] Implement auth tools
   - [ ] Add integration tests
   - [ ] Verify against real Mythic

4. **Implementation Phases 3-N: Feature Categories**
   - [ ] Implement tools by category
   - [ ] Add comprehensive tests for each
   - [ ] Maintain >90% coverage

---

**Status:** Design Complete - Ready for Implementation
**Next Document:** [02-API-MAPPING.md](02-API-MAPPING.md)
