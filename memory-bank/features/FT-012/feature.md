---
title: "FT-012: Explicit agent and model selection"
doc_kind: feature
doc_function: canonical
purpose: "Canonical feature document for adding explicit agent/model selection and model-aware adapter validation to start-issue. Owns only the problem space and verification contract."
derived_from:
  - https://github.com/dapi/start-issue/issues/12
  - ../../../README.md
  - ../../../README.ru.md
  - ../../../doc/spec.md
  - ../../../docs/agent-examples.md
  - ../../../docs/agent-examples.ru.md
status: active
delivery_status: in_progress
audience: humans_and_agents
must_not_define:
  - selected_design
  - implementation_sequence
---

# FT-012: Explicit agent and model selection

## What

### Problem

`start-issue` already resolves an agent from CLI, project config, user config, and environment, but model choice is not a first-class user-facing setting. Current help and no-issue output only partially surface agent selection, `init` cannot write a model config, and some internal agent operations still hardcode a Claude model instead of following an explicit user choice.

### Outcome

`start-issue` resolves both agent and model explicitly, documents both precedence chains, shows the effective values in user-facing output, writes model config during `init`, and validates unsupported agent/model combinations through the adapter boundary.

### Scope

- `REQ-01` Add explicit model selection inputs alongside the existing agent inputs: `--model`, `.start-issue/model`, `~/.config/start-issue/model`, and `START_ISSUE_MODEL`.
- `REQ-02` Preserve existing agent precedence exactly as documented today while adding model precedence `CLI -> project -> user -> environment -> built-in unset`.
- `REQ-03` Surface the resolved agent and resolved model in runtime-facing output, and document the supported agent/model flags, environment variables, and config file locations in `--help`.
- `REQ-04` Extend `start-issue init` so it can write both agent and model config at project or user scope.
- `REQ-05` Route model-aware behavior through the agent adapter contract so launch commands and other explicit model-aware adapter operations either pass the resolved model correctly or fail with a clear validation error.
- `REQ-06` Preserve backward-compatible behavior when no model is configured: no model should be passed to agent launch commands, and `--agent none` must never pass a model to a launch command.
- `REQ-07` Update docs and automated coverage for precedence, init writes, dry-run visibility, unsupported combinations, and no-model compatibility.

### Non-Scope

- `NS-01` Do not add new agents beyond `claude`, `codex`, `kimi`, `pi`, and `none`.
- `NS-02` Do not define or validate a universal catalog of model names across vendors.
- `NS-03` Do not redesign prompt-template precedence or prompt-improvement UX beyond what model-aware adapter support requires.
- `NS-04` Do not change worktree, branch naming, or issue-fetch semantics except where model-aware adapter validation must hook into existing flows.

### Constraints

- `CON-01` Existing agent precedence remains `CLI -> project -> user -> environment -> built-in default claude`.
- `CON-02` Model precedence must mirror the issue statement: `CLI -> project -> user -> environment -> built-in unset`.
- `CON-03` Unsupported agent/model combinations must fail clearly instead of silently ignoring the configured model.
- `CON-04` User-facing launch behavior without a configured model must remain backward-compatible.
- `CON-05` The feature package must stay consistent with `README.md`, `README.ru.md`, `doc/spec.md`, and `docs/agent-examples*.md` once implementation lands.

## Verify

### Exit Criteria

- `EC-01` `start-issue` resolves agent and model explicitly from documented sources and shows the effective values in user-facing output.
- `EC-02` `start-issue init` can write agent and model config for project or user scope.
- `EC-03` Agent adapters either pass the configured model in their supported command shape or reject unsupported combinations clearly.
- `EC-04` Existing behavior stays backward-compatible when no model is configured.
- `EC-05` Docs and automated coverage reflect the new precedence, init, dry-run, and validation behavior.

### Acceptance Scenarios

- `SC-01` Given `start-issue 123 --agent codex --model gpt-5.2 --dry-run`, when configuration resolves, then the output shows `Agent: codex`, `Model: gpt-5.2`, and a Codex launch command that includes the adapter-specific model argument.
- `SC-02` Given `start-issue 123 --agent claude --model sonnet --dry-run`, when configuration resolves, then the output shows `Agent: claude`, `Model: sonnet`, and a Claude launch command that includes the adapter-specific model argument.
- `SC-03` Given no issue argument, when the command prints guidance, then it displays the current agent/model configuration plus the supported config locations for both.
- `SC-04` Given `--help`, when the command prints usage, then it documents `--agent`, `--model`, `START_ISSUE_AGENT`, `START_ISSUE_MODEL`, `.start-issue/agent`, `.start-issue/model`, `~/.config/start-issue/agent`, and `~/.config/start-issue/model`.
- `SC-05` Given `start-issue init --project --agent codex --model gpt-5.2`, when init runs, then it writes `.start-issue/agent` and `.start-issue/model` without changing prompt-config behavior outside this feature.
- `SC-06` Given a configured model with an agent adapter that does not support explicit model selection, when the workflow reaches validation, then the command exits with a clear validation error instead of silently dropping the model.
- `SC-07` Given no model is configured, when `start-issue` runs, then launch commands omit model arguments and current non-model behavior remains intact.
- `SC-08` Given `--agent none`, when configuration includes a model, then user-facing output may still display the resolved model but no launch command receives it.

### Traceability Matrix

| Requirement ID | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- |
| `REQ-01` | `EC-01`, `SC-01`, `SC-02`, `SC-03`, `SC-04` | `CHK-01`, `CHK-02` | `EVID-01` |
| `REQ-02` | `EC-01`, `SC-03`, `SC-07` | `CHK-02` | `EVID-01` |
| `REQ-03` | `EC-01`, `SC-01`, `SC-02`, `SC-03`, `SC-04`, `SC-08` | `CHK-02`, `CHK-03` | `EVID-01` |
| `REQ-04` | `EC-02`, `SC-05` | `CHK-02` | `EVID-01` |
| `REQ-05` | `EC-03`, `SC-01`, `SC-02`, `SC-06` | `CHK-01`, `CHK-02` | `EVID-01` |
| `REQ-06` | `EC-04`, `SC-07`, `SC-08` | `CHK-02` | `EVID-01` |
| `REQ-07` | `EC-05`, `SC-03`, `SC-04`, `SC-05`, `SC-06`, `SC-07` | `CHK-02`, `CHK-03` | `EVID-01` |

### Checks

| Check ID | Covers | How to check | Expected |
| --- | --- | --- | --- |
| `CHK-01` | `EC-03` | `bash -n scripts/start-issue scripts/lib/start_issue/*.sh` | CLI/config/adapter changes remain syntactically valid. |
| `CHK-02` | `EC-01`, `EC-02`, `EC-03`, `EC-04`, `EC-05`, `SC-01` - `SC-08` | `mise exec -- bats test` | Automated coverage passes for precedence, dry-run, init writes, unsupported combinations, and no-model compatibility. |
| `CHK-03` | `EC-01`, `EC-05`, `SC-03`, `SC-04` | `git diff --check` | Docs/help/spec/example updates are internally consistent and whitespace-clean. |

### Test Matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | Local terminal output from syntax checks. |
| `CHK-02` | `EVID-01` | Local terminal output from Bats runs and CI output after PR update. |
| `CHK-03` | `EVID-01` | Local terminal output from `git diff --check`. |

### Evidence

- `EVID-01` Verification command output showing explicit agent/model resolution, init writes, adapter validation, and backward-compatible no-model behavior.

### Evidence Contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | Local and CI command output | implementer / CI | Terminal output and GitHub Actions job | `CHK-01`, `CHK-02`, `CHK-03` |
