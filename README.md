# start-issue

[![CI](https://github.com/dapi/start-issue/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/dapi/start-issue/actions/workflows/ci.yml)

[Русская версия](README.ru.md)

Start work on a GitHub issue from the terminal.

`start-issue` fetches issue metadata with `gh`, creates a git worktree with a branch name based on the issue, optionally runs `init.sh`, renames the current zellij tab, and starts a configurable coding agent session.

## Install

```bash
make install
```

This installs `scripts/start-issue` to `~/.local/bin/start-issue`.

Make sure `~/.local/bin` is in your `PATH`.

## Usage

```bash
start-issue 123
start-issue https://github.com/owner/repo/issues/123
start-issue 123 --repo owner/repo --base develop
start-issue 123 --agent codex
start-issue 123 --agent kimi --prompt-file .start-issue/prompt.md
start-issue 123 --no-agent
start-issue 123 --dry-run
```

## Agents

Supported agent values are `claude`, `codex`, `kimi`, `pi`, and `none`.

Agent selection precedence:

1. CLI: `--agent codex`, `--no-agent`, or legacy `--no-claude`
2. Project config: `.start-issue/agent` in the git root
3. User config: `~/.config/start-issue/agent`
4. Environment: `START_ISSUE_AGENT`
5. Built-in default: `claude`

Claude remains the default for compatibility. `--no-claude` still works as an alias for `--no-agent`.

Related Claude Code marketplace workflows:

- [task-router](https://github.com/dapi/claude-code-marketplace/tree/master/task-router)
- [zellij-workflow](https://github.com/dapi/claude-code-marketplace/tree/master/zellij-workflow)

## Agent Examples

### Claude

Claude is the default agent, so these commands are equivalent:

```bash
start-issue 123
start-issue 123 --agent claude
```

By default, Claude receives the repository-native task-router command:

```text
/task-router:route-task {ISSUE_URL}
```

Use `--command` to keep the Claude slash-command style but change the command prefix:

```bash
start-issue 123 --agent claude --command "/debug"
```

Use project config when Claude should be the default for this repository:

```bash
mkdir -p .start-issue
echo claude > .start-issue/agent
start-issue 123
```

### Codex

Launch Codex for a single issue:

```bash
start-issue 123 --agent codex
```

The script creates the worktree, renders the portable prompt, and launches:

```bash
codex --cd "$WORKTREE_PATH" --dangerously-bypass-approvals-and-sandbox "$PROMPT"
```

Use Codex as your project default:

```bash
mkdir -p .start-issue
echo codex > .start-issue/agent
start-issue 123
```

Use a custom prompt file with Codex:

```bash
start-issue 123 --agent codex --prompt-file .start-issue/prompt.md
```

### Kimi

Launch Kimi for a single issue:

```bash
start-issue 123 --agent kimi
```

The script creates the worktree, renders the portable prompt, and launches:

```bash
kimi --work-dir "$WORKTREE_PATH" --yolo -p "$PROMPT"
```

Use Kimi from the environment without changing project files:

```bash
START_ISSUE_AGENT=kimi start-issue 123
```

Use an inline prompt for one run:

```bash
start-issue 123 --agent kimi --prompt "Implement {ISSUE_URL} in {WORKTREE_PATH}. Keep changes scoped and run tests."
```

### Pi

Launch Pi for a single issue:

```bash
start-issue 123 --agent pi
```

The script changes into the worktree and launches:

```bash
cd "$WORKTREE_PATH"
pi "$PROMPT"
```

Use Pi as your user default for all repositories:

```bash
mkdir -p ~/.config/start-issue
echo pi > ~/.config/start-issue/agent
start-issue 123
```

Preview exactly what would happen before launching Pi:

```bash
start-issue 123 --agent pi --dry-run
```

## Prompt Configuration

Claude uses the legacy plugin-native command by default:

```text
/task-router:route-task {ISSUE_URL}
```

Other agents use a portable prompt by default. You can override the launch prompt with:

1. CLI: `--prompt-file path/to/prompt.md` or `--prompt "..."`
2. Project config: `.start-issue/prompt.md`
3. User config: `~/.config/start-issue/prompt.md`
4. Environment: `START_ISSUE_PROMPT_FILE` or `START_ISSUE_PROMPT`
5. Built-in default

Prompt templates support:

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

Unknown placeholders are left unchanged.

## Worktree Directory

The environment variable for the default worktree parent directory is `START_ISSUE_WORKTREE_DIR`.

CLI `--worktree-dir` has the highest priority. If neither is set, `start-issue` uses `~/worktrees`.

## Requirements

- `bash`
- `git`
- `gh` CLI with authenticated GitHub session
- `jq`
- selected agent CLI unless `--agent none` is used

## Specification

The workflow diagram is in [docs/start-issue-workflow.md](docs/start-issue-workflow.md).

The script specification is in [docs/specs/start-issue-spec.md](docs/specs/start-issue-spec.md).

## License

MIT
