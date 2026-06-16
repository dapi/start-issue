# shellcheck shell=bash disable=SC2034
sanitize_branch_slug() {
    local input="$1"
    local slug

    slug=$(printf "%s" "$input" | LC_ALL=en_US.UTF-8 sed 's/^\(\[[^]]*\][[:space:]-]*\)*//' | \
        LC_ALL=en_US.UTF-8 sed '
s/Щ/Shch/g; s/щ/shch/g;
s/Ж/Zh/g;   s/ж/zh/g;
s/Х/Kh/g;   s/х/kh/g;
s/Ц/Ts/g;   s/ц/ts/g;
s/Ч/Ch/g;   s/ч/ch/g;
s/Ш/Sh/g;   s/ш/sh/g;
s/Ю/Yu/g;   s/ю/yu/g;
s/Я/Ya/g;   s/я/ya/g;
s/Ё/Yo/g;   s/ё/yo/g;
s/А/A/g; s/а/a/g;
s/Б/B/g; s/б/b/g;
s/В/V/g; s/в/v/g;
s/Г/G/g; s/г/g/g;
s/Д/D/g; s/д/d/g;
s/Е/E/g; s/е/e/g;
s/З/Z/g; s/з/z/g;
s/И/I/g; s/и/i/g;
s/Й/Y/g; s/й/y/g;
s/К/K/g; s/к/k/g;
s/Л/L/g; s/л/l/g;
s/М/M/g; s/м/m/g;
s/Н/N/g; s/н/n/g;
s/О/O/g; s/о/o/g;
s/П/P/g; s/п/p/g;
s/Р/R/g; s/р/r/g;
s/С/S/g; s/с/s/g;
s/Т/T/g; s/т/t/g;
s/У/U/g; s/у/u/g;
s/Ф/F/g; s/ф/f/g;
s/Ы/Y/g; s/ы/y/g;
s/Э/E/g; s/э/e/g;
s/Ъ//g;  s/ъ//g;
s/Ь//g;  s/ь//g;
' | \
        tr '[:upper:]' '[:lower:]' | \
        sed 's/[^a-z0-9]/-/g' | sed 's/--*/-/g' | cut -c1-40 | sed 's/^-*//' | sed 's/-*$//')

    if [[ -z "$slug" ]]; then
        slug="work"
    fi

    printf "%s" "$slug"
}

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

    short_name=$(sanitize_branch_slug "$ISSUE_TITLE")

    BRANCH_NAME="$branch_type/issue-$ISSUE_NUMBER-$short_name"
}

validate_branch_name_or_fallback() {
    local elapsed="$1"

    if [[ "$BRANCH_NAME" =~ ^(feature|fix|hotfix|refactor|docs|test|chore)/issue-[0-9]+-[a-z0-9]([a-z0-9-]*[a-z0-9])?$ ]]; then
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
    local existing_worktree=""
    local path_branch=""

    plan_worktree_path

    log_info "📁 Creating worktree..."
    echo "   Path: $WORKTREE_PATH"
    echo "   Base: $BASE_BRANCH"

    if git show-ref --verify --quiet "refs/heads/$BRANCH_NAME" 2>/dev/null; then
        existing_worktree=$(find_worktree_path_by_branch "$BRANCH_NAME" || true)
        if plan_branch_reuse_resolution "$existing_worktree"; then
            return
        fi
    fi

    if [[ -d "$WORKTREE_PATH" ]]; then
        path_branch=$(find_worktree_branch_by_path "$WORKTREE_PATH" || true)
        if plan_path_reuse_resolution "$path_branch"; then
            return
        fi
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

worktree_name_for_branch() {
    local branch_name="$1"
    local worktree_name="$branch_name"

    if [[ "$FLAT_WORKTREE" == "true" ]]; then
        worktree_name="${branch_name//\//-}"
    fi

    printf "%s" "$worktree_name"
}

plan_worktree_path() {
    WORKTREE_PATH="$WORKTREE_DIR/$(worktree_name_for_branch "$BRANCH_NAME")"
}

worktree_branch_ref() {
    local branch_name="$1"
    printf "refs/heads/%s" "$branch_name"
}

find_worktree_path_by_branch() {
    local target_ref
    local line
    local current_path=""

    target_ref=$(worktree_branch_ref "$1")

    while IFS= read -r line; do
        case "$line" in
            worktree\ *)
                current_path="${line#worktree }"
                ;;
            branch\ "$target_ref")
                printf "%s" "$current_path"
                return 0
                ;;
        esac
    done < <(git worktree list --porcelain)

    return 1
}

