# start-issue

Start work on a GitHub issue from the terminal.

`start-issue` fetches issue metadata with `gh`, creates a git worktree with a branch name based on the issue, optionally runs `init.sh`, renames the current zellij tab, and starts an agent session.

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
start-issue 123 --no-claude
start-issue 123 --dry-run
```

## Requirements

- `bash`
- `git`
- `gh` CLI with authenticated GitHub session
- `jq`
- `claude` CLI for the current implementation

## Roadmap

The next step is making the launcher agent-agnostic so it can start `claude`, `codex`, `kimi`, or `pi` and use a configurable portable prompt.

Track that work in the repository issues.

## License

MIT
