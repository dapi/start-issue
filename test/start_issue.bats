#!/usr/bin/env bats

setup() {
  REPO_ROOT="$(cd "$BATS_TEST_DIRNAME/.." && pwd)"
  TEST_TMPDIR="${BATS_TEST_TMPDIR:-${BATS_TMPDIR:?}}"
  export REPO_ROOT
  export PATH="$REPO_ROOT/test/helpers/fake-bin:$PATH"
  export START_ISSUE_FAKE_ISSUE_JSON="$REPO_ROOT/test/fixtures/issue-1.json"

  export HOME="$TEST_TMPDIR/home"
  mkdir -p "$HOME"

  TEST_REPO="$TEST_TMPDIR/repo"
  mkdir -p "$TEST_REPO"
  git -C "$TEST_REPO" init -q -b master
  git -C "$TEST_REPO" config user.email "ci@example.invalid"
  git -C "$TEST_REPO" config user.name "CI"
  printf "fixture\n" > "$TEST_REPO/README.md"
  git -C "$TEST_REPO" add README.md
  git -C "$TEST_REPO" commit -q -m "Initial commit"
  git -C "$TEST_REPO" remote add origin git@github.com:owner/repo.git
  cd "$TEST_REPO"

  unset START_ISSUE_AGENT
  unset START_ISSUE_PROMPT
  unset START_ISSUE_PROMPT_FILE
  unset START_ISSUE_WORKTREE_DIR
  unset START_ISSUE_FAKE_BRANCH_NAME
  unset START_ISSUE_FAKE_AGENT_FAIL
}

run_start_issue() {
  run "$REPO_ROOT/scripts/start-issue" "$@"
}

assert_success() {
  [ "$status" -eq 0 ]
}

assert_failure() {
  [ "$status" -ne 0 ]
}

assert_output_contains() {
  [[ "$output" == *"$1"* ]]
}

install_fake_zellij_tab_status() {
  ZELLIJ_FAKE_BIN="$TEST_TMPDIR/zellij-bin"
  mkdir -p "$ZELLIJ_FAKE_BIN"
  {
    printf "%s\n" "#!/usr/bin/env bash"
    printf "%s\n" "set -euo pipefail"
    printf "%s\n" "printf '%s\\n' \"\$*\" > \"\${START_ISSUE_ZELLIJ_LOG:?}\""
  } > "$ZELLIJ_FAKE_BIN/zellij-tab-status"
  chmod +x "$ZELLIJ_FAKE_BIN/zellij-tab-status"
  export PATH="$ZELLIJ_FAKE_BIN:$PATH"
  export START_ISSUE_ZELLIJ_LOG="$TEST_TMPDIR/zellij-tab-status.log"
}

@test "default agent is claude and SSH origin remote is parsed" {
  run_start_issue 1 --dry-run --no-init

  assert_success
  assert_output_contains "Agent: claude (built-in default)"
  assert_output_contains "Fetching issue #1 from owner/repo"
  assert_output_contains "Prompt source: built-in Claude command"
  assert_output_contains "Default prompt files:"
  assert_output_contains "Project:"
  assert_output_contains ".start-issue/prompt.md"
  assert_output_contains "User: $HOME/.config/start-issue/prompt.md"
  assert_output_contains "claude --dangerously-skip-permissions"
  assert_output_contains "/task-router:route-task"
}

@test "missing issue prints selected default agent and prompt details" {
  run_start_issue

  assert_failure
  assert_output_contains "Usage: start-issue <issue-url-or-number> [options]"
  assert_output_contains "Current configuration:"
  assert_output_contains "Agent: claude (built-in default)"
  assert_output_contains "Prompt source: built-in Claude command"
  assert_output_contains "Prompt location: $REPO_ROOT/scripts/start-issue"
  assert_output_contains "Default prompt files:"
  assert_output_contains "Project:"
  assert_output_contains ".start-issue/prompt.md"
  assert_output_contains "User: $HOME/.config/start-issue/prompt.md"
  assert_output_contains "Prompt preview: /task-router:route-task {ISSUE_URL}"
  [[ "$output" != *"Fetching issue"* ]]
}

