# Operators

User account management, preferences, secrets, invite links.

!!! info "Auto-generated"
    This page is **generated from source code** by the schema extractor.
    Do not edit by hand — changes will be overwritten.

---


## `mythic_get_operators`

Get a list of all operators (users) in the Mythic instance

_No parameters._

---


## `mythic_get_operator`

Get details of a specific operator by ID

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `operator_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the operator to retrieve |


---


## `mythic_create_operator`

Create a new operator (user) account

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `username` | `string` | :material-check-bold:{ title="Required" } **required** | Username for the new operator |
| `password` | `string` | :material-check-bold:{ title="Required" } **required** | Password (minimum 12 characters) |
| `email` | `*string` | _optional_ | Email address |
| `bot` | `*bool` | _optional_ | Create as bot account |


---


## `mythic_update_operator_status`

Update operator status (active/inactive, admin privileges, deleted)

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `operator_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the operator to update |
| `active` | `*bool` | _optional_ | Set operator active/inactive |
| `admin` | `*bool` | _optional_ | Grant/revoke admin privileges |
| `deleted` | `*bool` | _optional_ | Mark operator as deleted |


---


## `mythic_update_password_email`

Update operator password and/or email address

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `operator_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the operator |
| `old_password` | `string` | :material-check-bold:{ title="Required" } **required** | Current password |
| `new_password` | `*string` | _optional_ | New password (min 12 chars) |
| `email` | `*string` | _optional_ | New email address |


---


## `mythic_get_operator_preferences`

Get UI preferences for an operator

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `operator_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the operator |


---


## `mythic_update_operator_preferences`

Update UI preferences for an operator

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `operator_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the operator |
| `preferences` | `map[string]any` | :material-check-bold:{ title="Required" } **required** | Preferences to update (key-value pairs) |


---


## `mythic_get_operator_secrets`

Get secrets/keys associated with an operator

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `operator_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the operator |


---


## `mythic_update_operator_secrets`

Update secrets/keys for an operator

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `operator_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the operator |
| `secrets` | `map[string]any` | :material-check-bold:{ title="Required" } **required** | Secrets to update (key-value pairs) |


---


## `mythic_get_invite_links`

Get all invitation links for new operators

_No parameters._

---


## `mythic_create_invite_link`

Create a new invitation link for operator registration

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `operation_id` | `*int` | _optional_ | Operation to associate link with |
| `operation_role` | `*string` | _optional_ | Role for new users (operator/spectator) |
| `max_uses` | `*int` | _optional_ | Maximum number of uses |
| `name` | `*string` | _optional_ | Human-readable name for the link |
| `short_code` | `*string` | _optional_ | Custom short code |


---


## `mythic_update_operator_operation`

Add or remove operators from an operation

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `operation_id` | `int` | :material-check-bold:{ title="Required" } **required** | Operation to modify |
| `add_users` | `*[]int` | _optional_ | Operator IDs to add with full access |
| `remove_users` | `*[]int` | _optional_ | Operator IDs to remove |
| `view_mode_operators` | `*[]int` | _optional_ | Operator IDs to set as view-only |
| `view_mode_spectators` | `*[]int` | _optional_ | Operator IDs to set as spectators |


---


