---
title: "FT-004: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution plan for FT-004. Tracks discovery context, implementation steps, risks, and test strategy without redefining canonical problem or solution facts."
derived_from:
  - feature.md
  - solution.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_004_scope
  - ft_004_selected_design
  - ft_004_acceptance_criteria
  - ft_004_blocker_state
---

# FT-004: Implementation Plan

## Current Goal

Replace the earlier prompt-fragment direction with the prompt-template improvement workflow expected by issue #4, then verify locally and in PR CI.

## Grounding / Support References

| Document | Role in this plan | Facts reused | Conflict action |
| --- | --- | --- | --- |
| `feature.md` | canonical problem / verify owner | `REQ-*`, `SC-01`, `CHK-01`, `EVID-01` | Update `feature.md` first |
| `solution.md` | canonical solution owner | `SOL-*`, `SD-*`, `CTR-*`, `FM-*`, `RB-*` | Update `solution.md` first |
| `runtime-surfaces.md` / `none` | optional grounding | none | Not used |
| `ui-reference/README.md` / `none` | optional interface reference | none | Not used |
| `use-cases/README.md` / `none` | optional scenario companion | none | Not used |

## Current State / Reference Points

| Path / module | Current role | Why relevant | Reuse / mirror |
| --- | --- | --- | --- |
| `scripts/start-issue` | Resolves prompt template, renders placeholders, builds launch command. | Primary change surface. | Reuse existing prompt precedence and agent adapter patterns. |
| `test/start_issue.bats` | Bats integration coverage with fake `gh` and fake agents. | Required verification surface. | Add prompt-improvement tests near prompt-template tests. |
| `test/helpers/fake-bin/*` | Fake agent CLIs for Bats. | Needed to emulate prompt-improvement generation. | Preserve existing branch-name behavior and add improvement output path. |
| `README.md`, `README.ru.md` | User-facing CLI/config docs. | New mode must be discoverable. | Preserve current tables and precedence explanation. |
| `doc/spec.md` | Russian script specification and acceptance criteria. | New behavior needs canonical spec coverage. | Update prompt section, algorithm, dry-run, examples, and acceptance criteria. |

## Test Strategy

| Test surface | Canonical refs | Existing coverage | Planned automated coverage | Required local suites / commands | Required CI suites / jobs | Manual-only gap / justification | Manual-only approval ref |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Prompt improvement proposal workflow | `REQ-01`, `REQ-02`, `REQ-03`, `REQ-04`, `SC-01`, `SOL-01` - `SOL-04` | Prompt file rendering already covered. | Bats tests for project prompt proposal output, built-in prompt proposal output, dry-run no-write, and `--agent none` rejection. | `bash -n scripts/start-issue`, `shellcheck scripts/start-issue`, `mise exec -- bats test` | Existing CI `test` job | none | none |
| Documentation consistency | `REQ-05` | README/spec document existing prompt config. | Static review plus `git diff --check`. | `git diff --check` | Existing CI static checks | none | none |

## Open Questions / Ambiguities

| Open Question ID | Question | Why unresolved | Blocks | Default action / escalation owner |
| --- | --- | --- | --- | --- |
| `OQ-01` | Should there be an auto-apply mode? | Issue asks for a mechanism to improve the prompt, but silent mutation is risky. | none | Keep proposal-only per `NS-01`; add apply later only with explicit user demand. |

## Environment Contract

| Area | Contract | Used by | Failure symptom |
| --- | --- | --- | --- |
| setup | Worktree is clean enough to amend the existing PR branch. | `STEP-01` - `STEP-06` | Unrelated dirty files overlap change surface. |
| test | Bats, shellcheck, bash, and mise are available per `mise.toml` and CI. | `CHK-01` | Missing tool blocks local verification and must be reported. |
| access / network / secrets | GitHub issue and PR are accessible through `gh`; implementation needs no further network access until push/CI. | `STEP-06` | `gh` push or PR edit fails. |

## Preconditions

| Precondition ID | Canonical ref | Required state | Used by steps | Blocks start |
| --- | --- | --- | --- | --- |
| `PRE-01` | `CON-01`, `SD-01`, `SD-02`, `SD-03` | Normal start workflow remains unchanged without `--improve-prompt`. | `STEP-02`, `STEP-04`, `STEP-05` | yes |

## Workstreams

