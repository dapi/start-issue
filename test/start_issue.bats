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
  unset START_ISSUE_MODEL
  unset START_ISSUE_PROMPT
  unset START_ISSUE_PROMPT_FILE
  unset START_ISSUE_WORKTREE_DIR
  unset START_ISSUE_FAKE_BRANCH_NAME
  unset START_ISSUE_FAKE_AGENT_FAIL
  unset START_ISSUE_FAKE_EXPECT_MODEL
  unset START_ISSUE_FAKE_FORBID_MODEL
  unset START_ISSUE_FAKE_LATEST_RELEASE_JSON
  unset START_ISSUE_REPOSITORY
}

run_start_issue() {
  run "$REPO_ROOT/scripts/start-issue" "$@"
}

run_install_script() {
  run env \
    HOME="$HOME" \
    PREFIX="$TEST_TMPDIR/install-prefix" \
    BINDIR="$TEST_TMPDIR/install-prefix/bin" \
    TARGET="$TEST_TMPDIR/install-prefix/bin/start-issue" \
    START_ISSUE_REPOSITORY="test/local" \
    START_ISSUE_ASSET_URL="file://$TEST_TMPDIR/install-fixture/start-issue" \
    START_ISSUE_CHECKSUM_URL="file://$TEST_TMPDIR/install-fixture/start-issue.sha256" \
    bash "$REPO_ROOT/install.sh" "$@"
}

build_installed_start_issue() {
  local output_path="$1"
  local version="$2"
  local tmpfile

  bash "$REPO_ROOT/scripts/build-start-issue" "$output_path" >/dev/null
  tmpfile="$TEST_TMPDIR/versioned-start-issue"
  awk -v version="$version" '
    BEGIN { replaced = 0 }
    /^VERSION="/ && replaced == 0 {
      print "VERSION=\"" version "\""
      replaced = 1
      next
    }
    { print }
  ' "$output_path" > "$tmpfile"
  mv "$tmpfile" "$output_path"
  chmod +x "$output_path"
}

create_fake_release_assets() {
  local dir="$1"
  local version="$2"
  local asset_path="$dir/start-issue"
  local checksum_path="$dir/start-issue.sha256"
  local json_path="$dir/release.json"
  local asset_url
  local checksum_url
  local checksum

  mkdir -p "$dir"
  cat > "$asset_path" <<EOF
#!/usr/bin/env bash
set -euo pipefail
if [[ "\${1:-}" == "--version" || "\${1:-}" == "-v" ]]; then
  echo "start-issue v$version"
  exit 0
fi
echo "updated test binary"
EOF
  chmod +x "$asset_path"
  checksum="$(shasum -a 256 "$asset_path" | awk '{ print $1 }')"
  printf "%s  start-issue\n" "$checksum" > "$checksum_path"
  asset_url="file://$asset_path"
  checksum_url="file://$checksum_path"
  cat > "$json_path" <<EOF
{"tag_name":"v$version","assets":[
  {"name":"start-issue","browser_download_url":"$asset_url"},
  {"name":"start-issue.sha256","browser_download_url":"$checksum_url"}
]}
EOF
  export START_ISSUE_FAKE_LATEST_RELEASE_JSON="$json_path"
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
  assert_output_contains "Agent: claude"
  assert_output_contains "Agent source: built-in default"
  assert_output_contains "Model: <unset>"
  assert_output_contains "Model source: built-in default"
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
  assert_output_contains "Error: missing issue URL or issue number"
  assert_output_contains 'Run `start-issue --help` for full usage and prompt variables.'
  assert_output_contains "Usage: start-issue <issue-url-or-number> [options]"
  assert_output_contains "Examples:"
  assert_output_contains "Prompt variables:"
  assert_output_contains "Current configuration:"
  assert_output_contains "Agent: claude"
  assert_output_contains "Agent source: built-in default"
  assert_output_contains "Model: <unset>"
  assert_output_contains "Model source: built-in default"
  assert_output_contains "Prompt source: built-in Claude command"
  assert_output_contains "Prompt location: $REPO_ROOT/scripts/start-issue"
  assert_output_contains "Project model:"
  assert_output_contains ".start-issue/model"
  assert_output_contains "User model:"
  assert_output_contains "Worktree dir: $HOME/worktrees (built-in default)"
  assert_output_contains "Default prompt files:"
  assert_output_contains "Project:"
  assert_output_contains ".start-issue/prompt.md"
  assert_output_contains "User: $HOME/.config/start-issue/prompt.md"
  [[ "$output" != *"Prompt preview:"* ]]
  [[ "$output" != *"Options:"* ]]
  [[ "$output" != *"Agent selection precedence:"* ]]
  [[ "$output" != *"Fetching issue"* ]]
}

