---
title: "FT-014: Solution"
doc_kind: feature
doc_function: canonical
purpose: "Canonical solution document for FT-014. Defines the selected Codex batch human-gate design without redefining feature scope or acceptance criteria."
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

- `SOL-01` Add a dedicated `--human-gate` mode for issue-starting workflows and validate it only for resolved agent `codex`.
- `SOL-02` Reuse the normal issue-start pipeline through prompt rendering, then branch only at the agent-launch stage into a Codex batch path.
- `SOL-03` Run Codex batch mode through `codex exec` with JSON event output, saved last-message output, explicit worktree cwd, and no ephemeral session semantics.
- `SOL-04` Persist one predictable per-run state directory under the prepared worktree and store at least `events.jsonl`, `last-message.txt`, and `thread-id`.
- `SOL-05` Parse the final status from the saved last-message artifact and recognize exactly two success-path statuses: `DONE` and `HUMAN_GATE`.
- `SOL-06` Treat `DONE` as terminal success without opening interactive Codex.
- `SOL-07` Treat `HUMAN_GATE` as an immediate transition into `codex resume --include-non-interactive "$thread_id"` and use exit code `2` only when the resume handoff was required but could not be opened.
- `SOL-08` Expose dedicated human-gate documentation through a help flag surface instead of a new positional subcommand, and mention the mode briefly in normal help.
- `SOL-09` Extend fake Codex coverage so tests can simulate JSON events, last-message statuses, missing thread ids, and resume failures deterministically.

## Requirement Mapping

| Requirement ID | Solution / architecture refs | Notes |
| --- | --- | --- |
| `REQ-01` | `SOL-01`, `CTR-01`, `DL-01` | Human-gate mode is intentionally Codex-only in this feature. |
| `REQ-02` | `SOL-02`, `CTR-02` | Preparation stages stay aligned with the normal issue path. |
| `REQ-03` | `SOL-03`, `CTR-03`, `DL-03` | Batch mode uses resumable `codex exec` plus machine-readable artifacts. |
| `REQ-04` | `SOL-04`, `CTR-03`, `DL-03` | Thread id and related artifacts share one predictable state directory. |
| `REQ-05` | `SOL-05`, `CTR-04`, `DL-04` | Final status parsing is strict and artifact-backed. |
| `REQ-06` | `SOL-06`, `CTR-04` | `DONE` is a pure non-interactive success path. |
| `REQ-07` | `SOL-07`, `CTR-05`, `DL-05` | Resume uses explicit thread id and documented exit code `2` on resume-open failure. |
| `REQ-08` | `SOL-08`, `CTR-06`, `DL-02` | Dedicated help is a first-class documentation surface. |
| `REQ-09` | `SOL-09` | Fake Codex and Bats own the regression net. |

## To-Be Flow

1. Parse CLI input and detect `--human-gate` or `--human-gate-help`.
2. If dedicated help was requested, print the dedicated human-gate help surface and exit `0`.
3. Resolve the normal issue-start context: repo, base branch, selected agent/model, prompt source, issue metadata, branch name, worktree path, optional `init.sh`, and rendered prompt.
4. If `--human-gate` is not enabled, preserve the existing agent-launch behavior unchanged.
5. If `--human-gate` is enabled, validate that the resolved agent is `codex`.
6. Create a per-run state directory under `<worktree>/.start-issue/runs/<timestamp>`.
7. Run `codex exec` in the prepared worktree with JSON event output redirected to `events.jsonl` and the final message saved to `last-message.txt`.
8. Parse `thread_id` from the JSON event stream, save it into `thread-id`, and fail clearly if a resumable session was not captured.
9. Parse the saved last-message file for the final status line.
10. If status is `DONE`, print success and exit `0`.
11. If status is `HUMAN_GATE`, open `codex resume --include-non-interactive "$thread_id"`.
12. If status is missing or unrecognized, or if resume handoff cannot be opened, exit with the documented error path and print the relevant artifact location or resume command.

## Contracts

| Contract ID | Related refs | Input / Output | Producer / Consumer | Notes |
| --- | --- | --- | --- | --- |
| `CTR-01` | `SOL-01`, `REQ-01` | Input: resolved agent plus `--human-gate`; output: codex-only validation result. | CLI/pipeline / human-gate workflow | Rejects non-Codex agents instead of silently degrading. |
| `CTR-02` | `SOL-02`, `REQ-02` | Input: ordinary issue-start state; output: prepared worktree and rendered prompt. | main pipeline / Codex batch launch | Human-gate mode branches late to minimize workflow drift. |
| `CTR-03` | `SOL-03`, `SOL-04`, `REQ-03`, `REQ-04` | Input: worktree path, prompt, optional model; output: `codex exec` command plus saved `events.jsonl`, `last-message.txt`, and `thread-id`. | Codex adapter / filesystem | State artifacts must be predictable and resumable. |
| `CTR-04` | `SOL-05`, `SOL-06`, `REQ-05`, `REQ-06` | Input: saved `last-message.txt`; output: recognized final status or explicit failure. | parser helper / pipeline | Status parsing is grounded in the saved artifact, not terminal text heuristics. |
| `CTR-05` | `SOL-07`, `REQ-07` | Input: captured `thread_id`; output: interactive resume command or exit code `2` with printed fallback command. | pipeline / user terminal | Uses `thread_id` as the primary identity surface. |
| `CTR-06` | `SOL-08`, `REQ-08` | Input: help invocation; output: normal-help mention or dedicated human-gate help page. | output layer / user | Dedicated help owns prompt contract, exit codes, and troubleshooting. |

