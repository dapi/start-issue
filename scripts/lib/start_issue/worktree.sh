# shellcheck shell=bash disable=SC2034
generate_fast_branch_name() {
    local branch_type="feature"
    local short_name

    if [[ "$ISSUE_LABELS" =~ (hotfix|critical|urgent) ]]; then
        branch_type="hotfix"
    elif [[ "$ISSUE_LABELS" =~ (bug|fix|bugfix|error) ]]; then
        branch_type="fix"
    elif [[ "$ISSUE_LABELS" =~ (docs|documentation) ]]; then
        branch_type="docs"
    elif [[ "$ISSUE_LABELS" =~ (refactor|tech-debt|cleanup|technical) ]]; then
        branch_type="refactor"
    elif [[ "$ISSUE_LABELS" =~ (test|testing|tests) ]]; then
        branch_type="test"
    elif [[ "$ISSUE_LABELS" =~ (chore|ci|build|infra) ]]; then
        branch_type="chore"
    fi

    short_name=$(echo "$ISSUE_TITLE" | tr '[:upper:]' '[:lower:]' | \
        sed 's/[^a-z0-9]/-/g' | sed 's/--*/-/g' | sed 's/^-//' | sed 's/-$//' | cut -c1-40)

    BRANCH_NAME="$branch_type/issue-$ISSUE_NUMBER-$short_name"
}

validate_branch_name_or_fallback() {
    local elapsed="$1"

    if [[ "$BRANCH_NAME" =~ ^(feature|fix|hotfix|refactor|docs|test|chore)/issue-[0-9]+-[a-z0-9-]+$ ]]; then
        log_success "   Branch: $BRANCH_NAME (${elapsed}s, ai:$AGENT)"
        return
    fi

    log_warn "Generated branch name doesn't match expected format: $BRANCH_NAME"
    generate_fast_branch_name
    log_info "   Using fallback: $BRANCH_NAME (${elapsed}s)"
}

generate_branch_name() {
    log_info "🧠 Generating branch name..."

    local start_time=$SECONDS
    local elapsed

    if [[ "$FAST_MODE" == "true" ]]; then
        generate_fast_branch_name
        elapsed=$((SECONDS - start_time))
        log_success "   Branch: $BRANCH_NAME (${elapsed}s, fast)"
        return
    fi

    if generate_ai_branch_name; then
        elapsed=$((SECONDS - start_time))
        validate_branch_name_or_fallback "$elapsed"
    else
        elapsed=$((SECONDS - start_time))
        log_warn "Could not generate branch name with $AGENT; falling back to fast heuristic"
        generate_fast_branch_name
        log_info "   Branch: $BRANCH_NAME (${elapsed}s, fast fallback)"
    fi
}