@test "help lists environment variables for prompt and config sources" {
  run_start_issue --help

  assert_success
  assert_output_contains "Environment variables:"
  assert_output_contains "START_ISSUE_AGENT"
  assert_output_contains "START_ISSUE_PROMPT"
  assert_output_contains "START_ISSUE_PROMPT_FILE"
  assert_output_contains "START_ISSUE_WORKTREE_DIR"
  assert_output_contains "START_ISSUE_DUMP_PROMPT"
}

@test "help documents update entry points" {
  run_start_issue --help

  assert_success
  assert_output_contains "start-issue update [options]"
  assert_output_contains "--update"
  assert_output_contains "start-issue update"
  assert_output_contains "start-issue --update"
}

@test "missing issue prints project agent and prompt file location" {
  mkdir -p .start-issue
  printf "codex\n" > .start-issue/agent
  printf "Project prompt for {ISSUE_URL}\n" > .start-issue/prompt.md

  run_start_issue

  assert_failure
  assert_output_contains "Error: missing issue URL or issue number"
  assert_output_contains "Agent: codex"
  assert_output_contains "Agent source: "
  assert_output_contains ".start-issue/agent"
  assert_output_contains "Prompt source:"
  assert_output_contains ".start-issue/prompt.md"
  assert_output_contains "Prompt location:"
  [[ "$output" != *"Prompt preview:"* ]]
  [[ "$output" != *"Fetching issue"* ]]
}

