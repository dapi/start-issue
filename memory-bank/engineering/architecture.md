---
title: Engineering Architecture Patterns
doc_kind: engineering
doc_function: canonical
purpose: "Architecture rules for the start-issue Bash CLI, module boundaries, adapters, failures, and configuration ownership."
derived_from:
  - ../dna/governance.md
  - ../domain/context-map.md
  - ../../README.md
  - ../../doc/spec.md
status: active
audience: humans_and_agents
---

# Engineering Architecture Patterns

`start-issue` is implemented as modular Bash source under
`scripts/lib/start_issue/` and distributed as a bundled single-file executable
through `scripts/build-start-issue`.

## Module Boundaries

| Module / Layer | Owns | Must not depend on directly |
| --- | --- | --- |
| `scripts/start-issue` | Entrypoint/bootstrap and version constant | Feature-specific behavior that belongs in modules |
| `cli.sh` | Argument parsing, mode flags, normalized workflow state | Agent command syntax, GitHub API calls |
| `config.sh` | Agent/model/prompt/worktree config resolution and prompt rendering | Worktree side effects, release update |
| `github.sh` | Issue parsing, repo/base detection, issue metadata fetch | Agent launch, config file writes |
| `worktree.sh` | Branch naming, worktree planning, safe create/reuse/delete, optional zellij/init side effects | Agent internals |
| `agent.sh` | Adapter validation, launch commands, AI branch naming, prompt improvement, Codex human-gate helpers | CLI precedence, GitHub issue fetching |
| `release.sh` | Download/checksum/version helpers shared by installer/update | Normal issue workflow |
| `update.sh` | Self-update mode orchestration | Git worktree or issue state |
| `init.sh` | `init`, `setup`, and first-run onboarding helpers | GitHub issue fetch, agent launch |
| `output.sh` | Help, status, dry-run, human-gate help, user-facing messages | Core side effects |
| `pipeline.sh` | Top-level orchestration order | Adapter-specific command construction |

## Adapter Boundary

Agent-specific behavior must stay centralized in `agent.sh`:

- supported agent validation;
- launch command construction;
- model argument handling;
- AI branch-name helper commands;
- prompt-improvement helper commands;
- Codex-specific human-gate batch/resume command construction.

Do not add ad hoc `case "$AGENT"` branches in unrelated modules unless the
change is only routing to the adapter boundary.

## Configuration Ownership

Configuration precedence is part of the public contract:

1. CLI
2. project config
3. user config
4. environment
5. built-in default

The owner module is `config.sh`; user-facing descriptions must stay aligned in
[ops/config.md](../ops/config.md), [README.md](../../README.md), and
[doc/spec.md](../../doc/spec.md).

## Failure Handling

- Fail fast for invalid user intent: unknown agents, prompt source conflicts,
  unsupported human-gate agent, missing required update dependencies.
- Treat optional integrations as warnings when documented as optional:
  `zellij-tab-status` and non-zero `init.sh` do not abort the normal workflow.
- Preserve diagnostic artifacts for Codex human-gate failures under the run state
  directory.
- Never continue after a worktree safety validation failure.

## Bash Boundary

Bash remains the core while:

- CLI modes stay small and explicit;
- configuration remains file/env/string based;
- local tests can cover behavior deterministically;
- output is primarily human-readable.

Reevaluate a Python core if future work requires nested configuration, richer
lifecycle commands such as `resume/list/cleanup`, structured machine-readable
output, or complex state persistence.
