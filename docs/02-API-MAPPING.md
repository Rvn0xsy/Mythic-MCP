# Mythic SDK to MCP Tools API Mapping

**Author:** Claude Code
**Date:** 2026-01-24
**Purpose:** Complete mapping of Mythic SDK methods to MCP tools
**Coverage:** 204+ Mythic SDK methods

---

## Table of Contents

1. [Mapping Overview](#mapping-overview)
2. [Authentication & Session](#authentication--session)
3. [Operations Management](#operations-management)
4. [Callbacks](#callbacks)
5. [Tasks & Responses](#tasks--responses)
6. [Payloads](#payloads)
7. [File Operations](#file-operations)
8. [Credentials & Artifacts](#credentials--artifacts)
9. [C2 Profiles](#c2-profiles)
10. [MITRE ATT&CK](#mitre-attck)
11. [Operators & Users](#operators--users)
12. [Tags & Categorization](#tags--categorization)
13. [Advanced Features](#advanced-features)
14. [Tool Naming Convention](#tool-naming-convention)

---

## Mapping Overview

### Design Principles

1. **One-to-One Mapping** - Each Mythic SDK method becomes one MCP tool
2. **Consistent Naming** - `mythic_{category}_{action}` pattern
3. **Type Preservation** - All SDK types mapped to JSON schemas
4. **Full Coverage** - No SDK functionality omitted

### Mapping Statistics

| Category | SDK Methods | MCP Tools | Status |
|----------|-------------|-----------|--------|
| Authentication & Session | 7 | 7 | Designed |
| Operations Management | 11 | 11 | Designed |
| Callbacks | 14 | 14 | Designed |
| Tasks & Responses | 20 | 20 | Designed |
| Payloads | 14 | 14 | Designed |
| File Operations | 10 | 10 | Designed |
| Credentials & Artifacts | 12 | 12 | Designed |
| C2 Profiles | 9 | 9 | Designed |
| MITRE ATT&CK | 7 | 7 | Designed |
| Operators & Users | 12 | 12 | Designed |
| Tags & Categorization | 9 | 9 | Designed |
| Processes | 6 | 6 | Designed |
| Hosts | 6 | 6 | Designed |
| Screenshots | 6 | 6 | Designed |
| Keylogs | 3 | 3 | Designed |
| Tokens | 3 | 3 | Designed |
| File Browser | 3 | 3 | Designed |
| RPFWD/Proxy | 6 | 6 | Designed |
| Eventing | 14 | 14 | Designed |
| Containers | 4 | 4 | Designed |
| Utilities | 8 | 8 | Designed |
| Alerts | 6 | 6 | Designed |
| Reporting | 3 | 3 | Designed |
| Browser Scripts | 2 | 2 | Designed |
| Build Parameters | 6 | 6 | Designed |
| Commands | 4 | 4 | Designed |
| **TOTAL** | **204** | **204** | **100%** |

---

## Authentication & Session

### SDK Methods → MCP Tools

| SDK Method | MCP Tool | Input Schema | Output |
|------------|----------|--------------|--------|
| `Login(ctx, username, password)` | `mythic_login` | `{username, password}` | Session info |
| `Logout(ctx)` | `mythic_logout` | `{}` | Success status |
| `IsAuthenticated()` | `mythic_is_authenticated` | `{}` | Boolean |
| `GetMe(ctx)` | `mythic_get_current_user` | `{}` | Operator object |
| `CreateAPIToken(ctx, req)` | `mythic_create_api_token` | `{token_type}` | API token |
| `DeleteAPIToken(ctx, id)` | `mythic_delete_api_token` | `{token_id}` | Success status |
| `RefreshAccessToken(ctx)` | `mythic_refresh_token` | `{}` | New token |

### Example Tool Definition

```json
{
  "name": "mythic_login",
  "description": "Authenticate with Mythic server using username and password",
  "inputSchema": {
    "type": "object",
    "properties": {
      "username": {
        "type": "string",
        "description": "Mythic username"
      },
      "password": {
        "type": "string",
        "description": "Mythic password"
      }
    },
    "required": ["username", "password"]
  }
}
```

---

## Operations Management

### SDK Methods → MCP Tools

| SDK Method | MCP Tool | Key Inputs | Purpose |
|------------|----------|------------|---------|
| `GetOperations(ctx)` | `mythic_get_operations` | - | List all operations |
| `GetOperationByID(ctx, id)` | `mythic_get_operation` | `operation_id` | Get specific operation |
| `CreateOperation(ctx, req)` | `mythic_create_operation` | `name, webhook, channel` | Create new operation |
| `UpdateOperation(ctx, id, req)` | `mythic_update_operation` | `operation_id, updates` | Modify operation |
| `SetCurrentOperation(id)` | `mythic_set_current_operation` | `operation_id` | Switch operation context |
| `GetCurrentOperation()` | `mythic_get_current_operation` | - | Get active operation |
| `GetOperatorsByOperation(ctx, id)` | `mythic_get_operation_operators` | `operation_id` | List operators in operation |
| `CreateOperationEventLog(ctx, req)` | `mythic_create_event_log` | `operation_id, message` | Log operation event |
| `GetOperationEventLog(ctx, id)` | `mythic_get_event_log` | `operation_id` | Get event logs |
| `GetGlobalSettings(ctx)` | `mythic_get_global_settings` | - | Get Mythic settings |
| `UpdateGlobalSettings(ctx, req)` | `mythic_update_global_settings` | `settings` | Update global config |

---

## Callbacks

### SDK Methods → MCP Tools

| SDK Method | MCP Tool | Key Inputs | Purpose |
|------------|----------|------------|---------|
| `GetAllCallbacks(ctx)` | `mythic_get_all_callbacks` | - | List all callbacks |
| `GetAllActiveCallbacks(ctx)` | `mythic_get_active_callbacks` | - | List active callbacks |
| `GetCallbackByID(ctx, id)` | `mythic_get_callback` | `callback_id` | Get callback details |
| `UpdateCallback(ctx, id, req)` | `mythic_update_callback` | `callback_id, updates` | Modify callback |
| `DeleteCallback(ctx, id)` | `mythic_delete_callback` | `callback_id` | Remove callback |
| `GetCallbacksForHost(ctx, host)` | `mythic_get_host_callbacks` | `hostname` | List callbacks on host |
| `GetLoadedCommands(ctx, id)` | `mythic_get_loaded_commands` | `callback_id` | List available commands |
| `ExportCallbackConfig(ctx, id)` | `mythic_export_callback_config` | `callback_id` | Export callback JSON |
| `CreateCallback(ctx, req)` | `mythic_create_callback` | `payload_uuid, ...` | Manual callback creation |
| `GetCallbackTokensByCallback(ctx, id)` | `mythic_get_callback_tokens` | `callback_id` | List callback tokens |
| `AddCallbackGraphEdge(ctx, req)` | `mythic_add_callback_edge` | `parent, child, c2` | Link P2P callbacks |
| `RemoveCallbackGraphEdge(ctx, req)` | `mythic_remove_callback_edge` | `parent, child` | Unlink P2P callbacks |
| `HideCallback(ctx, id)` | `mythic_hide_callback` | `callback_id` | Hide from UI |
| `UnhideCallback(ctx, id)` | `mythic_unhide_callback` | `callback_id` | Show in UI |

### Example Tool: Issue Task

```json
{
  "name": "mythic_issue_task",
  "description": "Execute a command on a callback",
  "inputSchema": {
    "type": "object",
    "properties": {
      "callback_id": {
        "type": "integer",
        "description": "Callback ID to task"
      },
      "command": {
        "type": "string",
        "description": "Command name (e.g., 'shell', 'download')"
      },
      "params": {
        "type": "string",
        "description": "Command parameters (JSON string)"
      },
      "files": {
        "type": "array",
        "items": {"type": "string"},
        "description": "File UUIDs to include"
      }
    },
    "required": ["callback_id", "command"]
  }
}
```

---

## Tasks & Responses

### Task Operations

| SDK Method | MCP Tool | Key Inputs | Purpose |
|------------|----------|------------|---------|
| `IssueTask(ctx, req)` | `mythic_issue_task` | `callback_id, command, params` | Execute command |
| `GetTask(ctx, id)` | `mythic_get_task` | `task_id` | Get task details |
| `UpdateTask(ctx, id, req)` | `mythic_update_task` | `task_id, updates` | Modify task |
| `GetTasksForCallback(ctx, id)` | `mythic_get_callback_tasks` | `callback_id` | List callback tasks |
| `GetTasksByStatus(ctx, status)` | `mythic_get_tasks_by_status` | `status` | Filter by status |
| `WaitForTaskComplete(ctx, id, timeout)` | `mythic_wait_for_task` | `task_id, timeout` | Wait for completion |
| `GetTaskOutput(ctx, id)` | `mythic_get_task_output` | `task_id` | Get task results |
| `ReissueTask(ctx, id)` | `mythic_reissue_task` | `task_id` | Re-execute task |
| `ReissueTaskWithHandler(ctx, id, handler)` | `mythic_reissue_task_with_handler` | `task_id` | Re-execute with callback |
| `GetTaskArtifacts(ctx, id)` | `mythic_get_task_artifacts` | `task_id` | Get task artifacts |

### Response Operations

| SDK Method | MCP Tool | Key Inputs | Purpose |
|------------|----------|------------|---------|
| `GetResponsesByTask(ctx, id)` | `mythic_get_task_responses` | `task_id` | Get task responses |
| `GetResponsesByCallback(ctx, id)` | `mythic_get_callback_responses` | `callback_id` | Get callback responses |
| `GetResponseByID(ctx, id)` | `mythic_get_response` | `response_id` | Get specific response |
| `GetLatestResponses(ctx, limit)` | `mythic_get_latest_responses` | `limit` | Get recent responses |
| `SearchResponses(ctx, query)` | `mythic_search_responses` | `query` | Search response text |
| `GetResponseStatistics(ctx)` | `mythic_get_response_statistics` | - | Get response stats |

---

## Payloads

### SDK Methods → MCP Tools

| SDK Method | MCP Tool | Key Inputs | Purpose |
|------------|----------|------------|---------|
| `GetPayloadTypes(ctx)` | `mythic_get_payload_types` | - | List available types |
| `GetPayloads(ctx)` | `mythic_get_payloads` | - | List all payloads |
| `GetPayloadByUUID(ctx, uuid)` | `mythic_get_payload` | `payload_uuid` | Get payload info |
| `CreatePayload(ctx, req)` | `mythic_create_payload` | `type, os, c2profiles, commands` | Build new payload |
| `UpdatePayload(ctx, uuid, req)` | `mythic_update_payload` | `payload_uuid, updates` | Modify payload |
| `DeletePayload(ctx, uuid)` | `mythic_delete_payload` | `payload_uuid` | Remove payload |
| `DownloadPayload(ctx, uuid)` | `mythic_download_payload` | `payload_uuid` | Download binary |
| `ExportPayloadConfig(ctx, uuid)` | `mythic_export_payload_config` | `payload_uuid` | Export JSON config |
| `RebuildPayload(ctx, uuid)` | `mythic_rebuild_payload` | `payload_uuid` | Rebuild payload |
| `WaitForPayloadComplete(ctx, uuid, timeout)` | `mythic_wait_for_payload` | `payload_uuid, timeout` | Wait for build |
| `GetPayloadCommands(ctx, type)` | `mythic_get_payload_commands` | `payload_type` | List available commands |
| `GetPayloadOnHost(ctx, host)` | `mythic_get_host_payloads` | `hostname` | Find payloads on host |
| `GetBuildParameters(ctx)` | `mythic_get_build_parameters` | - | List all build params |
| `GetBuildParametersByPayloadType(ctx, type)` | `mythic_get_payload_build_parameters` | `payload_type` | Get type params |

---

## File Operations

### SDK Methods → MCP Tools

| SDK Method | MCP Tool | Key Inputs | Purpose |
|------------|----------|------------|---------|
| `GetFiles(ctx)` | `mythic_get_files` | - | List all files |
| `GetFileByID(ctx, id)` | `mythic_get_file` | `file_id` | Get file info |
| `GetDownloadedFiles(ctx)` | `mythic_get_downloaded_files` | - | List downloaded files |
| `UploadFile(ctx, data, filename)` | `mythic_upload_file` | `file_data, filename` | Upload file |
| `DownloadFile(ctx, uuid)` | `mythic_download_file` | `file_uuid` | Download file |
| `DeleteFile(ctx, id)` | `mythic_delete_file` | `file_id` | Remove file |
| `BulkDownloadFiles(ctx, ids)` | `mythic_bulk_download_files` | `file_ids[]` | Download multiple files (ZIP) |
| `PreviewFile(ctx, id)` | `mythic_preview_file` | `file_id` | Preview file content |
| `RegisterFile(ctx, data, filename)` | `mythic_register_file` | `file_data, filename` | Register for upload |
| `GetFileMetadata(ctx, uuid)` | `mythic_get_file_metadata` | `file_uuid` | Get file metadata |

---

## Credentials & Artifacts

### Credentials

| SDK Method | MCP Tool | Key Inputs | Purpose |
|------------|----------|------------|---------|
| `GetCredentials(ctx)` | `mythic_get_credentials` | - | List all credentials |
| `GetCredentialsByOperation(ctx, id)` | `mythic_get_operation_credentials` | `operation_id` | Filter by operation |
| `CreateCredential(ctx, req)` | `mythic_create_credential` | `type, account, credential` | Add credential |
| `UpdateCredential(ctx, id, req)` | `mythic_update_credential` | `credential_id, updates` | Modify credential |
| `DeleteCredential(ctx, id)` | `mythic_delete_credential` | `credential_id` | Remove credential |

### Artifacts

| SDK Method | MCP Tool | Key Inputs | Purpose |
|------------|----------|------------|---------|
| `GetArtifacts(ctx)` | `mythic_get_artifacts` | - | List all artifacts |
| `GetArtifactsByOperation(ctx, id)` | `mythic_get_operation_artifacts` | `operation_id` | Filter by operation |
| `GetArtifactsByHost(ctx, host)` | `mythic_get_host_artifacts` | `hostname` | Filter by host |
| `GetArtifactsByType(ctx, type)` | `mythic_get_artifacts_by_type` | `artifact_type` | Filter by type |
| `CreateArtifact(ctx, req)` | `mythic_create_artifact` | `type, artifact, host` | Add artifact |
| `UpdateArtifact(ctx, id, req)` | `mythic_update_artifact` | `artifact_id, updates` | Modify artifact |
| `DeleteArtifact(ctx, id)` | `mythic_delete_artifact` | `artifact_id` | Remove artifact |

---

## C2 Profiles

### SDK Methods → MCP Tools

| SDK Method | MCP Tool | Key Inputs | Purpose |
|------------|----------|------------|---------|
| `GetC2Profiles(ctx)` | `mythic_get_c2_profiles` | - | List all C2 profiles |
| `GetC2ProfileByID(ctx, id)` | `mythic_get_c2_profile` | `profile_id` | Get profile info |
| `CreateC2Instance(ctx, req)` | `mythic_create_c2_instance` | `profile, parameters` | Create C2 instance |
| `ImportC2Instance(ctx, data)` | `mythic_import_c2_instance` | `config_json` | Import C2 config |
| `StartStopProfile(ctx, name, start)` | `mythic_control_c2_profile` | `profile_name, action` | Start/stop profile |
| `GetProfileOutput(ctx, name)` | `mythic_get_c2_output` | `profile_name` | Get profile logs |
| `C2HostFile(ctx, req)` | `mythic_c2_host_file` | `profile, file_uuid` | Host file on C2 |
| `C2SampleMessage(ctx, req)` | `mythic_c2_sample_message` | `profile` | Generate sample message |
| `C2GetIOC(ctx, name)` | `mythic_c2_get_iocs` | `profile_name` | Get profile IOCs |

---

## MITRE ATT&CK

### SDK Methods → MCP Tools

| SDK Method | MCP Tool | Key Inputs | Purpose |
|------------|----------|------------|---------|
| `GetAttackTechniques(ctx)` | `mythic_get_attack_techniques` | - | List all techniques |
| `GetAttackTechniqueByID(ctx, id)` | `mythic_get_attack_technique` | `technique_id` | Get technique details |
| `GetAttackTechniqueByTNum(ctx, tnum)` | `mythic_get_attack_technique_by_tnum` | `t_number` | Get by T-number (e.g., T1059) |
| `GetAttackByCommand(ctx, cmd)` | `mythic_get_command_attack_mappings` | `command_name` | Get command mappings |
| `GetAttackByTask(ctx, id)` | `mythic_get_task_attack_mappings` | `task_id` | Get task mappings |
| `AddMITREAttackToTask(ctx, taskID, techID)` | `mythic_add_attack_to_task` | `task_id, technique_id` | Add mapping |
| `GetAttacksByOperation(ctx, id)` | `mythic_get_operation_attack_stats` | `operation_id` | Get operation stats |

---

## Operators & Users

### SDK Methods → MCP Tools

| SDK Method | MCP Tool | Key Inputs | Purpose |
|------------|----------|------------|---------|
| `GetOperators(ctx)` | `mythic_get_operators` | - | List all operators |
| `GetOperatorByID(ctx, id)` | `mythic_get_operator` | `operator_id` | Get operator info |
| `CreateOperator(ctx, req)` | `mythic_create_operator` | `username, password` | Create new operator |
| `UpdateOperatorStatus(ctx, id, active)` | `mythic_update_operator_status` | `operator_id, active` | Activate/deactivate |
| `UpdateOperatorOperation(ctx, id, opID)` | `mythic_update_operator_operation` | `operator_id, operation_id` | Assign to operation |
| `GetOperatorPreferences(ctx, id)` | `mythic_get_operator_preferences` | `operator_id` | Get preferences |
| `UpdateOperatorPreferences(ctx, id, prefs)` | `mythic_update_operator_preferences` | `operator_id, preferences` | Update preferences |
| `GetOperatorSecrets(ctx, id)` | `mythic_get_operator_secrets` | `operator_id` | Get secrets |
| `UpdateOperatorSecrets(ctx, id, secrets)` | `mythic_update_operator_secrets` | `operator_id, secrets` | Update secrets |
| `CreateInviteLink(ctx)` | `mythic_create_invite_link` | - | Generate invite link |
| `GetInviteLinks(ctx)` | `mythic_get_invite_links` | - | List invite links |
| `UpdatePasswordAndEmail(ctx, req)` | `mythic_update_password_email` | `current_password, new_password, email` | Update credentials |

---

## Tags & Categorization

### SDK Methods → MCP Tools

| SDK Method | MCP Tool | Key Inputs | Purpose |
|------------|----------|------------|---------|
| `CreateTag(ctx, req)` | `mythic_create_tag` | `tag_type_id, data` | Create tag |
| `GetTags(ctx)` | `mythic_get_tags` | - | List all tags |
| `GetTagsByOperation(ctx, id)` | `mythic_get_operation_tags` | `operation_id` | Filter by operation |
| `DeleteTag(ctx, id)` | `mythic_delete_tag` | `tag_id` | Remove tag |
| `CreateTagType(ctx, req)` | `mythic_create_tag_type` | `name, color, description` | Create tag type |
| `GetTagTypes(ctx)` | `mythic_get_tag_types` | - | List tag types |
| `GetTagTypesByOperation(ctx, id)` | `mythic_get_operation_tag_types` | `operation_id` | Filter by operation |
| `UpdateTagType(ctx, id, req)` | `mythic_update_tag_type` | `tag_type_id, updates` | Modify tag type |
| `DeleteTagType(ctx, id)` | `mythic_delete_tag_type` | `tag_type_id` | Remove tag type |

---

## Advanced Features

### Processes

| SDK Method | MCP Tool | Purpose |
|------------|----------|---------|
| `GetProcesses(ctx)` | `mythic_get_processes` | List all processes |
| `GetProcessTree(ctx, id)` | `mythic_get_process_tree` | Get process tree |
| `GetProcessesByCallback(ctx, id)` | `mythic_get_callback_processes` | Filter by callback |
| `GetProcessesByOperation(ctx, id)` | `mythic_get_operation_processes` | Filter by operation |
| `GetProcessesByHost(ctx, host)` | `mythic_get_host_processes` | Filter by host |
| `GetProcessByID(ctx, id)` | `mythic_get_process` | Get specific process |

### Hosts

| SDK Method | MCP Tool | Purpose |
|------------|----------|---------|
| `GetHosts(ctx)` | `mythic_get_hosts` | List all hosts |
| `GetHostByID(ctx, id)` | `mythic_get_host` | Get host details |
| `GetHostByHostname(ctx, hostname)` | `mythic_get_host_by_name` | Get by hostname |
| `GetCallbacksForHost(ctx, host)` | `mythic_get_host_callbacks` | List host callbacks |
| `GetHostNetworkMap(ctx)` | `mythic_get_network_map` | Get network topology |
| `GetProcessesByHost(ctx, host)` | `mythic_get_host_processes` | List host processes |

### Screenshots

| SDK Method | MCP Tool | Purpose |
|------------|----------|---------|
| `GetScreenshots(ctx)` | `mythic_get_screenshots` | List all screenshots |
| `GetScreenshotByID(ctx, id)` | `mythic_get_screenshot` | Get screenshot |
| `GetScreenshotThumbnail(ctx, id)` | `mythic_get_screenshot_thumbnail` | Get thumbnail |
| `GetScreenshotTimeline(ctx)` | `mythic_get_screenshot_timeline` | Get timeline |
| `DownloadScreenshot(ctx, id)` | `mythic_download_screenshot` | Download image |
| `DeleteScreenshot(ctx, id)` | `mythic_delete_screenshot` | Remove screenshot |

### Keylogs

| SDK Method | MCP Tool | Purpose |
|------------|----------|---------|
| `GetKeylogs(ctx)` | `mythic_get_keylogs` | List all keylogs |
| `GetKeylogsByCallback(ctx, id)` | `mythic_get_callback_keylogs` | Filter by callback |
| `GetKeylogsByOperation(ctx, id)` | `mythic_get_operation_keylogs` | Filter by operation |

### File Browser

| SDK Method | MCP Tool | Purpose |
|------------|----------|---------|
| `GetFileBrowserObjects(ctx)` | `mythic_get_file_browser_objects` | List browsed files |
| `GetFileBrowserObjectsByCallback(ctx, id)` | `mythic_get_callback_file_browser` | Filter by callback |
| `GetFileBrowserObjectsByHost(ctx, host)` | `mythic_get_host_file_browser` | Filter by host |

### RPFWD/Proxy

| SDK Method | MCP Tool | Purpose |
|------------|----------|---------|
| `CreateRPFWD(ctx, req)` | `mythic_create_port_forward` | Create port forward |
| `GetRPFWDs(ctx)` | `mythic_get_port_forwards` | List forwards |
| `GetRPFWDStatus(ctx, id)` | `mythic_get_port_forward_status` | Get forward status |
| `DeleteRPFWD(ctx, id)` | `mythic_delete_port_forward` | Remove forward |
| `TestProxy(ctx, id)` | `mythic_test_proxy` | Test proxy connection |
| `ToggleProxy(ctx, id, enabled)` | `mythic_toggle_proxy` | Enable/disable proxy |

### Eventing & Workflows

| SDK Method | MCP Tool | Purpose |
|------------|----------|---------|
| `EventingExportWorkflow(ctx, id)` | `mythic_export_workflow` | Export workflow |
| `EventingImportContainerWorkflow(ctx, data)` | `mythic_import_workflow` | Import workflow |
| `EventingTestFile(ctx, data)` | `mythic_test_workflow_file` | Test workflow |
| `EventingTriggerManual(ctx, id)` | `mythic_trigger_workflow` | Manual trigger |
| `EventingTriggerManualBulk(ctx, ids)` | `mythic_trigger_workflows_bulk` | Bulk trigger |
| `EventingTriggerUpdate(ctx, id, req)` | `mythic_update_workflow_trigger` | Update trigger |
| `EventingTriggerRetry(ctx, id)` | `mythic_retry_workflow` | Retry workflow |
| `EventingTriggerRetryFromStep(ctx, id, step)` | `mythic_retry_workflow_from_step` | Retry from step |
| `EventingTriggerRunAgain(ctx, id)` | `mythic_run_workflow_again` | Run again |
| `EventingTriggerCancel(ctx, id)` | `mythic_cancel_workflow` | Cancel workflow |
| `UpdateEventGroupApproval(ctx, id, approved)` | `mythic_approve_event_group` | Approve event |
| `SendExternalWebhook(ctx, req)` | `mythic_send_webhook` | Send webhook |
| `ConsumingServicesTestWebhook(ctx, name)` | `mythic_test_consuming_webhook` | Test webhook |
| `ConsumingServicesTestLog(ctx, name)` | `mythic_test_consuming_log` | Test log |

### Container Operations

| SDK Method | MCP Tool | Purpose |
|------------|----------|---------|
| `ContainerListFiles(ctx, name)` | `mythic_container_list_files` | List container files |
| `ContainerDownloadFile(ctx, name, path)` | `mythic_container_download_file` | Download file |
| `ContainerWriteFile(ctx, name, path, data)` | `mythic_container_write_file` | Write file |
| `ContainerRemoveFile(ctx, name, path)` | `mythic_container_remove_file` | Remove file |

### Alerts

| SDK Method | MCP Tool | Purpose |
|------------|----------|---------|
| `GetAlerts(ctx)` | `mythic_get_alerts` | List all alerts |
| `GetAlertByID(ctx, id)` | `mythic_get_alert` | Get alert details |
| `GetUnresolvedAlerts(ctx)` | `mythic_get_unresolved_alerts` | List unresolved |
| `ResolveAlert(ctx, id)` | `mythic_resolve_alert` | Mark resolved |
| `CreateCustomAlert(ctx, req)` | `mythic_create_alert` | Create custom alert |
| `GetAlertStatistics(ctx)` | `mythic_get_alert_statistics` | Get alert stats |

### Reporting

| SDK Method | MCP Tool | Purpose |
|------------|----------|---------|
| `GenerateReport(ctx, req)` | `mythic_generate_report` | Generate report |
| `GetRedirectRules(ctx)` | `mythic_get_redirect_rules` | Get redirect rules |
| `CustomBrowserExport(ctx, req)` | `mythic_custom_browser_export` | Export browser data |

### Browser Scripts

| SDK Method | MCP Tool | Purpose |
|------------|----------|---------|
| `GetBrowserScripts(ctx)` | `mythic_get_browser_scripts` | List browser scripts |
| `GetBrowserScriptsByOperation(ctx, id)` | `mythic_get_operation_browser_scripts` | Filter by operation |

### Utilities

| SDK Method | MCP Tool | Purpose |
|------------|----------|---------|
| `GetConfig()` | `mythic_get_client_config` | Get client config |
| `ConfigCheck(ctx)` | `mythic_check_config` | Validate config |
| `CreateRandom(ctx, length)` | `mythic_create_random` | Generate random data |
| `DynamicQueryFunction(ctx, query)` | `mythic_dynamic_query` | Execute dynamic query |
| `DynamicBuildParameter(ctx, param)` | `mythic_dynamic_build_parameter` | Dynamic build param |
| `TypedArrayParseFunction(ctx, data)` | `mythic_parse_typed_array` | Parse typed array |
| `GetStagingInfo(ctx)` | `mythic_get_staging_info` | Get staging info |
| `Close()` | `mythic_close_client` | Close client |

---

## Tool Naming Convention

### Pattern

```
mythic_{category}_{action}[_{target}]
```

### Examples

| Category | Action | Target | Tool Name |
|----------|--------|--------|-----------|
| callback | get | all | `mythic_get_all_callbacks` |
| callback | get | one | `mythic_get_callback` |
| callback | update | - | `mythic_update_callback` |
| task | issue | - | `mythic_issue_task` |
| task | get | output | `mythic_get_task_output` |
| payload | create | - | `mythic_create_payload` |
| file | upload | - | `mythic_upload_file` |
| operation | get | operators | `mythic_get_operation_operators` |

### Categories

- `auth` - Authentication
- `operation` - Operations
- `callback` - Callbacks
- `task` - Tasks
- `payload` - Payloads
- `file` - Files
- `credential` - Credentials
- `artifact` - Artifacts
- `c2` - C2 Profiles
- `attack` - MITRE ATT&CK
- `operator` - Operators
- `tag` - Tags
- `process` - Processes
- `host` - Hosts
- `screenshot` - Screenshots
- `keylog` - Keylogs
- `token` - Tokens
- `browser` - File Browser
- `proxy` - Proxies/Forwards
- `workflow` - Eventing
- `container` - Containers
- `alert` - Alerts
- `report` - Reporting

### Actions

- `get` - Retrieve data
- `create` - Create new resource
- `update` - Modify existing resource
- `delete` - Remove resource
- `upload` - Upload data
- `download` - Download data
- `issue` - Issue command/task
- `wait` - Wait for completion
- `export` - Export configuration
- `import` - Import configuration

---

## Implementation Priority

### Phase 1: Core Operations (15 tools)
1. Authentication (7 tools)
2. Operations basics (3 tools)
3. Basic queries (5 tools)

### Phase 2: Essential Operations (40 tools)
4. Callbacks (14 tools)
5. Tasks (10 tools)
6. Files (10 tools)
7. Operators (6 tools)

### Phase 3: Advanced Operations (60 tools)
8. Payloads (14 tools)
9. C2 Profiles (9 tools)
10. Processes & Hosts (12 tools)
11. Credentials & Artifacts (12 tools)
12. MITRE ATT&CK (7 tools)
13. Tags (9 tools)

### Phase 4: Specialized Features (89 tools)
14. Screenshots, Keylogs, Tokens (12 tools)
15. File Browser, RPFWD (9 tools)
16. Eventing & Workflows (14 tools)
17. Containers (4 tools)
18. Alerts & Reporting (9 tools)
19. Browser Scripts (2 tools)
20. Build Parameters (6 tools)
21. Commands (4 tools)
22. Responses (6 tools)
23. Utilities (8 tools)
24. Remaining operations (15 tools)

---

## Testing Strategy

Each tool must have:

1. **Input Validation Test** - Verify schema enforcement
2. **Happy Path Test** - Normal operation succeeds
3. **Error Handling Test** - Graceful failure on errors
4. **Integration Test** - End-to-end against real Mythic

**Goal:** >90% coverage matching upstream SDK quality

---

**Status:** Complete Mapping - Ready for Implementation
**Next Document:** [03-TEST-STRATEGY.md](03-TEST-STRATEGY.md)
