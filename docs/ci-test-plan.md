# CI Test Plan for start-issue

## Goal

Add CI coverage that proves `scripts/start-issue` keeps working across agent selection, prompt configuration, worktree directory resolution, dry-run output, and failure paths without launching real agents or mutating a developer machine.

The first CI version should focus on deterministic shell tests with fake `git`, `gh`, and agent CLIs where needed. Real GitHub API calls and real agent sessions should stay out of CI.

## Quality Gates

Run on every pull request and push to the main branch:

- `bash -n scripts/start-issue`
- `shellcheck scripts/start-issue`
- shell integration tests for `--dry-run`
- documentation drift checks for removed legacy Claude-specific worktree environment names

## Recommended Workflow

Create `.github/workflows/ci.yml` with a Linux matrix:

```yaml
name: CI

on:
  pull_request:
  push:
    branches: [master, main]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        bash: ["system"]
    steps:
      - uses: actions/checkout@v4
      - name: Install test dependencies
        run: sudo apt-get update && sudo apt-get install -y shellcheck bats
      - name: Static checks
        run: |
          bash -n scripts/start-issue
          shellcheck scripts/start-issue
          git diff --check
      - name: Integration tests
        run: bats test
```

The matrix starts small. Add macOS after the fake-command harness is stable, because `sed`, `readlink`, and shell behavior can differ.

## Test Harness

Use Bats for shell integration tests:

```text
test/
  helpers/
    fake-bin/
      gh
      claude
      codex
      kimi
      pi
  fixtures/
    issue-1.json
  start_issue.bats
```

The tests should prepend `test/helpers/fake-bin` to `PATH`.

Fake `gh` should support:

- `gh auth status`
- `gh api repos/{owner}/{repo}/issues/{number}`

The fake `gh api` command should return fixture JSON with title, body, labels, and issue metadata.

Fake agents should only record their argv and cwd to files in `$BATS_TEST_TMPDIR`. They must not call external services.

Each test should create a temporary git repo with:

- an `origin` remote such as `git@github.com:owner/repo.git`
- a base branch
- enough refs for `git worktree add` dry-run paths, or use `--dry-run` for most scenarios

Use `--dry-run --no-init` as the default test mode. Add a small number of non-dry-run tests only after the worktree setup is reliable.

## Initial Test Scenarios

### Static and Documentation

- Script passes `bash -n`.
- Script passes `shellcheck`.
- Repository has no trailing whitespace via `git diff --check`.
- removed legacy Claude-specific worktree environment names do not appear in `scripts/`, `README.md`, `docs/`, or `github-workflow/`.

### Issue and Repository Parsing

- Issue number input uses repo from `origin`.
- Full issue URL overrides repo and issue number.
- SSH remote `git@github.com:owner/repo.git` parses correctly.
- HTTPS remote `https://github.com/owner/repo.git` parses correctly.
- Missing `origin` fails with a clear message unless `--repo` is provided.

### Agent Selection

- Default agent is `claude`.
- CLI `--agent codex` wins over env and config files.
- `--no-agent` selects `none`.
- Legacy `--no-claude` selects `none`.
- Project `.start-issue/agent` wins over user config and env.
- User `~/.config/start-issue/agent` wins over env.
- `START_ISSUE_AGENT` works when no higher-priority source exists.
- Unknown agent fails before issue mutation or worktree creation.

### Worktree Directory

- Default is `~/worktrees`.
- `START_ISSUE_WORKTREE_DIR` overrides default.
- CLI `--worktree-dir` overrides `START_ISSUE_WORKTREE_DIR`.
- legacy Claude-specific worktree environment names are ignored.
- `--flat` replaces `/` in the branch path with `-`.

### Prompt Selection and Rendering

- Default Claude prompt renders `/task-router:route-task {ISSUE_URL}`.
- Non-Claude agents use the portable prompt.
- CLI `--prompt` wins over all other prompt sources.
- CLI `--prompt-file` wins over project/user/env prompt sources.
- Project `.start-issue/prompt.md` wins over user/env prompt sources.
- User `~/.config/start-issue/prompt.md` wins over env prompt sources.
- `START_ISSUE_PROMPT_FILE` and `START_ISSUE_PROMPT` work when active.
- Providing both `--prompt-file` and `--prompt` fails fast.
- Providing both env prompt forms fails fast when env prompt is active.
- Known placeholders render:
  - `{ISSUE_URL}`
  - `{ISSUE_NUMBER}`
  - `{ISSUE_TITLE}`
  - `{ISSUE_BODY}`
  - `{ISSUE_LABELS}`
  - `{REPO}`
  - `{BRANCH_NAME}`
  - `{WORKTREE_PATH}`
  - `{BASE_BRANCH}`
- Unknown placeholders remain unchanged.

### Launch Commands

In `--dry-run`, assert the rendered command for each agent:

- `claude`: `cd "$WORKTREE_PATH" && claude --dangerously-skip-permissions "$PROMPT"`
- `codex`: `codex --cd "$WORKTREE_PATH" --dangerously-bypass-approvals-and-sandbox "$PROMPT"`
- `kimi`: `kimi --work-dir "$WORKTREE_PATH" --yolo -p "$PROMPT"`
- `pi`: `cd "$WORKTREE_PATH" && pi "$PROMPT"`
- `none`: prints manual next steps and does not build an agent launch command

### Branch Naming

- Fast branch naming maps labels to branch prefixes:
  - `bug` -> `fix/`
  - `hotfix` -> `hotfix/`
  - `docs` -> `docs/`
  - `refactor` -> `refactor/`
  - `test` -> `test/`
  - `chore` -> `chore/`
  - no matching label -> `feature/`
- Branch title slug is lowercase kebab case.
- `--ai` uses the selected fake agent command and accepts a valid branch name.
- `--ai` falls back to fast naming when the fake agent fails.
- `--ai` falls back to fast naming when the fake agent returns an invalid branch name.

### Initialization and Zellij

- `--no-init` skips `init.sh`.
- Missing `init.sh` is a warning, not a failure.
- Existing `init.sh` is shown in dry-run.
- Non-zero `init.sh` is a warning in non-dry-run smoke tests.
- Missing `zellij-rename-tab-to-issue-number` is ignored.
- Existing zellij helper is called with the issue number.

## Implementation Phases

1. Add `shellcheck` and `bash -n` CI only.
2. Add Bats with fake `gh` and dry-run tests for agent/worktree/prompt precedence.
3. Add fake agent launch tests for all supported agents.
4. Add branch naming and failure-path tests.
5. Add limited non-dry-run worktree tests using temporary repositories.
6. Add macOS CI once Linux tests are stable.

## Open Decisions

- Whether to vendor Bats through a GitHub Action or install it through apt/Homebrew.
- Whether to keep all tests in Bats or use a small Python harness for easier fixture management.
- Whether CI should test real `gh` behavior in a scheduled job against a public fixture issue. The default answer should be no until there is a clear need.
