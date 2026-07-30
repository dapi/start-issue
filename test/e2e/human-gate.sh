#!/usr/bin/env bash

# Opt-in smoke test for a real local Codex human-gate session.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
fixture_repo="dapi/start-issue-e2e-fixture"
fixture_issue="1"
scenario="done"

usage() {
    cat <<'EOF'
Usage: START_ISSUE_E2E=1 test/e2e/human-gate.sh [--scenario done|human-gate]

Runs start-issue against a real Codex CLI using the private fixture repository
dapi/start-issue-e2e-fixture and its control issue #1. It deletes the temporary
clone after a successful run; set START_ISSUE_E2E_KEEP=1 to retain it. The
HUMAN_GATE scenario opens Codex resume interactively; exit it to continue.
EOF
}

fail() {
    printf 'E2E human-gate: %s\n' "$*" >&2
    exit 1
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --scenario)
            scenario="${2:-}"
            shift 2
            ;;
        --help|-h)
            usage
            exit 0
            ;;
        *)
            fail "unknown argument: $1"
            ;;
    esac
done

[[ "$scenario" == "done" || "$scenario" == "human-gate" ]] || fail "scenario must be done or human-gate"
[[ "${START_ISSUE_E2E:-}" == "1" ]] || fail "set START_ISSUE_E2E=1 to authorize a real Codex session"
start_issue_bin="${START_ISSUE_E2E_BINARY:-$repo_root/scripts/start-issue}"

[[ -x "$start_issue_bin" ]] || fail "start-issue executable not found: $start_issue_bin"

codex_path="$(command -v codex || true)"
[[ -n "$codex_path" ]] || fail "codex is not on PATH"
[[ "$codex_path" != "$repo_root/test/helpers/fake-bin/codex" ]] || fail "PATH resolves codex to the test fake"
gh auth status >/dev/null || fail "gh is not authenticated"

if codex exec --help 2>&1 | grep -q -- '--ask-for-approval'; then
    fail "installed codex exec still advertises --ask-for-approval; use a current Codex CLI"
fi

fixture_root="$(mktemp -d "${TMPDIR:-/tmp}/start-issue-human-gate.XXXXXX")"
fixture_dir="$fixture_root/fixture"
worktree_parent="$fixture_root/worktrees"
log_path="$fixture_root/e2e.log"
expected_status="DONE"
if [[ "$scenario" == "human-gate" ]]; then
    expected_status="HUMAN_GATE"
fi

prompt=$(cat <<EOF
This is a controlled local human-gate E2E test.
Do not use tools, run commands, inspect files, modify files, create commits, or call external services.
Reply with exactly this one line and nothing else:
STATUS: $expected_status
EOF
)

printf 'E2E human-gate scenario: %s\n' "$scenario"
printf 'Fixture repository: %s (issue #%s)\n' "$fixture_repo" "$fixture_issue"
printf 'Worktree parent: %s\n' "$worktree_parent"
printf 'Log: %s\n' "$log_path"

gh repo clone "$fixture_repo" "$fixture_dir" -- --depth 1 || fail "could not clone fixture repository"
mkdir -p "$worktree_parent"
[[ -z "$(git -C "$fixture_dir" status --porcelain)" ]] || fail "fixture clone is unexpectedly dirty"

set +e
(
    cd "$fixture_dir"
    "$start_issue_bin" "$fixture_issue" \
        --agent codex \
        --human-gate \
        --no-init \
        --worktree-dir "$worktree_parent" \
        --prompt "$prompt"
) 2>&1 | tee "$log_path"
command_status=${PIPESTATUS[0]}
set -e

[[ $command_status -eq 0 ]] || fail "start-issue exited $command_status; inspect $log_path"

state_dir="$(awk '/^[[:space:]]*State dir: / { sub(/^[[:space:]]*State dir: /, ""); print; exit }' "$log_path")"
[[ -n "$state_dir" ]] || fail "state directory was not reported; inspect $log_path"
last_message_path="$state_dir/last-message.txt"
worktree_path="$(dirname "$(dirname "$(dirname "$state_dir")")")"

[[ -f "$state_dir/events.jsonl" ]] || fail "events.jsonl is missing: $state_dir/events.jsonl"
[[ -f "$last_message_path" ]] || fail "last-message.txt is missing: $last_message_path"
[[ -f "$state_dir/thread-id" ]] || fail "thread-id is missing: $state_dir/thread-id"
jq -e 'select(.type == "thread.started") | .thread_id | strings' "$state_dir/events.jsonl" >/dev/null || \
    fail "thread.started event is missing from $state_dir/events.jsonl"
grep -Fx "STATUS: $expected_status" "$last_message_path" >/dev/null || \
    fail "expected STATUS: $expected_status in $last_message_path"

if [[ "$scenario" == "human-gate" ]]; then
    grep -F 'Resume command: codex resume --include-non-interactive ' "$log_path" >/dev/null || \
        fail "resume command was not reported; inspect $log_path"
fi

unexpected_changes="$(git -C "$worktree_path" status --porcelain | awk '$0 !~ /^\?\? \.start-issue\// { print }')"
[[ -z "$unexpected_changes" ]] || fail "fixture worktree has unexpected changes: $unexpected_changes"

printf 'PASS: real Codex human-gate %s scenario. State: %s\n' "$scenario" "$state_dir"
if [[ "${START_ISSUE_E2E_KEEP:-}" == "1" ]]; then
    printf 'The temporary fixture is preserved at: %s\n' "$fixture_root"
else
    git -C "$fixture_dir" worktree remove --force "$worktree_path"
    rm -rf -- "$fixture_root"
    printf 'Temporary fixture clone and worktree removed. Set START_ISSUE_E2E_KEEP=1 to preserve them.\n'
fi