@test "missing issue prints project agent and prompt file location" {
  mkdir -p .start-issue
  printf "codex\n" > .start-issue/agent
  printf "Project prompt for {ISSUE_URL}\n" > .start-issue/prompt.md

  run_start_issue

  assert_failure
  assert_output_contains "Agent: codex ("
  assert_output_contains ".start-issue/agent)"
  assert_output_contains "Prompt source:"
  assert_output_contains ".start-issue/prompt.md"
  assert_output_contains "Prompt location:"
  assert_output_contains "Prompt preview: Project prompt for {ISSUE_URL}"
  [[ "$output" != *"Fetching issue"* ]]
}

@test "prompt improvement without issue prints explicit error" {
  run_start_issue --improve-prompt

  assert_failure
  assert_output_contains "--improve-prompt requires <issue-url-or-number>"
  assert_output_contains "Example: start-issue 123 --improve-prompt"
  [[ "$output" != *"Current configuration:"* ]]
  [[ "$output" != *"Fetching issue"* ]]
}

@test "full issue URL overrides detected repository" {
  run_start_issue https://github.com/other/project/issues/1 --dry-run --no-init --no-agent

  assert_success
  assert_output_contains "Fetching issue #1 from other/project"
  assert_output_contains "Selected agent: none (CLI)"
}

@test "HTTPS origin remote is parsed" {
  git remote set-url origin https://github.com/https-owner/https-repo.git

  run_start_issue 1 --dry-run --no-init --no-agent

  assert_success
  assert_output_contains "Fetching issue #1 from https-owner/https-repo"
}

@test "CLI agent wins over project config and environment" {
  mkdir -p .start-issue
  printf "kimi\n" > .start-issue/agent
  export START_ISSUE_AGENT=pi

  run_start_issue 1 --agent codex --dry-run --no-init

  assert_success
  assert_output_contains "Agent: codex (CLI)"
  assert_output_contains "codex --cd"
}

@test "project agent config wins over environment" {
  mkdir -p .start-issue
  printf "codex\n" > .start-issue/agent
  export START_ISSUE_AGENT=kimi

  run_start_issue 1 --dry-run --no-init

  assert_success
  assert_output_contains "Agent: codex ("
  assert_output_contains ".start-issue/agent)"
}

@test "user agent config wins over environment" {
  mkdir -p "$HOME/.config/start-issue"
  printf "pi\n" > "$HOME/.config/start-issue/agent"
  export START_ISSUE_AGENT=kimi

  run_start_issue 1 --dry-run --no-init

  assert_success
  assert_output_contains "Agent: pi ($HOME/.config/start-issue/agent)"
  assert_output_contains "cd $HOME/worktrees/feature/issue-1-add-login-button && pi"
}

@test "init writes project config with selected agent and portable prompt" {
  run_start_issue init --project --agent codex

  assert_success
  assert_output_contains "Scope: project config"
  assert_output_contains "Agent: codex (CLI)"
  [[ "$(cat .start-issue/agent)" == "codex" ]]
  [[ "$(cat .start-issue/prompt.md)" == *"Implement GitHub issue {ISSUE_URL} in this worktree."* ]]
  [[ "$output" != *"Fetching issue"* ]]
}

@test "init prompts for user config when scope is omitted" {
  run bash -c 'printf "2\n" | "$REPO_ROOT/scripts/start-issue" init'

  assert_success
  assert_output_contains "User config"
  [[ "$(cat "$HOME/.config/start-issue/agent")" == "claude" ]]
  [[ "$(cat "$HOME/.config/start-issue/prompt.md")" == "/task-router:route-task {ISSUE_URL}" ]]
}

@test "init keeps existing config unless forced" {
  mkdir -p .start-issue
  printf "kimi\n" > .start-issue/agent
  printf "custom\n" > .start-issue/prompt.md

  run_start_issue init --project --agent codex --prompt inline
  assert_success
  [[ "$(cat .start-issue/agent)" == "kimi" ]]
  [[ "$(cat .start-issue/prompt.md)" == "custom" ]]

  run_start_issue init --project --agent codex --prompt inline --force
  assert_success
  [[ "$(cat .start-issue/agent)" == "codex" ]]
  [[ "$(cat .start-issue/prompt.md)" == "inline" ]]
}

