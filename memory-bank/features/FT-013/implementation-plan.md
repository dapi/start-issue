---
title: "FT-013: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution plan for FT-013. Tracks discovery context, implementation steps, risks, and verification strategy without redefining canonical feature or solution facts."
derived_from:
  - feature.md
  - solution.md
  - decision-log.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_013_scope
  - ft_013_selected_design
  - ft_013_acceptance_criteria
  - ft_013_blocker_state
---

# FT-013: Implementation Plan

## Current Goal

Add a self-update workflow to `start-issue` that upgrades the running installation from the latest published GitHub release without changing the existing issue-starting workflow.

## Grounding / Support References

| Document | Role in this plan | Facts reused | Conflict action |
| --- | --- | --- | --- |
| `feature.md` | canonical problem / verify owner | `REQ-*`, `SC-*`, `CHK-*`, `EVID-01` | Update `feature.md` first |
| `solution.md` | canonical solution owner | `SOL-*`, `CTR-*`, `FM-*`, `RB-*`, `SD-*` | Update `solution.md` first |
| `decision-log.md` | ambiguity and conflict owner | `DL-*` | Update `decision-log.md` first |
| `../../../install.sh` | release-backed install contract | asset URLs, checksum verification, install mode | Update only if contract changes intentionally |
| `../../../README.md`, `../../../README.ru.md` | public install and CLI docs | install path, current CLI docs | Update if user-facing behavior changes |
| `../../../doc/spec.md` | canonical script spec | workflow requirements, CLI flags, algorithm details | Update if implementation or workflow contract changes |

## Current State / Reference Points

| Path / module | Current role | Why relevant | Reuse / mirror |
| --- | --- | --- | --- |
| `scripts/start-issue` | CLI entrypoint with version constant and module bootstrap. | Update mode starts at CLI parsing and version reporting. | Preserve ordinary entry behavior while adding update mode. |
| `scripts/lib/start_issue/cli.sh` | Parses flags and positional arguments. | Must normalize `update` and `--update` without breaking issue input parsing. | Extend, do not fork parsing logic. |
| `scripts/lib/start_issue/output.sh` | Renders help and user-facing status. | Must document the new mode and its output states. | Keep output style consistent. |
| `scripts/lib/start_issue/github.sh` | Fetches GitHub issue metadata. | Existing `gh` integration shows the current network integration style. | Avoid coupling update mode to issue-fetch semantics. |
| `install.sh` | Release-backed installer implementation. | Strongest existing source for asset, checksum, and install behavior. | Reuse or mirror its contract. |
| `test/start_issue.bats`, `test/helpers/fake-bin/gh` | End-to-end and fake CLI coverage. | Update mode needs deterministic coverage for release lookup and install outcomes. | Extend helpers to simulate release APIs and failures. |

## Test Strategy

| Test surface | Canonical refs | Existing coverage | Planned automated coverage | Required local suites / commands | Required CI suites / jobs | Manual-only gap / justification | Manual-only approval ref |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Update command entry parsing | `REQ-01`, `SC-01`, `SC-02` | No current coverage. | Add Bats coverage for both `update` and `--update`. | `mise exec -- bats test` | Existing CI `test` job | none | none |
| Release resolution and install/no-op/failure behavior | `REQ-02` - `REQ-06`, `SC-01` - `SC-04` | No current coverage. | Add fake release API and asset flows for update-needed, already-current, and failure scenarios. | `mise exec -- bats test` | Existing CI `test` job | none | none |
| Version normalization | `REQ-03`, `REQ-04`, `SC-07` | No current coverage. | Add Bats coverage for equivalent installed/release versions such as `1.11.1` and `v1.11.1`. | `mise exec -- bats test` | Existing CI `test` job | none | none |
| No-downgrade behavior | `REQ-03`, `REQ-04`, `SC-08` | No current coverage. | Add Bats coverage for an installed version newer than the latest published release. | `mise exec -- bats test` | Existing CI `test` job | none | none |
| Outside-git execution for update mode | `REQ-03`, `REQ-07`, `SC-06` | Existing workflow assumes git repo for issue flow. | Add Bats coverage that runs update mode outside a repository. | `mise exec -- bats test` | Existing CI `test` job | none | none |
| Existing issue workflow regression | `REQ-07`, `SC-05` | Existing Bats suite covers core flow. | Re-run the suite and add assertions only if update parsing creates edge regressions. | `mise exec -- bats test` | Existing CI `test` job | none | none |
| Docs/spec consistency | `REQ-08`, `EC-04` | Existing docs in-tree. | Update docs plus whitespace check. | `git diff --check` | Existing CI static checks | none | none |

