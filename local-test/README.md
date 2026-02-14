# Local test harness (safe)

This folder is a **local development harness** for running the Mythic MCP server in a container and exposing its stdio-based MCP transport over a TCP port for simple host-side smoke testing.

## Important note

I can’t help you with instructions or automation to deploy Mythic as a C2 framework or to pre-install/launch agents (e.g., Poseidon). That kind of setup can be used to enable unauthorized access.

What this harness *does* provide:

- A container image for the MCP server
- A TCP bridge (`socat`) so you can reach the MCP server from outside the container/VM
- A tiny smoke test client to validate the MCP handshake + `tools/list`

## Prereqs

- Docker Desktop (or Docker Engine) with `docker compose`

## Configure

Copy the example env file and fill it out:

- `copy .env.example .env` (PowerShell) or `cp .env.example .env` (bash)

At minimum you must set:

- `MYTHIC_URL`
- either `MYTHIC_API_TOKEN` **or** `MYTHIC_USERNAME` + `MYTHIC_PASSWORD`

## Run

From this directory:

- `docker compose up --build -d`

The MCP server will be bridged on:

- `127.0.0.1:3333` (from the VM)

If you need host->VM access, ensure your Hyper-V networking/port-forwarding allows reaching TCP 3333 on the VM.

## Smoke test

On the VM (or from the Hyper-V host if it can reach the VM’s 3333/TCP):

- `python .\scripts\smoke_test_mcp_tcp.py --host 127.0.0.1 --port 3333`

## Stop

- `docker compose down`
