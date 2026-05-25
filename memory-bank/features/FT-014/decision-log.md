---
title: "FT-014: Decision Log"
doc_kind: feature
doc_function: decision_log
purpose: "Resolved local decisions and conflict handling for FT-014. Stores only decisions grounded in the current feature package and in-scope artifacts."
derived_from:
  - https://github.com/dapi/start-issue/issues/25
  - feature.md
  - solution.md
  - ../../../scripts/lib/start_issue/init.sh
  - ../../../scripts/lib/start_issue/config.sh
  - ../../../scripts/lib/start_issue/output.sh
status: active
audience: humans_and_agents
---

# FT-014: Decision Log

## Decision Entries

| Decision ID | Date | Status | Topic | Decision |
| --- | --- | --- | --- | --- |
| `DL-01` | 2026-05-24 | accepted | `setup` versus `init` | Keep `init` for backward-compatible manual and project-scoped initialization, and add `setup` as the friendly user-only onboarding entry point for `~/.config/start-issue`. |
| `DL-02` | 2026-05-24 | accepted | Skip-agent semantics | In setup, "skip" means do not create `~/.config/start-issue/agent`; it does not write `none` or another sentinel value. |
| `DL-03` | 2026-05-24 | accepted | Prompt baseline for setup | Setup derives the previewed default prompt from the same built-in prompt rules already used by config initialization for the effective onboarding agent state. |
| `DL-04` | 2026-05-24 | accepted | First-run continuation | On ordinary launches, the first-run onboarding gate resolves before the normal pipeline, but after accept or decline the original command continues instead of exiting. |
| `DL-05` | 2026-05-24 | accepted | Explicit setup runtime boundary | `setup` is repo-independent and issue-independent because it owns only user-level configuration under `~/.config/start-issue`. |

## FPF-Closed Questions

### `FPF-01`: Should `setup` replace `init`, or coexist with it?

#### Why this mattered

Issue #25 explicitly names two possible directions: keep `setup` separate from `init`, or make `setup` the recommended UX while keeping `init` for technical/manual use. The document set needed one stable contract before feature scope, solution, and implementation plan could align.

#### Available facts

- Issue #25 says the task is to add a new `setup` command and `--setup` flag for interactive first-run configuration in `~/.config/start-issue`.
- The same issue notes a preferred direction: keep `init` for backward compatibility and make `setup` the new friendly onboarding entry point.
- Current code in `scripts/lib/start_issue/init.sh` already owns both project and user scopes, plus `--project`, `--user`, and `--force` semantics.
- Current docs in `README.md`, `README.ru.md`, and `doc/spec.md` already describe `init` as an existing public command.

#### FPF reasoning summary

- Bounded contexts: user-first onboarding and manual/project-capable initialization are related but not identical contexts.
- Evidence discipline: the issue provides a stated preference and current code/docs prove that `init` already owns behavior beyond the issue's requested onboarding scope.
- Decision rule: prefer the smallest change that satisfies the new UX without breaking a documented public command unnecessarily.

#### Resolution

Add `setup` as a user-only onboarding workflow for `~/.config/start-issue`, and keep `init` as the backward-compatible/manual initializer, including project scope.

#### Conflict handled

This resolves the potential conflict between "new onboarding command" and the already documented `init` contract. `setup` becomes the preferred UX for user config, while `init` remains the technical/manual initializer rather than being silently redefined.

### `FPF-02`: If the user skips agent selection during setup, what should be written?

#### Why this mattered

Issue #25 requires setup to let the user leave the default agent unspecified and says that in that case the `agent` file must not be created. The feature docs needed a precise meaning so later config resolution, docs, and tests would not diverge.

#### Available facts

- Issue #25 explicitly says "без указания / пропустить" must be allowed and that then the `agent` file is not created.
- Current config resolution in `scripts/lib/start_issue/config.sh` falls back from missing config files to `START_ISSUE_AGENT`, then to the built-in default `claude`.
- Current supported agent list includes `none`, but `none` means "do not launch an agent", not "unspecified default agent".

#### FPF reasoning summary

- Strict distinction: "no configured default agent" and "configured agent is none" are different states with different runtime meaning.
- Evidence discipline: the current resolver already gives a clear meaning to absent user config files, so inventing a sentinel would add a second meaning unnecessarily.
- Decision rule: preserve existing omission semantics when the issue explicitly allows omission.

#### Resolution

Implement setup skip as absence of `~/.config/start-issue/agent`. Do not write `none` or any other placeholder.

#### Conflict handled

This resolves the conflict between the existence of the `none` agent value in current CLI behavior and the issue's explicit "leave unspecified" requirement. The two states stay distinct.

