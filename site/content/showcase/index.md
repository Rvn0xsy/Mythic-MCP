# Showcase

See the Mythic MCP Server in action.

## Lab Walkthrough

A guided walkthrough of a full red team workflow driven entirely through
AI + MCP:

[:octicons-arrow-right-24: Lab Walkthrough](lab-walkthrough.md)

## What Can You Do With This?

Here are some real-world examples of what an AI assistant can do once
connected to Mythic via MCP:

### :dart: Automated Payload Deployment

> *"Build a Xenon payload using httpx with AESPSK encryption, wait for
> it to build, and give me the download link."*

The AI calls `mythic_get_payload_types` → `mythic_get_c2_profile_parameters`
→ `mythic_get_documentation` (reads C2 config requirements) →
`mythic_create_payload` → `mythic_wait_for_payload` →
`mythic_download_payload` — all automatically.

### :mag: Interactive Callback Triage

> *"List all active callbacks, then for each Windows host run `whoami`
> and `ipconfig` and summarise the results."*

The AI calls `mythic_get_active_callbacks` → iterates → `mythic_issue_task`
per callback → `mythic_wait_for_task` → `mythic_get_task_output` → presents
a summary table.

### :shield: MITRE ATT&CK Coverage Analysis

> *"What ATT&CK techniques have we used so far in the current operation?"*

`mythic_get_current_operation` → `mythic_get_attacks_by_operation` →
cross-references with the full technique list → displays coverage gaps.

### :file_folder: Credential and Artifact Tracking

> *"Show me all credentials we've collected, grouped by host and realm."*

`mythic_get_credentials` → formats and groups automatically.
