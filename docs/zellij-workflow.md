# Zellij Workflow

`start-issue` keeps the zellij integration separate from agent selection.

After the issue is fetched successfully, the script looks for an executable helper next to `scripts/start-issue`:

```bash
zellij-rename-tab-to-issue-number "{ISSUE_NUMBER}"
```

If the helper exists, the current zellij tab is renamed to the issue number. If it does not exist, the workflow continues without error.

This behavior is unchanged for all agents:

```bash
start-issue 123 --agent claude
start-issue 123 --agent codex
start-issue 123 --agent kimi
start-issue 123 --agent pi
start-issue 123 --no-agent
```

`--dry-run` prints the helper command that would run, then prints the selected agent and launch command.
