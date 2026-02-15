# Callbacks

Active agent sessions — list, update, P2P edges, tokens.

!!! info "Auto-generated"
    This page is **generated from source code** by the schema extractor.
    Do not edit by hand — changes will be overwritten.

---


## `mythic_get_all_callbacks`

Get a list of all callbacks (active agent connections) in Mythic

_No parameters._

---


## `mythic_get_active_callbacks`

Get a list of all active callbacks

_No parameters._

---


## `mythic_get_callback`

Get details of a specific callback by ID

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `callback_id` | `int` | :material-check-bold:{ title="Required" } **required** | Display ID of the callback to retrieve |


---


## `mythic_update_callback`

Update a callback's properties (description, active status, etc.)

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `callback_id` | `int` | :material-check-bold:{ title="Required" } **required** | Display ID of the callback to update |
| `active` | `*bool` | _optional_ | Set callback active/inactive status |
| `locked` | `*bool` | _optional_ | Lock/unlock callback for tasking |
| `description` | `*string` | _optional_ | Set callback description |
| `ips` | `[]string` | _optional_ | Update IP addresses |
| `user` | `*string` | _optional_ | Update username |
| `host` | `*string` | _optional_ | Update hostname |


---


## `mythic_delete_callback`

Delete one or more callbacks from Mythic

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `callback_ids` | `[]int` | :material-check-bold:{ title="Required" } **required** | Array of callback IDs to delete |


---


## `mythic_get_loaded_commands`

Get list of commands loaded in a specific callback

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `callback_id` | `int` | :material-check-bold:{ title="Required" } **required** | Display ID of the callback |


---


## `mythic_export_callback_config`

Export a callback's configuration as JSON

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `agent_callback_id` | `string` | :material-check-bold:{ title="Required" } **required** | Agent callback UUID to export |


---


## `mythic_import_callback_config`

Import a callback configuration from JSON

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `config` | `string` | :material-check-bold:{ title="Required" } **required** | JSON configuration string to import |


---


## `mythic_get_callback_tokens`

Get list of tokens associated with a callback

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `callback_id` | `int` | :material-check-bold:{ title="Required" } **required** | Display ID of the callback |


---


## `mythic_add_callback_edge`

Add a P2P connection between two callbacks in the callback graph

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `source_id` | `int` | :material-check-bold:{ title="Required" } **required** | Source callback ID |
| `destination_id` | `int` | :material-check-bold:{ title="Required" } **required** | Destination callback ID |
| `c2_profile_name` | `string` | :material-check-bold:{ title="Required" } **required** | C2 profile name for the connection |


---


## `mythic_remove_callback_edge`

Remove a P2P connection between callbacks

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `edge_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the callback graph edge to remove |


---


