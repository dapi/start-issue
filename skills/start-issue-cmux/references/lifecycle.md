# start-issue + cmux lifecycle

## Start

1. Confirm the issue URL/number, repository, agent, and whether focus is wanted.
2. From the repository root, run `scripts/start-in-cmux <issue> --agent <agent>`.
3. Capture the workspace reference printed by cmux and the worktree/branch printed by `start-issue`.
4. `start-issue` owns all GitHub, branch, worktree, init-hook, prompt, and agent operations.

## Resume

1. List workspaces and identify the one named for the issue.
2. Read its screen and inspect the expected worktree path.
3. Confirm the worktree branch matches the issue branch before interacting.
4. If an agent session is resumable, use the supported start-issue/agent resume path. Otherwise report the blocker; do not launch a duplicate agent automatically.

## Monitor

Use explicit workspace refs:

```bash
cmux read-screen --workspace workspace:16 --scrollback --lines 120
cmux sidebar-state --workspace workspace:16 --json
```

A terminal screen is evidence of what the agent displayed, not proof that GitHub or CI state has changed. Verify commits, PRs, and checks through their canonical tools.

## Finish

When the agent reports `STATUS: DONE`, summarize the exact artifacts it reported. Keep the workspace open unless the user asks to close it. A completed agent run does not authorize merge, release, deploy, or production actions.
