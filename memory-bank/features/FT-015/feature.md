---
title: "FT-015: Codex batch human-gate mode with resume flow"
doc_kind: feature
doc_function: canonical
purpose: "Canonical feature document for adding a Codex batch human-gate workflow to start-issue. Owns only the problem space and verification contract."
derived_from:
  - https://github.com/dapi/start-issue/issues/26
  - ../../../README.md
  - ../../../README.ru.md
  - ../../../doc/spec.md
  - ../../../scripts/lib/start_issue/agent.sh
  - ../../../scripts/lib/start_issue/cli.sh
  - ../../../scripts/lib/start_issue/output.sh
  - ../../../scripts/lib/start_issue/pipeline.sh
status: active
delivery_status: in_progress
audience: humans_and_agents
must_not_define:
  - selected_design
  - implementation_sequence
---

# FT-015: Codex batch human-gate mode with resume flow

## What

### Problem

`start-issue` can already prepare an issue worktree and launch Codex interactively, and the repository already uses `codex exec` for non-interactive helper operations such as branch-name generation and prompt improvement. But there is no first-class workflow that lets Codex work unattended on an issue and then reopen the exact same saved session only when a human decision is required.

Issue #26 asks for a Codex-oriented mode that:

1. prepares the issue worktree and prompt exactly like the normal issue flow;
2. runs Codex non-interactively;
3. exits successfully without opening an interactive session when the final status is `DONE`;
4. resumes the same Codex session interactively when the final status is `HUMAN_GATE`.

### Outcome

`start-issue 123 --agent codex --human-gate` becomes a first-class Codex-only workflow that runs `codex exec` in a resumable, non-ephemeral mode, captures the resulting `thread_id`, interprets the final status from the saved last-message file, and either exits `0` on `DONE` or resumes the same session by explicit thread id on `HUMAN_GATE`.

### Scope

- `REQ-01` Add a `--human-gate` mode for issue-starting workflows and make it valid only when the resolved agent is `codex`.
- `REQ-02` Preserve the normal issue-start preparation flow in this mode: resolve issue/repo/base branch, create or reuse the worktree, run optional `init.sh`, and render the selected prompt before invoking Codex.
- `REQ-03` In `--human-gate` mode, replace the normal interactive Codex launch with a non-interactive `codex exec` invocation that emits JSON events, saves the last message, and remains resumable.
- `REQ-04` Capture `thread_id` from the `codex exec --json` event stream and store it alongside other per-run state artifacts in a predictable directory under the prepared worktree.
- `REQ-05` Parse the final status from the saved last-message file, recognize `STATUS: DONE` and `STATUS: HUMAN_GATE`, and fail clearly when the status or required thread id is missing.
- `REQ-06` If the final status is `DONE`, exit `0` without opening an interactive Codex TUI.
- `REQ-07` If the final status is `HUMAN_GATE`, resume the same Codex session by explicit thread id using `codex resume --include-non-interactive`, not `--last`.
- `REQ-08` Document the mode in normal help and provide dedicated help that explains the flow, prompt contract, examples, exit codes, state artifacts, and troubleshooting.
- `REQ-09` Cover the mode with automated tests for command generation, `DONE`, `HUMAN_GATE`, missing status, missing thread id, and Codex-only validation.

### Non-Scope

- `NS-01` Do not generalize the mode to Claude, Kimi, Pi, or `agent=none` in this feature.
- `NS-02` Do not redesign normal interactive Codex launches when `--human-gate` is not selected.
- `NS-03` Do not add generic session listing, resume browsing, or cleanup subcommands beyond what is required for this Codex-only flow.
- `NS-04` Do not invent a new prompt templating system; the mode reuses the existing prompt resolution and rendering workflow.

### Constraints

- `CON-01` The mode must be resumable, so it must not rely on ephemeral Codex execution.
- `CON-02` The resumed interactive session must target the exact saved `thread_id`, not whichever session happens to be most recent.
- `CON-03` Unknown or missing final status is a hard failure and must point the user to the saved last-message artifact.
- `CON-04` The normal `start-issue ISSUE` workflow remains behaviorally unchanged when `--human-gate` is absent.
- `CON-05` Dedicated help must describe the exact final-status contract instead of leaving it implicit in example prompts.

## Verify

### Exit Criteria

