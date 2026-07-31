# shellcheck shell=bash disable=SC2034
parse_issue_input() {
    local input="$1"

    if [[ "$input" =~ ^https://github\.com/([^/]+)/([^/]+)/issues/([0-9]+) ]]; then
        REPO="${BASH_REMATCH[1]}/${BASH_REMATCH[2]}"
        ISSUE_NUMBER="${BASH_REMATCH[3]}"
    elif [[ "$input" =~ ^[0-9]+$ ]]; then
        ISSUE_NUMBER="$input"
    else
        die "Invalid issue format: $input. Use issue number or full GitHub URL."
    fi
}

detect_repo_from_remote() {
    if [[ -n "$REPO" ]]; then
        return
    fi

    local remote_url
    remote_url=$(git remote get-url origin 2>/dev/null || echo "")

    if [[ -z "$remote_url" ]]; then
        die "Cannot detect repository. No 'origin' remote found. Use --repo flag."
    fi

    if [[ "$remote_url" =~ git@github\.com:([^/]+)/(.+)$ ]]; then
        REPO="${BASH_REMATCH[1]}/${BASH_REMATCH[2]}"
        REPO="${REPO%.git}"
    elif [[ "$remote_url" =~ https://github\.com/([^/]+)/(.+)$ ]]; then
        REPO="${BASH_REMATCH[1]}/${BASH_REMATCH[2]}"
        REPO="${REPO%.git}"
    else
        die "Cannot parse repository from remote URL: $remote_url. Use --repo flag."
    fi
}

detect_base_branch() {
    if [[ -n "$BASE_BRANCH" ]]; then
        return
    fi

    local remote_head
    remote_head=$(git symbolic-ref refs/remotes/origin/HEAD 2>/dev/null | sed 's|refs/remotes/origin/||' || true)
    if [[ -n "$remote_head" ]]; then
        BASE_BRANCH="$remote_head"
        return
    fi

    BASE_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "HEAD")
    log_info "Could not detect default branch, using current: $BASE_BRANCH"
}

fetch_issue() {
    log_info "🔍 Fetching issue #$ISSUE_NUMBER from $REPO..."

    ISSUE_JSON=$(gh api "repos/$REPO/issues/$ISSUE_NUMBER" 2>/dev/null) || \
        die "Issue #$ISSUE_NUMBER not found in $REPO"

    ISSUE_TITLE=$(echo "$ISSUE_JSON" | jq -r '.title')
    ISSUE_BODY=$(echo "$ISSUE_JSON" | jq -r '.body // ""')
    ISSUE_LABELS=$(echo "$ISSUE_JSON" | jq -r '[.labels[].name] | join(", ")')
    ISSUE_URL="https://github.com/$REPO/issues/$ISSUE_NUMBER"

    echo "   Title: $ISSUE_TITLE"
    if [[ -n "$ISSUE_LABELS" ]]; then
        echo "   Labels: $ISSUE_LABELS"
    fi
}
