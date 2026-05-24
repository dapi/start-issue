---
title: "FT-014: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution plan for FT-014. Tracks discovery context, implementation steps, risks, and verification strategy without redefining canonical problem or solution facts."
derived_from:
  - feature.md
  - solution.md
  - decision-log.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_014_scope
  - ft_014_selected_design
  - ft_014_acceptance_criteria
  - ft_014_blocker_state
---

# FT-014: Implementation Plan

## Current Goal

Add a Codex-only batch human-gate mode that finishes unattended on `DONE` and reopens the same saved Codex session on `HUMAN_GATE`, without regressing the ordinary issue-start workflow.

## Grounding / Support References

| Document | Role in this plan | Facts reused | Conflict action |
| --- | --- | --- | --- |
| `feature.md` | canonical problem / verify owner | `REQ-*`, `SC-*`, `CHK-*`, `EVID-01` | Update `feature.md` first |
| `solution.md` | canonical solution owner | `SOL-*`, `CTR-*`, `SD-*`, `FM-*`, `RB-*` | Update `solution.md` first |
| `decision-log.md` | resolved local decisions | `DL-*`, `FPF-*` resolutions | Update `decision-log.md` first |
| `../../../scripts/lib/start_issue/cli.sh` | current flag parser | `init`/`update` flag patterns, positional parsing constraints | Extend parser without breaking issue input handling |
| `../../../scripts/lib/start_issue/output.sh` | current help and dry-run output | Normal help surface, dry-run command display | Add human-gate help without fragmenting existing help surfaces |
| `../../../scripts/lib/start_issue/pipeline.sh` | current orchestration | Agent launch happens after worktree and prompt rendering | Branch late so issue preparation stays shared |
| `../../../scripts/lib/start_issue/agent.sh` | Codex adapter boundary | Existing interactive Codex launch and non-interactive `codex exec` helper use | Reuse Codex-specific command ownership instead of scattering it |
| `../../../doc/spec.md`, `../../../README.md`, `../../../README.ru.md` | user-facing contract | Existing launch semantics and help/docs style | Update together once the mode contract is fixed |
| `../../../test/start_issue.bats`, `../../../test/helpers/fake-bin/codex` | regression net | Existing Codex dry-run and non-interactive fake behavior | Extend fake Codex rather than introducing live Codex dependencies |

## Current State / Reference Points

| Path / module | Current role | Why relevant | Reuse / mirror |
| --- | --- | --- | --- |
| `scripts/lib/start_issue/cli.sh` | Parses flags and positional subcommands. | Must accept `--human-gate` and `--human-gate-help` without breaking `ISSUE`, `init`, or `update`. | Mirror the current flag-driven parsing style. |
| `scripts/lib/start_issue/output.sh` | Owns `--help`, missing-issue help, dry-run rendering, and user-facing summaries. | Must mention the new mode briefly in normal help and own dedicated human-gate help text. | Keep one output owner for user contract text. |
| `scripts/lib/start_issue/pipeline.sh` | Runs the main issue workflow and launches the selected agent. | Human-gate mode should branch at the launch step, not earlier. | Preserve current order through prompt rendering. |
| `scripts/lib/start_issue/agent.sh` | Builds launch commands and already uses `codex exec` for non-interactive helper tasks. | Best current home for Codex batch and resume command construction plus status parsing helpers. | Keep Codex-specific behavior centralized. |
| `test/helpers/fake-bin/codex` | Minimal Codex fake for dry-run, branch naming, and prompt improvement. | Needs richer behavior for JSON events, last-message writing, and resume failure simulation. | Extend deterministically through env vars. |
| `test/start_issue.bats` | End-to-end coverage of CLI and agent behavior. | Main place to cover `DONE`, `HUMAN_GATE`, missing-state failures, and dedicated help. | Reuse current issue fixture and fake binary harness. |

## Test Strategy

| Test surface | Canonical refs | Existing coverage | Planned automated coverage | Required local suites / commands | Required CI suites / jobs | Manual-only gap / justification | Manual-only approval ref |
| --- | --- | --- | --- | --- | --- | --- | --- |
| CLI parsing and validation | `REQ-01`, `REQ-08`, `SC-07`, `SC-08` | Existing coverage for `init`, `update`, help, and Codex dry-run. | Add Bats coverage for `--human-gate`, `--human-gate-help`, and non-Codex rejection. | `mise exec -- bats test/start_issue.bats` | Existing CI `test` job | none | none |
| Codex batch command generation | `REQ-02`, `REQ-03`, `SC-01` | Existing dry-run covers only interactive Codex launch. | Add assertions for the exact batch command shape and saved-artifact paths. | `mise exec -- bats test/start_issue.bats` | Existing CI `test` job | none | none |
| Thread-id capture and final-status handling | `REQ-04` - `REQ-07`, `SC-02` - `SC-06` | No current coverage. | Simulate JSON event streams and last-message files for `DONE`, `HUMAN_GATE`, missing status, missing thread id, and resume failure. | `mise exec -- bats test/start_issue.bats` | Existing CI `test` job | none | none |
| Normal issue-workflow regression | `CON-04`, `REQ-02`, `SC-01` | Existing Bats suite already covers ordinary launches. | Re-run existing suite and add assertions only where the new flags affect shared surfaces. | `mise exec -- bats test/start_issue.bats` | Existing CI `test` job | none | none |
| Docs/help/spec consistency | `REQ-08`, `REQ-09`, `SC-08` | Existing docs in tree. | Update docs/help/spec and run whitespace check. | `git diff --check` | Existing CI static checks | none | none |