@test "init derives missing prompt from kept existing agent" {
  mkdir -p .start-issue
  printf "codex\n" > .start-issue/agent

  run_start_issue init --project

  assert_success
  assert_output_contains "Agent: codex ("
  assert_output_contains ".start-issue/agent (existing)"
  [[ "$(cat .start-issue/agent)" == "codex" ]]
  [[ "$(cat .start-issue/prompt.md)" == *"Implement GitHub issue {ISSUE_URL} in this worktree."* ]]
  [[ "$(cat .start-issue/prompt.md)" != *"/task-router:route-task"* ]]
}

@test "init dry-run does not write config files" {
  run_start_issue init --project --agent codex --dry-run

  assert_success
  assert_output_contains "[DRY-RUN] Would write agent config"
  [[ ! -e .start-issue/agent ]]
  [[ ! -e .start-issue/prompt.md ]]
}

@test "--no-agent prints manual next steps" {
  run_start_issue 1 --no-agent --dry-run --no-init

  assert_success
  assert_output_contains "Selected agent: none (CLI)"
  assert_output_contains "To start working:"
  assert_output_contains "codex --cd $HOME/worktrees/feature/issue-1-add-login-button"
}

@test "zellij-tab-status dry-run rename is shown when installed" {
  install_fake_zellij_tab_status

  run_start_issue 1 --agent none --dry-run --no-init

  assert_success
  assert_output_contains "Would run: zellij-tab-status --set-name \\#1"
}

@test "worktree directory priority uses environment and CLI override" {
  export START_ISSUE_WORKTREE_DIR="$TEST_TMPDIR/env-worktrees"

  run_start_issue 1 --agent none --dry-run --no-init
  assert_success
  assert_output_contains "Worktree directory: $TEST_TMPDIR/env-worktrees (START_ISSUE_WORKTREE_DIR)"

  run_start_issue 1 --agent none --dry-run --no-init --worktree-dir "$TEST_TMPDIR/cli-worktrees"
  assert_success
  assert_output_contains "Worktree directory: $TEST_TMPDIR/cli-worktrees (CLI)"
}

@test "legacy Claude worktree environment name is ignored" {
  legacy_name="CLAUDE""_WORKTREE_DIR"
  export "$legacy_name=$TEST_TMPDIR/legacy-worktrees"

  run_start_issue 1 --agent none --dry-run --no-init

  assert_success
  assert_output_contains "Worktree directory: $HOME/worktrees (built-in default)"
}

@test "codex, kimi, and pi launch commands are rendered in dry-run" {
  run_start_issue 1 --agent codex --dry-run --no-init
  assert_success
  assert_output_contains "codex --cd $HOME/worktrees/feature/issue-1-add-login-button"
  assert_output_contains "--dangerously-bypass-approvals-and-sandbox"

  run_start_issue 1 --agent kimi --dry-run --no-init
  assert_success
  assert_output_contains "kimi --work-dir $HOME/worktrees/feature/issue-1-add-login-button --yolo -p"

  run_start_issue 1 --agent pi --dry-run --no-init
  assert_success
  assert_output_contains "cd $HOME/worktrees/feature/issue-1-add-login-button && pi"
}

@test "prompt template from project file is rendered" {
  mkdir -p .start-issue
  printf "Prompt-{ISSUE_NUMBER}-{REPO}-{BASE_BRANCH}-{UNKNOWN}\n" > .start-issue/prompt.md

  run_start_issue 1 --agent codex --dry-run --no-init

  assert_success
  assert_output_contains "Prompt source:"
  assert_output_contains ".start-issue/prompt.md"
  assert_output_contains "Prompt-1-owner/repo-master"
  assert_output_contains "UNKNOWN"
}

@test "environment prompt file overrides user and project prompt files" {
  mkdir -p .start-issue "$HOME/.config/start-issue"
  printf "Project prompt {ISSUE_NUMBER}\n" > .start-issue/prompt.md
  printf "User prompt {ISSUE_NUMBER}\n" > "$HOME/.config/start-issue/prompt.md"
  printf "Env prompt {ISSUE_NUMBER}\n" > "$TEST_TMPDIR/env-prompt.md"
  export START_ISSUE_PROMPT_FILE="$TEST_TMPDIR/env-prompt.md"

  run_start_issue 1 --agent codex --dry-run --no-init

  assert_success
  assert_output_contains "Prompt source: START_ISSUE_PROMPT_FILE: $TEST_TMPDIR/env-prompt.md"
  assert_output_contains "Env\\ prompt\\ 1"
  [[ "$output" != *"Project prompt 1"* ]]
  [[ "$output" != *"User prompt 1"* ]]
}

