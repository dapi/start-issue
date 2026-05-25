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

Add a dedicated user-level `setup` onboarding workflow plus one-time first-run prompting for `~/.config/start-issue`, while preserving the existing `init` and ordinary issue-starting behavior.

## Grounding / Support References

| Document | Role in this plan | Facts reused | Conflict action |
| --- | --- | --- | --- |
| `feature.md` | canonical problem / verify owner | `REQ-*`, `SC-*`, `CHK-*`, `EVID-01` | Update `feature.md` first |
| `solution.md` | canonical solution owner | `SOL-*`, `CTR-*`, `SD-*`, `FM-*`, `RB-*` | Update `solution.md` first |
| `decision-log.md` | resolved local decisions | `DL-*`, `FPF-*` resolutions | Update `decision-log.md` first |
| `../../../scripts/lib/start_issue/init.sh` | current config-init owner | built-in agent/prompt defaults, user/project scope mechanics, force semantics | Reuse prompt/default logic instead of inventing new onboarding rules, without inheriting repo-only scope requirements. |
| `../../../scripts/lib/start_issue/config.sh` | current config resolution | fallback behavior when files are absent | Preserve omission semantics for skipped agent and declined prompt |
| `../../../scripts/lib/start_issue/output.sh` | help and compact usage owner | existing short usage and config-surface reporting | Extend without replacing compact missing-issue behavior |
| `../../../README.md`, `../../../README.ru.md`, `../../../doc/spec.md` | public contract | current `init`, config paths, and missing-issue UX | Update if user-visible onboarding behavior changes |

## Current State / Reference Points

| Path / module | Current role | Why relevant | Reuse / mirror |
| --- | --- | --- | --- |
| `scripts/lib/start_issue/cli.sh` | Parses subcommands and flags, currently including `init`, `update`, and issue inputs. | Needs setup-mode normalization without breaking existing mode conflicts. | Mirror existing `init` / `update` parsing pattern. |
| `scripts/lib/start_issue/init.sh` | Owns interactive scope selection and user/project config writes. | Strongest existing source for default agent/prompt logic and user-config write behavior. | Reuse prompt derivation and file-writing helpers where possible. |
| `scripts/lib/start_issue/config.sh` | Resolves user/project agent, model, and prompt config. | Defines the meaning of omitted files on later runs. | Preserve fallback-to-default behavior when setup skips writes. |
| `scripts/lib/start_issue/output.sh` | Prints full help, compact missing-issue usage, and config summaries. | Must add compact first-run onboarding messaging and setup documentation. | Extend current compact usage instead of printing full help during onboarding. |
| `scripts/lib/start_issue/pipeline.sh` | Owns missing-issue mode and normal workflow sequencing. | First-run gate must happen before the normal pipeline while allowing continuation. | Insert onboarding gate without consuming the original command. |
| `test/start_issue.bats` | Existing integration coverage for config, missing-issue, init, update, and agent/model behavior. | Primary regression net for interactive onboarding and file omission semantics. | Extend current fake-CLI and piped-input test style. |

## Test Strategy

| Test surface | Canonical refs | Existing coverage | Planned automated coverage | Required local suites / commands | Required CI suites / jobs | Manual-only gap / justification | Manual-only approval ref |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Setup command entry parsing | `REQ-01`, `SC-01`, `SC-02` | No current setup coverage. | Add Bats coverage for both `setup` and `--setup`. | `mise exec -- bats test` | Existing CI `test` job | none | none |
| Outside-git setup execution | `REQ-09`, `SC-09` | `init --user` already provides nearby behavior. | Add Bats coverage that runs explicit setup outside a repository and proves no issue/repo prerequisites are required. | `mise exec -- bats test` | Existing CI `test` job | none | none |
| User-config onboarding writes and omissions | `REQ-02`, `REQ-03`, `REQ-04`, `SC-01`, `SC-03`, `SC-04` | `init` covers adjacent behavior only. | Add Bats scenarios for agent selection, skip, prompt preview/save, and omitted files. | `mise exec -- bats test` | Existing CI `test` job | none | none |
| First-run accept/decline continuation | `REQ-05`, `REQ-06`, `REQ-07`, `SC-05`, `SC-06`, `SC-07`, `SC-10` | Missing-issue mode currently has no onboarding gate. | Add Bats coverage for compact usage, accept path, decline path, one-time directory marker, and continued normal execution. | `mise exec -- bats test` | Existing CI `test` job | none | none |
| `init` compatibility | `REQ-08`, `SC-08` | Existing `init` coverage already exists. | Re-run suite and add assertions only if setup parsing or shared helpers create regressions. | `mise exec -- bats test` | Existing CI `test` job | none | none |
| Docs/help/spec consistency | `REQ-10`, `SC-05`, `SC-08`, `SC-10` | Existing docs in tree. | Update docs and run whitespace check. | `bash -n scripts/start-issue scripts/lib/start_issue/*.sh`, `git diff --check` | Existing CI static checks | none | none |

