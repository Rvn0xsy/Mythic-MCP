# Screenshots

Screenshot capture timeline, thumbnails, downloads.

!!! info "Auto-generated"
    This page is **generated from source code** by the schema extractor.
    Do not edit by hand — changes will be overwritten.

---


## `mythic_get_screenshots`

Get screenshots captured by a specific callback

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `callback_id` | `int` | :material-check-bold:{ title="Required" } **required** | Display ID of the callback |
| `limit` | `int` | :material-check-bold:{ title="Required" } **required** | Maximum number of screenshots to retrieve |


---


## `mythic_get_screenshot_by_id`

Get detailed information about a specific screenshot by its ID

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `screenshot_id` | `int` | :material-check-bold:{ title="Required" } **required** | ID of the screenshot |


---


## `mythic_get_screenshot_timeline`

Get screenshots from a callback within a specific time range

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `callback_id` | `int` | :material-check-bold:{ title="Required" } **required** | Display ID of the callback |
| `start_time` | `string` | :material-check-bold:{ title="Required" } **required** | Start time in RFC3339 format |
| `end_time` | `string` | :material-check-bold:{ title="Required" } **required** | End time in RFC3339 format |


---


## `mythic_get_screenshot_thumbnail`

Download thumbnail image of a screenshot

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `agent_file_id` | `string` | :material-check-bold:{ title="Required" } **required** | Agent file ID of the screenshot |


---


## `mythic_download_screenshot`

Download the full resolution screenshot image

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `agent_file_id` | `string` | :material-check-bold:{ title="Required" } **required** | Agent file ID of the screenshot |


---


## `mythic_delete_screenshot`

Delete a screenshot from Mythic (destructive operation)

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `agent_file_id` | `string` | :material-check-bold:{ title="Required" } **required** | Agent file ID of the screenshot to delete |


---


