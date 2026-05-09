---
title: "FT-009: Modularize start-issue"
doc_kind: feature
doc_function: canonical
purpose: "Canonical feature document for modularizing start-issue. Owns only the problem space and verification contract."
derived_from:
  - https://github.com/dapi/start-issue/issues/9
  - ../../../README.md
  - ../../../doc/spec.md
status: active
delivery_status: in_progress
audience: humans_and_agents
must_not_define:
  - selected_design
  - implementation_sequence
---

# FT-009: Modularize start-issue

## What

### Problem

`scripts/start-issue` currently carries CLI parsing, configuration resolution, GitHub access, branch/worktree planning, output formatting, and agent-specific behavior inside one large Bash script. That shape makes feature work harder to isolate, increases regression risk, and raises the cost of adding future commands, richer configuration, or structured output.

### Outcome

`start-issue` keeps the same CLI contract while moving to explicit shell modules with a clearer orchestration pipeline and a normalized agent-adapter contract.

### Scope

- `REQ-01` Split the monolithic implementation into focused shell modules behind the existing CLI entrypoint.
- `REQ-02` Express the workflow as an explicit internal pipeline: parse input, resolve config, fetch issue, plan branch/worktree, execute plan, launch agent.
- `REQ-03` Centralize agent-specific behavior behind a uniform internal contract for validation, launch-command construction, branch-name generation, and prompt improvement.
- `REQ-04` Shape internal output/state handling so future structured output or adjacent commands can be added without another monolithic rewrite.
- `REQ-05` Document the resulting architecture, extension seams, and threshold for a later Bash-to-Python migration.
- `REQ-06` Preserve the existing user-facing CLI contract and functional behavior unless a documented defect is intentionally corrected.

### Non-Scope

- `NS-01` Do not rewrite the tool in Python as part of this issue.
- `NS-02` Do not add new user-facing commands such as `resume`, `list`, `cleanup`, or `--json` in this issue.
- `NS-03` Do not intentionally change prompt precedence, branch naming semantics, worktree creation semantics, or agent launch semantics beyond what modularization requires to preserve current behavior.

### Constraints

- `CON-01` Modularization must remain compatible with the current CLI entrypoint `scripts/start-issue`.
- `CON-02` Existing automated coverage must remain green after the refactor.
- `CON-03` New internal boundaries must be simple enough to keep Bash maintainable for the next increment.

## Verify

### Exit Criteria

- `EC-01` The codebase has a modular shell layout with an explicit orchestration pipeline and centralized agent adapters, while the CLI contract and existing behavior remain intact.
- `EC-02` Documentation explains the new module boundaries, the internal pipeline, and the conditions that would justify a later Python migration.
- `EC-03` Local syntax, static checks, and automated tests pass.

### Acceptance Scenarios

- `SC-01` Given the current repository configuration, when a user runs the documented `start-issue` commands that were supported before modularization, then the commands keep the same user-facing behavior while the implementation now delegates to focused modules.
- `SC-02` Given an agent-specific operation such as launch, branch-name generation, or prompt improvement, when `start-issue` needs that operation, then the orchestration layer calls a consistent adapter interface instead of branching ad hoc through the main script.

### Traceability Matrix

| Requirement ID | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- |
| `REQ-01` | `EC-01`, `SC-01` | `CHK-01`, `CHK-02` | `EVID-01` |
| `REQ-02` | `EC-01`, `SC-01`, `SC-02` | `CHK-01`, `CHK-02` | `EVID-01` |
| `REQ-03` | `EC-01`, `SC-02` | `CHK-01`, `CHK-02` | `EVID-01` |
| `REQ-04` | `EC-01` | `CHK-01`, `CHK-02` | `EVID-01` |
| `REQ-05` | `EC-02` | `CHK-02`, `CHK-03` | `EVID-01` |
| `REQ-06` | `EC-01`, `EC-03`, `SC-01` | `CHK-01`, `CHK-02`, `CHK-03` | `EVID-01` |

### Checks

| Check ID | Covers | How to check | Expected |
| --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `EC-03`, `SC-01`, `SC-02` | `bash -n scripts/start-issue && shellcheck scripts/start-issue` | Entry script and sourced modules remain syntactically valid and shellcheck-clean. |
| `CHK-02` | `EC-01`, `EC-03`, `SC-01`, `SC-02` | `mise exec -- bats test` | Existing workflow behavior stays green through the integration suite. |
| `CHK-03` | `EC-02`, `EC-03` | `git diff --check` | Documentation and refactor edits are internally consistent and whitespace-clean. |

### Test Matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | Local terminal output from syntax and shellcheck runs. |
| `CHK-02` | `EVID-01` | Local terminal output from Bats test runs and CI output after PR update. |
| `CHK-03` | `EVID-01` | Local terminal output from `git diff --check`. |

### Evidence

- `EVID-01` Verification command output showing the modularized implementation passes local checks and CI.

### Evidence Contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | Local and CI command output | implementer / CI | Terminal output and GitHub Actions job | `CHK-01`, `CHK-02`, `CHK-03` |