## Open Questions / Ambiguities

None currently open. Blocking ambiguities from issue #13 were resolved in `decision-log.md` as `DL-01` through `DL-06`.

## Environment Contract

| Area | Contract | Used by | Failure symptom |
| --- | --- | --- | --- |
| setup | Worktree may contain unrelated user changes and must not be reset. | all steps | Auto-edits overwrite unrelated work. |
| test | `bash`, `shellcheck`, `mise`, and `bats` are available per repo tooling. | verification steps | Local verification cannot complete and must be reported. |
| access / network / secrets | Real network access is not required for automated tests because fake release responses can be injected. | update tests | Tests become flaky or depend on live GitHub state. |

## Preconditions

| Precondition ID | Canonical ref | Required state | Used by steps | Blocks start |
| --- | --- | --- | --- | --- |
| `PRE-01` | `DL-01` - `DL-06` | The update target, comparison rule, no-downgrade rule, and installer contract are fixed. | `STEP-02` - `STEP-06` | yes |
| `PRE-02` | `CON-04`, `REQ-07` | Existing issue-starting workflow remains the regression baseline. | `STEP-02`, `STEP-05`, `STEP-06` | yes |

## Workstreams

| Workstream | Implements | Result | Owner | Dependencies |
| --- | --- | --- | --- | --- |
| `WS-01` | `REQ-01`, `REQ-07` | CLI parsing and orchestration gain an isolated update mode. | agent | `PRE-01`, `PRE-02` |
| `WS-02` | `REQ-02` - `REQ-06` | Release-backed update execution resolves latest release, normalizes versions, and installs safely. | agent | `WS-01` |
| `WS-03` | `REQ-08` | Public docs and spec describe the new workflow consistently. | agent | `WS-01`, `WS-02` |

## Approval Gates

None. The plan does not require destructive repo actions, but permission errors on the target executable must surface as user-facing command failures rather than silent fallback behavior.

## Work Order

| Step ID | Actor | Implements | Goal | Touchpoints | Artifact | Verifies | Evidence IDs | Check command / procedure | Blocked by | Needs approval | Escalate if |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | agent | `REQ-08` | Establish the FT-013 feature package and working contract. | `memory-bank/features/FT-013/*` | Feature docs | `CHK-03` | `EVID-01` | Review document boundaries and issue alignment. | none | none | Feature package drifts from issue #13 intent. |
| `STEP-02` | agent | `REQ-01`, `REQ-07` | Extend CLI parsing and top-level flow for isolated update mode. | `scripts/start-issue`, `scripts/lib/start_issue/cli.sh`, orchestration modules | Code changes | `CHK-01`, `CHK-02` | `EVID-01` | Syntax check and regression suite. | `PRE-01`, `PRE-02` | none | Update mode alters issue-input parsing or normal workflow behavior. |
| `STEP-03` | agent | `REQ-02` - `REQ-06` | Implement release metadata lookup, normalized version comparison, checksum verification, and install flow. | update helpers, `install.sh`, relevant modules | Code changes | `CHK-01`, `CHK-02` | `EVID-01` | Automated update-path coverage. | `STEP-02` | none | The release-backed contract must diverge from `install.sh` to work. |
| `STEP-04` | agent | `REQ-02` - `REQ-07` | Extend test doubles and Bats scenarios for metadata lookup, version-normalized no-op, failure, and outside-git update runs. | `test/start_issue.bats`, `test/helpers/fake-bin/gh`, fixtures | Tests | `CHK-02` | `EVID-01` | `mise exec -- bats test` | `STEP-02`, `STEP-03` | none | Fake release behavior cannot model the needed flows deterministically. |
| `STEP-05` | agent | `REQ-08` | Update English/Russian docs and canonical spec with the new workflow and expected outputs. | `README.md`, `README.ru.md`, `doc/spec.md` | Docs | `CHK-03` | `EVID-01` | `git diff --check` | `STEP-02`, `STEP-03` | none | Docs and implementation disagree about update target or release source. |
| `STEP-06` | agent | `EC-01` - `EC-05` | Run verification and record completion status. | full change surface | Verification output | `CHK-01`, `CHK-02`, `CHK-03` | `EVID-01` | `bash -n scripts/start-issue`, `shellcheck install.sh scripts/start-issue scripts/lib/start_issue/*.sh`, `git diff --check`, `mise exec -- bats test` | `STEP-02` - `STEP-05` | none | Required tools are unavailable or a regression remains unresolved. |

