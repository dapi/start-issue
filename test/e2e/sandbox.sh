#!/usr/bin/env bash

# Deterministic, network-free E2E smoke test for the built Go CLI.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
start_issue_bin="${START_ISSUE_SANDBOX_BINARY:-$repo_root/.build/start-issue}"
fixture_root="$(mktemp -d "${TMPDIR:-/tmp}/start-issue-sandbox-e2e.XXXXXX")"
fixture_root="$(cd "$fixture_root" && pwd -P)"
trap 'rm -rf -- "$fixture_root"' EXIT

fail() {
    printf 'E2E sandbox: %s\n' "$*" >&2
    exit 1
}

[[ -x "$start_issue_bin" ]] || fail "start-issue executable not found: $start_issue_bin"

repo="$fixture_root/repo"
remote="$fixture_root/remote.git"
worktrees="$fixture_root/worktrees"
home="$fixture_root/home"
bin="$fixture_root/bin"
kimi_log="$fixture_root/kimi.log"
init_marker="$fixture_root/init.marker"
mkdir -p "$repo" "$worktrees" "$home/.config/start-issue" "$bin"

git -C "$repo" init -q
git -C "$repo" config user.email sandbox@example.invalid
git -C "$repo" config user.name 'start-issue sandbox'
printf '# sandbox fixture\n' > "$repo/README.md"
cat > "$repo/init.sh" <<EOF
#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "\$PWD" > "$init_marker"
EOF
chmod +x "$repo/init.sh"
git -C "$repo" add README.md init.sh
git -C "$repo" commit -q -m 'sandbox fixture'
git init --bare -q "$remote"
git -C "$repo" remote add origin "$remote"
git -C "$repo" push -q -u origin HEAD

cat > "$bin/gh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
case "${1:-}" in
    auth)
        [[ "${2:-}" == "status" ]] || exit 1
        ;;
    api)
        printf '%s\n' '{"number":42,"title":"Sandbox E2E issue","body":"controlled fixture","labels":[{"name":"feature"}]}'
        ;;
    *)
        printf 'unexpected gh invocation: %s\n' "$*" >&2
        exit 1
        ;;
esac
EOF

cat > "$bin/kimi" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
printf 'cwd=%s\nargs=%s\n' "$PWD" "$*" > "$KIMI_LOG"
EOF

chmod +x "$bin/gh" "$bin/kimi"

export HOME="$home"
export KIMI_LOG="$kimi_log"
export START_ISSUE_REPOSITORY=acme/sandbox
export PATH="$bin:$PATH"

printf 'Sandbox E2E: issue start with real git worktree and fake Kimi\n'
(
    cd "$repo"
    "$start_issue_bin" 42 \
        --repo acme/sandbox \
        --agent kimi \
        --model sandbox-model \
        --worktree-dir "$worktrees" \
        --prompt 'Implement {ISSUE_URL} in {WORKTREE_PATH}'
)

worktree="$worktrees/feature/issue-42-sandbox-e2e-issue"
[[ -d "$worktree" ]] || fail "worktree was not created: $worktree"
[[ -f "$init_marker" ]] || fail "init.sh was not executed"
[[ "$(cat "$init_marker")" == "$worktree" ]] || fail "init.sh ran outside worktree: got $(cat "$init_marker"), want $worktree"
[[ "$(sed -n '1p' "$kimi_log")" == "cwd=$worktree" ]] || fail "Kimi cwd is wrong: $(sed -n '1p' "$kimi_log")"
grep -F -- '--model sandbox-model -p Implement https://github.com/acme/sandbox/issues/42 in' "$kimi_log" >/dev/null || \
    fail "Kimi arguments do not contain rendered prompt"
git -C "$worktree" status --porcelain | grep -v '^?? \.start-issue/' >/dev/null && \
    fail "worktree contains unexpected changes"

printf 'Sandbox E2E: dry-run no-agent path\n'
dry_run_log="$fixture_root/dry-run.log"
(
    cd "$repo"
    "$start_issue_bin" 43 \
        --repo acme/sandbox \
        --agent none \
        --no-init \
        --dry-run \
        --flat \
        --worktree-dir "$fixture_root/dry-run-worktrees"
) > "$dry_run_log"
grep -F '[DRY-RUN] Would run: git worktree add' "$dry_run_log" >/dev/null || fail "dry-run did not plan worktree creation"
grep -F 'Agent: none' "$dry_run_log" >/dev/null || fail "dry-run did not resolve no-agent"
[[ ! -d "$fixture_root/dry-run-worktrees" ]] || fail "dry-run created worktree directory"

printf 'PASS: sandbox E2E scenarios\n'
