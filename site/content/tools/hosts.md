# Hosts

Host inventory, network topology mapping.

!!! info "Auto-generated"
    This page is **generated from source code** by the schema extractor.
    Do not edit by hand — changes will be overwritten.

---


## `mythic_get_hosts`

Get all hosts tracked in a specific operation

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `operation_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the operation |


---


## `mythic_get_host_by_id`

Get detailed information about a specific host by its ID

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `host_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the host |


---


## `mythic_get_host_by_hostname`

Get detailed information about a specific host by its hostname

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `hostname` | `string` | :material-check-bold:{ title="Required" } **required** | Hostname to search for |


---


## `mythic_get_host_network_map`

Get network topology map showing host relationships in an operation

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `operation_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the operation |


---


## `mythic_get_callbacks_for_host`

Get all callbacks (agents) running on a specific host

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `host_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the host |


---


