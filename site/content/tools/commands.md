# Commands

Command schema and parameter introspection.

!!! info "Auto-generated"
    This page is **generated from source code** by the schema extractor.
    Do not edit by hand — changes will be overwritten.

---


## `mythic_get_commands`

Get a list of all commands available in Mythic

_No parameters._

---


## `mythic_get_command_parameters`

Get a list of all command parameters across all commands

_No parameters._

---


## `mythic_get_command_with_parameters`

Get details of a specific command including its parameters and helper methods

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `payload_type_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the payload type |
| `command_name` | `string` | :material-check-bold:{ title="Required" } **required** | Name of the command |


---