- `EC-01` `--human-gate` is accepted only for resolved agent `codex` and leaves other agents on a clear failure path.
- `EC-02` The mode prepares the issue worktree normally, then runs Codex through a resumable non-interactive `codex exec` path that records JSON events, last message, and thread id.
- `EC-03` `STATUS: DONE` exits `0` without opening interactive Codex, while `STATUS: HUMAN_GATE` resumes the same session by explicit thread id.
- `EC-04` Missing or unrecognized final status, missing thread id, or failed resume produce the documented non-zero exit behavior and actionable diagnostics.
- `EC-05` Help, spec, and automated coverage document the mode consistently, including prompt contract, examples, exit codes, and state-artifact location.

### Acceptance Scenarios

- `SC-01` Given `start-issue 123 --agent codex --human-gate`, when the issue flow reaches agent launch, then the command runs `codex exec` rather than the normal interactive Codex launch and passes the rendered prompt through the batch path.
- `SC-02` Given `start-issue 123 --agent codex --human-gate`, when `codex exec --json` emits `thread.started`, then the workflow stores the resulting `thread_id` in the per-run state directory.
- `SC-03` Given the saved last message starts with `STATUS: DONE`, when the batch run completes, then `start-issue` exits `0` without opening interactive Codex.
- `SC-04` Given the saved last message starts with `STATUS: HUMAN_GATE` and a `thread_id` was captured, when the batch run completes, then `start-issue` resumes `codex resume --include-non-interactive "$thread_id"`.
- `SC-05` Given `STATUS: HUMAN_GATE` but interactive resume cannot be opened, when the batch run completes, then the command exits `2` after printing the resume command and thread id.
- `SC-06` Given the saved last message has no recognized final status, when the batch run completes, then the command exits `1` and points to the saved last-message file.
- `SC-07` Given `--human-gate` is used with resolved agent `claude`, `kimi`, `pi`, or `none`, when parsing and validation complete, then the command exits with a clear Codex-only validation error.
- `SC-08` Given `start-issue --human-gate-help`, when the command prints dedicated help, then the output explains the mode flow, final-status contract, exit codes, state artifacts, troubleshooting, and examples.

### Traceability Matrix

| Requirement ID | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- |
| `REQ-01` | `EC-01`, `SC-07` | `CHK-01`, `CHK-02` | `EVID-01` |
| `REQ-02` | `EC-02`, `SC-01` | `CHK-02` | `EVID-01` |
| `REQ-03` | `EC-02`, `SC-01` | `CHK-01`, `CHK-02` | `EVID-01` |
| `REQ-04` | `EC-02`, `SC-02` | `CHK-02` | `EVID-01` |
| `REQ-05` | `EC-03`, `EC-04`, `SC-03`, `SC-04`, `SC-06` | `CHK-01`, `CHK-02` | `EVID-01` |
| `REQ-06` | `EC-03`, `SC-03` | `CHK-02` | `EVID-01` |
| `REQ-07` | `EC-03`, `EC-04`, `SC-04`, `SC-05` | `CHK-01`, `CHK-02` | `EVID-01` |
| `REQ-08` | `EC-05`, `SC-08` | `CHK-03` | `EVID-01` |
| `REQ-09` | `EC-05`, `SC-01` - `SC-08` | `CHK-02` | `EVID-01` |

### Checks

| Check ID | Covers | How to check | Expected |
| --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `EC-02`, `EC-03`, `EC-04` | `bash -n scripts/start-issue scripts/lib/start_issue/*.sh` | CLI/help/pipeline/Codex batch helpers remain syntactically valid. |
| `CHK-02` | `EC-01`, `EC-02`, `EC-03`, `EC-04`, `EC-05`, `SC-01` - `SC-08` | `mise exec -- bats test/start_issue.bats` | Automated coverage passes for command generation, status parsing, resume behavior, missing-state failures, dedicated help, and Codex-only validation. |
| `CHK-03` | `EC-05`, `SC-08` | `git diff --check` | Feature docs, help, and spec edits are internally consistent and whitespace-clean. |

### Test Matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | Local terminal output from syntax checks. |
| `CHK-02` | `EVID-01` | Local terminal output from Bats runs and CI output after branch update. |
| `CHK-03` | `EVID-01` | Local terminal output from `git diff --check`. |

### Evidence

- `EVID-01` Verification command output showing Codex batch command generation, thread-id capture, final-status handling, resume behavior, and dedicated-help coverage.

### Evidence Contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | Local and CI command output | implementer / CI | Terminal output and GitHub Actions job | `CHK-01`, `CHK-02`, `CHK-03` |
