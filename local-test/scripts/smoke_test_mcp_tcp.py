#!/usr/bin/env python3
"""Minimal MCP-over-TCP smoke test.

This speaks MCP JSON-RPC over a plain TCP stream (newline-delimited JSON objects).
It performs:
  1) initialize
  2) notifications/initialized
  3) tools/list

It’s intentionally tiny and dependency-free.
"""

import argparse
import json
import socket
import sys
import time


def send_json(sock: socket.socket, obj: dict) -> None:
    data = (json.dumps(obj) + "\n").encode("utf-8")
    sock.sendall(data)


class JsonStream:
    def __init__(self):
        self.buf = ""
        self.decoder = json.JSONDecoder()

    def feed(self, chunk: bytes):
        self.buf += chunk.decode("utf-8", errors="replace")

    def next_obj(self):
        # Skip leading whitespace/newlines
        s = self.buf.lstrip()
        if not s:
            self.buf = ""
            return None
        # Track how many chars were stripped
        stripped = len(self.buf) - len(s)
        try:
            obj, idx = self.decoder.raw_decode(s)
        except json.JSONDecodeError:
            return None
        # Advance buffer
        self.buf = s[idx:]
        # Re-attach any trailing data already in s[idx:] is kept;
        # stripped chars were removed; fine.
        return obj


def recv_until_response(sock: socket.socket, request_id: int, timeout_s: float):
    sock.settimeout(0.5)
    stream = JsonStream()
    deadline = time.time() + timeout_s

    while time.time() < deadline:
        try:
            chunk = sock.recv(4096)
            if not chunk:
                raise RuntimeError("connection closed")
            stream.feed(chunk)
        except socket.timeout:
            pass

        while True:
            obj = stream.next_obj()
            if obj is None:
                break

            # Ignore notifications/requests from server
            if isinstance(obj, dict) and obj.get("id") == request_id and "result" in obj:
                return obj
            if isinstance(obj, dict) and obj.get("id") == request_id and "error" in obj:
                raise RuntimeError(f"server error: {obj['error']}")

    raise TimeoutError("timed out waiting for response")


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--host", default="127.0.0.1")
    ap.add_argument("--port", type=int, default=3333)
    args = ap.parse_args()

    with socket.create_connection((args.host, args.port), timeout=10) as sock:
        # initialize
        init_id = 1
        send_json(
            sock,
            {
                "jsonrpc": "2.0",
                "id": init_id,
                "method": "initialize",
                "params": {
                    "protocolVersion": "2024-11-05",
                    "capabilities": {},
                    "clientInfo": {"name": "SmokeTest", "version": "1.0.0"},
                },
            },
        )
        init_resp = recv_until_response(sock, init_id, timeout_s=10)
        print("initialize: ok")

        # initialized notification
        send_json(sock, {"jsonrpc": "2.0", "method": "notifications/initialized"})

        # tools/list
        tools_id = 2
        send_json(sock, {"jsonrpc": "2.0", "id": tools_id, "method": "tools/list", "params": {}})
        tools_resp = recv_until_response(sock, tools_id, timeout_s=10)

        tools = tools_resp.get("result", {}).get("tools", [])
        print(f"tools/list: {len(tools)} tools")
        # Print a few tool names for sanity
        for t in tools[:10]:
            name = t.get("name") if isinstance(t, dict) else None
            if name:
                print(f"- {name}")

    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as exc:
        print(f"ERROR: {exc}", file=sys.stderr)
        raise SystemExit(1)