## Open Questions / Ambiguities

None currently open. Blocking ambiguities from issue #26 were resolved in [decision-log.md](decision-log.md) as `DL-01` through `DL-05`.

## Environment Contract

| Area | Contract | Used by | Failure symptom |
| --- | --- | --- | --- |
| setup | Worktree may contain unrelated user changes and must not be reset. | all steps | Auto-edits overwrite unrelated work. |
| test | `bash`, `mise`, and `bats` are available per repo tooling. | verification steps | Local verification cannot complete and must be reported. |
| access / network / secrets | Tests must stay fake-Codex based and not depend on a real Codex account or live session state. | human-gate tests | CI becomes flaky or requires real credentials. |

## Preconditions

| Precondition ID | Canonical ref | Required state | Used by steps | Blocks start |
| --- | --- | --- | --- | --- |
| `PRE-01` | `DL-01`, `CON-04` | Human-gate mode is fixed as Codex-only and ordinary issue flow remains the regression baseline. | `STEP-02` - `STEP-06` | yes |
| `PRE-02` | `DL-02`, `REQ-08` | Dedicated help shape is fixed as a flag surface rather than an additional positional subcommand. | `STEP-02`, `STEP-05`, `STEP-06` | yes |
| `PRE-03` | `DL-03`, `DL-05`, `CON-01`, `CON-02` | Run-state artifacts and explicit-thread resume contract are fixed. | `STEP-03` - `STEP-06` | yes |

## Workstreams

| Workstream | Implements | Result | Owner | Dependencies |
| --- | --- | --- | --- | --- |
| `WS-01` | `REQ-01`, `REQ-08` | CLI/help surfaces gain human-gate flags, Codex-only validation, and dedicated help. | agent | `PRE-01`, `PRE-02` |
| `WS-02` | `REQ-02` - `REQ-07` | Main pipeline and Codex adapter gain batch execution, state capture, final-status parsing, and resume handoff. | agent | `WS-01`, `PRE-03` |
| `WS-03` | `REQ-09` | Tests, docs, and spec describe and verify the same mode contract. | agent | `WS-01`, `WS-02` |

## Approval Gates

None. The feature changes terminal behavior, but it does not require destructive repo actions. Resume-handling failures must surface as explicit command failures rather than silent fallbacks.

## Work Order

| Step ID | Actor | Implements | Goal | Touchpoints | Artifact | Verifies | Evidence IDs | Check command / procedure | Blocked by | Needs approval | Escalate if |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | agent | `REQ-08` | Establish the FT-014 feature package and working contract. | `memory-bank/features/FT-014/*` | Feature docs | `CHK-03` | `EVID-01`, `EVID-14` | Review document boundaries and issue alignment. | none | none | Feature package drifts from issue #26 intent. |
| `STEP-02` | agent | `REQ-01`, `REQ-08` | Add human-gate flags, dedicated help, and Codex-only validation. | `cli.sh`, `output.sh`, `pipeline.sh` | Code changes | `CHK-01`, `CHK-02`, `CHK-03` | `EVID-01` | Syntax check, Bats help/validation tests, and doc diff check. | `PRE-01`, `PRE-02` | none | Help surface and real CLI behavior diverge. |
| `STEP-03` | agent | `REQ-02` - `REQ-07` | Implement the Codex batch execution path, run-state artifacts, status parsing, and explicit-thread resume handoff. | `pipeline.sh`, `agent.sh`, run-state helper module if needed | Code changes | `CHK-01`, `CHK-02` | `EVID-01` | Bats coverage for `DONE`, `HUMAN_GATE`, missing status, missing thread id, and resume failure. | `STEP-02`, `PRE-03` | none | Codex batch output cannot provide a stable `thread_id` or last-message contract without guessing. |
| `STEP-04` | agent | `REQ-09` | Extend fake Codex and test scenarios for deterministic batch and resume behavior. | `test/start_issue.bats`, `test/helpers/fake-bin/codex` | Tests | `CHK-02` | `EVID-01` | `mise exec -- bats test/start_issue.bats` | `STEP-03` | none | Tests require live Codex sessions or nondeterministic timing. |
| `STEP-05` | agent | `REQ-08`, `REQ-09` | Update README, Russian README, and spec with the new mode contract. | `README.md`, `README.ru.md`, `doc/spec.md` | Docs | `CHK-03` | `EVID-01` | `git diff --check` plus doc review. | `STEP-02`, `STEP-03` | none | Docs and actual exit/status behavior disagree. |
| `STEP-06` | agent | `EC-01` - `EC-05` | Run verification and record completion state. | full change surface | Verification output | `CHK-01`, `CHK-02`, `CHK-03` | `EVID-01` | `bash -n scripts/start-issue scripts/lib/start_issue/*.sh`, `git diff --check`, `mise exec -- bats test/start_issue.bats` | `STEP-02` - `STEP-05` | none | Required tools are unavailable or regressions remain unresolved. |