@test "prompt improvement writes proposal next to project prompt and exits before worktree" {
  mkdir -p .start-issue
  printf "Prompt {ISSUE_NUMBER}\n" > .start-issue/prompt.md
  export START_ISSUE_FAKE_IMPROVED_PROMPT="Improved prompt {ISSUE_URL} {ISSUE_NUMBER}"

  run_start_issue 1 --agent codex --improve-prompt --no-init

  assert_success
  assert_output_contains "Improving prompt template"
  assert_output_contains "Prompt source:"
  assert_output_contains ".start-issue/prompt.md"
  assert_output_contains "Proposal path:"
  assert_output_contains ".start-issue/prompt.improved.md"
  assert_output_contains "Prompt improvement written"
  [[ "$(cat .start-issue/prompt.improved.md)" == "Improved prompt {ISSUE_URL} {ISSUE_NUMBER}" ]]
  [[ "$output" != *"Creating worktree"* ]]
}

@test "built-in prompt improvement writes project proposal by default" {
  export START_ISSUE_FAKE_IMPROVED_PROMPT="Improved built-in prompt {ISSUE_URL}"

  run_start_issue 1 --agent codex --improve-prompt --no-init

  assert_success
  assert_output_contains "Prompt source: built-in portable prompt"
  assert_output_contains "Proposal path:"
  assert_output_contains ".start-issue/prompt.improved.md"
  [[ "$(cat .start-issue/prompt.improved.md)" == "Improved built-in prompt {ISSUE_URL}" ]]
}

@test "prompt improvement dry-run does not write proposal or call agent" {
  mkdir -p .start-issue
  printf "Prompt {ISSUE_NUMBER}\n" > .start-issue/prompt.md
  export START_ISSUE_FAKE_AGENT_FAIL=1

  run_start_issue 1 --agent codex --improve-prompt --dry-run --no-init

  assert_success
  assert_output_contains "[DRY-RUN] Would ask codex to generate an improved prompt proposal."
  [[ ! -e .start-issue/prompt.improved.md ]]
}

@test "prompt improvement uses custom output path and refuses overwrite" {
  mkdir -p .start-issue
  printf "Prompt {ISSUE_NUMBER}\n" > .start-issue/prompt.md
  export START_ISSUE_FAKE_IMPROVED_PROMPT="Improved custom prompt"

  run_start_issue 1 --agent codex --improve-prompt --prompt-output-file .start-issue/prompt.next.md --no-init

  assert_success
  assert_output_contains "Proposal path: .start-issue/prompt.next.md"
  [[ "$(cat .start-issue/prompt.next.md)" == "Improved custom prompt" ]]

  run_start_issue 1 --agent codex --improve-prompt --prompt-output-file .start-issue/prompt.next.md --no-init

  assert_failure
  assert_output_contains "Prompt improvement output already exists: .start-issue/prompt.next.md"
  [[ "$(cat .start-issue/prompt.next.md)" == "Improved custom prompt" ]]
}

@test "prompt improvement rejects agent none before fetching issue" {
  run_start_issue 1 --agent none --improve-prompt --dry-run --no-init

  assert_failure
  assert_output_contains "--improve-prompt requires an agent"
  [[ "$output" != *"Fetching issue"* ]]
}

@test "prompt conflict fails fast" {
  run_start_issue 1 --agent none --dry-run --prompt inline --prompt-file prompt.md

  assert_failure
  assert_output_contains "Use either --prompt-file or --prompt, not both."
  [[ "$output" != *"Fetching issue"* ]]
}

@test "unknown agent fails fast" {
  run_start_issue 1 --agent unknown --dry-run

  assert_failure
  assert_output_contains "Unknown agent: unknown"
  [[ "$output" != *"Fetching issue"* ]]
}