@test "help documents model option and config surfaces" {
  run_start_issue --help

  assert_success
  assert_output_contains "--model <name>"
  assert_output_contains "START_ISSUE_MODEL"
  assert_output_contains ".start-issue/model"
  assert_output_contains "~/.config/start-issue/model"
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

@test "update subcommand installs the latest release into the running executable path" {
  installed_script="$TEST_TMPDIR/bin/start-issue"
  mkdir -p "$(dirname "$installed_script")"
  build_installed_start_issue "$installed_script" "1.11.0"
  create_fake_release_assets "$TEST_TMPDIR/release-assets" "1.11.1"
  expected_path="$(cd "$(dirname "$installed_script")" && pwd -P)/$(basename "$installed_script")"

  run "$installed_script" update

  assert_success
  assert_output_contains "Installed version: v1.11.0"
  assert_output_contains "Latest release: v1.11.1"
  assert_output_contains "Updated start-issue at: $expected_path"
  assert_output_contains "Version: start-issue v1.11.1"
  run "$installed_script" --version
  assert_success
  assert_output_contains "start-issue v1.11.1"
}

@test "--update is equivalent to the update subcommand" {
  installed_script="$TEST_TMPDIR/bin/start-issue"
  mkdir -p "$(dirname "$installed_script")"
  build_installed_start_issue "$installed_script" "1.11.0"
  create_fake_release_assets "$TEST_TMPDIR/release-assets-flag" "1.11.1"
  expected_path="$(cd "$(dirname "$installed_script")" && pwd -P)/$(basename "$installed_script")"

  run "$installed_script" --update

  assert_success
  assert_output_contains "Latest release: v1.11.1"
  assert_output_contains "Updated start-issue at: $expected_path"
}

@test "update exits successfully when already on the latest release tag" {
  installed_script="$TEST_TMPDIR/bin/start-issue"
  mkdir -p "$(dirname "$installed_script")"
  build_installed_start_issue "$installed_script" "1.11.1"
  create_fake_release_assets "$TEST_TMPDIR/release-assets-current" "1.11.1"

  run "$installed_script" update

  assert_success
  assert_output_contains "Installed version: v1.11.1"
  assert_output_contains "Latest release: v1.11.1"
  assert_output_contains "already up to date"
}

@test "update treats bare and v-prefixed versions as equivalent" {
  installed_script="$TEST_TMPDIR/bin/start-issue"
  mkdir -p "$(dirname "$installed_script")"
  build_installed_start_issue "$installed_script" "1.11.1"
  create_fake_release_assets "$TEST_TMPDIR/release-assets-normalized" "1.11.1"

  run "$installed_script" --update

  assert_success
  assert_output_contains "Installed version: v1.11.1"
  assert_output_contains "Latest release: v1.11.1"
  assert_output_contains "already up to date"
}

@test "update does not downgrade when installed version is newer than the latest release" {
  installed_script="$TEST_TMPDIR/bin/start-issue"
  mkdir -p "$(dirname "$installed_script")"
  build_installed_start_issue "$installed_script" "1.12.0"
  create_fake_release_assets "$TEST_TMPDIR/release-assets-older" "1.11.1"

  run "$installed_script" update

  assert_success
  assert_output_contains "Installed version: v1.12.0"
  assert_output_contains "Latest release: v1.11.1"
  assert_output_contains "newer than the latest published release"
  run "$installed_script" --version
  assert_success
  assert_output_contains "start-issue v1.12.0"
}

@test "update works outside a git repository" {
  installed_script="$TEST_TMPDIR/bin/start-issue"
  outside_dir="$TEST_TMPDIR/outside"
  mkdir -p "$(dirname "$installed_script")" "$outside_dir"
  build_installed_start_issue "$installed_script" "1.11.1"
  create_fake_release_assets "$TEST_TMPDIR/release-assets-outside" "1.11.1"
  expected_path="$(cd "$(dirname "$installed_script")" && pwd -P)/$(basename "$installed_script")"

  run bash -c "cd '$outside_dir' && '$installed_script' update"

  assert_success
  assert_output_contains "Executable: $expected_path"
  assert_output_contains "already up to date"
}

@test "update fails clearly when latest release lookup fails" {
  installed_script="$TEST_TMPDIR/bin/start-issue"
  mkdir -p "$(dirname "$installed_script")"
  build_installed_start_issue "$installed_script" "1.11.0"

  run "$installed_script" update

  assert_failure
  assert_output_contains "Failed to resolve the latest GitHub release"
}

@test "update rejects mixing update mode with issue input" {
  run_start_issue 1 --update

  assert_failure
  assert_output_contains "Use either update or <issue-url-or-number>, not both."
  [[ "$output" != *"Fetching issue"* ]]
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
  assert_output_contains "Agent: codex"
  assert_output_contains "Agent source: CLI"
  assert_output_contains "codex --cd"
}

@test "CLI model wins over project, user, and environment" {
  mkdir -p .start-issue
  mkdir -p "$HOME/.config/start-issue"
  printf "project-model\n" > .start-issue/model
  printf "user-model\n" > "$HOME/.config/start-issue/model"
  export START_ISSUE_MODEL=env-model

  run_start_issue 1 --agent codex --model cli-model --dry-run --no-init

  assert_success
  assert_output_contains "Model: cli-model"
  assert_output_contains "Model source: CLI"
  assert_output_contains "codex --model cli-model --cd"
}

@test "project model config wins over environment" {
  mkdir -p .start-issue
  printf "project-model\n" > .start-issue/model
  export START_ISSUE_MODEL=env-model

  run_start_issue 1 --agent codex --dry-run --no-init

  assert_success
  assert_output_contains "Model: project-model"
  assert_output_contains "Model source: "
  assert_output_contains ".start-issue/model"
  assert_output_contains "codex --model project-model --cd"
}

@test "user model config wins over environment" {
  mkdir -p "$HOME/.config/start-issue"
  printf "user-model\n" > "$HOME/.config/start-issue/model"
  export START_ISSUE_MODEL=env-model

  run_start_issue 1 --agent pi --dry-run --no-init

  assert_success
  assert_output_contains "Model: user-model"
  assert_output_contains "Model source: $HOME/.config/start-issue/model"
  assert_output_contains "pi --model user-model"
}

@test "project agent config wins over environment" {
  mkdir -p .start-issue
  printf "codex\n" > .start-issue/agent
  export START_ISSUE_AGENT=kimi

  run_start_issue 1 --dry-run --no-init

  assert_success
  assert_output_contains "Agent: codex"
  assert_output_contains "Agent source: "
  assert_output_contains ".start-issue/agent"
}

@test "user agent config wins over environment" {
  mkdir -p "$HOME/.config/start-issue"
  printf "pi\n" > "$HOME/.config/start-issue/agent"
  export START_ISSUE_AGENT=kimi

  run_start_issue 1 --dry-run --no-init

  assert_success
  assert_output_contains "Agent: pi"
  assert_output_contains "Agent source: $HOME/.config/start-issue/agent"
  assert_output_contains "cd $HOME/worktrees/feature/issue-1-add-login-button && pi"
}

@test "init writes project config with selected agent and portable prompt" {
  run_start_issue init --project --agent codex

  assert_success
  assert_output_contains "Scope: project config"
  assert_output_contains "Agent: codex"
  assert_output_contains "Agent source: CLI"
  [[ "$(cat .start-issue/agent)" == "codex" ]]
  [[ "$(cat .start-issue/prompt.md)" == *"Implement GitHub issue {ISSUE_URL} in this worktree."* ]]
  [[ "$(cat .start-issue/prompt.md)" == *"target the base branch {BASE_BRANCH}."* ]]
  [[ "$output" != *"Fetching issue"* ]]
}

@test "init writes project model config when selected" {
  run_start_issue init --project --agent codex --model gpt-5.2

  assert_success
  assert_output_contains "Model: gpt-5.2"
  assert_output_contains "Model source: CLI"
  [[ "$(cat .start-issue/model)" == "gpt-5.2" ]]
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
  assert_output_contains "Agent: codex"
  assert_output_contains "Agent source: "
  assert_output_contains ".start-issue/agent (existing)"
  [[ "$(cat .start-issue/agent)" == "codex" ]]
  [[ "$(cat .start-issue/prompt.md)" == *"Implement GitHub issue {ISSUE_URL} in this worktree."* ]]
  [[ "$(cat .start-issue/prompt.md)" == *"target the base branch {BASE_BRANCH}."* ]]
  [[ "$(cat .start-issue/prompt.md)" != *"/task-router:route-task"* ]]
}

@test "init dry-run does not write config files" {
  run_start_issue init --project --agent codex --dry-run

  assert_success
  assert_output_contains "[DRY-RUN] Would write agent config"
  assert_output_contains "No model config to write"
  [[ ! -e .start-issue/agent ]]
  [[ ! -e .start-issue/model ]]
  [[ ! -e .start-issue/prompt.md ]]
}

@test "init --force removes existing model config when no model is selected" {
  mkdir -p .start-issue
  printf "old-model\n" > .start-issue/model

  run_start_issue init --project --agent codex --force

  assert_success
  [[ ! -e .start-issue/model ]]
}

@test "--no-agent prints manual next steps" {
  run_start_issue 1 --no-agent --dry-run --no-init

  assert_success
  assert_output_contains "Selected agent: none (CLI)"
  assert_output_contains "To start working:"
  assert_output_contains "codex --cd $HOME/worktrees/feature/issue-1-add-login-button"
}

@test "--no-agent displays resolved model without passing launch args" {
  run_start_issue 1 --agent none --model sonnet --dry-run --no-init

  assert_success
  assert_output_contains "Selected agent: none (CLI)"
  assert_output_contains "Resolved model: sonnet (CLI)"
  [[ "$output" != *"--model sonnet"* ]]
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

@test "dry-run renders explicit model for supported launch adapters" {
  run_start_issue 1 --agent codex --model gpt-5.2 --dry-run --no-init
  assert_success
  assert_output_contains "Model: gpt-5.2"
  assert_output_contains "Model source: CLI"
  assert_output_contains "codex --model gpt-5.2 --cd $HOME/worktrees/feature/issue-1-add-login-button"

  run_start_issue 1 --agent claude --model sonnet --dry-run --no-init
  assert_success
  assert_output_contains "Model: sonnet"
  assert_output_contains "Model source: CLI"
  assert_output_contains "claude --model sonnet --dangerously-skip-permissions"
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

@test "prompt improvement passes explicit model to the selected agent" {
  export START_ISSUE_FAKE_IMPROVED_PROMPT="Improved prompt"
  export START_ISSUE_FAKE_EXPECT_MODEL=sonnet

  run_start_issue 1 --agent claude --model sonnet --improve-prompt --no-init

  assert_success
  [[ "$(cat .start-issue/prompt.improved.md)" == "Improved prompt" ]]
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

@test "AI branch naming passes explicit model to the selected agent" {
  export START_ISSUE_FAKE_BRANCH_NAME=fix/issue-1-ai-generated-name
  export START_ISSUE_FAKE_EXPECT_MODEL=gpt-5.2

  run_start_issue 1 --agent codex --model gpt-5.2 --ai --dry-run --no-init

  assert_success
  assert_output_contains "Branch: fix/issue-1-ai-generated-name"
  assert_output_contains "Model: gpt-5.2"
  assert_output_contains "Model source: CLI"
}

@test "no-model branch naming keeps existing adapter behavior" {
  export START_ISSUE_FAKE_BRANCH_NAME=fix/issue-1-ai-generated-name
  export START_ISSUE_FAKE_FORBID_MODEL=1

  run_start_issue 1 --agent codex --ai --dry-run --no-init

  assert_success
}

@test "unsupported model selection helper returns a clear error" {
  run bash -lc 'set -euo pipefail; RED=""; NC=""; source "'"$REPO_ROOT"'/scripts/lib/start_issue/utils.sh"; source "'"$REPO_ROOT"'/scripts/lib/start_issue/agent.sh"; AGENT="unsupported"; MODEL="x"; validate_model_selection_support launch'

  assert_failure
  assert_output_contains "does not support explicit model selection"
}

@test "AI branch naming falls back when selected agent returns invalid output" {
  export START_ISSUE_FAKE_BRANCH_NAME="not a branch"

  run_start_issue 1 --agent codex --ai --dry-run --no-init

  assert_success
  assert_output_contains "Generated branch name doesn't match expected format"
  assert_output_contains "Using fallback: feature/issue-1-add-login-button"
}

@test "AI branch naming falls back when branch ends with a trailing dash" {
  export START_ISSUE_FAKE_BRANCH_NAME="feature/issue-1-ai-generated-"

  run_start_issue 1 --agent codex --ai --dry-run --no-init

  assert_success
  assert_output_contains "Generated branch name doesn't match expected format"
  assert_output_contains "Using fallback: feature/issue-1-add-login-button"
}

@test "fast branch naming trims a trailing dash after truncation" {
  local long_prefix
  long_prefix="$(printf 'a%.0s' {1..39})"
  cat > "$TEST_TMPDIR/issue-trailing-dash.json" <<EOF
{
  "number": 1,
  "title": "${long_prefix} b",
  "body": "Trailing dash truncation case.",
  "html_url": "https://github.com/owner/repo/issues/1",
  "labels": [
    {
      "name": "enhancement"
    }
  ]
}
EOF
  export START_ISSUE_FAKE_ISSUE_JSON="$TEST_TMPDIR/issue-trailing-dash.json"

  run_start_issue 1 --agent none --dry-run --no-init

  assert_success
  assert_output_contains "Path: $HOME/worktrees/feature/issue-1-${long_prefix}"
  [[ "$output" != *"feature/issue-1-${long_prefix}-"* ]]
}

@test "make test target exists and runs local verification stack" {
  run make -n -f "$REPO_ROOT/Makefile" test

  assert_success
  assert_output_contains "bash -n scripts/start-issue"
  assert_output_contains "shellcheck install.sh scripts/start-issue"
  assert_output_contains "git diff --check"
  assert_output_contains "bats test"
}

@test "install.sh installs local release asset" {
  mkdir -p "$TEST_TMPDIR/install-fixture"
  cat > "$TEST_TMPDIR/install-fixture/start-issue" <<'EOF'
#!/usr/bin/env bash
printf 'start-issue test-build\n'
EOF
  chmod +x "$TEST_TMPDIR/install-fixture/start-issue"
  shasum -a 256 "$TEST_TMPDIR/install-fixture/start-issue" | awk '{ print $1 "  start-issue" }' > "$TEST_TMPDIR/install-fixture/start-issue.sha256"

  run_install_script

  assert_success
  assert_output_contains "Downloading latest release from test/local"
  assert_output_contains "Installed: $TEST_TMPDIR/install-prefix/bin/start-issue"
  assert_output_contains "Version: start-issue test-build"
  [[ -x "$TEST_TMPDIR/install-prefix/bin/start-issue" ]]
}

@test "install.sh debug mode prints diagnostic details" {
  mkdir -p "$TEST_TMPDIR/install-fixture"
  cat > "$TEST_TMPDIR/install-fixture/start-issue" <<'EOF'
#!/usr/bin/env bash
printf 'start-issue debug-build\n'
EOF
  chmod +x "$TEST_TMPDIR/install-fixture/start-issue"
  shasum -a 256 "$TEST_TMPDIR/install-fixture/start-issue" | awk '{ print $1 "  start-issue" }' > "$TEST_TMPDIR/install-fixture/start-issue.sha256"

  run_install_script --debug

  assert_success
  assert_output_contains "DEBUG: Repository: test/local"
  assert_output_contains "DEBUG: Asset URL: file://$TEST_TMPDIR/install-fixture/start-issue"
  assert_output_contains "DEBUG: Verifying checksum"
  assert_output_contains "+ install.sh:"
  assert_output_contains "Version: start-issue debug-build"
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
