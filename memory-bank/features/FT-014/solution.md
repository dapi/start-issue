---
title: "FT-014: Solution"
doc_kind: feature
doc_function: canonical
purpose: "Canonical solution document for FT-014. Defines the selected setup-onboarding design without redefining feature scope or acceptance criteria."
derived_from:
  - feature.md
  - decision-log.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_014_scope
  - ft_014_acceptance_criteria
  - ft_014_delivery_status
  - implementation_sequence
---

# FT-014: Solution

## Selected Design

- `SOL-01` Add a dedicated setup mode that can be entered either by the `setup` subcommand or the `--setup` flag and normalize both forms into one workflow.
- `SOL-02` Scope `setup` to user-level onboarding only: it targets `~/.config/start-issue`, creates that directory if missing, and never writes project config.
- `SOL-03` Reuse the current config-init prompt contract for onboarding: agent selection drives the default prompt template, skip leaves user agent config absent, and prompt persistence remains an explicit confirmation step.
- `SOL-04` Add a first-run gate for ordinary non-setup launches that checks only for existence of `~/.config/start-issue`, prints compact usage plus an onboarding question when absent, and materializes the directory even if the user declines setup.
- `SOL-05` After the first-run gate resolves, continue the originally requested non-setup workflow using the same command mode and config-resolution pipeline that would normally follow.
- `SOL-06` Keep `init` as the backward-compatible manual/project-capable initializer, while documenting `setup` as the friendlier user-level onboarding entry point.
- `SOL-07` Keep explicit setup independent of git-repository discovery and issue-fetch prerequisites because it owns only user-level config state.
- `SOL-08` Update help text, README, Russian README, spec, and tests together so setup and first-run behavior are described as one coherent contract.

## Requirement Mapping

| Requirement ID | Solution / architecture refs | Notes |
| --- | --- | --- |
| `REQ-01` | `SOL-01`, `CTR-01` | Both command-entry forms converge into one setup mode. |
| `REQ-02` | `SOL-02`, `CTR-02`, `DL-01` | Setup owns user config only. |
| `REQ-03` | `SOL-03`, `CTR-03`, `DL-02` | Agent selection permits skip without writing a sentinel config. |
| `REQ-04` | `SOL-03`, `CTR-03`, `DL-03` | Prompt preview and persistence reuse the current default-prompt contract. |
| `REQ-05` | `SOL-04`, `CTR-04` | First-run gate is compact and directory-based. |
| `REQ-06` | `SOL-04`, `SOL-05`, `DL-04` | Decline path writes only the directory and suppresses future auto-onboarding. |
| `REQ-07` | `SOL-05`, `DL-04` | Ordinary commands continue after onboarding resolves. |
| `REQ-08` | `SOL-06`, `DL-01` | `init` remains available and distinct. |
| `REQ-09` | `SOL-02`, `SOL-07` | Setup is user-config only and therefore repo-independent. |
| `REQ-10` | `SOL-08` | User-visible docs/tests move together. |

## To-Be Flow

1. Parse CLI input and normalize `setup` and `--setup` into setup mode.
2. If setup mode is selected, run user-config onboarding against `~/.config/start-issue` without repo discovery or issue fetching, then exit after reporting the result.
3. If setup mode is not selected and the command is an ordinary non-init, non-update launch, check whether `~/.config/start-issue` exists before the rest of the normal pipeline.
4. When the directory is absent, print compact usage, explain that configuration is not initialized, and ask whether to run setup now.
5. If the user accepts, run the same setup workflow used by explicit setup mode.
6. If the user declines, create an empty `~/.config/start-issue` directory and leave `agent` and `prompt.md` absent.
7. After either accept or decline completes, continue the original non-setup command flow without reopening onboarding.
8. In setup mode, ask for agent choice (`claude`, `codex`, `kimi`, `pi`, `skip`), derive the default prompt for the effective agent, show the prompt, and persist only the files the user confirms.

## Contracts

| Contract ID | Related refs | Input / Output | Producer / Consumer | Notes |
| --- | --- | --- | --- | --- |
| `CTR-01` | `SOL-01`, `REQ-01` | Input: raw CLI args; output: normalized setup-mode state. | CLI parser / orchestration | `setup` and `--setup` are equivalent entry points. |
| `CTR-02` | `SOL-02`, `REQ-02`, `REQ-08` | Input: setup mode; output: user-config directory and optional `agent` / `prompt.md` files only. | setup workflow / filesystem | Project-scoped `.start-issue` remains owned by `init`. |
| `CTR-03` | `SOL-03`, `REQ-03`, `REQ-04` | Input: selected onboarding agent state; output: optional agent file plus previewed prompt and optional prompt file. | setup workflow / user | Skip means "leave agent file absent", not `agent=none`. |
| `CTR-04` | `SOL-04`, `SOL-05`, `REQ-05`, `REQ-06`, `REQ-07` | Input: ordinary command state plus user response; output: resolved first-run gate and continued original flow. | top-level orchestration / normal pipeline | The directory existence is the one-time onboarding marker. |
| `CTR-05` | `SOL-07`, `REQ-09` | Input: explicit setup invocation; output: completed user onboarding without repo or issue prerequisites. | setup workflow / CLI entrypoint | Setup must remain callable from outside a repository. |

