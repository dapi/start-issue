---
title: "FT-012: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution plan for FT-012. Tracks discovery context, implementation steps, risks, and verification strategy without redefining canonical problem or solution facts."
derived_from:
  - feature.md
  - solution.md
  - decision-log.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_012_scope
  - ft_012_selected_design
  - ft_012_acceptance_criteria
  - ft_012_blocker_state
---

# FT-012: Implementation Plan

## Current Goal

Add explicit, documented model selection to `start-issue` while preserving the existing agent-selection contract and backward-compatible no-model behavior.

## Grounding / Support References

| Document | Role in this plan | Facts reused | Conflict action |
| --- | --- | --- | --- |
| `feature.md` | canonical problem / verify owner | `REQ-*`, `SC-*`, `CHK-*`, `EVID-01` | Update `feature.md` first |
| `solution.md` | canonical solution owner | `SOL-*`, `CTR-*`, `SD-*`, `FM-*`, `RB-*` | Update `solution.md` first |
| `decision-log.md` | resolved local decisions | `DL-*`, `FPF-*` resolutions | Update `decision-log.md` first |
| `../../../README.md`, `../../../README.ru.md` | CLI and config contract | User-visible usage, init examples, precedence text | Preserve public behavior unless FT-012 explicitly changes it |
| `../../../doc/spec.md` | detailed script specification | Defaults, algorithm, help obligations, acceptance checklist | Update after implementation details settle |
| `../../../docs/agent-examples.md`, `../../../docs/agent-examples.ru.md` | user examples | Agent command examples that must grow model examples consistently | Update if launch syntax changes are documented |

## Current State / Reference Points

| Path / module | Current role | Why relevant | Reuse / mirror |
| --- | --- | --- | --- |
| `scripts/lib/start_issue/cli.sh` | Parses CLI flags. | Needs `--model` support. | Mirror `--agent` value parsing and validation shape where applicable. |
| `scripts/lib/start_issue/config.sh` | Resolves agent and prompt config. | Needs model resolution and source tracking. | Keep precedence reporting consistent with existing patterns. |
| `scripts/lib/start_issue/output.sh` | Prints help, config summaries, and dry-run output. | Must show resolved model and model config locations. | Reuse current selected-config / current-config display flow. |
| `scripts/lib/start_issue/init.sh` | Writes init config. | Must persist model config next to agent config. | Reuse current project/user scope behavior and force semantics. |
| `scripts/lib/start_issue/agent.sh` | Central adapter logic. | Must own model-aware validation and command building. | Preserve adapter ownership instead of scattering support checks. |
| `scripts/lib/start_issue/pipeline.sh` | Orchestrates normal and missing-issue flows. | Must thread resolved model through display, validation, and execution. | Keep current stage order unless validation needs an additional gate. |
| `test/start_issue.bats`, `test/helpers/fake-bin/*` | End-to-end workflow coverage and fake CLIs. | Main regression net for new config precedence and command shape. | Extend existing agent tests rather than creating a separate harness. |
| `README.md`, `README.ru.md`, `doc/spec.md`, `docs/agent-examples*.md` | User-facing docs. | Must describe agent/model inputs together. | Update in one pass to avoid contract drift. |

## Test Strategy

| Test surface | Canonical refs | Existing coverage | Planned automated coverage | Required local suites / commands | Required CI suites / jobs | Manual-only gap / justification | Manual-only approval ref |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Agent/model precedence | `REQ-01`, `REQ-02`, `REQ-03`, `SC-01`, `SC-02`, `SC-03`, `SC-04`, `SC-07` | Agent precedence already covered. | Add model CLI/project/user/env/default cases and mixed agent/model assertions. | `mise exec -- bats test` | Existing CI `test` job | none | none |
| Init writes and dry-run output | `REQ-03`, `REQ-04`, `SC-01`, `SC-02`, `SC-05`, `SC-08` | Init and dry-run agent coverage already exists. | Extend assertions for `.start-issue/model`, user config, and rendered command lines. | `mise exec -- bats test` | Existing CI `test` job | none | none |
| Adapter validation and no-model compatibility | `REQ-05`, `REQ-06`, `SC-06`, `SC-07`, `SC-08` | Current adapter coverage is indirect. | Add focused tests for unsupported combinations and no-model fallbacks. | `bash -n scripts/start-issue scripts/lib/start_issue/*.sh`, `mise exec -- bats test` | Existing CI `test` job | none | none |
| Documentation consistency | `REQ-07`, `SC-03`, `SC-04` | Docs already exist in tree. | Static review and whitespace check after updates. | `git diff --check` | Existing CI static checks | none | none |

