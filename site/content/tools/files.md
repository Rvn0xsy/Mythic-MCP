# Files

Upload, download, preview, and bulk-export files.

!!! info "Auto-generated"
    This page is **generated from source code** by the schema extractor.
    Do not edit by hand — changes will be overwritten.

---


## `mythic_get_files`

Get a list of all files in Mythic (uploaded and downloaded)

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `limit` | `int` | _optional_ | Maximum number of files to return (default 100) |


---


## `mythic_get_file`

Get metadata and information about a specific file by its UUID

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `file_id` | `string` | :material-check-bold:{ title="Required" } **required** | UUID of the file to retrieve |


---


## `mythic_get_downloaded_files`

Get a list of files downloaded from agents

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `limit` | `int` | _optional_ | Maximum number of files to return (default 100) |


---


## `mythic_upload_file`

Upload a file to Mythic for later use (tasking, payload building, etc.)

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `filename` | `string` | :material-check-bold:{ title="Required" } **required** | Name of the file |
| `file_data` | `string` | :material-check-bold:{ title="Required" } **required** | Base64-encoded file content |


---


## `mythic_download_file`

Get a one-time download URL for a file stored in Mythic. Use curl or wget with the returned download_url to fetch the file content. The URL token is single-use and expires after 5 minutes.

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `file_uuid` | `string` | :material-check-bold:{ title="Required" } **required** | UUID of the file to download |


---


## `mythic_delete_file`

Delete a file from Mythic by its UUID

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `file_id` | `string` | :material-check-bold:{ title="Required" } **required** | UUID of the file to delete |


---


## `mythic_bulk_download_files`

Get a download URL for multiple files bundled as a single ZIP archive. Use curl or wget with the returned URL to fetch the ZIP.

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `file_uuids` | `[]string` | :material-check-bold:{ title="Required" } **required** | Array of file UUIDs to download |


---


## `mythic_preview_file`

Preview a file's content (for text files, limited size)

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `file_id` | `string` | :material-check-bold:{ title="Required" } **required** | UUID of the file to preview |


---


