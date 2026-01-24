# Mythic MCP Server Test Strategy

**Author:** Claude Code
**Date:** 2026-01-24
**Purpose:** Comprehensive test design following CI-First philosophy
**Goal:** >90% test coverage with 0% skip rate

---

## Table of Contents

1. [Testing Philosophy](#testing-philosophy)
2. [Test Architecture](#test-architecture)
3. [Test Environment](#test-environment)
4. [E2E Test Workflows](#e2e-test-workflows)
5. [Test Data Management](#test-data-management)
6. [Test Implementation Patterns](#test-implementation-patterns)
7. [Coverage Requirements](#coverage-requirements)

---

## Testing Philosophy

### CI-First Principles Applied

**1. Integration Over Isolation**
- Test against REAL Mythic instances, not mocks
- Test against REAL MCP clients, not fake transports
- Test complete workflows, not isolated functions

**2. No Test Skips**
- Every test creates its own dependencies
- Use `EnsureXExists()` helper patterns
- Tests FAIL (not skip) when infrastructure unavailable

**3. CI as Source of Truth**
- If tests pass in CI, the MCP server works
- Local tests may lie (wrong env, cached state)
- All PRs require green CI

**4. Systematic Approach**
- Break complex tests into phases
- Each phase validates a layer
- Progressive validation (auth → core → advanced)

### Test Pyramid (Inverted for Integration)

```
┌─────────────────────────────┐
│    E2E MCP Workflow Tests   │  ← MOST IMPORTANT
│  (Full MCP client → Mythic) │     80% of test effort
├─────────────────────────────┤
│   Integration Tests         │  ← CORE VALIDATION
│   (MCP server → Mythic SDK) │     15% of test effort
├─────────────────────────────┤
│      Unit Tests             │  ← TARGETED ONLY
│   (Formatters, validators)  │     5% of test effort
└─────────────────────────────┘
```

---

## Test Architecture

### Test Structure

```
tests/
├── integration/
│   ├── e2e_helpers.go              # E2E test infrastructure
│   │   - MCPTestSetup struct
│   │   - Mythic server setup
│   │   - MCP client creation
│   │   - Cleanup functions
│   │
│   ├── e2e_auth_test.go            # Phase 1: Authentication
│   │   - Login/logout workflows
│   │   - Token management
│   │   - Session handling
│   │
│   ├── e2e_operations_test.go      # Phase 2: Operations
│   │   - Operation CRUD
│   │   - Global settings
│   │   - Event logging
│   │
│   ├── e2e_callbacks_tasks_test.go # Phase 3: Core C2
│   │   - Payload build
│   │   - Agent deployment
│   │   - Callback management
│   │   - Task execution
│   │   - Output retrieval
│   │
│   ├── e2e_files_test.go           # Phase 4: File Ops
│   │   - File upload/download
│   │   - Bulk operations
│   │   - File management
│   │
│   ├── e2e_advanced_test.go        # Phase 5: Advanced
│   │   - Credentials & artifacts
│   │   - MITRE ATT&CK
│   │   - Tags & categorization
│   │   - Processes & hosts
│   │
│   └── e2e_specialized_test.go     # Phase 6: Specialized
│       - Screenshots, keylogs
│       - Eventing & workflows
│       - Container operations
│       - Reporting
│
├── unit/
│   ├── formatters_test.go          # Output formatting
│   ├── validators_test.go          # Input validation
│   └── errors_test.go              # Error translation
│
└── testdata/
    ├── payloads/                   # Test payload configs
    ├── files/                      # Test files
    └── workflows/                  # Test workflow definitions
```

### Test Levels

**Level 1: Unit Tests (5%)**
- Format converters (Mythic types → MCP responses)
- Input validators (MCP params → Mythic requests)
- Error translators (Mythic errors → MCP errors)
- **Coverage Target:** >95%
- **Execution Time:** <5 seconds
- **Dependencies:** None

**Level 2: Integration Tests (15%)**
- MCP server initialization
- Tool registration
- Mythic client integration
- **Coverage Target:** >90%
- **Execution Time:** <30 seconds
- **Dependencies:** None (mocked transport)

**Level 3: E2E Workflow Tests (80%)**
- Complete MCP client workflows
- Real Mythic server interactions
- Agent deployment and tasking
- **Coverage Target:** >90% (all 204 tools)
- **Execution Time:** 5-10 minutes
- **Dependencies:** Docker (Mythic + Poseidon agent)

---

## Test Environment

### Infrastructure Setup

**Docker Compose Test Environment:**

```yaml
# docker-compose.test.yml
version: '3.8'

services:
  mythic:
    image: itsafeature/mythic_server:latest
    container_name: mythic_test
    environment:
      MYTHIC_ADMIN_PASSWORD: TestPassword123!
      MYTHIC_SERVER_PORT: 7443
      MYTHIC_SERVER_BIND_IP: 0.0.0.0
    ports:
      - "7443:7443"
    volumes:
      - mythic_data:/Mythic
    networks:
      - mythic_test

  postgres:
    image: postgres:13
    container_name: mythic_postgres_test
    environment:
      POSTGRES_DB: mythic
      POSTGRES_USER: mythic
      POSTGRES_PASSWORD: mythic
    networks:
      - mythic_test

  rabbitmq:
    image: rabbitmq:3-management
    container_name: mythic_rabbitmq_test
    networks:
      - mythic_test

  # Poseidon agent container (for E2E tests)
  poseidon:
    image: itsafeature/poseidon:latest
    container_name: poseidon_test
    depends_on:
      - mythic
    networks:
      - mythic_test

networks:
  mythic_test:
    driver: bridge

volumes:
  mythic_data:
```

### Test Helpers (`tests/integration/e2e_helpers.go`)

```go
package integration

import (
    "context"
    "fmt"
    "testing"
    "time"

    "github.com/modelcontextprotocol/go-sdk/mcp"
    "github.com/nbaertsch/mythic-sdk-go/pkg/mythic"
    "github.com/stretchr/testify/require"

    "github.com/YOUR_ORG/mythic-mcp/pkg/server"
)

// MCPTestSetup provides E2E test infrastructure
type MCPTestSetup struct {
    T             *testing.T
    Ctx           context.Context

    // MCP server components
    MCPServer     *server.Server
    MCPClient     *mcp.Client

    // Mythic SDK client (for verification)
    MythicClient  *mythic.Client

    // Test resources
    OperationID   int
    PayloadUUID   string
    CallbackID    int
    AgentProcess  *os.Process

    // Cleanup functions
    cleanupFuncs  []func()
}

// SetupE2ETest creates complete E2E test environment
func SetupE2ETest(t *testing.T) *MCPTestSetup {
    ctx := context.Background()

    // Get Mythic credentials from environment
    mythicURL := os.Getenv("MYTHIC_URL")
    if mythicURL == "" {
        mythicURL = "https://127.0.0.1:7443"
    }

    mythicPassword := os.Getenv("MYTHIC_PASSWORD")
    require.NotEmpty(t, mythicPassword, "MYTHIC_PASSWORD must be set")

    // Create Mythic SDK client for verification
    mythicClient, err := mythic.NewClient(&mythic.Config{
        ServerURL:     mythicURL,
        Username:      "mythic_admin",
        Password:      mythicPassword,
        SSL:           true,
        SkipTLSVerify: true,
    })
    require.NoError(t, err)

    // Authenticate Mythic client
    err = mythicClient.Login(ctx)
    require.NoError(t, err)

    // Create MCP server
    mcpServer, err := server.NewServer(&server.Config{
        MythicURL:     mythicURL,
        Username:      "mythic_admin",
        Password:      mythicPassword,
        SSL:           true,
        SkipTLSVerify: true,
    })
    require.NoError(t, err)

    // Create in-memory transport for testing
    transport := createTestTransport()

    // Start MCP server
    go func() {
        _ = mcpServer.Run(ctx, transport)
    }()

    // Wait for server ready
    time.Sleep(2 * time.Second)

    // Create MCP client
    mcpClient := mcp.NewClient(transport)

    setup := &MCPTestSetup{
        T:            t,
        Ctx:          ctx,
        MCPServer:    mcpServer,
        MCPClient:    mcpClient,
        MythicClient: mythicClient,
        cleanupFuncs: []func(){},
    }

    // Register cleanup
    t.Cleanup(setup.Cleanup)

    return setup
}

// Cleanup runs all registered cleanup functions
func (s *MCPTestSetup) Cleanup() {
    for i := len(s.cleanupFuncs) - 1; i >= 0; i-- {
        s.cleanupFuncs[i]()
    }

    // Close clients
    if s.MythicClient != nil {
        s.MythicClient.Close()
    }
}

// CallMCPTool executes an MCP tool and returns result
func (s *MCPTestSetup) CallMCPTool(toolName string, args map[string]interface{}) (*mcp.ToolResult, error) {
    return s.MCPClient.CallTool(s.Ctx, toolName, args)
}

// EnsurePayloadExists creates a test payload if needed
func (s *MCPTestSetup) EnsurePayloadExists() string {
    if s.PayloadUUID != "" {
        return s.PayloadUUID
    }

    // Create payload using MCP tool
    result, err := s.CallMCPTool("mythic_create_payload", map[string]interface{}{
        "payload_type": "poseidon",
        "os":           "linux",
        "c2_profiles": []map[string]interface{}{
            {
                "name": "http",
                "parameters": map[string]interface{}{
                    "callback_host": "http://127.0.0.1:80",
                },
            },
        },
        "commands": []string{"shell", "download", "upload", "ps"},
    })
    require.NoError(s.T, err)

    // Extract UUID from result
    payload := parsePayloadResult(result)
    s.PayloadUUID = payload.UUID

    // Register cleanup
    s.cleanupFuncs = append(s.cleanupFuncs, func() {
        s.CallMCPTool("mythic_delete_payload", map[string]interface{}{
            "payload_uuid": s.PayloadUUID,
        })
    })

    // Wait for payload build
    s.WaitForPayloadBuild(s.PayloadUUID, 90)

    return s.PayloadUUID
}

// EnsureCallbackExists deploys agent and waits for callback
func (s *MCPTestSetup) EnsureCallbackExists() int {
    if s.CallbackID != 0 {
        return s.CallbackID
    }

    // Ensure payload exists
    payloadUUID := s.EnsurePayloadExists()

    // Download payload
    result, err := s.CallMCPTool("mythic_download_payload", map[string]interface{}{
        "payload_uuid": payloadUUID,
    })
    require.NoError(s.T, err)

    // Save payload to temp file
    payloadPath := fmt.Sprintf("/tmp/agent_%s", payloadUUID)
    payloadData := extractPayloadData(result)
    err = os.WriteFile(payloadPath, payloadData, 0755)
    require.NoError(s.T, err)

    // Start agent
    cmd := exec.Command(payloadPath)
    err = cmd.Start()
    require.NoError(s.T, err)

    s.AgentProcess = cmd.Process

    // Register cleanup
    s.cleanupFuncs = append(s.cleanupFuncs, func() {
        if s.AgentProcess != nil {
            s.AgentProcess.Kill()
        }
        os.Remove(payloadPath)
    })

    // Wait for callback
    callbackID := s.WaitForCallback(60)
    s.CallbackID = callbackID

    // Register callback cleanup
    s.cleanupFuncs = append(s.cleanupFuncs, func() {
        s.CallMCPTool("mythic_delete_callback", map[string]interface{}{
            "callback_id": s.CallbackID,
        })
    })

    return s.CallbackID
}

// WaitForCallback polls for new callback
func (s *MCPTestSetup) WaitForCallback(timeoutSeconds int) int {
    start := time.Now()
    timeout := time.Duration(timeoutSeconds) * time.Second

    for {
        if time.Since(start) > timeout {
            s.T.Fatal("Timeout waiting for callback")
        }

        // Get active callbacks
        result, err := s.CallMCPTool("mythic_get_active_callbacks", map[string]interface{}{})
        require.NoError(s.T, err)

        callbacks := parseCallbacksResult(result)
        if len(callbacks) > 0 {
            return callbacks[0].ID
        }

        time.Sleep(5 * time.Second)
    }
}

// WaitForPayloadBuild waits for payload build completion
func (s *MCPTestSetup) WaitForPayloadBuild(uuid string, timeoutSeconds int) {
    start := time.Now()
    timeout := time.Duration(timeoutSeconds) * time.Second

    for {
        if time.Since(start) > timeout {
            s.T.Fatal("Timeout waiting for payload build")
        }

        // Check payload status
        result, err := s.CallMCPTool("mythic_get_payload", map[string]interface{}{
            "payload_uuid": uuid,
        })
        require.NoError(s.T, err)

        payload := parsePayloadResult(result)
        if payload.BuildStatus == "success" {
            return
        }

        if payload.BuildStatus == "error" {
            s.T.Fatalf("Payload build failed: %s", payload.BuildMessage)
        }

        time.Sleep(5 * time.Second)
    }
}

// ExecuteTask issues a task and waits for completion
func (s *MCPTestSetup) ExecuteTask(callbackID int, command string, params string) *TaskResult {
    // Issue task
    result, err := s.CallMCPTool("mythic_issue_task", map[string]interface{}{
        "callback_id": callbackID,
        "command":     command,
        "params":      params,
    })
    require.NoError(s.T, err)

    task := parseTaskResult(result)

    // Wait for completion
    _, err = s.CallMCPTool("mythic_wait_for_task", map[string]interface{}{
        "task_id": task.ID,
        "timeout": 30,
    })
    require.NoError(s.T, err)

    // Get output
    result, err = s.CallMCPTool("mythic_get_task_output", map[string]interface{}{
        "task_id": task.ID,
    })
    require.NoError(s.T, err)

    return parseTaskResult(result)
}
```

---

## E2E Test Workflows

### Workflow 1: Authentication & Session (Phase 0)

**File:** `tests/integration/e2e_auth_test.go`
**Duration:** ~30 seconds
**Dependencies:** Mythic server only

```go
func TestE2E_AuthenticationLifecycle(t *testing.T) {
    setup := SetupE2ETest(t)

    t.Run("Login", func(t *testing.T) {
        result, err := setup.CallMCPTool("mythic_login", map[string]interface{}{
            "username": "mythic_admin",
            "password": os.Getenv("MYTHIC_PASSWORD"),
        })
        require.NoError(t, err)
        require.NotNil(t, result)
    })

    t.Run("GetCurrentUser", func(t *testing.T) {
        result, err := setup.CallMCPTool("mythic_get_current_user", nil)
        require.NoError(t, err)

        user := parseOperatorResult(result)
        assert.Equal(t, "mythic_admin", user.Username)
    })

    t.Run("CreateAPIToken", func(t *testing.T) {
        result, err := setup.CallMCPTool("mythic_create_api_token", map[string]interface{}{
            "token_type": "User",
        })
        require.NoError(t, err)

        token := parseTokenResult(result)
        assert.NotEmpty(t, token.Value)

        // Cleanup
        setup.CallMCPTool("mythic_delete_api_token", map[string]interface{}{
            "token_id": token.ID,
        })
    })

    t.Run("Logout", func(t *testing.T) {
        result, err := setup.CallMCPTool("mythic_logout", nil)
        require.NoError(t, err)
        assert.True(t, result.IsSuccess)
    })
}
```

### Workflow 2: Operations Management (Phase 1)

**File:** `tests/integration/e2e_operations_test.go`
**Duration:** ~45 seconds
**Dependencies:** Authenticated session

```go
func TestE2E_OperationsManagement(t *testing.T) {
    setup := SetupE2ETest(t)

    t.Run("GetOperations", func(t *testing.T) {
        result, err := setup.CallMCPTool("mythic_get_operations", nil)
        require.NoError(t, err)

        ops := parseOperationsResult(result)
        assert.NotEmpty(t, ops)
    })

    t.Run("CreateOperation", func(t *testing.T) {
        result, err := setup.CallMCPTool("mythic_create_operation", map[string]interface{}{
            "name": "Test Operation E2E",
        })
        require.NoError(t, err)

        op := parseOperationResult(result)
        assert.Equal(t, "Test Operation E2E", op.Name)

        // Verify creation via Mythic SDK
        sdkOp, err := setup.MythicClient.GetOperationByID(setup.Ctx, op.ID)
        require.NoError(t, err)
        assert.Equal(t, "Test Operation E2E", sdkOp.Name)
    })

    t.Run("UpdateOperation", func(t *testing.T) {
        // Create operation first
        createResult, _ := setup.CallMCPTool("mythic_create_operation", map[string]interface{}{
            "name": "Test Op for Update",
        })
        op := parseOperationResult(createResult)

        // Update it
        result, err := setup.CallMCPTool("mythic_update_operation", map[string]interface{}{
            "operation_id": op.ID,
            "webhook":      "https://example.com/webhook",
        })
        require.NoError(t, err)

        updated := parseOperationResult(result)
        assert.Equal(t, "https://example.com/webhook", updated.Webhook)
    })
}
```

### Workflow 3: Callback & Task Execution (Phase 3)

**File:** `tests/integration/e2e_callbacks_tasks_test.go`
**Duration:** ~3-5 minutes
**Dependencies:** Poseidon agent container

```go
func TestE2E_CallbackTaskLifecycle(t *testing.T) {
    if os.Getenv("SKIP_AGENT_TESTS") != "" {
        t.Skip("Skipping agent tests")
    }

    setup := SetupE2ETest(t)

    var callbackID int

    t.Run("Part1_PayloadBuild", func(t *testing.T) {
        result, err := setup.CallMCPTool("mythic_create_payload", map[string]interface{}{
            "payload_type": "poseidon",
            "os":           "linux",
            "c2_profiles": []map[string]interface{}{
                {
                    "name": "http",
                    "parameters": map[string]interface{}{
                        "callback_host": "http://127.0.0.1:80",
                    },
                },
            },
            "commands": []string{"shell", "ps", "download"},
        })
        require.NoError(t, err)

        payload := parsePayloadResult(result)
        assert.NotEmpty(t, payload.UUID)

        // Wait for build
        setup.WaitForPayloadBuild(payload.UUID, 90)
    })

    t.Run("Part2_CallbackEstablishment", func(t *testing.T) {
        callbackID = setup.EnsureCallbackExists()
        assert.Greater(t, callbackID, 0)

        // Get callback details
        result, err := setup.CallMCPTool("mythic_get_callback", map[string]interface{}{
            "callback_id": callbackID,
        })
        require.NoError(t, err)

        callback := parseCallbackResult(result)
        assert.Equal(t, "active", callback.Status)
    })

    t.Run("Part3_ShellTaskExecution", func(t *testing.T) {
        taskResult := setup.ExecuteTask(callbackID, "shell", "whoami")

        assert.Equal(t, "completed", taskResult.Status)
        assert.NotEmpty(t, taskResult.Output)
        assert.Contains(t, taskResult.Output, "root") // or current user
    })

    t.Run("Part4_ProcessListing", func(t *testing.T) {
        taskResult := setup.ExecuteTask(callbackID, "ps", "")

        assert.Equal(t, "completed", taskResult.Status)

        // Verify processes populated
        result, err := setup.CallMCPTool("mythic_get_callback_processes", map[string]interface{}{
            "callback_id": callbackID,
        })
        require.NoError(t, err)

        processes := parseProcessesResult(result)
        assert.NotEmpty(t, processes)
    })
}
```

### Workflow 4: File Operations (Phase 4)

**File:** `tests/integration/e2e_files_test.go`
**Duration:** ~40 seconds
**Dependencies:** None

```go
func TestE2E_FileOperations(t *testing.T) {
    setup := SetupE2ETest(t)

    var fileUUID string

    t.Run("UploadFile", func(t *testing.T) {
        testData := []byte("This is test file content")

        result, err := setup.CallMCPTool("mythic_upload_file", map[string]interface{}{
            "filename":  "test.txt",
            "file_data": base64.StdEncoding.EncodeToString(testData),
        })
        require.NoError(t, err)

        file := parseFileResult(result)
        fileUUID = file.UUID
        assert.NotEmpty(t, fileUUID)
    })

    t.Run("DownloadFile", func(t *testing.T) {
        result, err := setup.CallMCPTool("mythic_download_file", map[string]interface{}{
            "file_uuid": fileUUID,
        })
        require.NoError(t, err)

        downloadedData := parseFileDataResult(result)
        assert.Equal(t, "This is test file content", string(downloadedData))
    })

    t.Run("DeleteFile", func(t *testing.T) {
        result, err := setup.CallMCPTool("mythic_delete_file", map[string]interface{}{
            "file_uuid": fileUUID,
        })
        require.NoError(t, err)
        assert.True(t, result.IsSuccess)
    })
}
```

---

## Test Data Management

### Shared vs Per-Test Resources

**Shared Resources (Expensive to Create):**
- Mythic server instance
- Poseidon agent container
- Database connections
- MCP server instance

**Per-Test Resources (Created in Each Test):**
- Operations
- Payloads
- Callbacks
- Tasks
- Files
- Credentials

### EnsureX Pattern

```go
// EnsureXExists creates resource if doesn't exist, returns if does
func (s *MCPTestSetup) EnsureMythicRunning() {
    // Check if Mythic is accessible
    ctx, cancel := context.WithTimeout(s.Ctx, 5*time.Second)
    defer cancel()

    result, err := s.CallMCPTool("mythic_get_current_user", nil)
    if err == nil && result != nil {
        return // Already running and authenticated
    }

    // Wait for Mythic to be ready
    s.T.Fatal("Mythic server not accessible")
}

// Shared resource - reuse across tests in same suite
var sharedPayloadUUID string

func (s *MCPTestSetup) EnsureTestPayload() string {
    if sharedPayloadUUID != "" {
        return sharedPayloadUUID
    }

    // Create payload
    result, err := s.CallMCPTool("mythic_create_payload", map[string]interface{}{
        "payload_type": "poseidon",
        // ...
    })
    require.NoError(s.T, err)

    payload := parsePayloadResult(result)
    sharedPayloadUUID = payload.UUID

    // Wait for build
    s.WaitForPayloadBuild(sharedPayloadUUID, 90)

    return sharedPayloadUUID
}
```

---

## Test Implementation Patterns

### Pattern 1: Verify via Both MCP and SDK

```go
t.Run("CreateOperation", func(t *testing.T) {
    // Create via MCP
    mcpResult, err := setup.CallMCPTool("mythic_create_operation", map[string]interface{}{
        "name": "Test Operation",
    })
    require.NoError(t, err)

    op := parseOperationResult(mcpResult)

    // Verify via Mythic SDK
    sdkOp, err := setup.MythicClient.GetOperationByID(setup.Ctx, op.ID)
    require.NoError(t, err)
    assert.Equal(t, "Test Operation", sdkOp.Name)
})
```

### Pattern 2: Progressive Validation

```go
func TestE2E_TaskExecution_Progressive(t *testing.T) {
    setup := SetupE2ETest(t)

    // Phase 1: Issue task
    issueResult, err := setup.CallMCPTool("mythic_issue_task", args)
    require.NoError(t, err)
    task := parseTaskResult(issueResult)

    // Phase 2: Verify task created
    getResult, err := setup.CallMCPTool("mythic_get_task", map[string]interface{}{
        "task_id": task.ID,
    })
    require.NoError(t, err)

    // Phase 3: Wait for completion
    _, err = setup.CallMCPTool("mythic_wait_for_task", map[string]interface{}{
        "task_id": task.ID,
        "timeout": 30,
    })
    require.NoError(t, err)

    // Phase 4: Get output
    outputResult, err := setup.CallMCPTool("mythic_get_task_output", map[string]interface{}{
        "task_id": task.ID,
    })
    require.NoError(t, err)
    assert.NotEmpty(t, outputResult)
}
```

### Pattern 3: Error Testing

```go
t.Run("ErrorHandling_InvalidCallbackID", func(t *testing.T) {
    result, err := setup.CallMCPTool("mythic_get_callback", map[string]interface{}{
        "callback_id": 99999, // Non-existent
    })

    // Should return MCP error, not panic
    assert.Error(t, err)
    assert.Nil(t, result)

    // Error should be user-friendly
    assert.Contains(t, err.Error(), "not found")
})
```

---

## Coverage Requirements

### Coverage Targets

| Component | Target | Measurement |
|-----------|--------|-------------|
| MCP Tools | >90% | All 204 tools tested |
| Error Paths | >80% | Error scenarios covered |
| Input Validation | 100% | All tool inputs validated |
| Output Formatting | >90% | All output formats tested |
| Mythic SDK Integration | >90% | All SDK methods called |

### Coverage Tracking

```bash
# Generate coverage report
go test -v -coverprofile=coverage.out ./tests/integration/...

# View coverage by package
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Fail if coverage below threshold
go test -coverprofile=coverage.out ./tests/integration/... && \
  go tool cover -func=coverage.out | grep total | awk '{print $3}' | \
  awk '{if (int($1) < 90) exit 1}'
```

### Skip Rate Tracking

```bash
# Count skips in test output
go test -v ./tests/integration/... 2>&1 | grep -c "SKIP"

# Fail if any skips detected
SKIPS=$(go test -v ./tests/integration/... 2>&1 | grep -c "SKIP")
if [ $SKIPS -gt 0 ]; then
  echo "ERROR: $SKIPS tests skipped - all tests must run!"
  exit 1
fi
```

---

## Success Criteria

### Quantitative Metrics

- [ ] **Tool Coverage:** 204/204 tools tested (100%)
- [ ] **Test Coverage:** >90% code coverage
- [ ] **Skip Rate:** 0% (no skipped tests)
- [ ] **Pass Rate:** 100% in CI
- [ ] **Execution Time:** <10 minutes full suite
- [ ] **Flake Rate:** <1% (tests are reliable)

### Qualitative Goals

- [ ] Tests mirror realistic operator workflows
- [ ] No manual setup required
- [ ] Tests fail fast with clear errors
- [ ] Comprehensive logging for debugging
- [ ] Reproducible in CI and locally
- [ ] Self-contained cleanup
- [ ] Easy to add new tests

---

**Status:** Test Strategy Complete
**Next Document:** [04-CI-CD-DESIGN.md](04-CI-CD-DESIGN.md)
