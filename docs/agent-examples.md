# Agent Examples

## Claude

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

## Codex

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

## Kimi

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

## Pi

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