find_worktree_branch_by_path() {
    local target_path="$1"
    local line
    local current_path=""

    target_path=$(canonicalize_existing_path "$target_path")

    while IFS= read -r line; do
        case "$line" in
            worktree\ *)
                current_path="${line#worktree }"
                ;;
            branch\ *)
                if [[ "$current_path" == "$target_path" ]]; then
                    printf "%s" "${line#branch }"
                    return 0
                fi
                ;;
        esac
    done < <(git worktree list --porcelain)

    return 1
}

validate_reused_worktree() {
    local registered_branch

    if [[ ! -d "$WORKTREE_PATH" ]]; then
        die "Cannot reuse worktree path '$WORKTREE_PATH': directory does not exist."
    fi

    registered_branch=$(find_worktree_branch_by_path "$WORKTREE_PATH" || true)

    if [[ -z "$registered_branch" ]]; then
        die "Cannot reuse worktree path '$WORKTREE_PATH': path exists but is not a git worktree for this repository."
    fi

    if [[ "$registered_branch" != "$(worktree_branch_ref "$BRANCH_NAME")" ]]; then
        die "Cannot reuse worktree path '$WORKTREE_PATH': it belongs to branch '${registered_branch#refs/heads/}', not '$BRANCH_NAME'."
    fi
}

prompt_branch_conflict_resolution() {
    local existing_worktree="$1"

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
    read -r -n 1 -p "Choice: " CONFLICT_CHOICE
    echo ""
}

plan_branch_reuse_resolution() {
    local existing_worktree="$1"

    prompt_branch_conflict_resolution "$existing_worktree"

    case "$CONFLICT_CHOICE" in
        1)
            if [[ -z "$existing_worktree" ]]; then
                die "No existing worktree found for branch '$BRANCH_NAME'. Use 3 to delete and recreate."
            fi
            WORKTREE_PATH="$existing_worktree"
            validate_reused_worktree
            log_info "   Using existing worktree: $WORKTREE_PATH"
            return 0
            ;;
        2)
            local version=2
            local new_branch="${BRANCH_NAME}-v${version}"

            while git show-ref --verify --quiet "refs/heads/$new_branch" 2>/dev/null; do
                ((version++))
                new_branch="${BRANCH_NAME}-v${version}"
            done

            BRANCH_NAME="$new_branch"
            plan_worktree_path
            log_info "   New branch name: $BRANCH_NAME"
            return 1
            ;;
        3)
            log_info "   Removing existing branch/worktree..."
            if [[ -n "$existing_worktree" ]]; then
                git worktree remove --force "$existing_worktree" 2>/dev/null || rm -rf "$existing_worktree"
            fi
            git branch -D "$BRANCH_NAME" 2>/dev/null || true
            log_success "   ✅ Cleaned up"
            return 1
            ;;
        *)
            die "Aborted"
            ;;
    esac
}

prompt_path_conflict_resolution() {
    local path_branch="$1"

    echo ""
    log_warn "Worktree path already exists: $WORKTREE_PATH"
    if [[ -n "$path_branch" ]]; then
        echo "   Registered branch: ${path_branch#refs/heads/}"
    else
        echo "   Registered branch: none"
    fi
    echo ""
    echo "  1) Use existing worktree"
    echo "  2) Delete and recreate"
    echo "  0) Exit"
    echo ""
    read -r -n 1 -p "Choice: " CONFLICT_CHOICE
    echo ""
}

plan_path_reuse_resolution() {
    local path_branch="$1"

    prompt_path_conflict_resolution "$path_branch"

    case "$CONFLICT_CHOICE" in
        1)
            validate_reused_worktree
            log_info "   Using existing worktree"
            return 0
            ;;
        2)
            log_info "   Removing existing worktree..."
            git worktree remove --force "$WORKTREE_PATH" 2>/dev/null || rm -rf "$WORKTREE_PATH"
            if [[ -n "$path_branch" ]]; then
                git branch -D "${path_branch#refs/heads/}" 2>/dev/null || true
            else
                git branch -D "$BRANCH_NAME" 2>/dev/null || true
            fi
            return 1
            ;;
        *)
            die "Aborted"
            ;;
    esac
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
