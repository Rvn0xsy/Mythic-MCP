# Tags

Flexible tagging system for any Mythic object.

!!! info "Auto-generated"
    This page is **generated from source code** by the schema extractor.
    Do not edit by hand — changes will be overwritten.

---


## `mythic_get_tag_types`

Get a list of all tag types (categories for tags)

_No parameters._

---


## `mythic_get_tag_types_by_operation`

Get tag types filtered by operation

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `operation_id` | `int` | :material-check-bold:{ title="Required" } **required** | Operation ID to filter tag types |


---


## `mythic_get_tag_type`

Get details of a specific tag type by ID

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tag_type_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the tag type to retrieve |


---


## `mythic_create_tag_type`

Create a new tag type (category)

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | `string` | :material-check-bold:{ title="Required" } **required** | Name of the tag type |
| `description` | `*string` | _optional_ | Description of the tag type |
| `color` | `*string` | _optional_ | Hex color code (e.g. #FF5733) |


---


## `mythic_update_tag_type`

Update an existing tag type's properties

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tag_type_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the tag type to update |
| `name` | `*string` | _optional_ | New name for the tag type |
| `description` | `*string` | _optional_ | New description |
| `color` | `*string` | _optional_ | New hex color code |
| `deleted` | `*bool` | _optional_ | Mark as deleted |


---


## `mythic_delete_tag_type`

Delete a tag type

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tag_type_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the tag type to delete |


---


## `mythic_create_tag`

Create a tag and apply it to an object (task, callback, file, etc.)

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tag_type_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the tag type to use |
| `source_type` | `string` | :material-check-bold:{ title="Required" } **required** | Type of object to tag (task/callback/filemeta/payload/artifact/process/keylog) |
| `source_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the object to tag |


---


## `mythic_get_tag`

Get details of a specific tag by ID

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tag_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the tag to retrieve |


---


## `mythic_get_tags`

Get all tags for a specific object

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `source_type` | `string` | :material-check-bold:{ title="Required" } **required** | Type of object (task/callback/filemeta/etc.) |
| `source_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the object |


---


## `mythic_get_tags_by_operation`

Get all tags in an operation

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `operation_id` | `int` | :material-check-bold:{ title="Required" } **required** | Operation ID to get tags for |


---


## `mythic_delete_tag`

Delete a tag from an object

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tag_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the tag to delete |


---


