# Documentation

Browse installed agent and C2 profile documentation.

!!! info "Auto-generated"
    This page is **generated from source code** by the schema extractor.
    Do not edit by hand — changes will be overwritten.

---


## `mythic_list_documentation`

List available documentation for installed Mythic agents, C2 profiles, and wrappers. Returns a tree of documentation pages. Use this to discover what documentation is available before retrieving specific pages with mythic_get_documentation. Each entry has a path field you can pass to mythic_get_documentation.

_No parameters._

---


## `mythic_get_documentation`

Retrieve the full markdown content of a specific documentation page. Use mythic_list_documentation first to discover available pages. Pass the path value from the listing (e.g. Agents/poseidon, C2 Profiles/httpx/examples, Agents/poseidon/commands/shell). IMPORTANT: Always read C2 profile documentation before creating payloads - profiles may require specific configuration files or parameters not obvious from the API alone.

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `path` | `string` | :material-check-bold:{ title="Required" } **required** | Documentation path from the listing (e.g. Agents/poseidon or C2 Profiles/httpx/examples or Agents/poseidon/commands/shell) |


---


