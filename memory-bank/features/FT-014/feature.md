---
title: "FT-014: First-run setup onboarding for user config"
doc_kind: feature
doc_function: canonical
purpose: "Canonical feature document for adding `setup` and first-run user-config onboarding to start-issue. Owns only the problem space and verification contract."
derived_from:
  - https://github.com/dapi/start-issue/issues/25
  - ../../../README.md
  - ../../../README.ru.md
  - ../../../doc/spec.md
  - ../../../scripts/lib/start_issue/init.sh
  - ../../../scripts/lib/start_issue/config.sh
  - ../../../scripts/lib/start_issue/output.sh
status: active
delivery_status: in_progress
audience: humans_and_agents
must_not_define:
  - selected_design
  - implementation_sequence
---

# FT-014: First-run setup onboarding for user config

## What

### Problem

`start-issue` already has `init`, and user-scoped config can be stored under `~/.config/start-issue`, but the current UX is oriented toward a technical/manual setup flow. There is no dedicated first-run onboarding entry point, and ordinary launches do not detect that the user config directory has never been initialized.

### Outcome

`start-issue` exposes a friendly user-level onboarding flow through both `start-issue setup` and `start-issue --setup`, and ordinary non-setup launches offer a one-time first-run prompt when `~/.config/start-issue` does not exist yet.

### Scope

- `REQ-01` Support both `start-issue setup` and `start-issue --setup` as equivalent entry points for interactive user-config onboarding.
- `REQ-02` `setup` works against `~/.config/start-issue` only: it creates the directory when missing and never writes project config under `.start-issue`.
- `REQ-03` `setup` asks the user to choose a default agent from `claude`, `codex`, `kimi`, `pi`, or skip, and if the user skips, no `agent` file is created.
- `REQ-04` `setup` derives the default prompt from the same built-in prompt contract used by current config initialization, shows that prompt to the user, and writes `prompt.md` only when the user explicitly confirms.
- `REQ-05` During an ordinary non-setup launch, if `~/.config/start-issue` does not exist, the command prints compact usage, reports that configuration is not initialized, and offers to run setup.
- `REQ-06` If the user declines first-run onboarding, `start-issue` creates an empty `~/.config/start-issue` directory, does not create `agent` or `prompt.md`, and does not show first-run onboarding automatically again on later launches.
- `REQ-07` After the first-run gate is resolved, the current command continues in the originally selected mode instead of silently exiting early.
- `REQ-08` Keep `start-issue init` available for backward compatibility and manual config workflows; `setup` becomes the friendly user-onboarding entry point rather than replacing `init`.
- `REQ-09` `setup` does not require an issue argument, git repository context, `gh`, or issue-fetch dependencies; it is a user-config workflow only.
- `REQ-10` Update `README.md`, `README.ru.md`, `doc/spec.md`, help text, and automated coverage for the new setup flow, first-run prompt, user-config file behavior, and `init`/`setup` relationship.

### Non-Scope

- `NS-01` Do not redesign project-scoped `.start-issue` initialization.
- `NS-02` Do not change issue fetching, worktree creation, prompt-improvement behavior, or self-update behavior beyond what the first-run gate requires.
- `NS-03` Do not add a richer config editor for model, worktree directory, or other settings not named in issue #25.
- `NS-04` Do not remove `init` or change its existing scope-selection semantics except where docs must distinguish it from `setup`.

### Constraints

- `CON-01` The setup target directory is `~/.config/start-issue`.
- `CON-02` Skipping agent selection must leave the user-level agent config absent rather than writing a sentinel value.
- `CON-03` Declining prompt persistence must leave `prompt.md` absent.
- `CON-04` The first-run experience must be compact and must not print the full `--help`.
- `CON-05` First-run onboarding is one-time per presence of `~/.config/start-issue`; the directory itself is the completion marker.
- `CON-06` Ordinary issue-starting behavior must remain intact after the first-run gate completes.
- `CON-07` Explicit setup must remain usable outside a git repository because it owns only user-level configuration.

## Verify

### Exit Criteria

- `EC-01` `start-issue setup` and `start-issue --setup` both enter the same interactive user-config onboarding workflow.
- `EC-02` Setup writes only the confirmed user-level files in `~/.config/start-issue` and preserves omission semantics for skipped agent and declined prompt persistence.
- `EC-03` An ordinary first run without `~/.config/start-issue` shows compact onboarding, creates the directory on acceptance or decline, and does not repeat onboarding automatically once the directory exists.
- `EC-04` Declining first-run onboarding still allows the originally requested non-setup command to continue with existing defaults.
- `EC-05` Docs, help text, spec, and tests describe the same `setup` behavior and the same relationship between `setup` and `init`.
- `EC-06` Explicit setup remains usable outside a git repository and without issue-fetch prerequisites.

