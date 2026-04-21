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
  assert_output_contains "claude --dangerously-skip-permissions"
  assert_output_contains "/task-router:route-task"
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