## Open Questions / Ambiguities

None currently open. Blocking ambiguities were resolved in [decision-log.md](decision-log.md).

## Environment Contract

| Area | Contract | Used by | Failure symptom |
| --- | --- | --- | --- |
| setup | Worktree may already contain user changes outside this feature and must not be reset. | all steps | Config/doc edits accidentally overwrite unrelated work. |
| test | `bash`, `mise`, and `bats` are available per repo tooling; `shellcheck` may be available locally. | verification steps | Local verification cannot complete and must be reported. |
| access / network / secrets | Existing tests rely on fake CLIs; no new external secret is required for FT-012 docs or tests. | adapter/test steps | Tests start depending on real vendor CLIs or remote state. |

## Preconditions

| Precondition ID | Canonical ref | Required state | Used by steps | Blocks start |
| --- | --- | --- | --- | --- |
| `PRE-01` | `CON-01`, `CON-02`, `SD-01` | Agent precedence and model precedence are understood as separate but parallel config chains. | `STEP-02` - `STEP-06` | yes |
| `PRE-02` | `CON-03`, `SD-02` | Adapter boundary remains the single source of truth for explicit model support. | `STEP-03` - `STEP-06` | yes |
| `PRE-03` | `CON-04`, `SD-03` | No-model launch compatibility is preserved. | `STEP-03` - `STEP-06` | yes |

## Workstreams

| Workstream | Implements | Result | Owner | Dependencies |
| --- | --- | --- | --- | --- |
| `WS-1` | `REQ-01`, `REQ-02`, `REQ-03` | Model resolution and reporting are added across CLI, config, and output flows. | agent | `PRE-01` |
| `WS-2` | `REQ-04`, `REQ-05`, `REQ-06` | Init and adapter flows persist and validate model selection correctly. | agent | `WS-1`, `PRE-02`, `PRE-03` |
| `WS-3` | `REQ-07` | Docs and tests align to the combined agent/model contract. | agent | `WS-1`, `WS-2` |

## Approval Gates

None. The plan does not require destructive actions.

## Work Order

| Step ID | Actor | Implements | Goal | Touchpoints | Artifact | Verifies | Evidence IDs | Check command / procedure | Blocked by | Needs approval | Escalate if |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | agent | `REQ-07` | Establish the FT-012 feature package and working contract. | `memory-bank/features/FT-012/*` | Feature docs | `CHK-03` | `EVID-01`, `EVID-12` | Review document boundaries and issue alignment. | none | none | Feature package drifts from issue #12 intent. |
| `STEP-02` | agent | `REQ-01`, `REQ-02`, `REQ-03` | Add explicit model parsing, resolution, source tracking, and runtime/help reporting. | `cli.sh`, `config.sh`, `output.sh`, `pipeline.sh` | Code changes | `CHK-01`, `CHK-02` | `EVID-01` | Syntax check plus regression suite. | `PRE-01` | none | Help text, missing-issue output, and runtime output disagree about precedence or defaults. |
| `STEP-03` | agent | `REQ-05`, `REQ-06` | Extend adapter validation and command building for explicit model selection and no-model compatibility. | `agent.sh`, `pipeline.sh`, fake CLIs if needed | Code changes | `CHK-01`, `CHK-02` | `EVID-01` | Regression suite with dry-run command assertions and unsupported-combination tests. | `STEP-02`, `PRE-02`, `PRE-03` | none | One agent path cannot express model support or rejection through the adapter boundary. |
| `STEP-04` | agent | `REQ-04` | Extend init to write model config consistently with agent config. | `init.sh`, related tests | Code changes | `CHK-02` | `EVID-01` | Init and force-path tests. | `STEP-02` | none | Init can persist an invalid agent/model pair. |
| `STEP-05` | agent | `REQ-07` | Update docs and examples to describe agent/model behavior as one contract. | `README.md`, `README.ru.md`, `doc/spec.md`, `docs/agent-examples*.md` | Docs | `CHK-03` | `EVID-01` | `git diff --check` plus doc review. | `STEP-02` - `STEP-04` | none | Docs and real command shapes diverge. |
| `STEP-06` | agent | `EC-01` - `EC-05` | Run verification and record completion state. | full change surface | Verification output | `CHK-01`, `CHK-02`, `CHK-03` | `EVID-01` | `bash -n scripts/start-issue scripts/lib/start_issue/*.sh`, `git diff --check`, `mise exec -- bats test` | `STEP-02` - `STEP-05` | none | Required tools are unavailable or regressions remain unresolved. |

