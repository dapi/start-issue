---
title: "FT-008: Worktree Lifecycle Stabilization"
doc_kind: feature
doc_function: canonical
purpose: "Canonical feature document for stabilizing worktree lifecycle behavior and tightening local verification for start-issue."
derived_from:
  - https://github.com/dapi/start-issue/issues/8
  - ../../../doc/spec.md
  - ../../../README.md
status: active
delivery_status: in_progress
audience: humans_and_agents
must_not_define:
  - selected_design
  - implementation_sequence
---

# FT-008: Worktree Lifecycle Stabilization

## What

### Problem

`start-issue` currently mixes planning and side effects in its worktree startup path. Branch-to-worktree matching is not exact, path reuse can continue without proving that the directory belongs to the intended branch, and the local engineering loop is weaker than CI because `make test` is missing.

### Outcome

The worktree lifecycle becomes deterministic and safe to reuse, while local verification and the Bash structure become easier to extend without changing the user-facing CLI.

### Scope

- `REQ-01` `make test` runs the same local verification stack needed for day-to-day development.
- `REQ-02` Existing-branch reuse resolves the exact matching worktree, not a prefix or adjacent branch.
- `REQ-03` Reuse of an existing path is validated before `init.sh` or agent launch can run.
- `REQ-04` Worktree lifecycle logic is split into clearer planning and side-effect steps inside the Bash implementation.
- `REQ-05` Regression coverage expands for branch/path conflicts, reuse/delete flows, `--flat`, and base-branch fallback behavior.

### Non-Scope

- `NS-01` Do not change the external CLI flags, argument order, or overall workflow shape.
- `NS-02` Do not rewrite the tool into another language.
- `NS-03` Do not redesign issue fetching, prompt rendering, or agent command contracts beyond what lifecycle correctness requires.

### Constraints

- `CON-01` Existing successful workflows should remain compatible unless they relied on unsafe reuse behavior.
- `CON-02` Lifecycle safety checks must fail before side effects that assume a valid reused worktree.

## Verify

### Exit Criteria

- `EC-01` `make test` exists and runs successfully in a ready local environment.
- `EC-02` Reuse and delete/recreate flows cannot target the wrong worktree.
- `EC-03` New lifecycle regression tests cover the risky scenarios named in the issue.

### Acceptance Scenarios

- `SC-01` Given an existing branch and multiple similarly named worktrees, when the user chooses reuse, `start-issue` continues only with the exact worktree registered to the requested branch.
- `SC-02` Given an existing directory at the planned worktree path that is not the requested branch worktree, when the user chooses reuse, `start-issue` fails before `init.sh` or agent launch.

### Traceability Matrix

| Requirement ID | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- |
| `REQ-01` | `EC-01` | `CHK-01` | `EVID-01` |
| `REQ-02` | `EC-02`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-03` | `EC-02`, `SC-02` | `CHK-01` | `EVID-01` |
| `REQ-04` | `EC-02`, `EC-03` | `CHK-01` | `EVID-01` |
| `REQ-05` | `EC-03`, `SC-01`, `SC-02` | `CHK-01` | `EVID-01` |

### Checks

| Check ID | Covers | How to check | Expected |
| --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `EC-02`, `EC-03`, `SC-01`, `SC-02` | `make test` | Make target runs syntax, static checks, diff check, and Bats suite successfully. |

### Test Matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | Local terminal output from `make test` and equivalent direct commands if tool trust/setup blocks the wrapper. |

### Evidence

- `EVID-01` Verification command output showing the lifecycle regression suite and static checks pass.

### Evidence Contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | Local verification output | implementer | Terminal output for `make test` or its equivalent commands | `CHK-01` |