## Open Questions / Ambiguities

None currently open. Blocking ambiguities from issue #25 were resolved in [decision-log.md](decision-log.md).

## Environment Contract

| Area | Contract | Used by | Failure symptom |
| --- | --- | --- | --- |
| setup | Worktree may already contain unrelated user changes and must not be reset. | all steps | Auto-edits overwrite unrelated work. |
| test | `bash`, `mise`, and `bats` are available per repo tooling; `shellcheck` may be available locally. | verification steps | Local verification cannot complete and must be reported. |
| access / network / secrets | Setup/onboarding tests should rely on fake CLIs and local HOME fixtures rather than external services. | onboarding and regression tests | Tests become flaky or depend on live GitHub state. |

## Preconditions

| Precondition ID | Canonical ref | Required state | Used by steps | Blocks start |
| --- | --- | --- | --- | --- |
| `PRE-01` | `DL-01` - `DL-04` | The `setup`/`init` boundary, prompt baseline, and first-run continuation rule are fixed. | `STEP-02` - `STEP-06` | yes |
| `PRE-02` | `CON-05`, `SD-04` | Directory existence is the accepted one-time onboarding marker. | `STEP-03` - `STEP-06` | yes |
| `PRE-03` | `REQ-07`, `SD-05` | Ordinary command execution must resume after onboarding resolves. | `STEP-03` - `STEP-06` | yes |
| `PRE-04` | `REQ-09`, `SD-06` | Explicit setup must remain runnable without repo discovery or issue-fetch dependencies. | `STEP-02` - `STEP-06` | yes |

## Workstreams

| Workstream | Implements | Result | Owner | Dependencies |
| --- | --- | --- | --- | --- |
| `WS-1` | `REQ-01`, `REQ-02`, `REQ-08`, `REQ-09` | CLI parsing and onboarding helpers add explicit setup mode without replacing init or requiring repo context. | agent | `PRE-01`, `PRE-04` |
| `WS-2` | `REQ-03`, `REQ-04`, `REQ-05`, `REQ-06`, `REQ-07` | User onboarding flow and first-run gate handle accept/decline and resume the original command. | agent | `WS-1`, `PRE-02`, `PRE-03` |
| `WS-3` | `REQ-10` | Docs and tests describe the same onboarding contract. | agent | `WS-1`, `WS-2` |

## Approval Gates

None. The plan does not require destructive repo actions.

## Work Order

| Step ID | Actor | Implements | Goal | Touchpoints | Artifact | Verifies | Evidence IDs | Check command / procedure | Blocked by | Needs approval | Escalate if |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | agent | `REQ-10` | Establish the FT-014 feature package and working contract. | `memory-bank/features/FT-014/*` | Feature docs | `CHK-03` | `EVID-01`, `EVID-14` | Review document boundaries and issue alignment. | none | none | Feature package drifts from issue #25 intent. |
| `STEP-02` | agent | `REQ-01`, `REQ-02`, `REQ-08`, `REQ-09` | Add setup parsing and user-only onboarding entry flow while preserving `init` and repo-independent setup behavior. | `scripts/start-issue`, `cli.sh`, shared onboarding helpers | Code changes | `CHK-01`, `CHK-02` | `EVID-01` | Syntax check plus regression suite. | `PRE-01`, `PRE-04` | none | `setup`, `init`, and `update` mode parsing become ambiguous or setup still requires repo state. |
| `STEP-03` | agent | `REQ-03`, `REQ-04` | Implement onboarding agent selection, prompt preview, and selective file writes under `~/.config/start-issue`. | onboarding helpers, `init.sh` or new module, `output.sh` | Code changes | `CHK-01`, `CHK-02` | `EVID-01` | Focused Bats scenarios for save/skip behavior. | `STEP-02` | none | Prompt derivation cannot stay aligned with current init defaults. |
| `STEP-04` | agent | `REQ-05`, `REQ-06`, `REQ-07` | Add the one-time first-run gate and continue the original non-setup command after accept/decline. | `pipeline.sh`, `output.sh`, config helpers | Code changes | `CHK-01`, `CHK-02` | `EVID-01` | Bats scenarios for first-run accept/decline and continued execution. | `STEP-02`, `PRE-02`, `PRE-03` | none | First-run gating cannot resume the original command safely. |
| `STEP-05` | agent | `REQ-10` | Update help text, README, Russian README, and spec for setup/onboarding, outside-git setup, and `init` compatibility. | `README.md`, `README.ru.md`, `doc/spec.md`, `output.sh` | Docs | `CHK-03` | `EVID-01` | `git diff --check` plus doc review. | `STEP-02` - `STEP-04` | none | Docs and implementation disagree about setup/write semantics, repo independence, or first-run continuation. |
| `STEP-06` | agent | `EC-01` - `EC-06` | Run verification and record completion state. | full change surface | Verification output | `CHK-01`, `CHK-02`, `CHK-03` | `EVID-01` | `bash -n scripts/start-issue scripts/lib/start_issue/*.sh`, `git diff --check`, `mise exec -- bats test` | `STEP-02` - `STEP-05` | none | Required tools are unavailable or regressions remain unresolved. |

