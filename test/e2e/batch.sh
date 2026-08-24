#!/usr/bin/env bash

# Opt-in smoke test for a real local Codex batch session.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
fixture_repo="dapi/start-issue-e2e-fixture"
fixture_issue="1"
scenario="done"

usage() {
    cat <<'EOF'
Usage: START_ISSUE_E2E=1 test/e2e/batch.sh [--scenario done|human-gate|full-delivery]

Runs start-issue against a real Codex CLI using the private fixture repository
dapi/start-issue-e2e-fixture and its control issue #1. It deletes the temporary
clone after a successful run; set START_ISSUE_E2E_KEEP=1 to retain it. The
HUMAN_GATE scenario opens Codex resume interactively; exit it to continue.
FULL_DELIVERY also requires START_ISSUE_E2E_FULL_DELIVERY=1. It authorizes an
unsandboxed Codex run that creates a unique fixture commit, remote branch, and
pull request. Its temporary fixture and diagnostic artifacts are retained as
evidence automatically.
EOF
}

fail() {
    printf 'E2E batch: %s\n' "$*" >&2
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

[[ "$scenario" == "done" || "$scenario" == "human-gate" || "$scenario" == "full-delivery" ]] || \
    fail "scenario must be done, human-gate, or full-delivery"
[[ "${START_ISSUE_E2E:-}" == "1" ]] || fail "set START_ISSUE_E2E=1 to authorize a real Codex session"
if [[ "$scenario" == "full-delivery" ]]; then
    [[ "${START_ISSUE_E2E_FULL_DELIVERY:-}" == "1" ]] || \
        fail "set START_ISSUE_E2E_FULL_DELIVERY=1 to authorize unsandboxed GitHub delivery"
fi
start_issue_bin="${START_ISSUE_E2E_BINARY:-$repo_root/.build/start-issue}"

[[ -x "$start_issue_bin" ]] || fail "start-issue executable not found: $start_issue_bin"

codex_path="$(command -v codex || true)"
[[ -n "$codex_path" ]] || fail "codex is not on PATH"
[[ "$codex_path" != "$repo_root/test/helpers/fake-bin/codex" ]] || fail "PATH resolves codex to the test fake"
gh auth status >/dev/null || fail "gh is not authenticated"

codex_exec_help="$(codex exec --help 2>&1)" || fail "codex exec --help failed"
printf '%s' "$codex_exec_help" | grep -Fq 'Run Codex non-interactively' || \
    fail "resolved codex does not expose the real Codex exec interface"
printf '%s' "$codex_exec_help" | grep -Fq -- '--output-last-message' || \
    fail "resolved codex does not support --output-last-message"
if printf '%s' "$codex_exec_help" | grep -q -- '--ask-for-approval'; then
    fail "installed codex exec still advertises --ask-for-approval; use a current Codex CLI"
fi
if [[ "$scenario" == "full-delivery" ]]; then
    codex_help="$(codex --help 2>&1)" || fail "codex --help failed"
    printf '%s' "$codex_help" | grep -Fq -- '--dangerously-bypass-approvals-and-sandbox' || \
        fail "resolved codex does not support the full-delivery permission option"
fi

fixture_root="$(mktemp -d "${TMPDIR:-/tmp}/start-issue-batch.XXXXXX")"
fixture_dir="$fixture_root/fixture"
worktree_parent="$fixture_root/worktrees"
log_path="$fixture_root/e2e.log"
expected_status="DONE"
permission_args=()
if [[ "$scenario" == "human-gate" ]]; then
    expected_status="HUMAN_GATE"
fi

if [[ "$scenario" == "full-delivery" ]]; then
    permission_args=(--batch-permissions full-delivery)
    delivery_id="$(date -u +%Y%m%d%H%M%S)-$$"
    delivery_branch="e2e/batch-full-delivery-$delivery_id"
    delivery_file="full-delivery-$delivery_id.txt"
    fixture_base="$(gh repo view "$fixture_repo" --json defaultBranchRef --jq '.defaultBranchRef.name')"
    [[ -n "$fixture_base" ]] || fail "could not resolve fixture default branch"
fi

if [[ "$scenario" == "full-delivery" ]]; then
    prompt=$(cat <<EOF
This is an explicitly authorized full-delivery E2E test in the private fixture repository $fixture_repo.
Complete these exact steps without changing any other tracked file:
1. Create and switch to the new branch $delivery_branch.
2. Create $delivery_file containing exactly: full-delivery $delivery_id
3. Commit that file with message: Verify batch full delivery $delivery_id
4. Push the branch to origin.
5. Create a pull request into $fixture_base with title "Batch full delivery $delivery_id" and body "Automated retained evidence for start-issue batch full-delivery E2E."
6. Verify the pull request exists, then reply with STATUS: DONE and include its URL.
Do not merge or close the pull request. Do not modify production or any repository other than $fixture_repo.
EOF
    )
else
    prompt=$(cat <<EOF
This is a controlled local batch E2E test.
Do not use tools, run commands, inspect files, modify files, create commits, or call external services.
Reply with exactly this one line and nothing else:
STATUS: $expected_status
EOF
    )
fi

printf 'E2E batch scenario: %s\n' "$scenario"
printf 'Codex executable: %s\n' "$codex_path"
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
        --batch \
        "${permission_args[@]}" \
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

if [[ "$scenario" == "full-delivery" ]]; then
    current_branch="$(git -C "$worktree_path" branch --show-current)"
    [[ "$current_branch" == "$delivery_branch" ]] || \
        fail "full-delivery branch is $current_branch, want $delivery_branch"
    git -C "$worktree_path" show "HEAD:$delivery_file" | grep -Fx "full-delivery $delivery_id" >/dev/null || \
        fail "delivery commit does not contain the expected marker"
    git -C "$worktree_path" ls-remote --exit-code --heads origin "$delivery_branch" >/dev/null || \
        fail "remote delivery branch is missing: $delivery_branch"
    pr_url="$(gh pr list --repo "$fixture_repo" --state open --head "$delivery_branch" --json url --jq '.[0].url // empty')"
    [[ -n "$pr_url" ]] || fail "full-delivery pull request is missing for $delivery_branch"
    printf 'Full-delivery PR: %s\n' "$pr_url"
fi

unexpected_changes="$(git -C "$worktree_path" status --porcelain | awk '$0 !~ /^\?\? \.start-issue\// { print }')"
[[ -z "$unexpected_changes" ]] || fail "fixture worktree has unexpected changes: $unexpected_changes"

printf 'PASS: real Codex batch %s scenario. State: %s\n' "$scenario" "$state_dir"
if [[ "$scenario" == "full-delivery" || "${START_ISSUE_E2E_KEEP:-}" == "1" ]]; then
    printf 'The temporary fixture is preserved at: %s\n' "$fixture_root"
else
    git -C "$fixture_dir" worktree remove --force "$worktree_path"
    rm -rf -- "$fixture_root"
    printf 'Temporary fixture clone and worktree removed. Set START_ISSUE_E2E_KEEP=1 to preserve them.\n'
fi
