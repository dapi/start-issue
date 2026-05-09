---
title: "FT-008: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution plan for FT-008. Tracks implementation order, test strategy, and current delivery status without redefining the canonical feature or solution."
derived_from:
  - feature.md
  - solution.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_008_scope
  - ft_008_selected_design
  - ft_008_acceptance_criteria
  - ft_008_blocker_state
---

# FT-008: Implementation Plan

## Current Goal

Stabilize worktree lifecycle handling, add a usable local `make test`, and increase regression confidence for conflict-heavy startup flows.

## Grounding / Support References

| Document | Role in this plan | Facts reused | Conflict action |
| --- | --- | --- | --- |
| `feature.md` | canonical problem / verify owner | `REQ-*`, `SC-01`, `SC-02`, `CHK-01`, `EVID-01` | Update `feature.md` first |
| `solution.md` | canonical solution owner | `SOL-*`, `SD-*`, `CTR-*`, `FM-*`, `RB-*` | Update `solution.md` first |

## Current State / Reference Points

| Path / module | Current role | Why relevant | Reuse / mirror |
| --- | --- | --- | --- |
| `Makefile` | Only install/uninstall targets exist. | Issue explicitly requires `make test`. | Mirror CI verification steps where practical. |
| `scripts/start-issue` | Orchestrates config, fetch, worktree lifecycle, init, and agent launch in one Bash script. | Main lifecycle bug surface. | Preserve CLI contract while separating planning helpers. |
| `test/start_issue.bats` | Existing integration coverage for config, prompt, and dry-run flows. | Needs lifecycle regression scenarios. | Extend current fake-CLI test style. |

## Test Strategy

| Test surface | Canonical refs | Existing coverage | Planned automated coverage | Required local suites / commands | Manual-only gap / justification |
| --- | --- | --- | --- | --- | --- |
| Local verification entrypoint | `REQ-01` | none | Assert `make -n test` contains the expected verification stack. | `make test` | none |
| Lifecycle conflict handling | `REQ-02`, `REQ-03`, `REQ-05`, `SC-01`, `SC-02` | minimal | Bats scenarios for exact branch reuse, invalid path reuse, mismatched branch reuse, delete/recreate, `--flat`, and base fallback. | `make test` | none |
| Internal decomposition | `REQ-04` | none | Static review plus shell syntax/static checks. | `bash -n scripts/start-issue`, `shellcheck scripts/start-issue` | none |

## Environment Contract

| Area | Contract | Used by | Failure symptom |
| --- | --- | --- | --- |
| setup | Git, Bash, ShellCheck, and Bats are available locally. | `STEP-02` - `STEP-04` | Verification blocked and must be reported. |
| repo state | Existing user changes outside FT-008 are preserved. | all steps | Conflicting edits require careful merge, not revert. |

## Workstreams

| Workstream | Implements | Result | Owner | Dependencies |
| --- | --- | --- | --- | --- |
| `WS-1` | `REQ-02`, `REQ-03`, `REQ-04` | Safer lifecycle planner and reuse validation in `scripts/start-issue`. | agent | none |
| `WS-2` | `REQ-01`, `REQ-05` | Local test target and expanded Bats coverage. | agent | `WS-1` |
| `WS-3` | feature-flow process | FT-008 memory-bank package. | agent | none |

## Work Order

| Step ID | Actor | Implements | Goal | Touchpoints | Verifies | Check command / procedure |
| --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | agent | feature-flow process | Create FT-008 feature package from issue #8. | `memory-bank/features/FT-008/*` | document consistency | Review doc boundaries and issue alignment. |
| `STEP-02` | agent | `REQ-02`, `REQ-03`, `REQ-04` | Refactor worktree planning and harden reuse validation. | `scripts/start-issue` | `CHK-01` | `bash -n scripts/start-issue` |
| `STEP-03` | agent | `REQ-01`, `REQ-05` | Add local `make test` and lifecycle regression tests. | `Makefile`, `test/start_issue.bats` | `CHK-01` | `make test` |
| `STEP-04` | agent | `EC-01` - `EC-03` | Run verification and record status. | full change surface | `CHK-01` | `bash -n scripts/start-issue && shellcheck scripts/start-issue && bats test` and `make test` when environment permits |

## Execution Risks

| Risk ID | Risk | Impact | Mitigation |
| --- | --- | --- | --- |
| `ER-01` | Interactive conflict flows are easy to break while refactoring. | Wrong worktree can be reused or deleted. | Add Bats coverage for each user choice branch. |
| `ER-02` | Local wrapper target diverges from CI. | False confidence in `make test`. | Keep the target aligned with the same static checks and Bats suite. |

## Execution Status

- `STEP-01` completed on 2026-05-09.
- `STEP-02` and `STEP-03` implemented locally on 2026-05-09.
- `STEP-04` in progress pending full local verification output.