## Target Architecture

### Architecture Invariants

- `setup` never writes project config.
- Missing `~/.config/start-issue` is the only first-run trigger.
- Declining onboarding still materializes the directory so the prompt is one-time.
- Skip/decline semantics are represented by omitted files, not synthetic config values.
- Explicit setup never requires repo discovery or issue-fetch tooling.
- Existing `init` remains available for manual or project-scoped initialization.
- Ordinary issue-starting behavior continues after onboarding instead of being replaced by it.

### Target Shape

| Layer / responsibility | To-be role | Boundary / non-owner | Related refs |
| --- | --- | --- | --- |
| CLI parsing | Recognize `setup` and `--setup`, keep `init`, `update`, and issue-input parsing coherent. | Does not own onboarding I/O. | `SOL-01` |
| Setup workflow | Own interactive user onboarding, directory creation, file omission semantics, prompt preview/save confirmation, and repo-independent execution. | Does not initialize project config. | `SOL-02`, `SOL-03`, `SOL-07` |
| First-run gate | Detect missing user-config directory on ordinary launches and route into setup or decline handling once. | Does not replace the normal issue pipeline. | `SOL-04`, `SOL-05` |
| Existing init flow | Continue owning manual project/user initialization and `--project` / `--user` semantics. | Does not become the primary first-run UX. | `SOL-06` |
| Docs/tests | Describe and verify one combined onboarding contract. | Do not introduce behavior not present in implementation. | `SOL-08` |

## Accepted Local Decisions

- `SD-01` `setup` is a user-onboarding command, not a rename of `init`; `init` stays available because it already owns project-scoped config and force semantics.
- `SD-02` Onboarding "skip" is represented by absence of `~/.config/start-issue/agent`, because issue #25 explicitly allows leaving the agent unspecified.
- `SD-03` Onboarding derives the default prompt using the same built-in prompt logic already used by config initialization so setup does not invent a second prompt baseline.
- `SD-04` First-run decline creates only `~/.config/start-issue`; absence of nested files remains meaningful and compatible with current fallback behavior.
- `SD-05` The first-run gate resolves before the main workflow but does not consume the original command; after onboarding it resumes normal execution.
- `SD-06` Explicit setup is repo-independent because its contract is limited to user-level config files under `~/.config/start-issue`.

## Change Surface

| Ref | Surface | Type | Why it changes |
| --- | --- | --- | --- |
| `SOL-01`, `SOL-04`, `SOL-05` | `scripts/start-issue`, `scripts/lib/start_issue/cli.sh`, `pipeline.sh` | code | Add setup parsing, first-run gating, and continuation into the normal pipeline. |
| `SOL-02`, `SOL-03`, `SOL-06`, `SOL-07` | `scripts/lib/start_issue/init.sh` or new onboarding helper module, `config.sh`, `output.sh` | code | Add user-only onboarding prompts, prompt preview/save flow, repo-independent setup execution, and compact first-run messaging while preserving `init`. |
| `SOL-01` - `SOL-08` | `test/start_issue.bats`, fake helpers if needed | test | Cover both setup entry points, outside-git setup, first-run accept/decline, file omissions, and `init` compatibility. |
| `SOL-08` | `README.md`, `README.ru.md`, `doc/spec.md`, `memory-bank/features/FT-014/*` | doc | Document the new onboarding contract and `setup`/`init` split. |

## Failure Modes

- `FM-01` First-run onboarding fires repeatedly because the decline path does not create `~/.config/start-issue`.
- `FM-02` Skip or decline writes a synthetic config value such as `none` and changes current default-agent behavior.
- `FM-03` The first-run gate consumes the original command and exits instead of continuing the requested workflow.
- `FM-04` Setup and `init` drift into contradictory user-config contracts.
- `FM-05` Setup derives a prompt baseline that differs from current built-in init/config behavior for the same effective agent.
- `FM-06` Explicit setup accidentally depends on git or issue-fetch tooling and fails outside repositories.

## Rollout / Backout

- `RB-01` Roll out by landing setup mode, repo-independent onboarding, first-run gating, docs, and regression tests together.
- `RB-02` Back out by removing setup mode and the first-run gate while retaining the existing `init` behavior.

## ADR Dependencies

See [decision-log.md](decision-log.md).
