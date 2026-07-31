---
title: Configuration Guide
doc_kind: ops
doc_function: canonical
purpose: "Configuration sources, precedence, files, environment variables, and ownership for start-issue."
derived_from:
  - ../dna/governance.md
  - ../domain/rules.md
  - ../../README.md
  - ../../doc/spec.md
status: active
audience: humans_and_agents
---

# Configuration Guide

## Configuration Architecture

`start-issue` uses simple CLI flags, plaintext config files, and environment
variables. The canonical precedence is:

1. CLI flags
2. project config under `.start-issue/`
3. user config under `~/.config/start-issue/`
4. environment variables
5. built-in defaults

Config resolution is owned by the Go configuration helpers in
`cmd/start-issue`. User-facing
documentation must stay aligned in README and `doc/spec.md`.

## File Layout

Project config:

```text
.start-issue/
├── agent
├── model
└── prompt.md
```

User config:

```text
~/.config/start-issue/
├── agent
├── model
└── prompt.md
```

`agent` and `model` are plain strings. `prompt.md` is a prompt template.

## Environment Variables

| Variable | Description | Default | Owner |
| --- | --- | --- | --- |
| `START_ISSUE_AGENT` | Default agent when CLI and config files do not set one | built-in `claude` | Go configuration helpers |
| `START_ISSUE_MODEL` | Default model when CLI and config files do not set one | unset | Go configuration helpers |
| `START_ISSUE_PROMPT` | Inline prompt template | none | Go configuration helpers |
| `START_ISSUE_PROMPT_FILE` | Prompt template file | none | Go configuration helpers |
| `START_ISSUE_WORKTREE_DIR` | Parent directory for created worktrees | `~/worktrees` | Go configuration/worktree helpers |
| `START_ISSUE_DUMP_PROMPT` | Print full rendered prompt in dry-run when set to `1` | unset | Go output helpers |

`START_ISSUE_PROMPT` and `START_ISSUE_PROMPT_FILE` are mutually exclusive when
no CLI prompt override is provided.

## Setup And Init

- `start-issue setup` and `start-issue --setup` write only user config.
- First-run onboarding creates `~/.config/start-issue` as the marker that setup
  was offered.
- `start-issue init --project` writes project config and requires a git repo.
- `start-issue init --user` writes user config and can run outside a git repo.
- `--force` overwrites existing `agent` and `prompt.md`; omitted `--model` under
  `--force` removes an existing `model` file to return to built-in unset.

## Secrets

No secrets are stored by `start-issue` config files. GitHub authentication and
agent authentication remain owned by their external CLIs.