### Acceptance Scenarios

- `SC-01` Given `~/.config/start-issue` does not exist, when the user runs `start-issue setup`, then the command creates the directory, asks for agent selection, shows the derived default prompt, and writes only the files the user confirms.
- `SC-02` Given `~/.config/start-issue` does not exist, when the user runs `start-issue --setup`, then the command performs the same onboarding workflow as the subcommand form.
- `SC-03` Given setup is running and the user chooses `skip` for agent selection, when the workflow completes, then `~/.config/start-issue/agent` is absent.
- `SC-04` Given setup is running and the user declines prompt persistence, when the workflow completes, then `~/.config/start-issue/prompt.md` is absent.
- `SC-05` Given an ordinary non-setup launch and `~/.config/start-issue` does not exist, when the command starts, then it prints compact usage plus a setup prompt before the normal workflow continues.
- `SC-06` Given the user declines first-run onboarding during an ordinary non-setup launch, when the command continues, then `~/.config/start-issue` exists, `agent` and `prompt.md` remain absent, and future ordinary launches no longer show first-run onboarding automatically.
- `SC-07` Given the user accepts first-run onboarding during an ordinary non-setup launch, when setup finishes, then the original command continues with the normal workflow using the resolved config state.
- `SC-08` Given the user runs `start-issue init`, when config initialization is requested, then the existing `init` scope-selection and file-writing workflow remains available.
- `SC-09` Given the user runs `start-issue setup` outside a git repository, when onboarding starts, then it still works against `~/.config/start-issue` without requiring repo discovery or issue fetching.
- `SC-10` Given the user runs `start-issue` without an issue after first-run onboarding has already been completed or bypassed, when the command prints guidance, then it keeps the compact missing-issue output and does not reopen onboarding.

### Traceability Matrix

| Requirement ID | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- |
| `REQ-01` | `EC-01`, `SC-01`, `SC-02` | `CHK-01`, `CHK-02` | `EVID-01` |
| `REQ-02` | `EC-02`, `SC-01` | `CHK-02` | `EVID-01` |
| `REQ-03` | `EC-02`, `SC-01`, `SC-03` | `CHK-02` | `EVID-01` |
| `REQ-04` | `EC-02`, `SC-01`, `SC-04` | `CHK-02` | `EVID-01` |
| `REQ-05` | `EC-03`, `SC-05`, `SC-10` | `CHK-02`, `CHK-03` | `EVID-01` |
| `REQ-06` | `EC-03`, `EC-04`, `SC-06` | `CHK-02` | `EVID-01` |
| `REQ-07` | `EC-04`, `SC-06`, `SC-07` | `CHK-02` | `EVID-01` |
| `REQ-08` | `EC-05`, `SC-08` | `CHK-02`, `CHK-03` | `EVID-01` |
| `REQ-09` | `EC-06`, `SC-09` | `CHK-02` | `EVID-01` |
| `REQ-10` | `EC-05`, `SC-05`, `SC-08`, `SC-10` | `CHK-01`, `CHK-02`, `CHK-03` | `EVID-01` |

### Checks

| Check ID | Covers | How to check | Expected |
| --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `EC-05` | `bash -n scripts/start-issue scripts/lib/start_issue/*.sh` | CLI parsing, onboarding helpers, and help/output changes remain syntactically valid. |
| `CHK-02` | `EC-01`, `EC-02`, `EC-03`, `EC-04`, `EC-05`, `EC-06`, `SC-01` - `SC-10` | `mise exec -- bats test` | Automated coverage proves both setup entry points, outside-git setup, first-run accept/decline behavior, file omission semantics, continued normal execution, and `init` compatibility. |
| `CHK-03` | `EC-03`, `EC-05`, `SC-05`, `SC-08`, `SC-10` | `git diff --check` | Help, docs, spec, and feature docs remain internally consistent and whitespace-clean. |

### Test Matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | Local terminal output from syntax checks. |
| `CHK-02` | `EVID-01` | Local terminal output from Bats runs and CI output after branch update. |
| `CHK-03` | `EVID-01` | Local terminal output from `git diff --check`. |

### Evidence

- `EVID-01` Verification command output showing setup entry-point equivalence, outside-git setup, first-run accept/decline behavior, user-config file writes/omissions, continued normal execution, and unchanged `init` availability.

### Evidence Contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | Local and CI command output | implementer / CI | Terminal output and GitHub Actions job | `CHK-01`, `CHK-02`, `CHK-03` |
