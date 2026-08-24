---
title: Engineering Architecture Patterns
doc_kind: engineering
doc_function: canonical
purpose: "Architecture rules for the start-issue Go CLI, module boundaries, adapters, failures, and configuration ownership."
derived_from:
  - ../dna/governance.md
  - ../domain/context-map.md
  - ../../README.md
  - ../../doc/spec.md
status: active
audience: humans_and_agents
---

# Engineering Architecture Patterns

`start-issue` is a Go CLI rooted at `cmd/start-issue` and distributed as a
platform-specific compiled binary.

## Module Boundaries

| Module / Layer | Owns | Must not depend on directly |
| --- | --- | --- |
| `cmd/start-issue` | Entrypoint, CLI parsing, orchestration, and version injection | Native GitHub or agent protocol clients |
| configuration helpers | Agent/model/prompt/worktree config resolution and prompt rendering | Worktree side effects, release update |
| repository helpers | Issue parsing, repo/base detection, issue metadata fetch, and worktree lifecycle | Agent launch, config file writes |
| agent adapter helpers | Validation and launch command construction | CLI precedence and issue fetching |
| release helpers | Asset selection, checksum verification, version comparison, and self-update | Normal issue workflow |

## Adapter Boundary

Agent-specific behavior must stay centralized in Go adapter helpers:

- supported agent validation;
- launch command construction;
- model argument handling;
- AI branch-name helper commands;
- prompt-improvement helper commands;
- Codex-specific batch/resume command construction.

Do not add ad hoc `case "$AGENT"` branches in unrelated modules unless the
change is only routing to the adapter boundary.

## Configuration Ownership

Configuration precedence is part of the public contract:

1. CLI
2. project config
3. user config
4. environment
5. built-in default

The owner is the Go configuration layer; user-facing descriptions must stay aligned in
[ops/config.md](../ops/config.md), [README.md](../../README.md), and
[doc/spec.md](../../doc/spec.md).

## Failure Handling

- Fail fast for invalid user intent: unknown agents, prompt source conflicts,
  unsupported batch agent, missing required update dependencies.
- Treat optional integrations as warnings when documented as optional:
  `zellij-tab-status` and non-zero `init.sh` do not abort the normal workflow.
- Preserve diagnostic artifacts for Codex batch failures under the run state
  directory.
- Never continue after a worktree safety validation failure.

## Process Boundary

`git`, `gh`, and supported agent CLIs remain external processes. Go owns the
application lifecycle, filesystem operations, release verification, and tests.
