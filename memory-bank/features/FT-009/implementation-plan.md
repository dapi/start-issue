---
title: "FT-009: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution plan for FT-009. Tracks discovery context, implementation steps, risks, and verification strategy without redefining canonical problem or solution facts."
derived_from:
  - feature.md
  - solution.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_009_scope
  - ft_009_selected_design
  - ft_009_acceptance_criteria
  - ft_009_blocker_state
---

# FT-009: Implementation Plan

## Current Goal

Modularize `start-issue` into maintainable shell components while preserving the existing CLI and automated behavior.

## Grounding / Support References

| Document | Role in this plan | Facts reused | Conflict action |
| --- | --- | --- | --- |
| `feature.md` | canonical problem / verify owner | `REQ-*`, `SC-01`, `SC-02`, `CHK-*`, `EVID-01` | Update `feature.md` first |
| `solution.md` | canonical solution owner | `SOL-*`, `SD-*`, `CTR-*`, `FM-*`, `RB-*` | Update `solution.md` first |
| `../../../README.md` | CLI and user contract | existing commands, config precedence, workflow phases | Preserve public behavior unless a defect is intentionally fixed |
| `../../../doc/spec.md` | detailed script expectations | algorithm, defaults, and documentation obligations | Update if implementation changes documented internals |

## Current State / Reference Points

| Path / module | Current role | Why relevant | Reuse / mirror |
| --- | --- | --- | --- |
| `scripts/start-issue` | Monolithic CLI, orchestration, and adapter logic. | Primary refactor surface. | Preserve behavior while extracting cohesive functions. |
| `test/start_issue.bats` | End-to-end workflow coverage. | Main regression net for the refactor. | Keep scenarios green; add targeted coverage only if a gap appears. |
| `test/helpers/fake-bin/*` | Fake CLIs for agents and `gh`. | Needed to preserve adapter behavior expectations in tests. | Mirror current adapter contract as it is normalized. |
| `README.md`, `README.ru.md` | User-facing behavior docs. | Public contract must remain stable. | Update only where internal architecture or extension guidance becomes user-relevant. |
| `doc/spec.md` | Canonical script spec. | Must reflect any meaningful architecture-facing contract or constraints. | Update after the module/pipeline shape is settled. |

## Test Strategy

| Test surface | Canonical refs | Existing coverage | Planned automated coverage | Required local suites / commands | Required CI suites / jobs | Manual-only gap / justification | Manual-only approval ref |
| --- | --- | --- | --- | --- | --- | --- | --- |
| CLI compatibility after modularization | `REQ-01`, `REQ-02`, `REQ-06`, `SC-01` | Existing Bats integration suite covers core flows. | Re-run existing suite; add tests only if the refactor exposes an uncovered behavior edge. | `bash -n scripts/start-issue`, `shellcheck scripts/start-issue`, `mise exec -- bats test` | Existing CI `test` job | none | none |
| Agent adapter normalization | `REQ-03`, `SC-02` | Existing agent-path coverage is indirect. | Validate through current tests and add focused assertions if adapter extraction changes observable behavior. | `mise exec -- bats test` | Existing CI `test` job | none | none |
| Documentation / whitespace consistency | `REQ-05`, `REQ-06` | Existing docs are already in-tree. | Static review plus whitespace check. | `git diff --check` | Existing CI static checks | none | none |

## Open Questions / Ambiguities

| Open Question ID | Question | Why unresolved | Blocks | Default action / escalation owner |
| --- | --- | --- | --- | --- |
| `OQ-01` | Should module files live under `scripts/lib/` or another path? | The issue names module responsibilities but not layout. | none | Choose the simplest layout that keeps sourcing explicit and portable. |
| `OQ-02` | How much of future structured output should be made concrete now? | The issue asks to prepare for growth without adding user-facing modes yet. | none | Add internal seams only; do not invent public flags in this iteration. |

## Environment Contract

| Area | Contract | Used by | Failure symptom |
| --- | --- | --- | --- |
| setup | Worktree may already contain user changes outside this feature and must not be reset. | all steps | Refactor accidentally overwrites unrelated work. |
| test | `bash`, `shellcheck`, `mise`, and `bats` are available per repo tooling. | verification steps | Local verification cannot complete and must be reported. |
| access / network / secrets | `gh` access is only needed insofar as the existing tests or workflows already require it. | compatibility verification | Commands relying on issue metadata fail unexpectedly. |

## Preconditions

| Precondition ID | Canonical ref | Required state | Used by steps | Blocks start |
| --- | --- | --- | --- | --- |
| `PRE-01` | `CON-01`, `REQ-06`, `SD-02` | The existing CLI contract is understood well enough to preserve it during extraction. | `STEP-02` - `STEP-06` | yes |
| `PRE-02` | `CON-02` | The current test suite is the minimum regression baseline. | `STEP-05`, `STEP-06` | yes |

## Workstreams

| Workstream | Implements | Result | Owner | Dependencies |
| --- | --- | --- | --- | --- |
| `WS-1` | `REQ-01`, `REQ-02`, `REQ-06` | Entry script reduced to bootstrap/orchestration with explicit pipeline stages. | agent | `PRE-01` |
| `WS-2` | `REQ-03`, `REQ-04` | Agent and output/planning contracts normalized behind modules. | agent | `WS-1` |
| `WS-3` | `REQ-05` | Architecture and migration-threshold docs updated. | agent | `WS-1`, `WS-2` |

## Approval Gates

None. The plan does not require destructive actions.

## Work Order

