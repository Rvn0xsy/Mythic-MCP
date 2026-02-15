# Keylogs

Keylogger data retrieval by operation or callback.

!!! info "Auto-generated"
    This page is **generated from source code** by the schema extractor.
    Do not edit by hand — changes will be overwritten.

---


## `mythic_get_keylogs`

Get all keylogger data captured across all operations

_No parameters._

---


## `mythic_get_keylogs_by_operation`

Get all keylogger data captured in a specific operation

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `operation_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the operation |


---


## `mythic_get_keylogs_by_callback`

Get all keylogger data captured by a specific callback

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `callback_id` | `int` | :material-check-bold:{ title="Required" } **required** | Display ID of the callback |


---


