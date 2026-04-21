# start-issue Command

Use `start-issue` to start work on a GitHub issue from the terminal.

```bash
start-issue 123
start-issue https://github.com/owner/repo/issues/123
start-issue 123 --agent codex
start-issue 123 --no-agent
```

The command fetches issue metadata, creates a branch and git worktree, optionally runs `init.sh`, renames the current zellij tab when the helper is available, and launches the selected coding agent.

## Agent Selection

Supported values:

- `claude`
- `codex`
- `kimi`
- `pi`
- `none`

Selection precedence:

1. `--agent`, `--no-agent`, or legacy `--no-claude`
2. `.start-issue/agent` in the git root
3. `~/.config/start-issue/agent`
4. `START_ISSUE_AGENT`
5. default `claude`

## Prompt Selection

Prompt precedence:

1. `--prompt-file` or `--prompt`
2. `.start-issue/prompt.md` in the git root
3. `~/.config/start-issue/prompt.md`
4. `START_ISSUE_PROMPT_FILE` or `START_ISSUE_PROMPT`
5. built-in default

Prompt variables:

```text
{ISSUE_URL}
{ISSUE_NUMBER}
{ISSUE_TITLE}
{ISSUE_BODY}
{ISSUE_LABELS}
{REPO}
{BRANCH_NAME}
{WORKTREE_PATH}
{BASE_BRANCH}
```

Claude keeps the legacy default prompt:

```text
/task-router:route-task {ISSUE_URL}
```

Other agents use a portable implementation prompt by default.

## Worktree Directory

Use `START_ISSUE_WORKTREE_DIR` for the default worktree parent directory.

`--worktree-dir` overrides `START_ISSUE_WORKTREE_DIR`. If neither is set, `start-issue` uses `~/worktrees`.

## Dry Run

Use `--dry-run` to inspect the selected agent, prompt source, worktree path, and launch command without creating the worktree or starting an agent.
