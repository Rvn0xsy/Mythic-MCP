# Tool Reference

The Mythic MCP Server exposes **147 tools** across **18 categories**.
Each tool has a stable name, a description the AI reads to decide when
to use it, and a typed parameter schema.

!!! info "Auto-generated"
    These pages are generated directly from the Go source code.
    They stay in sync with the server automatically on every deploy.

---

| Category | Tools | Description |
|----------|:-----:|-------------|
| [Authentication](authentication.md) | 7 | Login, logout, API token lifecycle, session management. |
| [C2 Profiles](c2-profiles.md) | 10 | C2 profile lifecycle — start/stop listeners, IOCs, sample messages, configuration. |
| [Callbacks](callbacks.md) | 11 | Active agent sessions — list, update, P2P edges, tokens. |
| [Commands](commands.md) | 3 | Command schema and parameter introspection. |
| [Credentials & Artifacts](credentials.md) | 14 | Credential store and artifact (IOC / forensic evidence) tracking. |
| [Documentation](documentation.md) | 2 | Browse installed agent and C2 profile documentation. |
| [Files](files.md) | 8 | Upload, download, preview, and bulk-export files. |
| [Hosts](hosts.md) | 5 | Host inventory, network topology mapping. |
| [Keylogs](keylogs.md) | 3 | Keylogger data retrieval by operation or callback. |
| [MITRE ATT&CK](mitre-attack.md) | 6 | Technique lookup, task/command/operation mapping. |
| [Operations](operations.md) | 11 | Operation (campaign) CRUD, event logging, global settings. |
| [Operators](operators.md) | 12 | User account management, preferences, secrets, invite links. |
| [Payload Discovery](payload-discovery.md) | 3 | Introspect build parameters, C2 profile parameters, and available commands for each payload type. |
| [Payloads](payloads.md) | 12 | Build, download, manage, and inspect agent payload binaries. |
| [Processes](processes.md) | 5 | Process enumeration and tree views. |
| [Screenshots](screenshots.md) | 6 | Screenshot capture timeline, thumbnails, downloads. |
| [Tags](tags.md) | 11 | Flexible tagging system for any Mythic object. |
| [Tasks & Responses](tasks.md) | 18 | Issue commands to agents, read output, wait for completion, OPSEC bypass. |

