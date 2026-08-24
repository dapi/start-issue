---
title: Domain Context Map
doc_kind: domain
doc_function: canonical
purpose: "Bounded contexts and ownership boundaries for start-issue."
derived_from:
  - ../dna/governance.md
  - glossary.md
  - model.md
status: active
audience: humans_and_agents
canonical_for:
  - bounded_contexts
  - domain_context_map
---

# Domain Context Map

## Bounded Contexts

| Context | Owns language / rules for | Upstream contexts | Downstream contexts | Must not know |
| --- | --- | --- | --- | --- |
| `CLI input` | Modes, flags, argument normalization | Product docs | Config, issue workflow, update/setup/init routing | Agent-specific command details |
| `Configuration` | Agent/model/prompt/worktree precedence and source reporting | CLI input | Output, pipeline, adapters | GitHub issue internals |
| `GitHub issue intake` | Repo detection, issue URL/number parsing, issue metadata | CLI input, configuration | Branch naming, prompt rendering | Worktree side effects |
| `Worktree lifecycle` | Branch names, path planning, safe reuse/delete/create | Issue intake, configuration | Init, prompt rendering, agent launch | Agent CLI internals |
| `Agent adapter` | Agent validation, launch commands, AI branch names, prompt improvement | Configuration, worktree lifecycle | External agent CLI | Config precedence implementation |
| `Release/update` | Latest release lookup, version compare, checksum, install path | CLI input | Installed executable | Issue/worktree workflow |
| `Memory-bank governance` | Documentation process, flows, templates, reachability audit | Template source and local docs | Feature packages, AGENTS.md | Runtime config state |

## Context Relationships

| Relationship ID | Upstream | Downstream | Contract | Notes |
| --- | --- | --- | --- | --- |
| `REL-01` | `CLI input` | `Configuration` | Parsed flags and mode state | CLI intent wins precedence |
| `REL-02` | `Configuration` | `Agent adapter` | Resolved agent/model/prompt | Adapter must not recompute precedence |
| `REL-03` | `GitHub issue intake` | `Worktree lifecycle` | Issue number/title/labels and repo/base branch | Branch naming depends on issue facts |
| `REL-04` | `Worktree lifecycle` | `Agent adapter` | Verified worktree path and branch | Agent launch starts only after worktree is ready |
| `REL-05` | `Release/update` | `Installed executable` | Verified asset and target path | No git repo required |
| `REL-06` | `Memory-bank governance` | `FeaturePackage` | Flow rules, templates, stable IDs | New packages use current flow |

## Shared Kernel / Published Language

- `SK-01` Shared IDs in feature docs: `REQ-*`, `SC-*`, `CHK-*`, `EVID-*`,
  `SOL-*`, `STEP-*` as defined by feature-flow.
- `SK-02` CLI values for agents: `claude`, `codex`, `kimi`, `pi`, `none`.
- `PL-01` Public config files: `.start-issue/agent`, `.start-issue/model`,
  `.start-issue/prompt.md`, and user equivalents under
  `~/.config/start-issue/`.

## Boundary Rules

- CLI parsing should not branch on agent-specific command syntax.
- Agent adapters should not fetch GitHub issues or decide worktree reuse.
- Release/update code should not require a git repository or issue input.
- Memory-bank documents can guide implementation but must not become runtime
  configuration.

## Open Boundary Questions

- `OQ-01` How future lifecycle commands should be decomposed into focused Go
  helpers without weakening the existing ownership boundaries.
- `OQ-02` Whether non-Codex agents will expose enough resumable batch semantics
  to generalize batch mode.
