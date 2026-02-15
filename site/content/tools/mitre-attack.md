# MITRE ATT&CK

Technique lookup, task/command/operation mapping.

!!! info "Auto-generated"
    This page is **generated from source code** by the schema extractor.
    Do not edit by hand — changes will be overwritten.

---


## `mythic_get_attack_techniques`

Get a list of all MITRE ATT&CK techniques available in Mythic

_No parameters._

---


## `mythic_get_attack_technique_by_id`

Get details of a specific MITRE ATT&CK technique by internal ID

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `attack_id` | `int` | :material-check-bold:{ title="Required" } **required** | Internal ID of the MITRE ATT&CK technique |


---


## `mythic_get_attack_technique_by_tnum`

Get details of a specific MITRE ATT&CK technique by T-number (e.g., T1055)

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `t_number` | `string` | :material-check-bold:{ title="Required" } **required** | MITRE ATT&CK technique T-number (e.g. |


---


## `mythic_get_attack_by_task`

Get MITRE ATT&CK techniques associated with a specific task

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `task_id` | `int` | :material-check-bold:{ title="Required" } **required** | Internal ID of the task |


---


## `mythic_get_attack_by_command`

Get MITRE ATT&CK techniques associated with a specific command

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `command_id` | `int` | :material-check-bold:{ title="Required" } **required** | Internal ID of the command |


---


## `mythic_get_attacks_by_operation`

Get all MITRE ATT&CK techniques used in an operation

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `operation_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the operation |


---


