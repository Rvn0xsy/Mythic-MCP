# E2E Integration Test Failure Report

**Date:** 2026-02-14  
**CI Run:** [#22022536800](https://github.com/nbaertsch/Mythic-MCP/actions/runs/22022536800)  
**Commit:** `c84496f` (fix: add missing e2e build tags)  
**Mythic Version:** v3.4.23 (CI runner)

---

## Executive Summary

**83 tests total** across 6 phases:

| Result  | Count | Percentage |
|---------|-------|------------|
| PASS    | 18    | 21.7%      |
| FAIL    | 47    | 56.6%      |
| SKIP    | 18    | 21.7%      |

All 47 failures trace back to **4 root causes** in the test harness and MCP server — not in the Mythic framework itself:

| Root Cause | Assertion Pattern | Failing Assertions | Fix Scope |
|---|---|---|---|
| **RC-1: Response shape mismatch** | `Should be true` | 28 | Test harness or tool handlers |
| **RC-2: Error not propagated** | `An error is expected but got nil` | 49 | MCP tool handlers |
| **RC-3: Update operations not applied** | `Not equal` (value mismatch) | 4 | Tool handler logic |
| **RC-4: Schema mismatch / API limitation** | `Received unexpected error` | 2 | Tool schema or Mythic API |

> **Note:** Many tests have multiple assertions that fail, and later tests cascade from earlier failures. The 47 unique FAIL test functions produce ~83 individual assertion failures.

---

## Root Cause Analysis

### RC-1: Response Shape Mismatch — `content` is `map`, not `[]interface{}` (28 assertions)

**Problem:** List tools (e.g. `mythic_get_operations`, `mythic_get_callbacks`) return structured content as:

```json
{
  "content": {
    "count": 1,
    "items": [...]
  }
}
```

Tests expect `content` to be a bare `[]interface{}` array:

```go
content, ok := result["content"].([]interface{})
require.True(t, ok, "Expected content to be an array")  // FAILS — content is map[string]interface{}
```

**Mechanism:** The `CallMCPTool()` helper in `e2e_helpers.go` (line ~370) normalizes MCP responses by copying `structuredContent` → `content`. When the tool handler returns a wrapper object `{count, items}`, the `content` field is a `map[string]interface{}`, not a `[]interface{}`.

**Fix Options:**
1. **Fix tests:** Unwrap `result["content"].(map[string]interface{})["items"].([]interface{})` instead of `result["content"].([]interface{})`
2. **Fix tool handlers:** Return `items` array directly as `structuredContent` instead of wrapping in `{count, items}`
3. **Fix normalizer:** Have `CallMCPTool()` detect `{count, items}` wrappers and unwrap automatically

**Affected Tests (28 assertion failures across ~20 tests):**

| Phase | Test | Tool |
|-------|------|------|
| 1 | `TestE2E_Operations_GetOperations` | `mythic_get_operations` |
| 1 | `TestE2E_Operations_CreateAndManage` | `mythic_create_operation` (metadata shape) |
| 1 | `TestE2E_Operations_CurrentOperation` | `mythic_get_current_operation` |
| 1 | `TestE2E_Operations_EventLog` | `mythic_get_event_log` |
| 1 | `TestE2E_Operations_Operators` | `mythic_get_operation_operators` |
| 1 | `TestE2E_Operators_GetOperators` | `mythic_get_operators` |
| 1 | `TestE2E_Operators_InviteLinks` | `mythic_get_invite_links` |
| 1 | `TestE2E_Tags_GetTagTypes` | `mythic_get_tag_types` |
| 1 | `TestE2E_Tags_GetTagTypesByOperation` | `mythic_get_tag_types_by_operation` |
| 1 | `TestE2E_Tags_GetTagsByOperation` | `mythic_get_tags_by_operation` |
| 1 | `TestE2E_Tags_TagTypes` | `mythic_create_tag_type` (metadata shape) |
| 1 | `TestE2E_Tags_MultipleTagTypes` | `mythic_create_tag_type` (metadata shape) |
| 1 | `TestE2E_Tags_FullWorkflow` | `mythic_create_tag_type` (cascading) |
| 3 | `TestE2E_Callbacks_GetAllCallbacks` | `mythic_get_all_callbacks` |
| 3 | `TestE2E_Callbacks_GetActiveCallbacks` | `mythic_get_active_callbacks` |
| 4 | `TestE2E_Files_GetFiles` | `mythic_get_files` |
| 4 | `TestE2E_Files_GetDownloadedFiles` | `mythic_get_downloaded_files` |
| 4 | `TestE2E_Hosts_GetHosts` | `mythic_get_hosts` |
| 4 | `TestE2E_Keylogs_GetKeylogs` | `mythic_get_keylogs` |
| 4 | `TestE2E_Keylogs_GetKeylogsByOperation` | `mythic_get_keylogs_by_operation` |
| 4 | `TestE2E_Processes_GetProcesses` | `mythic_get_processes` |
| 4 | `TestE2E_Processes_GetProcessesByOperation` | `mythic_get_processes_by_operation` |
| 4 | `TestE2E_Responses_GetLatestResponses` | `mythic_get_latest_responses` |
| 4 | `TestE2E_Responses_SearchResponses` | `mythic_search_responses` |
| 5 | `TestE2E_Credentials_GetCredentials` | `mythic_get_credentials` |
| 5 | `TestE2E_Credentials_GetByOperation` | `mythic_get_operation_credentials` |
| 5 | `TestE2E_Artifacts_GetArtifacts` | `mythic_get_artifacts` |
| 5 | `TestE2E_Artifacts_GetByOperation` | `mythic_get_operation_artifacts` |
| 5 | `TestE2E_Artifacts_GetByHost` | `mythic_get_host_artifacts` |
| 5 | `TestE2E_Artifacts_GetByType` | `mythic_get_artifacts_by_type` |

---

### RC-2: Error Not Propagated — `isError: true` Not Surfaced as Go Error (49 assertions)

**Problem:** When MCP tools encounter invalid inputs (non-existent IDs, bad parameters), the tool handlers return a **successful MCP response** with `isError: true` in the content body, rather than returning a **JSON-RPC error**. The `CallMCPTool()` helper only returns a Go error when `resp.Error != nil` (a JSON-RPC level error), so `err` is always `nil` for these cases.

**CI Log Evidence:**

```
Error:    An error is expected but got nil.
Messages: Expected error when getting non-existent operation
```

**Mechanism in `e2e_helpers.go`:**

```go
// Line ~350 — only JSON-RPC errors are returned as Go errors
if resp.Error != nil {
    return nil, fmt.Errorf("MCP error: %v", resp.Error)
}
// isError in the content body is copied to normalizedResult but NOT returned as an error
```

**Fix Options:**
1. **Fix `CallMCPTool()` helper:** Check `isError: true` in the normalized result and return it as a Go error
2. **Fix tool handlers:** Return JSON-RPC errors (via the MCP SDK's error return mechanism) for invalid inputs instead of `isError: true` content
3. **Fix tests:** Use `assert.True(t, result["isError"].(bool))` instead of `assert.Error(t, err)`

**Affected Tests (49 assertion failures across ~18 tests):**

| Phase | Test | Tools Tested |
|-------|------|-------------|
| 0 | `TestE2E_Auth_ErrorHandling/LoginWithInvalidCredentials` | `mythic_login` |
| 0 | `TestE2E_Auth_ErrorHandling/GetCurrentUserWhenNotAuthenticated` | `mythic_get_current_user` |
| 1 | `TestE2E_Operations_ErrorHandling` | `mythic_get_operation`, `mythic_update_operation`, `mythic_create_operation` |
| 1 | `TestE2E_Operators_ErrorHandling` | `mythic_get_operator`, `mythic_create_operator`, `mythic_update_operator_status` |
| 1 | `TestE2E_Tags_ErrorHandling` | `mythic_get_tag_type`, `mythic_update_tag_type`, `mythic_delete_tag_type`, `mythic_get_tag`, `mythic_delete_tag`, `mythic_create_tag`, `mythic_create_tag_type` |
| 3 | `TestE2E_Callbacks_ErrorHandling` | `mythic_get_callback`, `mythic_update_callback`, `mythic_export_callback_config`, `mythic_get_callback_tasks` |
| 4 | `TestE2E_Files_ErrorHandling` | `mythic_get_file`, `mythic_delete_file`, `mythic_download_file`, `mythic_preview_file` |
| 4 | `TestE2E_Hosts_ErrorHandling` | `mythic_get_host_by_id`, `mythic_get_host_by_hostname`, `mythic_get_callbacks_for_host`, `mythic_get_host_artifacts` |
| 4 | `TestE2E_Keylogs_ErrorHandling` | `mythic_get_keylogs_by_operation`, `mythic_get_keylogs_by_callback` |
| 4 | `TestE2E_Processes_ErrorHandling` | `mythic_get_processes_by_operation`, `mythic_get_processes_by_callback`, `mythic_get_processes_by_host`, `mythic_get_process_tree` |
| 4 | `TestE2E_Screenshots_ErrorHandling` | `mythic_get_screenshots`, `mythic_get_screenshot_by_id`, `mythic_download_screenshot`, `mythic_get_screenshot_thumbnail` |
| 4 | `TestE2E_Tasks_ErrorHandling` | `mythic_get_task`, `mythic_get_task_output`, `mythic_get_task_artifacts`, `mythic_issue_task` |
| 5 | `TestE2E_CredentialsArtifacts_ErrorHandling` | `mythic_get_credential`, `mythic_update_credential`, `mythic_delete_credential`, `mythic_get_artifact`, `mythic_update_artifact`, `mythic_delete_artifact`, `mythic_get_host_artifacts`, `mythic_get_artifacts_by_type` |

---

### RC-3: Update Operations Not Applied — Mutation Results Don't Reflect Changes (4 assertions)

**Problem:** The `mythic_update_operation` and `mythic_set_current_operation` tools appear to succeed but the returned/verified values don't match what was set. Specifically in `TestE2E_Operations_CreateAndManage`:

| Expected | Actual | Field |
|----------|--------|-------|
| `"https://example.com/webhook"` | `""` | webhook URL after update |
| `"test-channel"` | `"#random"` | channel after update |
| `"Updated Test Operation"` | `"Test Operation E2E"` | name after update |

And in `TestE2E_Operations_CurrentOperation`:

| Expected | Actual | Field |
|----------|--------|-------|
| `3` | `1` | operations count after creating 2 new ones |

**Root Cause:** The tool handlers for `mythic_update_operation` likely:
1. Send the update mutation to Mythic but don't verify the response
2. Return the pre-update state rather than the post-update state
3. The underlying Mythic GraphQL mutation may not support updating all fields (webhook, channel) or the field names don't match

For `mythic_set_current_operation`, the operation count mismatch (`expected 3, actual 1`) suggests that `mythic_create_operation` is not actually creating new operations — it may be failing silently with `isError: true` content (linking back to RC-2).

**Affected Tests:**

| Phase | Test | Tool |
|-------|------|------|
| 1 | `TestE2E_Operations_CreateAndManage` | `mythic_update_operation` |
| 1 | `TestE2E_Operations_CurrentOperation` | `mythic_create_operation`, `mythic_set_current_operation` |

---

### RC-4: Schema Mismatch / API Limitation (2 assertions)

**Problem A — `CreateAPIToken` schema mismatch:**

```
MCP error: invalid params: validating "arguments": validating root:
unexpected additional properties ["token_type"]
```

The test sends `{"token_type": "User"}` but `createAPITokenArgs` is an empty struct — it doesn't declare any parameters. The tool's JSON schema rejects unknown properties.

**Problem B — `CreateTagType` not supported:**

```
CreateTagType: tag type creation not supported via GraphQL API: operation failed
```

The Mythic GraphQL API doesn't support creating tag types. The SDK method `CreateTagType` returns this error because the mutation doesn't exist in Mythic's schema.

**Affected Tests:**

| Phase | Test | Tool | Error |
|-------|------|------|-------|
| 0 | `TestE2E_Auth_APITokens/CreateAPIToken` | `mythic_create_api_token` | Schema rejects `token_type` param |
| 1 | `TestE2E_Tags_CreateAndApplyTags` | `mythic_create_tag_type` | Mythic API doesn't support this mutation |
| 1 | `TestE2E_Tags_MultipleTagTypes` | `mythic_create_tag_type` | Same as above |
| 1 | `TestE2E_Tags_FullWorkflow` | `mythic_create_tag_type` | Same — cascading failure |
| 1 | `TestE2E_Tags_TagTypes` | `mythic_create_tag_type` | Same — cascading failure |

---

## Skipped Tests (18)

All skipped tests require **active callbacks** which don't exist in the CI environment (no real agent is deployed/calling back). They correctly skip with messages like `"No active callbacks available"`.

| Phase | Test | Reason |
|-------|------|--------|
| 0 | `TestE2E_Auth_APITokens/DeleteAPIToken` | Cascading skip — CreateAPIToken failed |
| 3 | `TestE2E_Callbacks_GetCallback` | No active callbacks |
| 3 | `TestE2E_Callbacks_UpdateCallback` | No active callbacks |
| 3 | `TestE2E_Callbacks_GetCallbackTokens` | No active callbacks |
| 3 | `TestE2E_Callbacks_GetLoadedCommands` | No active callbacks |
| 3 | `TestE2E_Callbacks_GraphEdges` | No active callbacks |
| 3 | `TestE2E_Callbacks_ExportImportConfig` | No active callbacks |
| 3 | `TestE2E_Callbacks_FullWorkflow` | No active callbacks |
| 4 | `TestE2E_Hosts_GetHostByID` | No hosts available |
| 4 | `TestE2E_Hosts_GetHostByHostname` | No hosts available |
| 4 | `TestE2E_Hosts_GetCallbacksForHost` | No hosts available |
| 4 | `TestE2E_Hosts_HostDetails` | No hosts available |
| 4 | `TestE2E_Hosts_FullWorkflow` | No hosts available |
| 4 | `TestE2E_Keylogs_*` (3 tests) | No callbacks with keylogs |
| 4 | `TestE2E_Processes_*` (5 tests) | No callbacks with processes |
| 4 | `TestE2E_Screenshots_*` (8 tests) | No callbacks with screenshots |
| 4 | `TestE2E_Tasks_*` (8 tests) | No callbacks to issue tasks to |
| 4 | `TestE2E_Responses_*` (2 tests) | No task responses available |

---

## Passing Tests (18)

These tests pass because they either:
- Don't assert on response shape (they check for success/existence only)
- Use the Mythic SDK directly for verification instead of parsing MCP content
- Test create/delete operations that return simple success messages

| Phase | Test |
|-------|------|
| 0 | `TestE2E_Auth_LoginLogout` (+ 5 subtests) |
| 0 | `TestE2E_Auth_RefreshToken` (+ 1 subtest) |
| 1 | `TestE2E_Operations_FullWorkflow` |
| 1 | `TestE2E_Operations_GlobalSettings` |
| 1 | `TestE2E_Operators_CreateAndManage` |
| 1 | `TestE2E_Operators_PasswordAndEmail` |
| 1 | `TestE2E_Operators_Preferences` |
| 1 | `TestE2E_Operators_Secrets` |
| 1 | `TestE2E_Operators_UpdateOperatorOperation` |
| 1 | `TestE2E_Operators_MultipleOperators` |
| 1 | `TestE2E_Operators_FullWorkflow` |
| 4 | `TestE2E_Files_UploadDownloadDelete` |
| 4 | `TestE2E_Files_MultipleUploads` |
| 4 | `TestE2E_Files_PreviewFile` |
| 4 | `TestE2E_Files_BulkDownload` |
| 4 | `TestE2E_Files_FullWorkflow` |
| 4 | `TestE2E_Hosts_GetHostNetworkMap` |
| 5 | `TestE2E_Credentials_CreateAndManage` |
| 5 | `TestE2E_Artifacts_CreateAndManage` |
| 5 | `TestE2E_CredentialsArtifacts_FullWorkflow` |

---

## Recommended Fix Priority

### Priority 1: Fix `CallMCPTool()` normalizer in `e2e_helpers.go` (fixes ~28 tests)

Add response unwrapping logic to handle `{count, items}` wrapper objects and surface `isError: true` as Go errors:

```go
// After normalizing content, unwrap {count, items} wrappers
if contentMap, ok := normalizedResult["content"].(map[string]interface{}); ok {
    if items, hasItems := contentMap["items"]; hasItems {
        normalizedResult["content"] = items  // Unwrap to bare array
    }
}

// Surface isError as Go-level error
if isError, ok := normalizedResult["isError"].(bool); ok && isError {
    // Extract error text from mcp_content
    errorText := extractErrorText(normalizedResult)
    return normalizedResult, fmt.Errorf("MCP tool error: %s", errorText)
}
```

### Priority 2: Fix `createAPITokenArgs` schema (fixes 2 tests)

Add `token_type` parameter to the `createAPITokenArgs` struct:

```go
type createAPITokenArgs struct {
    TokenType string `json:"token_type,omitempty" jsonschema:"Token type (User or C2)"`
}
```

### Priority 3: Investigate `mythic_update_operation` handler (fixes 4 tests)

The update mutation may be:
- Using wrong field names in the GraphQL mutation
- Not returning the updated object
- Not supporting webhook/channel fields

### Priority 4: Handle `CreateTagType` API limitation (fixes 4 tests)

Options:
- Skip tag type creation tests with `t.Skip("CreateTagType not supported by Mythic GraphQL API")`
- Implement tag type creation via Mythic REST API if available
- Use pre-existing tag types in the test environment

---

## Phase-by-Phase Results

### Phase 0: Authentication (5 tests)
| Test | Result | Root Cause |
|------|--------|------------|
| `TestE2E_Auth_LoginLogout` | ✅ PASS | — |
| `TestE2E_Auth_APITokens` | ❌ FAIL | RC-4 (schema mismatch) |
| `TestE2E_Auth_RefreshToken` | ✅ PASS | — |
| `TestE2E_Auth_ErrorHandling` | ❌ FAIL | RC-2 (error not propagated) |
| **Subtotal** | **2 PASS, 2 FAIL, 1 SKIP** | |

### Phase 1: Core Setup (22 tests)
| Test | Result | Root Cause |
|------|--------|------------|
| `TestE2E_Operations_GetOperations` | ❌ FAIL | RC-1 |
| `TestE2E_Operations_CreateAndManage` | ❌ FAIL | RC-1, RC-3 |
| `TestE2E_Operations_CurrentOperation` | ❌ FAIL | RC-3 |
| `TestE2E_Operations_EventLog` | ❌ FAIL | RC-1 |
| `TestE2E_Operations_Operators` | ❌ FAIL | RC-1 |
| `TestE2E_Operations_ErrorHandling` | ❌ FAIL | RC-2 |
| `TestE2E_Operations_FullWorkflow` | ✅ PASS | — |
| `TestE2E_Operations_GlobalSettings` | ✅ PASS | — |
| `TestE2E_Operators_GetOperators` | ❌ FAIL | RC-1 |
| `TestE2E_Operators_InviteLinks` | ❌ FAIL | RC-1 |
| `TestE2E_Operators_ErrorHandling` | ❌ FAIL | RC-2 |
| `TestE2E_Operators_CreateAndManage` | ✅ PASS | — |
| `TestE2E_Operators_PasswordAndEmail` | ✅ PASS | — |
| `TestE2E_Operators_Preferences` | ✅ PASS | — |
| `TestE2E_Operators_Secrets` | ✅ PASS | — |
| `TestE2E_Operators_UpdateOperatorOperation` | ✅ PASS | — |
| `TestE2E_Operators_MultipleOperators` | ✅ PASS | — |
| `TestE2E_Operators_FullWorkflow` | ✅ PASS | — |
| `TestE2E_Tags_*` (8 tests) | ❌ ALL FAIL | RC-1, RC-2, RC-4 |
| **Subtotal** | **10 PASS, 17 FAIL** | |

### Phase 2: C2 & Payloads (0 tests ran — no failed tests listed)

All C2/payload tests appear to have passed or not been included in the failure logs.

### Phase 3: Agents & Callbacks (10 tests)
| Test | Result | Root Cause |
|------|--------|------------|
| `TestE2E_Callbacks_GetAllCallbacks` | ❌ FAIL | RC-1 |
| `TestE2E_Callbacks_GetActiveCallbacks` | ❌ FAIL | RC-1 |
| `TestE2E_Callbacks_ErrorHandling` | ❌ FAIL | RC-2 |
| 7 other callback tests | ⏭️ SKIP | No active callbacks |
| **Subtotal** | **0 PASS, 3 FAIL, 7 SKIP** | |

### Phase 4: Tasks & Responses (combined with Files, Hosts, etc.)
| Test | Result | Root Cause |
|------|--------|------------|
| `TestE2E_Files_GetFiles` | ❌ FAIL | RC-1 |
| `TestE2E_Files_GetDownloadedFiles` | ❌ FAIL | RC-1 |
| `TestE2E_Files_ErrorHandling` | ❌ FAIL | RC-2 |
| `TestE2E_Files_UploadDownloadDelete` | ✅ PASS | — |
| `TestE2E_Files_MultipleUploads` | ✅ PASS | — |
| `TestE2E_Files_PreviewFile` | ✅ PASS | — |
| `TestE2E_Files_BulkDownload` | ✅ PASS | — |
| `TestE2E_Files_FullWorkflow` | ✅ PASS | — |
| `TestE2E_Hosts_GetHosts` | ❌ FAIL | RC-1 |
| `TestE2E_Hosts_ErrorHandling` | ❌ FAIL | RC-2 |
| `TestE2E_Hosts_GetHostNetworkMap` | ✅ PASS | — |
| 5 other host tests | ⏭️ SKIP | No hosts |
| `TestE2E_Keylogs_GetKeylogs` | ❌ FAIL | RC-1 |
| `TestE2E_Keylogs_GetKeylogsByOperation` | ❌ FAIL | RC-1 |
| `TestE2E_Keylogs_ErrorHandling` | ❌ FAIL | RC-2 |
| 3 other keylog tests | ⏭️ SKIP | No keylogs |
| `TestE2E_Processes_GetProcesses` | ❌ FAIL | RC-1 |
| `TestE2E_Processes_GetProcessesByOperation` | ❌ FAIL | RC-1 |
| `TestE2E_Processes_ErrorHandling` | ❌ FAIL | RC-2 |
| 5 other process tests | ⏭️ SKIP | No processes |
| `TestE2E_Screenshots_ErrorHandling` | ❌ FAIL | RC-2 |
| 8 other screenshot tests | ⏭️ SKIP | No screenshots |
| `TestE2E_Tasks_ErrorHandling` | ❌ FAIL | RC-2 |
| `TestE2E_Responses_GetLatestResponses` | ❌ FAIL | RC-1 |
| `TestE2E_Responses_SearchResponses` | ❌ FAIL | RC-1 |
| 10 other task/response tests | ⏭️ SKIP | No tasks |

### Phase 5: Artifacts & Credentials (9 tests)
| Test | Result | Root Cause |
|------|--------|------------|
| `TestE2E_Credentials_CreateAndManage` | ✅ PASS | — |
| `TestE2E_Credentials_GetCredentials` | ❌ FAIL | RC-1 |
| `TestE2E_Credentials_GetByOperation` | ❌ FAIL | RC-1 |
| `TestE2E_Artifacts_CreateAndManage` | ✅ PASS | — |
| `TestE2E_Artifacts_GetArtifacts` | ❌ FAIL | RC-1 |
| `TestE2E_Artifacts_GetByOperation` | ❌ FAIL | RC-1 |
| `TestE2E_Artifacts_GetByHost` | ❌ FAIL | RC-1 |
| `TestE2E_Artifacts_GetByType` | ❌ FAIL | RC-1 |
| `TestE2E_CredentialsArtifacts_ErrorHandling` | ❌ FAIL | RC-2 |
| `TestE2E_CredentialsArtifacts_FullWorkflow` | ✅ PASS | — |
| **Subtotal** | **3 PASS, 7 FAIL** | |

---

## Appendix: Raw Error Evidence

### RC-1 Example (`mythic_get_operations`):

```
DEBUG: Normalized result for mythic_get_operations:
{
  "content": {
    "count": 1,
    "items": [
      {
        "admin": { "username": "mythic_admin", ... },
        "admin_id": 1,
        ...
      }
    ]
  },
  ...
}
DEBUG: content field type: map[string]interface {}

Error Trace: e2e_operations_test.go:25
Error:       Should be true
Messages:    Expected content to be an array
```

### RC-2 Example (`mythic_login` with invalid creds):

```
Error Trace: e2e_auth_test.go:149
Error:       An error is expected but got nil.
Test:        TestE2E_Auth_ErrorHandling/LoginWithInvalidCredentials
```

### RC-3 Example (`mythic_update_operation`):

```
Error:     Not equal:
           expected: "https://example.com/webhook"
           actual  : ""
Test:      TestE2E_Operations_CreateAndManage

Error:     Not equal:
           expected: "Updated Test Operation"
           actual  : "Test Operation E2E"
Test:      TestE2E_Operations_CreateAndManage
```

### RC-4 Example (`mythic_create_api_token`):

```
Error:     Received unexpected error:
           MCP error: invalid params: validating "arguments": validating root:
           unexpected additional properties ["token_type"]
Test:      TestE2E_Auth_APITokens/CreateAPIToken
```