## Parallelizable Work

- `PAR-01` Feature docs can be reviewed and refined before code changes.
- `PAR-02` Test-double updates should follow the chosen release-resolution contract.
- `PAR-03` Public docs should wait until update-target and install behavior are fixed in code.

## Checkpoints

| Checkpoint ID | Refs | Condition | Evidence IDs |
| --- | --- | --- | --- |
| `CP-01` | `STEP-02`, `STEP-03` | Update mode is isolated from the issue workflow and uses a release-backed install contract. | `EVID-01` |
| `CP-02` | `STEP-04`, `STEP-06` | Both update entry forms, no-op behavior, failure handling, and normal workflow regression coverage pass. | `EVID-01` |

## Execution Risks

| Risk ID | Risk | Impact | Mitigation | Trigger |
| --- | --- | --- | --- | --- |
| `ER-01` | Update mode reuses issue-workflow prerequisites and becomes unusable outside git repos. | Direct violation of issue #13. | Keep update flow separate from repo/issue planning. | Update code reads git context before mode dispatch. |
| `ER-02` | Version lookup compares mismatched formats such as `1.12.0` versus `v1.12.0`. | False updates or false no-op decisions. | Normalize version strings before comparison and test both states. | Output says update needed when versions are semantically equal. |
| `ER-03` | Installer logic diverges between `install.sh` and self-update. | Two incompatible installation contracts emerge. | Share or mirror the same release/checksum/install behavior deliberately. | One path verifies checksum or target mode differently. |

## Stop Conditions / Fallback

| Stop ID | Related refs | Trigger | Immediate action | Safe fallback state |
| --- | --- | --- | --- | --- |
| `STOP-01` | `CON-04`, `REQ-07`, `FM-01` | Adding update mode regresses the existing issue workflow. | Stop expanding update behavior and repair compatibility first. | Existing issue workflow remains intact even if update mode scope is temporarily reduced. |
| `STOP-02` | `DL-01`, `FM-04` | Update behavior cannot stay consistent with the release-backed installer contract. | Stop and re-open the decision instead of shipping a second install path. | Existing release installation remains the only supported path. |

## Plan-local Evidence

| Evidence ID | Artifact | Producer | Path contract | Reused by checkpoints |
| --- | --- | --- | --- | --- |
| `EVID-13` | Feature-flow working-contract summary | implementer | Final response summary | `CP-01`, `CP-02` |

## Execution Status

- `STEP-01` completed on 2026-05-24 by creating the FT-013 feature package and decision log.
- `STEP-02` - `STEP-06` not started in this document-only review pass.

## Ready For Acceptance

- The feature package is ready to drive implementation once the review-improve cycle closes without `critical` or `important` document issues.
- Final acceptance remains owned by `feature.md`.
