# Credentials & Artifacts

Credential store and artifact (IOC / forensic evidence) tracking.

!!! info "Auto-generated"
    This page is **generated from source code** by the schema extractor.
    Do not edit by hand — changes will be overwritten.

---


## `mythic_get_credentials`

Get a list of all credentials stored in Mythic

_No parameters._

---


## `mythic_get_credential`

Get details of a specific credential by ID

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `credential_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the credential to retrieve |


---


## `mythic_get_operation_credentials`

Get credentials filtered by operation

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `operation_id` | `int` | :material-check-bold:{ title="Required" } **required** | Operation ID to filter credentials |


---


## `mythic_create_credential`

Create a new credential entry

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `type` | `string` | :material-check-bold:{ title="Required" } **required** | Credential type (plaintext/hash/key/ticket/etc.) |
| `account` | `string` | :material-check-bold:{ title="Required" } **required** | Account/username |
| `realm` | `*string` | _optional_ | Domain/realm |
| `credential` | `string` | :material-check-bold:{ title="Required" } **required** | The actual credential (password/hash/key) |
| `comment` | `*string` | _optional_ | Additional notes about the credential |
| `task_id` | `*int` | _optional_ | Task ID that discovered this credential |


---


## `mythic_update_credential`

Update an existing credential's properties

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `credential_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the credential to update |
| `type` | `*string` | _optional_ | New credential type |
| `account` | `*string` | _optional_ | New account/username |
| `realm` | `*string` | _optional_ | New domain/realm |
| `credential` | `*string` | _optional_ | New credential value |
| `comment` | `*string` | _optional_ | New comment |


---


## `mythic_delete_credential`

Delete a credential from Mythic

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `credential_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the credential to delete |


---


## `mythic_get_artifacts`

Get a list of all artifacts (IOCs, forensic evidence)

_No parameters._

---


## `mythic_get_artifact`

Get details of a specific artifact by ID

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `artifact_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the artifact to retrieve |


---


## `mythic_get_operation_artifacts`

Get artifacts filtered by operation

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `operation_id` | `int` | :material-check-bold:{ title="Required" } **required** | Operation ID to filter artifacts |


---


## `mythic_get_host_artifacts`

Get artifacts filtered by host

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `host` | `string` | :material-check-bold:{ title="Required" } **required** | Hostname to filter artifacts |


---


## `mythic_get_artifacts_by_type`

Get artifacts filtered by artifact type

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `artifact_type` | `string` | :material-check-bold:{ title="Required" } **required** | Artifact type to filter (File Write/Registry Write/etc.) |


---


## `mythic_create_artifact`

Create a new artifact entry (IOC, forensic evidence)

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `artifact` | `string` | :material-check-bold:{ title="Required" } **required** | The artifact (file path/registry key/etc.) |
| `base_artifact` | `*string` | _optional_ | Base artifact for pattern matching |
| `host` | `*string` | _optional_ | Hostname where artifact was observed |
| `task_id` | `*int` | _optional_ | Task ID that created this artifact |


---


## `mythic_update_artifact`

Update an existing artifact's properties

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `artifact_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the artifact to update |
| `host` | `*string` | _optional_ | New hostname |


---


## `mythic_delete_artifact`

Delete an artifact from Mythic

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `artifact_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the artifact to delete |


---


