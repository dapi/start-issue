# shellcheck shell=bash disable=SC2034
agent_supports_explicit_model_selection() {
    local operation="$1"

    case "$operation" in
        launch|branch-name|prompt-improvement)
            ;;
        *)
            return 1
            ;;
    esac

    case "$AGENT" in
        claude|codex|kimi|pi)
            return 0
            ;;
        *)
            return 1
            ;;
    esac
}

validate_model_selection_support() {
    local operation="$1"

    if [[ -z "$MODEL" || "$AGENT" == "none" ]]; then
        return 0
    fi

    if agent_supports_explicit_model_selection "$operation"; then
        return 0
    fi

    die "Agent '$AGENT' does not support explicit model selection for $operation."
}

claude_noninteractive_model() {
    if [[ -n "$MODEL" ]]; then
        printf "%s" "$MODEL"
    else
        printf "%s" "haiku"
    fi
}

validate_prompt_improvement_mode() {
    if [[ "$IMPROVE_PROMPT" != "true" ]]; then
        return
    fi

    if [[ "$AGENT" == "none" ]]; then
        die "--improve-prompt requires an agent. Use --agent claude, codex, kimi, or pi."
    fi
}

default_prompt_improvement_output_path() {
    if [[ -n "$PROMPT_IMPROVEMENT_OUTPUT_FILE" ]]; then
        printf "%s" "$PROMPT_IMPROVEMENT_OUTPUT_FILE"
        return
    fi

    if [[ -n "$PROMPT_TEMPLATE_PATH" ]]; then
        local dir
        local file
        local stem

        dir=$(dirname "$PROMPT_TEMPLATE_PATH")
        file=$(basename "$PROMPT_TEMPLATE_PATH")
        if [[ "$file" == *.md ]]; then
            stem="${file%.md}"
            printf "%s/%s.improved.md" "$dir" "$stem"
        else
            printf "%s/%s.improved" "$dir" "$file"
        fi
        return
    fi

    printf "%s/.start-issue/prompt.improved.md" "$PROJECT_ROOT"
}

prompt_improvement_request() {
    cat << EOF
Improve the following start-issue prompt template.

Return ONLY the complete improved prompt template. Do not include commentary, code fences, diffs, or explanations.

Preserve any placeholders that are still useful. Supported placeholders:
{ISSUE_URL}, {ISSUE_NUMBER}, {ISSUE_TITLE}, {ISSUE_BODY}, {ISSUE_LABELS}, {REPO}, {BRANCH_NAME}, {WORKTREE_PATH}, {BASE_BRANCH}

Prompt source:
$PROMPT_SOURCE

Repository:
$REPO

Current issue used as improvement context:
- URL: $ISSUE_URL
- Number: $ISSUE_NUMBER
- Title: $ISSUE_TITLE
- Labels: $ISSUE_LABELS
- Body:
$ISSUE_BODY

Current prompt template:
--- START PROMPT TEMPLATE ---
$PROMPT_TEMPLATE
--- END PROMPT TEMPLATE ---
EOF
}

agent_supports_operation() {
    local operation="$1"

    case "$operation" in
        validate|launch|branch-name|prompt-improvement)
            ;;
        *)
            return 1
            ;;
    esac

    case "$AGENT" in
        claude|codex|kimi|pi)
            return 0
            ;;
        none)
            [[ "$operation" == "validate" ]] && return 0
            return 1
            ;;
        *)
            return 1
            ;;
    esac
}

generate_improved_prompt_template() {
    local request
    local output

    request=$(prompt_improvement_request)
    validate_model_selection_support "prompt-improvement"

    case "$AGENT" in
        claude)
            output=$(claude --print --model "$(claude_noninteractive_model)" --no-session-persistence \
                --disable-slash-commands "$request" 2>/dev/null) || return 1
            ;;
        codex)
            if [[ -n "$MODEL" ]]; then
                output=$(codex exec --model "$MODEL" --cd "$PROJECT_ROOT" --sandbox read-only \
                    --skip-git-repo-check "$request" 2>/dev/null) || return 1
            else
                output=$(codex exec --cd "$PROJECT_ROOT" --sandbox read-only \
                    --skip-git-repo-check "$request" 2>/dev/null) || return 1
            fi
            ;;
        kimi)
            if [[ -n "$MODEL" ]]; then
                output=$(kimi --model "$MODEL" --work-dir "$PROJECT_ROOT" --quiet -p "$request" 2>/dev/null) || return 1
            else
                output=$(kimi --work-dir "$PROJECT_ROOT" --quiet -p "$request" 2>/dev/null) || return 1
            fi
            ;;
        pi)
            if [[ -n "$MODEL" ]]; then
                output=$(pi --model "$MODEL" --print --no-tools --no-session "$request" 2>/dev/null) || return 1
            else
                output=$(pi --print --no-tools --no-session "$request" 2>/dev/null) || return 1
            fi
            ;;
        *)
            return 1
            ;;
    esac

    output=$(printf "%s" "$output" | sed '1{/^```[[:alnum:]_-]*$/d;}; ${/^```$/d;}')
    [[ -n "$(trim "$output")" ]] || return 1
    printf "%s" "$output"
}