### `FPF-03`: Which prompt should setup preview when the user skips agent selection?

#### Why this mattered

Issue #25 requires setup to generate and show a default prompt after agent selection, but it does not separately define the prompt baseline for the skip-agent case. The feature docs needed one grounded rule to avoid inconsistent docs and tests.

#### Available facts

- Current init logic in `scripts/lib/start_issue/init.sh` derives the default prompt from the selected effective agent: Claude gets the built-in Claude command; other agents get the portable prompt.
- Current config resolution in `scripts/lib/start_issue/config.sh` falls back to built-in default agent `claude` when no user/project/environment agent is configured.
- Issue #25 allows leaving the agent unspecified by omitting the `agent` file, not by changing the built-in default contract.

#### FPF reasoning summary

- Bounded contexts: setup preview should reflect the effective runtime baseline, not invent a third prompt contract.
- Trust boundary: the strongest in-repo facts are the existing built-in prompt derivation rules and the current built-in default agent.
- Decision rule: when the issue is silent, reuse the narrowest existing contract that keeps behavior predictable.

#### Resolution

Setup derives the previewed default prompt from the effective onboarding agent state using the current built-in prompt rules. If the user skips agent selection, the preview uses the built-in default agent behavior, which today is the Claude default prompt.

#### Conflict handled

This resolves the gap between the issue's prompt-preview requirement and the lack of an explicit skip-agent prompt rule, while staying aligned with current init/config behavior.

### `FPF-04`: What happens to the original command after first-run onboarding is accepted or declined?

#### Why this mattered

Issue #25 says ordinary launches should offer setup when the user config directory is missing, and also says that declining setup should create the empty directory so onboarding does not repeat. The documents needed to decide whether the command then exits or continues, because that materially affects orchestration, UX, and testing.

#### Available facts

- Issue #25 frames the behavior as "при обычном запуске" and describes setup as an offer during that launch, not as a separate required restart step.
- The issue's decline path creates only the directory and explicitly justifies that behavior as preventing repeated onboarding on every launch.
- Current `start-issue` behavior already distinguishes between ordinary command handling and missing-issue guidance in `scripts/lib/start_issue/pipeline.sh`; there is no existing "must restart after config change" contract.
- The expected UX example for first run shows compact usage and `Run setup now? [Y/n]`, which reads as an inline decision point inside the current launch.

#### FPF reasoning summary

- Object-of-talk discipline: the issue speaks about what happens during one ordinary launch, so the answer must preserve that launch as the unit of behavior.
- Evidence discipline: no source says the user must rerun the command after responding to the onboarding prompt.
- Decision rule: prefer the interpretation with the least extra friction and the fewest hidden user steps, unless contradicted by current behavior or issue text.

#### Resolution

The first-run gate runs inline during the original ordinary launch. After the user accepts or declines onboarding, the command continues in the originally requested non-setup mode.

#### Conflict handled

This resolves the ambiguity between "prompt inline and continue" versus "prompt then exit and require rerun". The inline-continuation rule is more consistent with the issue wording and with the explicit one-time-directory marker.

### `FPF-05`: Should explicit `setup` depend on repository or issue context?

#### Why this mattered

Issue #25 defines `setup` in terms of user-level config files under `~/.config/start-issue`, but current `start-issue` also has many repo-dependent workflows. The feature package needed to close whether explicit setup can run outside a repository before implementation and tests diverged.

#### Available facts

- Issue #25 defines the setup target as `~/.config/start-issue` and never mentions project scope, issue fetching, or git requirements for `setup`.
- The issue's acceptance criteria say `start-issue setup` and `start-issue --setup` should launch interactive setup; they do not require an issue argument.
- Current `init` behavior in `doc/spec.md` already distinguishes user scope from project scope and states that `start-issue init --user` can run outside a git repository.
- Current code in `scripts/lib/start_issue/init.sh` has separate handling for `user` scope without project-root requirements.

#### FPF reasoning summary

- Bounded contexts: explicit setup belongs to the same user-config context as `init --user`, not to the repo/issue execution context.
- Evidence discipline: every concrete setup fact in the issue points at user-level filesystem state, while repo dependence belongs to a different existing workflow.
- Decision rule: avoid introducing prerequisites that are not required by the issue and are not needed by the owned state transition.

#### Resolution

Explicit `setup` is repo-independent and issue-independent. It should run against `~/.config/start-issue` without requiring git discovery, repo detection, `gh`, or issue fetching.

#### Conflict handled

This resolves the architectural tension between current repo-dependent startup flows and the new user-only onboarding flow by assigning `setup` to the user-config bounded context.
