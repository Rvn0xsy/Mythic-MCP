# Authentication

Login, logout, API token lifecycle, session management.

!!! info "Auto-generated"
    This page is **generated from source code** by the schema extractor.
    Do not edit by hand — changes will be overwritten.

---


## `mythic_login`

Authenticate with Mythic server using username and password

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `username` | `string` | :material-check-bold:{ title="Required" } **required** | Mythic username |
| `password` | `string` | :material-check-bold:{ title="Required" } **required** | Mythic password |


---


## `mythic_logout`

End the current Mythic session and clear authentication

_No parameters._

---


## `mythic_is_authenticated`

Check if currently authenticated with Mythic server

_No parameters._

---


## `mythic_get_current_user`

Get information about the current authenticated user

_No parameters._

---


## `mythic_create_api_token`

Create a new API token for programmatic access

_No parameters._

---


## `mythic_delete_api_token`

Delete an existing API token

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `token_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the token to delete |


---


## `mythic_refresh_token`

Refresh the current access token to extend session

_No parameters._

---