generate_ai_branch_name() {
    local prompt="Git branch name for issue #$ISSUE_NUMBER: \"$ISSUE_TITLE\" (labels: $ISSUE_LABELS).
Format: {type}/issue-$ISSUE_NUMBER-{kebab-case-name}
Types: bug/fix -> fix, enhancement -> feature, hotfix -> hotfix, docs -> docs, refactor -> refactor, test -> test, chore -> chore, default -> feature.
Reply with ONLY the branch name."
    local output=""

    if ! agent_supports_operation "branch-name"; then
        return 1
    fi

    if ! command -v "$AGENT" &> /dev/null; then
        return 1
    fi

    validate_model_selection_support "branch-name"

    case "$AGENT" in
        claude)
            output=$(claude --print --model "$(claude_noninteractive_model)" --no-session-persistence \
                --disable-slash-commands "$prompt" 2>/dev/null) || return 1
            ;;
        codex)
            if [[ -n "$MODEL" ]]; then
                output=$(codex exec --model "$MODEL" --cd "$PROJECT_ROOT" --sandbox read-only \
                    --skip-git-repo-check "$prompt" 2>/dev/null | tail -n 1) || return 1
            else
                output=$(codex exec --cd "$PROJECT_ROOT" --sandbox read-only \
                    --skip-git-repo-check "$prompt" 2>/dev/null | tail -n 1) || return 1
            fi
            ;;
        kimi)
            if [[ -n "$MODEL" ]]; then
                output=$(kimi --model "$MODEL" --work-dir "$PROJECT_ROOT" --quiet -p "$prompt" 2>/dev/null) || return 1
            else
                output=$(kimi --work-dir "$PROJECT_ROOT" --quiet -p "$prompt" 2>/dev/null) || return 1
            fi
            ;;
        pi)
            if [[ -n "$MODEL" ]]; then
                output=$(pi --model "$MODEL" --print --no-tools --no-session "$prompt" 2>/dev/null | tail -n 1) || return 1
            else
                output=$(pi --print --no-tools --no-session "$prompt" 2>/dev/null | tail -n 1) || return 1
            fi
            ;;
        *)
            return 1
            ;;
    esac

    BRANCH_NAME=$(printf "%s" "$output" | tr -d '`"' | awk 'NF { last=$0 } END { print last }' | xargs)

    [[ -n "$BRANCH_NAME" ]]
}

improve_prompt_template() {
    local output_path
    output_path=$(default_prompt_improvement_output_path)

    log_info "📝 Improving prompt template..."
    echo "   Prompt source: $PROMPT_SOURCE"
    echo "   Proposal path: $output_path"

    if [[ "$DRY_RUN" == "true" ]]; then
        echo "   [DRY-RUN] Would ask $AGENT to generate an improved prompt proposal."
        return
    fi

    if [[ -e "$output_path" ]]; then
        die "Prompt improvement output already exists: $output_path"
    fi

    local improved_prompt
    improved_prompt=$(generate_improved_prompt_template) || \
        die "Could not generate improved prompt with $AGENT"

    mkdir -p "$(dirname "$output_path")"
    printf "%s\n" "$improved_prompt" > "$output_path"
    log_success "   ✅ Prompt improvement written"
    echo "   Review the proposal and copy it to the active prompt file if accepted."
}

build_launch_command() {
    LAUNCH_CWD=""
    LAUNCH_CMD=()
    validate_model_selection_support "launch"

    case "$AGENT" in
        claude)
            LAUNCH_CWD="$WORKTREE_PATH"
            if [[ -n "$MODEL" ]]; then
                LAUNCH_CMD=(claude --model "$MODEL" --dangerously-skip-permissions "$AGENT_PROMPT")
            else
                LAUNCH_CMD=(claude --dangerously-skip-permissions "$AGENT_PROMPT")
            fi
            ;;
        codex)
            if [[ -n "$MODEL" ]]; then
                LAUNCH_CMD=(codex --model "$MODEL" --cd "$WORKTREE_PATH" --dangerously-bypass-approvals-and-sandbox "$AGENT_PROMPT")
            else
                LAUNCH_CMD=(codex --cd "$WORKTREE_PATH" --dangerously-bypass-approvals-and-sandbox "$AGENT_PROMPT")
            fi
            ;;
        kimi)
            if [[ -n "$MODEL" ]]; then
                LAUNCH_CMD=(kimi --model "$MODEL" --work-dir "$WORKTREE_PATH" --yolo -p "$AGENT_PROMPT")
            else
                LAUNCH_CMD=(kimi --work-dir "$WORKTREE_PATH" --yolo -p "$AGENT_PROMPT")
            fi
            ;;
        pi)
            LAUNCH_CWD="$WORKTREE_PATH"
            if [[ -n "$MODEL" ]]; then
                LAUNCH_CMD=(pi --model "$MODEL" "$AGENT_PROMPT")
            else
                LAUNCH_CMD=(pi "$AGENT_PROMPT")
            fi
            ;;
        none)
            ;;
        *)
            die "Unknown agent: $AGENT"
            ;;
    esac
}
