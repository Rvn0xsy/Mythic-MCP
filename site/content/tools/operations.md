# Operations

Operation (campaign) CRUD, event logging, global settings.

!!! info "Auto-generated"
    This page is **generated from source code** by the schema extractor.
    Do not edit by hand — changes will be overwritten.

---


## `mythic_get_operations`

Get a list of all operations in the Mythic instance

_No parameters._

---


## `mythic_get_operation`

Get details of a specific operation by ID

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `operation_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the operation to retrieve |


---


## `mythic_create_operation`

Create a new operation (campaign/engagement)

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | `string` | :material-check-bold:{ title="Required" } **required** | Name of the new operation |
| `webhook` | `*string` | _optional_ | Webhook URL for notifications |
| `channel` | `*string` | _optional_ | Slack/Discord channel for notifications |
| `admin_id` | `*int` | _optional_ | Operator ID to set as admin |


---


## `mythic_update_operation`

Update an existing operation's properties

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `operation_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the operation to update |
| `name` | `*string` | _optional_ | New name for the operation |
| `webhook` | `*string` | _optional_ | Webhook URL for notifications |
| `channel` | `*string` | _optional_ | Slack/Discord channel |
| `complete` | `*bool` | _optional_ | Mark operation as complete |
| `admin_id` | `*int` | _optional_ | New admin operator ID |
| `banner_text` | `*string` | _optional_ | Banner text for operation |
| `banner_color` | `*string` | _optional_ | Banner color (hex code) |


---


## `mythic_set_current_operation`

Set the current operation context for the client

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `operation_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the operation to set as current |


---


## `mythic_get_current_operation`

Get the currently active operation context

_No parameters._

---


## `mythic_get_operation_operators`

Get list of operators (users) in a specific operation

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `operation_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the operation |


---


## `mythic_create_event_log`

Create an event log entry for an operation

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `operation_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the operation |
| `message` | `string` | :material-check-bold:{ title="Required" } **required** | Event log message |
| `level` | `*string` | _optional_ | Log level (info/warning/error) |
| `source` | `*string` | _optional_ | Source of the event |


---


## `mythic_get_event_log`

Get event log entries for an operation

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `operation_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the operation |
| `limit` | `int` | _optional_ | Maximum number of log entries to return (default 100) |


---


## `mythic_get_global_settings`

Get global Mythic server settings

_No parameters._

---


## `mythic_update_global_settings`

Update global Mythic server settings

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `settings` | `map[string]any` | :material-check-bold:{ title="Required" } **required** | Settings to update (key-value pairs) |


---


