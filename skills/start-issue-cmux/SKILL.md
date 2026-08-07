---
name: start-issue-cmux
description: Launch and monitor start-issue work in a dedicated cmux workspace. Use when a user wants a GitHub issue started through start-issue in a separate cmux tab/workspace, resumed without duplicate worktrees or agents, or monitored through cmux status and screen output.
---

# Start Issue in cmux

Use this skill as a thin orchestration layer between `start-issue` and cmux. `start-issue` remains the source of truth for GitHub issue resolution, branch naming, worktree creation/reuse, init hooks, prompt rendering, and agent launch. cmux provides the terminal workspace and optional monitoring surfaces.

## Preconditions

- The user explicitly requested cmux, a separate tab/workspace, or resume/monitoring of a start-issue run.
- The target repository and issue URL/number are known.
- `cmux`, `start-issue`, and `git` are available.
- Use the repository's normal environment loader. If the repository has `.envrc`, launch `start-issue` through `direnv exec .`.

Do not create a worktree manually. Do not invoke an agent directly when the request is to start an issue: delegate that lifecycle to `start-issue`.

## Quick start

From the target repository:

```bash
<installed-skill-dir>/scripts/start-in-cmux \
  https://github.com/OWNER/REPO/issues/123 \
  --agent codex
```

The installed skill directory depends on the installer (commonly `~/.codex/skills/start-issue-cmux` or `~/.agents/skills/start-issue-cmux`). When working from a checkout of this repository, the equivalent path is `skills/start-issue-cmux/scripts/start-in-cmux`.

The launcher creates a workspace named `start-issue #123`, keeps the current workspace focused by default, and runs:

```bash
direnv exec . start-issue <issue> --agent codex
```

If `.envrc` is absent, it runs `start-issue` directly. Pass `--focus` only when the user explicitly wants cmux to switch focus.

## Workflow

1. Resolve the target repository and issue. Prefer the full GitHub issue URL when available.
2. Inspect the current caller context with `cmux identify --json` if already inside cmux. Do not infer a target workspace from global focus.
3. Run `scripts/start-in-cmux` with the explicit issue and agent. Let `start-issue` resolve or reuse the worktree.
4. Record the returned cmux workspace reference and the worktree path printed by `start-issue`.
5. For status checks, read the named workspace with `cmux read-screen --workspace <ref> --scrollback --lines <n>` and inspect `cmux sidebar-state --workspace <ref> --json` when available.
6. Treat the agent's own terminal output and `start-issue` state as authoritative. Do not send follow-up input unless the user names the workspace and asks for it.
7. On resume, locate the existing workspace/worktree first. Never create a second worktree or launch a second agent for the same issue without explicit user direction.

## Status handling

- Agent working: leave the workspace running; do not interrupt or send speculative input.
- `STATUS: WAIT`: report the expected external event and keep the workspace available for resume.
- `STATUS: HUMAN_GATE`: report the exact decision or approval required; do not send an answer on the user's behalf.
- `STATUS: DONE`: report the worktree, branch, commit/PR evidence shown by the agent, and remaining review steps.

Use cmux notifications or sidebar status only as attention cues. They do not replace the issue, agent, CI, or repository state.

## Layout guidance

Start with one terminal surface. Add panes only when useful and only in the named workspace:

- terminal: the `start-issue` agent;
- optional terminal: logs or a test watcher;
- optional browser: local/stage verification;
- optional Markdown viewer: plan or handoff notes.

Prefer additive pane creation with `--focus false`. Do not close or modify unrelated workspaces. For browser automation, load the official `cmux-browser` skill; for settings/layout customization, load `cmux-customization`.

## Safety boundaries

- No merge, release, deploy, production mutation, or external message is implied by this skill.
- Do not bypass `start-issue` with manual `git worktree add`, direct agent launch, or a second orchestration runtime.
- Do not put tokens, credentials, or rendered secret environment values in workspace descriptions, cmux actions, prompts, logs, or issue comments.
- Do not focus, send input to, move, or close another workspace unless the user explicitly identifies it.
- If cmux is unavailable, report the diagnostic blocker and offer the normal `start-issue` command; do not silently fall back when the user specifically requested a separate cmux workspace.

## References

- Read [lifecycle.md](references/lifecycle.md) for start/resume/status/finish behavior.
- Read [cmux-layouts.md](references/cmux-layouts.md) when the user requests helper panes, browser surfaces, or a repeatable project layout.
