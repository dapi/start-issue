# Codex batch mode: autonomous work with a human gate

This guide is for someone seeing `start-issue --batch` for the first time. It
explains what batch mode does, why a run may stop at `HUMAN_GATE`, how the
operator continues the same Codex session, and when the agent needs restricted
or full-delivery permissions. For the complete CLI contract, see
[spec.md](spec.md).

The examples assume that `codex` is already selected by project or user
configuration. The released `--human-gate` flag remains available as a
compatibility alias for `--batch`.

## What human-gate is

Human-gate is a checkpoint where autonomous Codex work is either declared
complete or handed to a person for a decision.

A normal `start-issue` run opens an interactive Codex session. With `--batch`,
Codex runs in batch mode and works through the issue without the
operator remaining in the interactive UI. Its final response must contain one
explicit result:

- `STATUS: DONE` — the task is complete and no human action is needed;
- `STATUS: HUMAN_GATE` — continuing requires a human decision.

Human-gate is not an approval prompt before every command and it is not a
sandbox mode. It defines how an autonomous run ends and when control returns to
the operator.

## Why use it

Without this protocol, an exit code alone cannot reliably distinguish finished
work from a run that stopped because of a question, missing access, or an
unresolved failure. Human-gate lets Codex implement and test unattended, report
unambiguous completion, stop before decisions it must not make, and resume the
same saved thread instead of starting over with a new agent.

## How it works

1. `start-issue` resolves the issue, creates or reuses the worktree, and renders
   the prompt as usual.
2. It runs `codex exec` in batch mode instead of opening an interactive session.
3. It saves the event stream, final message, and `thread_id`.
4. Codex performs the task and ends with exactly one status line:
   `STATUS: DONE` or `STATUS: HUMAN_GATE`.
5. `DONE` exits successfully. `HUMAN_GATE` opens the saved session with
   `codex resume --include-non-interactive <thread_id>`, preserving all context
   for the operator.
6. A Codex error or missing status produces an error while preserving the
   diagnostic files.

Codex decides whether to stop according to the prompt rules; `start-issue`
does not inspect every command. It reads the final status and performs the
corresponding handoff.

## When Codex should stop

`STATUS: HUMAN_GATE` is appropriate for destructive operations, missing
credentials, incompatible product choices, unresolved failures that cannot be
safely fixed within the issue, and production or security decisions. Routine
technical choices and recoverable errors should be handled by Codex without
stopping the run.

## How permissions relate to human-gate

These are separate controls:

- `--batch` enables the batch run and `DONE` / `HUMAN_GATE` handoff;
- `--batch-permissions` controls which technical operations Codex can
  perform during that run.

Editing a worktree and delivering a PR cross different trust boundaries. A
sandboxed run can usually change and test code but may not be able to access
GitHub, write Git metadata, push, or create a PR. The safe boundary therefore
remains the default, while end-to-end delivery requires a visible opt-in.

## Choose a mode

Use `restricted` for normal working-tree implementation:

```bash
start-issue 123 --batch
```

`restricted` is the default. Codex runs with the `workspace-write` sandbox. It
can edit and test files in the prepared worktree, but network access, Git
metadata writes, push, and pull-request delivery are not guaranteed.

Use `full-delivery` only when the run must also read GitHub context, commit,
push, and create or update a pull request:

```bash
start-issue 123 --batch \
  --batch-permissions full-delivery
```

This is an explicit opt-in to unsandboxed Codex execution. It grants technical
capability, not permission for destructive Git operations, production or
security changes, or unresolved product decisions. Codex must still return
`STATUS: HUMAN_GATE` when one of those decisions requires the operator.

## Full-delivery preflight

Before using `full-delivery`, check the selected GitHub account, remote, and
repository access:

```bash
gh auth status
git remote get-url origin
gh repo view --json nameWithOwner,viewerPermission
```

Run a dry-run to verify the resolved mode and launch command without creating
a worktree or starting Codex:

```bash
start-issue 123 --batch \
  --batch-permissions full-delivery \
  --dry-run
```

The output should report:

```text
Batch permissions: full-delivery (CLI)
```

## Example: implement an issue and deliver its PR

Assume issue `123` is in the current repository and the authenticated GitHub
account has write access.

1. Inspect the planned run:

   ```bash
   start-issue 123 --batch \
     --batch-permissions full-delivery \
     --dry-run
   ```

2. Start the real run after reviewing the unsandboxed-execution warning:

   ```bash
   start-issue 123 --batch \
     --batch-permissions full-delivery
   ```

3. Codex can now implement and test the issue, commit the result, push the
   issue branch, and create or update its pull request when the task and
   repository state allow it.

4. A final `STATUS: DONE` exits successfully. A final
   `STATUS: HUMAN_GATE` opens the exact saved Codex thread for the operator.

The CLI keeps diagnostic state under:

```text
<worktree>/.start-issue/runs/<timestamp>/events.jsonl
<worktree>/.start-issue/runs/<timestamp>/last-message.txt
<worktree>/.start-issue/runs/<timestamp>/thread-id
```

## One-command environment override

The environment variable is useful for one command or a controlled automation
wrapper:

```bash
START_ISSUE_BATCH_PERMISSIONS=full-delivery \
  start-issue 123 --batch
```

The CLI option has higher precedence than the environment variable. Avoid
putting `full-delivery` in a global shell profile: keeping the opt-in visible at
the command site makes the unsandboxed boundary easier to review.

## Troubleshooting

- If `restricted` cannot access GitHub or write Git metadata, finish delivery
  manually or rerun with an explicitly reviewed `full-delivery` selection.
- If full delivery cannot push or create a PR, recheck `gh auth status`, the
  selected account, `origin`, and `viewerPermission`.
- If status parsing fails, inspect `events.jsonl` and `last-message.txt` in the
  printed state directory.
- If automatic resume fails, use the saved thread id:

  ```bash
  codex resume --include-non-interactive <thread_id>
  ```

For the built-in reference, run:

```bash
start-issue --batch-help
```
