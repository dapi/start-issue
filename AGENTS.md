# Repository Instructions

## Memory Bank

This repository uses `memory-bank/` from `https://github.com/dapi/memory-bank/`.
Before non-trivial product, documentation, architecture, or feature work, read:

1. [memory-bank/README.md](memory-bank/README.md)
2. [memory-bank/product/context.md](memory-bank/product/context.md)
3. [memory-bank/domain/glossary.md](memory-bank/domain/glossary.md)
4. [memory-bank/engineering/testing-policy.md](memory-bank/engineering/testing-policy.md)
5. [memory-bank/ops/development.md](memory-bank/ops/development.md)

For new medium or large feature work, use
[memory-bank/flows/feature-flow.md](memory-bank/flows/feature-flow.md). New
feature packages should use `brief.md -> optional design.md ->
implementation-plan.md`. Existing legacy packages with `feature.md` and
`solution.md` may stay as-is until they are touched for real work.

## Local Checks

Run `make test` before handoff when changes affect code, tests, release logic,
or memory-bank navigation. It includes shell syntax checks, shellcheck,
memory-bank link audit, whitespace checks, and the Bats suite.

## Live E2E Through cmux

Run a live agent E2E only when explicitly requested. Use a **terminal tab in
the workspace that invoked the agent**, not a new cmux workspace and not a tab
in an unrelated project. Resolve that workspace with `cmux identify`, which
returns both the invocation `caller` and the global `focused` workspace. Use
only `caller.workspace_ref` and `caller.pane_ref`; do not use
`cmux current-workspace`, which reports global focus and can be unrelated.

```bash
caller_context="$(cmux identify)"
workspace="$(printf '%s' "$caller_context" | jq -r '.caller.workspace_ref // empty')"
pane="$(printf '%s' "$caller_context" | jq -r '.caller.pane_ref // empty')"
test -n "$workspace" && test -n "$pane"
```

If `caller` is `null`, do not guess from a workspace name or global focus; ask
for an explicit cmux workspace/pane target. With a resolved caller, create the
surface explicitly:

```bash
cmux new-surface --type terminal \
  --workspace workspace:<start-issue-workspace> \
  --pane pane:<start-issue-pane> \
  --working-directory /absolute/path/to/start-issue-worktree \
  --focus true
cmux rename-tab --workspace workspace:<start-issue-workspace> \
  --surface surface:<new-surface> 'human-gate E2E'
cmux send --workspace workspace:<start-issue-workspace> \
  --surface surface:<new-surface> \
  'START_ISSUE_E2E=1 START_ISSUE_E2E_KEEP=1 make e2e-human-gate\n'
```

Poll `cmux read-screen` until `PASS` or a terminal failure is visible. Report
the exact terminal status, thread id, and retained artifact path. Do not claim
the suite passed before the tab output contains its terminal result.