| Step ID | Actor | Implements | Goal | Touchpoints | Artifact | Verifies | Evidence IDs | Check command / procedure | Blocked by | Needs approval | Escalate if |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | agent | `REQ-05` | Establish the FT-009 feature package and working contract. | `memory-bank/features/FT-009/*` | Feature docs | `CHK-03` | `EVID-01` | Review document boundaries and issue alignment. | none | none | Feature package drifts from issue #9 intent. |
| `STEP-02` | agent | `REQ-01`, `REQ-02`, `REQ-06` | Extract CLI/config/orchestration responsibilities into modules while keeping the entry script stable. | `scripts/start-issue`, module paths | Code changes | `CHK-01`, `CHK-02` | `EVID-01` | Syntax check and regression suite. | `PRE-01` | none | The extraction changes user-visible behavior. |
| `STEP-03` | agent | `REQ-03` | Normalize agent adapter responsibilities behind one internal contract. | agent-related code paths, fake agent helpers if needed | Code changes | `CHK-01`, `CHK-02` | `EVID-01` | Regression suite with agent scenarios. | `STEP-02` | none | One agent path needs special-case logic the contract cannot express cleanly. |
| `STEP-04` | agent | `REQ-04` | Introduce normalized workflow state / output seams for future structured output and adjacent commands. | orchestration and output code paths | Code changes | `CHK-01`, `CHK-02` | `EVID-01` | Regression suite and diff review. | `STEP-02` | none | The new seam becomes an abstraction with no real consumer. |
| `STEP-05` | agent | `REQ-05`, `REQ-06` | Update README/spec docs to explain the modular architecture and Python-migration threshold. | `README.md`, `README.ru.md`, `doc/spec.md` | Docs | `CHK-03` | `EVID-01` | `git diff --check` | `STEP-02`, `STEP-03`, `STEP-04` | none | Docs and implementation disagree about module or pipeline responsibilities. |
| `STEP-06` | agent | `EC-01`, `EC-02`, `EC-03` | Run verification and record completion status. | full change surface | Verification output | `CHK-01`, `CHK-02`, `CHK-03` | `EVID-01` | `bash -n scripts/start-issue`, `shellcheck scripts/start-issue`, `git diff --check`, `mise exec -- bats test` | `STEP-02` - `STEP-05` | none | Required tools are unavailable or regressions remain unresolved. |

## Parallelizable Work

- `PAR-01` Feature docs can be prepared before code changes, because they define the working contract.
- `PAR-02` Documentation edits should wait until the module layout and adapter contract settle.
- `PAR-03` Code extraction and adapter normalization are tightly coupled and should stay sequential.

## Checkpoints

| Checkpoint ID | Refs | Condition | Evidence IDs |
| --- | --- | --- | --- |
| `CP-01` | `STEP-02`, `STEP-03` | Entry script is thinner, module boundaries are explicit, and agent behavior remains covered. | `EVID-01` |
| `CP-02` | `STEP-04`, `STEP-06` | Pipeline/output seams exist without changing the CLI contract, and all automated checks pass. | `EVID-01` |

## Execution Risks

| Risk ID | Risk | Impact | Mitigation | Trigger |
| --- | --- | --- | --- | --- |
| `ER-01` | Bash module extraction increases hidden global coupling instead of reducing it. | Refactor adds complexity without improving maintainability. | Normalize shared state deliberately and keep module interfaces small. | Modules need pervasive implicit globals. |
| `ER-02` | Behavior-preservation assumptions are wrong because some flows are covered only implicitly. | Regression slips through during refactor. | Use the existing Bats suite as guardrail and add focused tests if a gap appears. | A refactor changes output or command shape unexpectedly. |
| `ER-03` | Output/planning seams are over-designed before a concrete `--json` or subcommand feature exists. | Unnecessary abstraction burden. | Limit the seam to normalized internal state already needed by the refactor. | New types/functions appear with no direct current caller value. |

## Stop Conditions / Fallback

| Stop ID | Related refs | Trigger | Immediate action | Safe fallback state |
| --- | --- | --- | --- | --- |
| `STOP-01` | `CON-01`, `CON-02`, `REQ-06`, `FM-01`, `FM-02` | CLI compatibility or test parity cannot be preserved with the proposed split. | Stop adding abstractions and repair compatibility first. | Existing user-facing behavior remains intact even if modularization depth is reduced. |

## Plan-local Evidence

| Evidence ID | Artifact | Producer | Path contract | Reused by checkpoints |
| --- | --- | --- | --- | --- |
| `EVID-09` | Feature-flow working-contract summary | implementer | Final response summary | `CP-01`, `CP-02` |

## Execution Status

- `STEP-01` completed on 2026-05-09 by creating the FT-009 feature package.
- `STEP-02` completed on 2026-05-09 by extracting the CLI/config/orchestration flow into sourced shell modules while keeping `scripts/start-issue` as the stable entrypoint.
- `STEP-03` completed on 2026-05-09 by centralizing agent-specific operations in `scripts/lib/start_issue/agent.sh`.
- `STEP-04` completed on 2026-05-09 by making the orchestration pipeline explicit in `scripts/lib/start_issue/pipeline.sh` and separating output concerns into `output.sh`.
- `STEP-05` completed on 2026-05-09 by documenting the modular architecture and Python-migration threshold in `README.md`, `README.ru.md`, and `doc/spec.md`.
- `STEP-06` completed locally on 2026-05-09 with passing `bash -n scripts/start-issue`, `shellcheck scripts/start-issue scripts/lib/start_issue/*.sh`, `git diff --check`, and `mise exec -- bats test`.
- `delivery_status` remains `in_progress` until CI evidence is available for the branch.

## Ready For Acceptance

- The feature package is ready to drive implementation.
- Final acceptance remains owned by `feature.md`.
