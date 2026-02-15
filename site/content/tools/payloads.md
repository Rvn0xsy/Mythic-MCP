# Payloads

Build, download, manage, and inspect agent payload binaries.

!!! info "Auto-generated"
    This page is **generated from source code** by the schema extractor.
    Do not edit by hand — changes will be overwritten.

---


## `mythic_get_payloads`

Get a list of all payloads in Mythic

_No parameters._

---


## `mythic_get_payload`

Get details of a specific payload by UUID

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `payload_uuid` | `string` | :material-check-bold:{ title="Required" } **required** | UUID of the payload to retrieve |


---


## `mythic_get_payload_types`

Get list of available payload types (agent types). Each payload type includes a supported_c2_profiles array listing which C2 profiles it can use. Use this to verify C2 profile compatibility BEFORE creating a payload.

_No parameters._

---


## `mythic_create_payload`

Create/build a new payload

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `payload_type` | `string` | :material-check-bold:{ title="Required" } **required** | Payload type name (agent type) |
| `description` | `string` | _optional_ | Description of the payload |
| `tag` | `string` | _optional_ | Tag for the payload |
| `filename` | `string` | _optional_ | Filename for the payload |
| `os` | `string` | _optional_ | Operating system for the payload |
| `selected_os` | `string` | _optional_ | Selected OS variant |
| `commands` | `[]string` | _optional_ | List of command names to include |
| `c2_profiles` | `[]map[string]any` | _optional_ | C2 profile configurations. Each entry is {\ |
| `build_parameters` | `map[string]any` | _optional_ | Build parameter key-value pairs |
| `wrapper_payload` | `string` | _optional_ | UUID of payload to wrap |


---


## `mythic_update_payload`

Update a payload's properties (description, tag, etc.)

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `payload_uuid` | `string` | :material-check-bold:{ title="Required" } **required** | UUID of the payload to update |
| `description` | `*string` | _optional_ | Update payload description |
| `callback_alert` | `*bool` | _optional_ | Update callback alert setting |
| `deleted` | `*bool` | _optional_ | Mark payload as deleted |


---


## `mythic_delete_payload`

Delete a payload from Mythic

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `payload_uuid` | `string` | :material-check-bold:{ title="Required" } **required** | UUID of the payload to delete |


---


## `mythic_rebuild_payload`

Rebuild/regenerate an existing payload

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `payload_uuid` | `string` | :material-check-bold:{ title="Required" } **required** | UUID of the payload to rebuild |


---


## `mythic_export_payload_config`

Export a payload's configuration as JSON

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `payload_uuid` | `string` | :material-check-bold:{ title="Required" } **required** | UUID of the payload to export |


---


## `mythic_get_payload_commands`

Get list of commands available in a payload

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `payload_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the payload |


---


## `mythic_get_payload_on_host`

Get list of payloads deployed on hosts in an operation

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `operation_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the operation |


---


## `mythic_wait_for_payload`

Wait for a payload build to complete with timeout

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `payload_uuid` | `string` | :material-check-bold:{ title="Required" } **required** | UUID of the payload to wait for |
| `timeout` | `int` | _optional_ | Timeout in seconds (default 60) |


---


## `mythic_download_payload`

Get a one-time download URL for a built payload binary. Use curl or wget with the returned download_url to fetch the file. The URL token is single-use and expires after 5 minutes.

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `payload_uuid` | `string` | :material-check-bold:{ title="Required" } **required** | UUID of the payload to download |


---