create_worktree() {
    local worktree_name="$BRANCH_NAME"
    local existing_worktree
    local choice

    if [[ "$FLAT_WORKTREE" == "true" ]]; then
        worktree_name="${BRANCH_NAME//\//-}"
    fi
    WORKTREE_PATH="$WORKTREE_DIR/$worktree_name"

    log_info "📁 Creating worktree..."
    echo "   Path: $WORKTREE_PATH"
    echo "   Base: $BASE_BRANCH"

    if git show-ref --verify --quiet "refs/heads/$BRANCH_NAME" 2>/dev/null; then
        existing_worktree=$(git worktree list --porcelain | grep -A2 "^worktree " | \
            awk -v branch="$BRANCH_NAME" '/^worktree /{path=$2} /^branch refs\/heads\// && $2 ~ branch {print path}' | head -1)

        echo ""
        log_warn "Branch '$BRANCH_NAME' already exists."
        if [[ -n "$existing_worktree" ]]; then
            echo "   Existing worktree: $existing_worktree"
        fi
        echo ""
        echo "  1) Use existing worktree and continue"
        echo "  2) Create new branch with different name"
        echo "  3) Delete branch/worktree and recreate"
        echo "  0) Exit"
        echo ""
        read -r -n 1 -p "Choice: " choice
        echo ""
        case "$choice" in
            1)
                if [[ -n "$existing_worktree" ]]; then
                    WORKTREE_PATH="$existing_worktree"
                    log_info "   Using existing worktree: $WORKTREE_PATH"
                    return
                else
                    die "No existing worktree found for branch '$BRANCH_NAME'. Use 3 to delete and recreate."
                fi
                ;;
            2)
                local version=2
                local new_branch="${BRANCH_NAME}-v${version}"
                while git show-ref --verify --quiet "refs/heads/$new_branch" 2>/dev/null; do
                    ((version++))
                    new_branch="${BRANCH_NAME}-v${version}"
                done
                BRANCH_NAME="$new_branch"
                worktree_name="$BRANCH_NAME"
                if [[ "$FLAT_WORKTREE" == "true" ]]; then
                    worktree_name="${BRANCH_NAME//\//-}"
                fi
                WORKTREE_PATH="$WORKTREE_DIR/$worktree_name"
                log_info "   New branch name: $BRANCH_NAME"
                ;;
            3)
                log_info "   Removing existing branch/worktree..."
                if [[ -n "$existing_worktree" ]]; then
                    git worktree remove --force "$existing_worktree" 2>/dev/null || rm -rf "$existing_worktree"
                fi
                git branch -D "$BRANCH_NAME" 2>/dev/null || true
                log_success "   ✅ Cleaned up"
                ;;
            *)
                die "Aborted"
                ;;
        esac
    fi

    if [[ -d "$WORKTREE_PATH" ]]; then
        echo ""
        log_warn "Worktree path already exists: $WORKTREE_PATH"
        echo ""
        echo "  1) Use existing worktree"
        echo "  2) Delete and recreate"
        echo "  0) Exit"
        echo ""
        read -r -n 1 -p "Choice: " choice
        echo ""
        case "$choice" in
            1)
                log_info "   Using existing worktree"
                return
                ;;
            2)
                log_info "   Removing existing worktree..."
                git worktree remove --force "$WORKTREE_PATH" 2>/dev/null || rm -rf "$WORKTREE_PATH"
                git branch -D "$BRANCH_NAME" 2>/dev/null || true
                ;;
            *)
                die "Aborted"
                ;;
        esac
    fi

    if [[ "$DRY_RUN" == "true" ]]; then
        echo "   [DRY-RUN] Would run: git worktree add -b $BRANCH_NAME $WORKTREE_PATH $BASE_BRANCH"
        return
    fi

    mkdir -p "$(dirname "$WORKTREE_PATH")"
    git fetch origin "$BASE_BRANCH" --quiet 2>/dev/null || true

    git worktree add -b "$BRANCH_NAME" "$WORKTREE_PATH" "origin/$BASE_BRANCH" || \
        git worktree add -b "$BRANCH_NAME" "$WORKTREE_PATH" "$BASE_BRANCH" || \
        die "Failed to create worktree"

    log_success "   ✅ Worktree created"
}

run_init_script() {
    local init_script

    if [[ "$NO_INIT" == "true" ]]; then
        log_info "⏭️  Skipping init.sh (--no-init)"
        return
    fi

    init_script="$WORKTREE_PATH/init.sh"

    if [[ ! -f "$init_script" ]]; then
        log_warn "init.sh not found, skipping initialization"
        return
    fi

    log_info "⚙️  Running init.sh..."

    if [[ "$DRY_RUN" == "true" ]]; then
        echo "   [DRY-RUN] Would run: $init_script"
        return
    fi

    if ! (cd "$WORKTREE_PATH" && bash ./init.sh); then
        log_warn "init.sh exited with non-zero code"
    else
        log_success "   ✅ Done"
    fi
}

rename_zellij_tab() {
    local tab_name="#$ISSUE_NUMBER"

    if [[ "$DRY_RUN" == "true" ]]; then
        if command -v zellij-tab-status &> /dev/null; then
            echo "   [DRY-RUN] Would run: zellij-tab-status --set-name $(shell_join "$tab_name")"
        else
            echo "   [DRY-RUN] Would skip zellij tab rename: zellij-tab-status not found"
        fi
        return
    fi

    if ! command -v zellij-tab-status &> /dev/null; then
        return
    fi

    log_info "📑 Renaming zellij tab..."
    if zellij-tab-status --set-name "$tab_name" &> /dev/null; then
        log_success "   ✅ Tab renamed to #$ISSUE_NUMBER"
    else
        log_warn "Could not rename zellij tab with zellij-tab-status"
    fi
}
