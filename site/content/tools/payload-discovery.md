# Payload Discovery

Introspect build parameters, C2 profile parameters, and available commands for each payload type.

!!! info "Auto-generated"
    This page is **generated from source code** by the schema extractor.
    Do not edit by hand — changes will be overwritten.

---


## `mythic_get_payload_type_build_parameters`

Get the build parameter schema for a payload type. Returns all configurable build parameters including name, type, required, default value, description, and available choices. Use this before creating a payload to discover what build_parameters are needed.

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `payload_type_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the payload type (use mythic_get_payload_types to find IDs) |


---


## `mythic_get_c2_profile_parameters`

Get the configuration parameter schema for a C2 profile. Returns all configurable parameters including name, type, required, default value, and description. IMPORTANT: Call this before creating a payload to learn what parameters to pass in the c2_profiles[].parameters field. Common parameters include callback_host, callback_port, etc. Also ensure the C2 profile is STARTED before deploying the payload.

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `c2_profile_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the C2 profile (use mythic_get_c2_profiles to find IDs) |


---


## `mythic_get_payload_type_commands`

Get all available commands for a specific payload type. Returns command names, descriptions, and help text. Use this to discover what commands can be included when creating a payload.

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `payload_type_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the payload type (use mythic_get_payload_types to find IDs) |


---


