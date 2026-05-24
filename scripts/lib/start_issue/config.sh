# shellcheck shell=bash disable=SC2034
check_selected_agent_dependency() {
    if [[ "$AGENT" == "none" || "$DRY_RUN" == "true" ]]; then
        return
    fi

    if ! command -v "$AGENT" &> /dev/null; then
        die "$AGENT CLI not found. Install it or use --agent none."
    fi
}

validate_agent() {
    case "$AGENT" in
        claude|codex|kimi|pi|none)
            ;;
        *)
            die "Unknown agent: $AGENT. Valid agents: claude, codex, kimi, pi, none."
            ;;
    esac
}

validate_model_config() {
    if [[ -z "$MODEL" ]]; then
        die "Model config is empty. Remove the empty model config or set a value."
    fi
}

resolve_agent() {
    local project_agent_file="$PROJECT_ROOT/.start-issue/agent"
    local user_agent_file="$HOME/.config/start-issue/agent"

    if [[ -n "$AGENT_CLI" ]]; then
        AGENT="$AGENT_CLI"
        AGENT_SOURCE="CLI"
    elif [[ -f "$project_agent_file" ]]; then
        AGENT=$(read_first_config_value "$project_agent_file")
        AGENT_SOURCE="$project_agent_file"
    elif [[ -f "$user_agent_file" ]]; then
        AGENT=$(read_first_config_value "$user_agent_file")
        AGENT_SOURCE="$user_agent_file"
    elif [[ -n "${START_ISSUE_AGENT:-}" ]]; then
        AGENT=$(trim "$START_ISSUE_AGENT")
        AGENT_SOURCE="START_ISSUE_AGENT"
    else
        AGENT="claude"
        AGENT_SOURCE="built-in default"
    fi

    if [[ -z "$AGENT" ]]; then
        die "Agent config is empty. Valid agents: claude, codex, kimi, pi, none."
    fi

    validate_agent
}

resolve_model() {
    local project_model_file="$PROJECT_ROOT/.start-issue/model"
    local user_model_file="$HOME/.config/start-issue/model"

    if [[ -n "$MODEL_CLI" ]]; then
        MODEL=$(trim "$MODEL_CLI")
        MODEL_SOURCE="CLI"
    elif [[ -f "$project_model_file" ]]; then
        MODEL=$(read_first_config_value "$project_model_file")
        MODEL_SOURCE="$project_model_file"
    elif [[ -f "$user_model_file" ]]; then
        MODEL=$(read_first_config_value "$user_model_file")
        MODEL_SOURCE="$user_model_file"
    elif [[ -n "${START_ISSUE_MODEL:-}" ]]; then
        MODEL=$(trim "$START_ISSUE_MODEL")
        MODEL_SOURCE="START_ISSUE_MODEL"
    else
        MODEL=""
        MODEL_SOURCE="built-in default"
    fi

    if [[ -n "$MODEL_SOURCE" && "$MODEL_SOURCE" != "built-in default" ]]; then
        validate_model_config
    fi
}

default_portable_prompt_template() {
    cat << 'EOF'
Implement GitHub issue {ISSUE_URL} in this worktree.

Context:
- Repo: {REPO}
- Issue: #{ISSUE_NUMBER}
- Title: {ISSUE_TITLE}
- Branch: {BRANCH_NAME}
- Worktree: {WORKTREE_PATH}

Start by reading the issue with gh if needed. Follow repository instructions. Keep changes scoped. Run relevant tests or checks. Summarize changed files and verification before finishing.
EOF
}

default_claude_prompt_template() {
    if [[ -n "$INITIAL_COMMAND" ]]; then
        printf "%s {ISSUE_URL}" "$INITIAL_COMMAND"
    else
        printf "/task-router:route-task {ISSUE_URL}"
    fi
}

read_prompt_file() {
    local path="$1"

    if [[ ! -f "$path" ]]; then
        die "Prompt file not found: $path"
    fi

    if [[ ! -r "$path" ]]; then
        die "Prompt file is not readable: $path"
    fi

    PROMPT_TEMPLATE=$(< "$path")
}

