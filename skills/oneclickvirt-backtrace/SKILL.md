---
name: oneclickvirt-backtrace
description: "Query RDAP/BGP metadata for an explicit IP; distinguish lookup data from an actual traceroute."
metadata:
  repository: oneclickvirt/backtrace
  integration-version: "1.0.0"
---

# backtrace

## Scope

Query RDAP/BGP metadata for an explicit IP; distinguish lookup data from an actual traceroute.

- Read probe_help from the installed binary before assuming flags are supported. MCP returns native JSON when the program provides JSON, otherwise native text.
- Keep partial/network errors distinct from confirmed test outcomes. Current release assets and their actual architecture are available through release_assets.
- The ip-info preset requires a literal target IP. Do not pass a host, proxy, shell expression or a network range.

## Install and update

```bash
curl -fL --retry 3 https://raw.githubusercontent.com/oneclickvirt/backtrace/main/scripts/install_agent.py -o install_agent.py
python3 install_agent.py --client codex
```

The installer is also mirrored at `https://cnb.cool/oneclickvirt/backtrace/-/git/raw/main/scripts/install_agent.py`; add `--source cnb` to install from an immutable CNB snapshot (requires Git).

Use `--client claude` or `--client openclaw` for another client. Download and run again to update; `--ref` pins a source tag/commit. The installer verifies all SHA-256 checksums before replacing its files, preserves extra local files/client settings, and records the source commit in `install.json`.

## MCP

Requires Python 3.9+. The adapter uses existing local commands and their existing permissions; it does not install a virtualization runtime. Merge the installed `mcp.example.json` into the client configuration. For a remote host, run its adapter over noninteractive SSH.

```bash
python3 install_agent.py --client codex --install-binary
```

`--version` selects the probe release tag; `--ref` selects Skill/MCP source independently. The default resolves the latest stable GitHub release once, selects a real asset for this OS/architecture, and verifies SHA-256. Missing assets fail without guessing. Reuse an existing executable with `--binary /absolute/path/to/program`.

Call `integration_info` first for capabilities and prerequisites.

Available probe presets:

- `ip-info`: `backtrace -json -timeout 12s -ip <target>`

`release_assets` returns the actual version, filenames, sizes, digests and download URLs. Pass an exact tag for a pinned release. After a product update, verify help and an actual call rather than treating file presence as a successful test.

## Source and documentation

- [`backtrace_install.sh`](https://github.com/oneclickvirt/backtrace/blob/main/backtrace_install.sh)
- [`cmd/main.go`](https://github.com/oneclickvirt/backtrace/blob/main/cmd/main.go)
- [`go.mod`](https://github.com/oneclickvirt/backtrace/blob/main/go.mod)

- [集成安装说明 / Integration guide](https://www.spiritlhl.net/guide/agent_integrations.html)

Check the relevant source and current documentation before giving commands. Preserve requested version pins, data paths and native host identifiers. Distinguish actual program execution, isolated tests and unverified environments.