| Workstream | Implements | Result | Owner | Dependencies |
| --- | --- | --- | --- | --- |
| `WS-1` | `REQ-01` - `REQ-04`, `SOL-01` - `SOL-04` | Script support for prompt-template improvement proposals. | agent | `PRE-01` |
| `WS-2` | `REQ-05` | Tests and docs updated. | agent | `WS-1` |

## Approval Gates

None. The plan does not require destructive or external production actions.

## Work Order

| Step ID | Actor | Implements | Goal | Touchpoints | Artifact | Verifies | Evidence IDs | Check command / procedure | Blocked by | Needs approval | Escalate if |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | agent | `REQ-05` | Rewrite feature package from prompt fragments to prompt improvement. | `memory-bank/features/FT-004/*` | Feature docs | `CHK-01` | `EVID-01` | Review document boundaries. | none | none | Feature docs drift from issue intent. |
| `STEP-02` | agent | `REQ-01` - `REQ-04` | Implement `--improve-prompt`, proposal output, and early exit. | `scripts/start-issue` | Code changes | `CHK-01` | `EVID-01` | `bash -n scripts/start-issue` | `PRE-01` | none | Bash portability issue appears. |
| `STEP-03` | agent | `REQ-02`, `REQ-03` | Update fake agents for deterministic proposal output. | `test/helpers/fake-bin/*` | Test helpers | `CHK-01` | `EVID-01` | Bats prompt-improvement tests | `STEP-02` | none | Fake branch-name behavior regresses. |
| `STEP-04` | agent | `REQ-05` | Add Bats coverage. | `test/start_issue.bats` | Tests | `CHK-01` | `EVID-01` | `mise exec -- bats test` | `STEP-02`, `STEP-03` | none | Existing tests fail for unchanged behavior. |
| `STEP-05` | agent | `REQ-05` | Update README/spec docs. | `README.md`, `README.ru.md`, `doc/spec.md` | Docs | `CHK-01` | `EVID-01` | `git diff --check` | `STEP-02` | none | Docs and implementation drift. |
| `STEP-06` | agent | `EC-01` | Run verification, amend PR branch, wait for CI. | full change surface | Verification output | `CHK-01` | `EVID-01` | `bash -n`, `shellcheck`, `mise exec -- bats test`, CI `test` | `STEP-02` - `STEP-05` | none | Required local tool or CI unavailable. |

## Parallelizable Work

- `PAR-01` Documentation updates can proceed after the script contract is stable.
- `PAR-02` Script and tests should stay sequential because test assertions depend on exact output.

## Checkpoints

| Checkpoint ID | Refs | Condition | Evidence IDs |
| --- | --- | --- | --- |
| `CP-01` | `STEP-02`, `STEP-03` | Script parses and proposal generation is deterministic in tests. | `EVID-01` |
| `CP-02` | `STEP-04`, `STEP-06` | Bats covers proposal write/dry-run/rejection and suite passes. | `EVID-01` |

## Execution Risks

| Risk ID | Risk | Impact | Mitigation | Trigger |
| --- | --- | --- | --- | --- |
| `ER-01` | Agent returns commentary instead of a clean prompt template. | Proposal file may need manual cleanup. | Request only the complete prompt template and strip simple code fences. | Real-agent output includes non-template prose. |
| `ER-02` | User expects automatic apply. | Feature may feel incomplete. | Keep proposal-only for safety and document copy/review step. | Follow-up request for apply mode. |

## Stop Conditions / Fallback

| Stop ID | Related refs | Trigger | Immediate action | Safe fallback state |
| --- | --- | --- | --- | --- |
| `STOP-01` | `CON-01`, `FM-01`, `FM-02` | Normal start behavior changes or proposal generation writes unsafe output. | Stop feature work and repair compatibility first. | Existing prompt/start behavior preserved. |

## Plan-local Evidence

| Evidence ID | Artifact | Producer | Path contract | Reused by checkpoints |
| --- | --- | --- | --- | --- |
| `EVID-09` | Simplify-review verdict | implementer | Final response summary | `CP-01`, `CP-02` |

## Execution Status

- `STEP-01` - `STEP-06` completed locally on 2026-05-09.
- `EVID-01` local checks passed: `bash -n scripts/start-issue`, `shellcheck scripts/start-issue`, `git diff --check`, `mise exec -- bats test`.
- `delivery_status` remains `in_progress` until the corrected PR branch has passing CI evidence.

## Ready For Acceptance

- All workstreams complete.
- `CHK-01` passes locally and in CI.
- Docs and implementation agree on `--improve-prompt`, proposal paths, and non-overwrite behavior.
- Final acceptance remains owned by `feature.md`.