resolve_prompt_template() {
    local project_prompt_file="$PROJECT_ROOT/.start-issue/prompt.md"
    local user_prompt_file="$HOME/.config/start-issue/prompt.md"

    PROMPT_LOCATION=""
    PROMPT_TEMPLATE_PATH=""

    if [[ -n "$PROMPT_FILE_CLI" && -n "$PROMPT_INLINE_CLI" ]]; then
        die "Use either --prompt-file or --prompt, not both."
    fi

    if [[ -n "$PROMPT_FILE_CLI" ]]; then
        read_prompt_file "$PROMPT_FILE_CLI"
        PROMPT_SOURCE="CLI --prompt-file: $PROMPT_FILE_CLI"
        PROMPT_LOCATION=$(absolute_path "$PROMPT_FILE_CLI")
        PROMPT_TEMPLATE_PATH="$PROMPT_LOCATION"
    elif [[ -n "$PROMPT_INLINE_CLI" ]]; then
        PROMPT_TEMPLATE="$PROMPT_INLINE_CLI"
        PROMPT_SOURCE="CLI --prompt"
        PROMPT_LOCATION="inline CLI argument"
    elif [[ -f "$project_prompt_file" ]]; then
        read_prompt_file "$project_prompt_file"
        PROMPT_SOURCE="$project_prompt_file"
        PROMPT_LOCATION="$project_prompt_file"
        PROMPT_TEMPLATE_PATH="$PROMPT_LOCATION"
    elif [[ -f "$user_prompt_file" ]]; then
        read_prompt_file "$user_prompt_file"
        PROMPT_SOURCE="$user_prompt_file"
        PROMPT_LOCATION="$user_prompt_file"
        PROMPT_TEMPLATE_PATH="$PROMPT_LOCATION"
    elif [[ -n "${START_ISSUE_PROMPT_FILE:-}" || -n "${START_ISSUE_PROMPT:-}" ]]; then
        if [[ -n "${START_ISSUE_PROMPT_FILE:-}" && -n "${START_ISSUE_PROMPT:-}" ]]; then
            die "Use either START_ISSUE_PROMPT_FILE or START_ISSUE_PROMPT, not both."
        fi

        if [[ -n "${START_ISSUE_PROMPT_FILE:-}" ]]; then
            read_prompt_file "$START_ISSUE_PROMPT_FILE"
            PROMPT_SOURCE="START_ISSUE_PROMPT_FILE: $START_ISSUE_PROMPT_FILE"
            PROMPT_LOCATION=$(absolute_path "$START_ISSUE_PROMPT_FILE")
            PROMPT_TEMPLATE_PATH="$PROMPT_LOCATION"
        else
            PROMPT_TEMPLATE="$START_ISSUE_PROMPT"
            PROMPT_SOURCE="START_ISSUE_PROMPT"
            PROMPT_LOCATION="START_ISSUE_PROMPT environment variable"
        fi
    else
        if [[ "$AGENT" == "claude" ]]; then
            PROMPT_TEMPLATE=$(default_claude_prompt_template)
            PROMPT_SOURCE="built-in Claude command"
        else
            PROMPT_TEMPLATE=$(default_portable_prompt_template)
            PROMPT_SOURCE="built-in portable prompt"
        fi
        PROMPT_LOCATION="$SCRIPT_PATH"
    fi
}

render_prompt_template() {
    local rendered="$PROMPT_TEMPLATE"

    rendered="${rendered//\{ISSUE_URL\}/$ISSUE_URL}"
    rendered="${rendered//\{ISSUE_NUMBER\}/$ISSUE_NUMBER}"
    rendered="${rendered//\{ISSUE_TITLE\}/$ISSUE_TITLE}"
    rendered="${rendered//\{ISSUE_BODY\}/$ISSUE_BODY}"
    rendered="${rendered//\{ISSUE_LABELS\}/$ISSUE_LABELS}"
    rendered="${rendered//\{REPO\}/$REPO}"
    rendered="${rendered//\{BRANCH_NAME\}/$BRANCH_NAME}"
    rendered="${rendered//\{WORKTREE_PATH\}/$WORKTREE_PATH}"
    rendered="${rendered//\{BASE_BRANCH\}/$BASE_BRANCH}"

    AGENT_PROMPT="$rendered"
}