@test "AI branch naming accepts selected agent output" {
  export START_ISSUE_FAKE_BRANCH_NAME=fix/issue-1-ai-generated-name

  run_start_issue 1 --agent codex --ai --dry-run --no-init

  assert_success
  assert_output_contains "Branch: fix/issue-1-ai-generated-name"
  assert_output_contains "ai:codex"
}

@test "AI branch naming falls back when selected agent returns invalid output" {
  export START_ISSUE_FAKE_BRANCH_NAME="not a branch"

  run_start_issue 1 --agent codex --ai --dry-run --no-init

  assert_success
  assert_output_contains "Generated branch name doesn't match expected format"
  assert_output_contains "Using fallback: feature/issue-1-add-login-button"
}

@test "make test target exists and runs local verification stack" {
  run make -n -f "$REPO_ROOT/Makefile" test

  assert_success
  assert_output_contains "bash -n scripts/start-issue"
  assert_output_contains "shellcheck scripts/start-issue"
  assert_output_contains "git diff --check"
  assert_output_contains "bats test"
}

@test "branch reuse matches the exact worktree path" {
  git worktree add "$TEST_TMPDIR/worktree-v2" -b feature/issue-1-add-login-button-v2 master
  git worktree add "$TEST_TMPDIR/worktree-exact" -b feature/issue-1-add-login-button master
  expected_worktree="$(cd "$TEST_TMPDIR/worktree-exact" && pwd -P)"

  run bash -c 'printf "1\n" | "$REPO_ROOT/scripts/start-issue" 1 --agent none --dry-run --no-init'

  assert_success
  assert_output_contains "Existing worktree: $expected_worktree"
  assert_output_contains "Using existing worktree: $expected_worktree"
  assert_output_contains "✅ Worktree ready at: $expected_worktree"
  [[ "$output" != *"worktree-v2"* ]]
}

@test "reusing a plain directory is rejected before init or agent launch" {
  mkdir -p "$HOME/worktrees/feature/issue-1-add-login-button"

  run bash -c 'printf "1\n" | "$REPO_ROOT/scripts/start-issue" 1 --agent none --no-init'

  assert_failure
  assert_output_contains "path exists but is not a git worktree for this repository"
  [[ "$output" != *"Worktree ready"* ]]
}

@test "reusing a path registered to a different branch is rejected" {
  git worktree add "$HOME/worktrees/feature/issue-1-add-login-button" -b chore/other-branch master

  run bash -c 'printf "1\n" | "$REPO_ROOT/scripts/start-issue" 1 --agent none --no-init'

  assert_failure
  assert_output_contains "Registered branch: chore/other-branch"
  assert_output_contains "Cannot reuse worktree path '$HOME/worktrees/feature/issue-1-add-login-button': it belongs to branch 'chore/other-branch', not 'feature/issue-1-add-login-button'."
  [[ "$output" != *"Worktree ready"* ]]
}

@test "delete and recreate flow replaces the conflicting branch worktree" {
  git worktree add "$TEST_TMPDIR/existing-worktree" -b feature/issue-1-add-login-button master

  run bash -c 'printf "3\n" | "$REPO_ROOT/scripts/start-issue" 1 --agent none --no-init'

  assert_success
  assert_output_contains "Removing existing branch/worktree"
  assert_output_contains "✅ Cleaned up"
  assert_output_contains "✅ Worktree created"
  [[ -d "$HOME/worktrees/feature/issue-1-add-login-button" ]]
  [[ ! -d "$TEST_TMPDIR/existing-worktree" ]]
}

@test "flat worktree mode uses flattened path" {
  run_start_issue 1 --agent none --dry-run --no-init --flat

  assert_success
  assert_output_contains "Path: $HOME/worktrees/feature-issue-1-add-login-button"
  assert_output_contains "Would run: git worktree add -b feature/issue-1-add-login-button $HOME/worktrees/feature-issue-1-add-login-button master"
}

@test "base branch falls back to the current branch when origin HEAD is missing" {
  git checkout -q -b develop

  run_start_issue 1 --agent none --dry-run --no-init

  assert_success
  assert_output_contains "Could not detect default branch, using current: develop"
  assert_output_contains "Base: develop"
  assert_output_contains "Would run: git worktree add -b feature/issue-1-add-login-button $HOME/worktrees/feature/issue-1-add-login-button develop"
}