## Parallelizable Work

- `PAR-01` Feature docs can be prepared before code changes because they define the onboarding contract.
- `PAR-02` Setup helper work can start before first-run gate wiring, but the continuation rule must stay fixed.
- `PAR-03` Public docs should wait until setup/write semantics and first-run continuation are fixed in code.

## Checkpoints

| Checkpoint ID | Refs | Condition | Evidence IDs |
| --- | --- | --- | --- |
| `CP-01` | `STEP-02`, `STEP-03` | Setup mode is explicit, user-only, repo-independent, and preserves skip/save omission semantics. | `EVID-01` |
| `CP-02` | `STEP-04`, `STEP-06` | First-run gating is one-time, compact, and resumes the requested command after accept or decline. | `EVID-01`, `EVID-14` |

## Execution Risks

| Risk ID | Risk | Impact | Mitigation | Trigger |
| --- | --- | --- | --- | --- |
| `ER-01` | Setup reuses `init` too loosely and accidentally writes project config, force-style behavior, or repo-only prerequisites. | Violates issue scope and changes existing semantics. | Keep user-only onboarding as a separate contract even if helpers are shared. | Setup writes into `.start-issue`, mentions project scope, or fails outside a repo. |
| `ER-02` | First-run decline does not create `~/.config/start-issue`. | Onboarding repeats forever. | Treat the directory itself as the completion marker and cover decline behavior with tests. | The second ordinary run still prompts setup. |
| `ER-03` | Skip agent is implemented as `none` instead of missing file. | Later runs change default-agent behavior. | Preserve omitted-file semantics from current config resolution. | Resolved agent becomes `none` after a skipped setup. |
| `ER-04` | First-run gating exits after setup/decline instead of continuing the original command. | Ordinary launches become two-step and contradict the issue. | Make continuation an explicit orchestration contract and test both accept and decline paths. | After first-run prompt, issue fetch/worktree steps never occur. |
| `ER-05` | Docs describe setup as repo-independent but implementation still runs git checks first. | User-facing contract and behavior diverge immediately. | Test setup outside a repository and keep repo checks out of the explicit setup path. | `start-issue setup` fails before creating `~/.config/start-issue` outside a repo. |

## Stop Conditions / Fallback

| Stop ID | Related refs | Trigger | Immediate action | Safe fallback state |
| --- | --- | --- | --- | --- |
| `STOP-01` | `REQ-07`, `FM-03` | The first-run gate cannot safely continue the original command without breaking existing mode handling. | Stop auto-fixing and raise a human gate with facts, options, and risk. | Keep explicit `setup` available while deferring automatic first-run continuation. |
| `STOP-02` | `REQ-08`, `FM-04` | Preserving `init` semantics and adding `setup` creates an unresolved user-facing conflict. | Stop and re-open the onboarding boundary decision from `DL-01`. | Keep `init` as the only documented initializer until the conflict is resolved. |

## Plan-local Evidence

| Evidence ID | Artifact | Producer | Path contract | Reused by checkpoints |
| --- | --- | --- | --- | --- |
| `EVID-14` | Feature-flow working-contract summary | implementer | Final response summary | `CP-02` |

## Execution Status

- `STEP-01` completed on 2026-05-24 by creating the FT-014 feature package and decision log.
- `STEP-02` - `STEP-06` not started in this document-only review pass.

## Ready For Acceptance

- The feature package is ready to guide implementation work for issue #25 once the review-improve cycle closes without `critical` or `important` document issues.
- Final acceptance remains owned by `feature.md`.