## Target Architecture

### Architecture Invariants

- Human-gate mode is Codex-only in FT-014.
- The normal issue-start pipeline remains the source of truth through prompt rendering.
- Batch Codex runs remain resumable and therefore never use ephemeral execution.
- Resume always targets the captured `thread_id`.
- Final status is read from saved artifacts, not inferred from incidental stdout text.
- Ordinary `start-issue ISSUE` behavior remains isolated from human-gate-specific logic.

### Target Shape

| Layer / responsibility | To-be role | Boundary / non-owner | Related refs |
| --- | --- | --- | --- |
| CLI parsing | Recognize `--human-gate` and `--human-gate-help`, then normalize them into workflow state. | Does not own Codex command building or status parsing. | `SOL-01`, `SOL-08` |
| Main pipeline | Reuse ordinary issue preparation and branch late into the Codex batch workflow. | Does not reimplement prompt resolution or Codex event parsing. | `SOL-02` |
| Codex adapter / batch helpers | Build batch and resume commands, create state directories, and capture machine-readable artifacts. | Does not own repo/issue/base-branch planning. | `SOL-03`, `SOL-04`, `SOL-07` |
| Status parser | Interpret the saved last-message file and return `DONE`, `HUMAN_GATE`, or a clear failure. | Does not read live terminal output or invent fallback statuses. | `SOL-05`, `SOL-06` |
| Help/docs/tests | Document the prompt contract and verify all success/failure surfaces deterministically. | Do not define behavior that code cannot support. | `SOL-08`, `SOL-09` |

## Internal Flow

1. `SOL-01` CLI parsing records whether human-gate mode or dedicated help was requested.
2. `SOL-08` Dedicated help exits before issue validation or repo work.
3. `SOL-02` The normal issue pipeline prepares the worktree and rendered prompt.
4. `SOL-01` A Codex-only validation gate runs before agent launch.
5. `SOL-03` - `SOL-04` Codex batch helpers create a run directory and execute `codex exec` with saved artifacts.
6. `SOL-05` - `SOL-06` Status parsing decides between terminal success and resume handoff.
7. `SOL-07` Resume is attempted only for `HUMAN_GATE`, by explicit thread id.

## Accepted Local Decisions

- `SD-01` Use a flag-based dedicated help entrypoint (`--human-gate-help`) rather than adding a new positional subcommand surface for help only.
- `SD-02` Reject `--human-gate` for any resolved non-Codex agent in FT-014 instead of silently ignoring the flag.
- `SD-03` Keep human-gate state artifacts under the prepared worktree so the work product and its resumable session metadata stay co-located.
- `SD-04` Parse only the documented status lines `STATUS: DONE` and `STATUS: HUMAN_GATE`; any other terminal state is an explicit failure.
- `SD-05` Print the exact resume command and thread id when the resume handoff was required but could not be opened.

## Change Surface

| Ref | Surface | Type | Why it changes |
| --- | --- | --- | --- |
| `SOL-01`, `SOL-08` | `scripts/lib/start_issue/cli.sh`, `scripts/lib/start_issue/output.sh` | code | Add human-gate flags, Codex-only validation messaging, and dedicated help. |
| `SOL-02` - `SOL-07` | `scripts/lib/start_issue/pipeline.sh`, `scripts/lib/start_issue/agent.sh` | code | Branch the launch stage into a resumable Codex batch workflow and parse final status. |
| `SOL-04` - `SOL-07` | `scripts/lib/start_issue/worktree.sh` or a new helper module | code | Create predictable run-state directories and manage saved artifacts. |
| `SOL-09` | `test/start_issue.bats`, `test/helpers/fake-bin/codex` | test | Simulate JSON events, last-message statuses, resume commands, and failure cases deterministically. |
| `SOL-08` - `SOL-09` | `README.md`, `README.ru.md`, `doc/spec.md`, `memory-bank/features/FT-014/*` | doc | Document the mode and keep feature-flow artifacts consistent with user-facing help. |

## Failure Modes

- `FM-01` `--human-gate` silently falls back to the normal interactive launch, so unattended workflows never happen.
- `FM-02` The workflow captures a status but loses `thread_id`, making `HUMAN_GATE` impossible to resume correctly.
- `FM-03` The implementation resumes `--last` or another implicit session selector and opens the wrong Codex session.
- `FM-04` Status parsing relies on loose text matching and misclassifies arbitrary summaries as `DONE` or `HUMAN_GATE`.
- `FM-05` Dedicated help and actual exit behavior drift apart, leaving prompt authors with an incorrect contract.

## Rollout / Backout

- `RB-01` Roll out by adding the Codex-only flag, dedicated help, batch helpers, and tests together.
- `RB-02` Back out by removing the human-gate mode and restoring the previous interactive-only Codex launch path.

## ADR Dependencies

See [decision-log.md](decision-log.md).
