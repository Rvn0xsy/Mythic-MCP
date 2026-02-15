# Processes

Process enumeration and tree views.

!!! info "Auto-generated"
    This page is **generated from source code** by the schema extractor.
    Do not edit by hand — changes will be overwritten.

---


## `mythic_get_processes`

Get a list of all processes enumerated in Mythic

_No parameters._

---


## `mythic_get_processes_by_operation`

Get all processes enumerated in a specific operation

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `operation_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the operation |


---


## `mythic_get_processes_by_callback`

Get all processes enumerated by a specific callback

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `callback_id` | `int` | :material-check-bold:{ title="Required" } **required** | Display ID of the callback |


---


## `mythic_get_process_tree`

Get process tree structure for a callback showing parent-child relationships

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `callback_id` | `int` | :material-check-bold:{ title="Required" } **required** | Display ID of the callback |


---


## `mythic_get_processes_by_host`

Get all processes enumerated on a specific host

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `host_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the host |


---


