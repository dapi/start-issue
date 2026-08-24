# Codex human-gate: purpose and permission modes

This guide explains how to choose the capability boundary for a Codex
human-gate run. For the complete CLI contract, see [spec.md](spec.md).

The examples assume that `codex` is already selected by project or user
configuration.

## Why use human-gate

A normal `start-issue` run hands control to an interactive agent session.
Human-gate instead runs Codex in batch mode so it can work through the issue
without the operator staying in the interactive UI. Codex then ends in one of
two explicit states:

- `STATUS: DONE` — the requested work is complete;
- `STATUS: HUMAN_GATE` — a real decision or missing permission requires the
  operator, so `start-issue` resumes the exact saved Codex thread.

This is useful for semi-autonomous issue work: implementation and tests can run
unattended, while destructive actions, missing credentials, and product
decisions still stop at a visible human gate.

Permission modes are needed because editing a worktree and delivering a pull
request cross different trust boundaries. A sandboxed Codex run can usually
change and test code, but may be unable to access GitHub, write Git metadata,
push, or create a PR. Giving every run unrestricted access would solve that
technical limitation by silently broadening privileges. `start-issue` therefore
keeps the safe boundary by default and requires a visible opt-in when the agent
must deliver the change end to end.

## Choose a mode

Use `restricted` for normal working-tree implementation:

```bash
start-issue 123 --human-gate
```

`restricted` is the default. Codex runs with the `workspace-write` sandbox. It
can edit and test files in the prepared worktree, but network access, Git
metadata writes, push, and pull-request delivery are not guaranteed.

Use `full-delivery` only when the run must also read GitHub context, commit,
push, and create or update a pull request:

```bash
start-issue 123 --human-gate \
  --human-gate-permissions full-delivery
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
start-issue 123 --human-gate \
  --human-gate-permissions full-delivery \
  --dry-run
```

The output should report:

```text
Human-gate permissions: full-delivery (CLI)
```

## Example: implement an issue and deliver its PR

Assume issue `123` is in the current repository and the authenticated GitHub
account has write access.

1. Inspect the planned run:

   ```bash
   start-issue 123 --human-gate \
     --human-gate-permissions full-delivery \
     --dry-run
   ```

2. Start the real run after reviewing the unsandboxed-execution warning:

   ```bash
   start-issue 123 --human-gate \
     --human-gate-permissions full-delivery
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
START_ISSUE_HUMAN_GATE_PERMISSIONS=full-delivery \
  start-issue 123 --human-gate
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
start-issue --human-gate-help
```
