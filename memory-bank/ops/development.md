---
title: Development Environment
doc_kind: ops
doc_function: canonical
purpose: "Local development setup, commands, dependencies, and verification workflow for start-issue."
derived_from:
  - ../dna/governance.md
  - ../engineering/testing-policy.md
  - ../../README.md
  - ../../doc/spec.md
status: active
audience: humans_and_agents
---

# Development Environment

## Setup

Required tools for normal development:

- Go 1.24+
- `git`
- `gh` with an authenticated GitHub session for issue/update flows

The Go CLI parses release metadata, downloads update assets, and verifies
checksums itself. `jq`, download tools, and checksum tools are not required for
the CLI update workflow. The optional Bash installer still requires `bash`,
`curl` or `wget`, and a SHA-256 tool.

Build the binary:

```bash
make build
```

Install locally from source:

```bash
make install
```

Run the opt-in real-Codex batch E2E smoke test with a usable issue in the
current repository:

```bash
START_ISSUE_E2E=1 make e2e-batch
```

Run the deterministic, network-free built-binary E2E used in CI:

```bash
make e2e-sandbox
```

It uses fake `gh` and agent commands but real local git/worktree operations.

When a live E2E must be visible in cmux, follow the canonical cmux-tab
procedure in [`../../AGENTS.md`](../../AGENTS.md): find the `start-issue`
workspace using `cmux tree --all`, create a terminal surface in its active pane,
and run the command with `START_ISSUE_E2E_KEEP=1`.

## Daily Commands

```bash
make test
make build
make install
```

Useful direct commands:

```bash
gofmt -w cmd/start-issue/*.go
go vet ./...
go test ./...
python3 scripts/check_memory_bank_index.py --max-depth 4
```

## Dry-Run Development

Use `--dry-run` to inspect config, prompt source, worktree path, and launch
command without creating a worktree or launching an agent:

```bash
.build/start-issue 123 --repo dapi/start-issue --agent codex --dry-run
```

Set `START_ISSUE_DUMP_PROMPT=1` when the full rendered prompt must be visible in
dry-run output.

## Browser Testing

Not applicable. This project has no browser UI. CLI output is verified through
Go tests and direct command output review.

## Local Services

No database or long-running local services are required. External command
dependencies (`gh`, selected agent CLIs, `zellij-tab-status`) are optional or
mode-specific as documented in README/spec.

## Development Safety

- Prefer `.build/start-issue --dry-run` when exploring behavior.
- Do not run release commands unless preparing an actual release.
- Do not delete worktrees manually while tests depend on fake worktree state;
  use isolated temp directories in tests.