## Parallelizable Work

- `PAR-01` Feature docs can be prepared before code changes because they define the working contract.
- `PAR-02` Precedence/output work can start before init and adapter work settle, but the final docs must wait for the adapter contract.
- `PAR-03` Adapter validation and init persistence should stay sequential so both consume one support contract.

## Checkpoints

| Checkpoint ID | Refs | Condition | Evidence IDs |
| --- | --- | --- | --- |
| `CP-01` | `STEP-02`, `STEP-03` | Resolved model state is threaded through config/output and adapter validation without launch regressions. | `EVID-01` |
| `CP-02` | `STEP-04`, `STEP-05`, `STEP-06` | Init, docs, and tests all reflect the same explicit agent/model contract. | `EVID-01`, `EVID-12` |

## Execution Risks

| Risk ID | Risk | Impact | Mitigation | Trigger |
| --- | --- | --- | --- | --- |
| `ER-01` | Model precedence is implemented differently in runtime, init, and help text. | Users cannot predict which model is active. | Centralize resolution and reuse one display contract. | Different outputs cite different winning sources. |
| `ER-02` | Adapter support is guessed instead of validated by code/tests. | One agent silently ignores model selection. | Keep support checks in `agent.sh` and add explicit failing tests. | A dry-run command omits the model despite a configured value. |
| `ER-03` | No-model compatibility regresses because the implementation forces model flags everywhere. | Existing workflows change unexpectedly. | Preserve the unset-model rule from `DL-03` and cover it with regression tests. | Existing dry-run outputs change when no model is configured. |
| `ER-04` | Docs/examples update only README/spec and leave example docs stale. | Feature package and public docs drift. | Treat `docs/agent-examples*.md` as in-scope doc surfaces. | README/spec mention `--model`, example docs do not. |

## Stop Conditions / Fallback

| Stop ID | Related refs | Trigger | Immediate action | Safe fallback state |
| --- | --- | --- | --- | --- |
| `STOP-01` | `CON-03`, `FM-02`, `FM-03` | Adapter support for one agent cannot be expressed or verified without guessing unsupported CLI behavior. | Stop auto-fixing and raise a human gate with facts, options, and risk. | Keep model support scoped to the validated agent set only after explicit approval. |
| `STOP-02` | `CON-04`, `FM-04` | No-model compatibility and explicit model support cannot both be preserved with current design. | Stop and re-evaluate the compatibility decision from `DL-03`. | Preserve existing no-model behavior until a human chooses the tradeoff. |

## Plan-local Evidence

| Evidence ID | Artifact | Producer | Path contract | Reused by checkpoints |
| --- | --- | --- | --- | --- |
| `EVID-12` | Feature-flow working-contract summary | implementer | Final response summary | `CP-02` |

## Execution Status

- `STEP-01` completed on 2026-05-24 by creating the FT-012 feature package.
- `STEP-02` completed on 2026-05-24 by adding explicit model parsing, resolution, source tracking, and runtime/help reporting across `cli.sh`, `config.sh`, `output.sh`, and `pipeline.sh`.
- `STEP-03` completed on 2026-05-24 by extending `agent.sh` with model-aware launch, branch-name, and prompt-improvement adapter behavior.
- `STEP-04` completed on 2026-05-24 by extending `init.sh` to write, keep, or remove `model` config consistently with force semantics.
- `STEP-05` completed on 2026-05-24 by updating `README.md`, `README.ru.md`, `doc/spec.md`, and `docs/agent-examples*.md` to describe the combined agent/model contract.
- `STEP-06` completed locally on 2026-05-24 with passing `bash -n scripts/start-issue scripts/lib/start_issue/*.sh test/helpers/fake-bin/*`, `shellcheck scripts/start-issue scripts/lib/start_issue/*.sh`, `git diff --check`, and `bats test/start_issue.bats`.
- `delivery_status` remains `in_progress` until branch CI and PR review complete.

## Ready For Acceptance

- The feature package is ready to guide implementation work for issue #12.
- Final acceptance remains owned by `feature.md`.