## Parallelizable Work

- `PAR-01` Feature docs can be reviewed and refined before code changes.
- `PAR-02` Fake Codex extensions can be prepared in parallel with batch-helper implementation once the artifact contract is fixed.
- `PAR-03` Public docs should wait until the dedicated-help and exit-code contract are settled in code.

## Checkpoints

| Checkpoint ID | Refs | Condition | Evidence IDs |
| --- | --- | --- | --- |
| `CP-01` | `STEP-02`, `STEP-03` | Human-gate mode branches late from the normal pipeline and resumes only by explicit thread id. | `EVID-01` |
| `CP-02` | `STEP-04`, `STEP-05`, `STEP-06` | Tests, help, README, and spec all reflect the same status and exit-code contract. | `EVID-01`, `EVID-14` |

## Execution Risks

| Risk ID | Risk | Impact | Mitigation | Trigger |
| --- | --- | --- | --- | --- |
| `ER-01` | CLI parsing for dedicated help or human-gate mode collides with existing issue positional parsing. | Help or issue start may break. | Keep dedicated help flag-based and extend the current parser conservatively. | `start-issue 123 --human-gate` or `start-issue --human-gate-help` is misparsed as an issue argument. |
| `ER-02` | The implementation captures stdout text but not a stable `thread_id`. | `HUMAN_GATE` cannot reopen the exact session. | Treat JSON event capture and saved thread-id as mandatory artifacts. | Resume logic reaches `--last` or has no id to use. |
| `ER-03` | Status parsing becomes too permissive. | Arbitrary agent summaries are misclassified as terminal statuses. | Parse only the documented `STATUS:` lines from the saved artifact. | A message without the exact contract still exits success. |
| `ER-04` | Dedicated help describes a prompt contract that tests do not enforce. | Prompt authors rely on stale or inaccurate documentation. | Keep help, spec, and fake-Codex tests aligned to one contract. | README/help says one exit/status behavior while tests prove another. |

## Stop Conditions / Fallback

| Stop ID | Related refs | Trigger | Immediate action | Safe fallback state |
| --- | --- | --- | --- | --- |
| `STOP-01` | `CON-01`, `CON-02`, `FM-02`, `FM-03` | A resumable Codex batch run cannot reliably expose or preserve `thread_id` from the available artifacts. | Stop auto-fixing and raise a human gate with the observed artifact gap, options, and risk. | Keep the existing interactive Codex launch as the only supported path. |
| `STOP-02` | `CON-04`, `FM-01` | Human-gate mode changes ordinary issue-start behavior even when the flag is absent. | Stop expanding the feature and repair the regression first. | Existing issue workflow remains the only supported launch path until isolated branching is restored. |

## Plan-local Evidence

| Evidence ID | Artifact | Producer | Path contract | Reused by checkpoints |
| --- | --- | --- | --- | --- |
| `EVID-14` | Feature-flow working-contract summary | implementer | Final response summary | `CP-02` |

## Execution Status

- `STEP-01` completed on 2026-05-24 by creating the FT-014 feature package and decision log.
- `STEP-02` completed on 2026-05-24 by adding `--human-gate` / `--human-gate-help`, Codex-only validation, and dedicated help surfaces.
- `STEP-03` completed on 2026-05-24 by adding the Codex batch execution path, run-state artifacts, final-status parsing, and explicit-thread resume handling.
- `STEP-04` completed on 2026-05-24 by extending fake Codex and Bats coverage for batch, resume, missing-status, and missing-thread-id scenarios.
- `STEP-05` completed on 2026-05-24 by updating `README.md`, `README.ru.md`, and `doc/spec.md` with the shipped human-gate contract.
- `STEP-06` completed locally on 2026-05-24 with passing `bash -n scripts/start-issue scripts/lib/start_issue/*.sh test/helpers/fake-bin/*`, `shellcheck install.sh scripts/start-issue scripts/lib/start_issue/*.sh test/helpers/fake-bin/*`, `git diff --check`, and `bats test/start_issue.bats`.

## Ready For Acceptance

- The feature package is ready to guide implementation work for issue #26 once the review-improve cycle closes without `critical` or `important` document issues.
- Final acceptance remains owned by `feature.md`.
