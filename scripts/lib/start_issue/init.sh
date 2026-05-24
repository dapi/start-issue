# shellcheck shell=bash disable=SC2034
resolve_init_model() {
    local model_file="$1"

    if [[ -f "$model_file" && "$INIT_FORCE" != "true" ]]; then
        MODEL=$(read_first_config_value "$model_file")
        MODEL_SOURCE="$model_file (existing)"
    elif [[ -n "$MODEL_CLI" ]]; then
        MODEL=$(trim "$MODEL_CLI")
        MODEL_SOURCE="CLI"
    else
        MODEL=""
        MODEL_SOURCE="built-in default"
    fi

    if [[ -n "$MODEL_SOURCE" && "$MODEL_SOURCE" != "built-in default" ]]; then
        validate_model_config
    fi
}

resolve_init_agent() {
    local agent_file="$1"

    if [[ -f "$agent_file" && "$INIT_FORCE" != "true" ]]; then
        AGENT=$(read_first_config_value "$agent_file")
        AGENT_SOURCE="$agent_file (existing)"
    elif [[ -n "$AGENT_CLI" ]]; then
        AGENT=$(trim "$AGENT_CLI")
        AGENT_SOURCE="CLI"
    else
        AGENT="claude"
        AGENT_SOURCE="built-in default"
    fi

    if [[ -z "$AGENT" ]]; then
        die "Agent config is empty. Valid agents: claude, codex, kimi, pi, none."
    fi

    validate_agent
}

resolve_init_prompt_template() {
    if [[ -n "$PROMPT_FILE_CLI" && -n "$PROMPT_INLINE_CLI" ]]; then
        die "Use either --prompt-file or --prompt, not both."
    fi

    if [[ -n "$PROMPT_FILE_CLI" ]]; then
        read_prompt_file "$PROMPT_FILE_CLI"
        PROMPT_SOURCE="CLI --prompt-file: $PROMPT_FILE_CLI"
    elif [[ -n "$PROMPT_INLINE_CLI" ]]; then
        PROMPT_TEMPLATE="$PROMPT_INLINE_CLI"
        PROMPT_SOURCE="CLI --prompt"
    elif [[ "$AGENT" == "claude" ]]; then
        PROMPT_TEMPLATE=$(default_claude_prompt_template)
        PROMPT_SOURCE="built-in Claude command"
    else
        PROMPT_TEMPLATE=$(default_portable_prompt_template)
        PROMPT_SOURCE="built-in portable prompt"
    fi
}

select_init_scope() {
    local project_available=false
    local choice=""

    if git rev-parse --git-dir &> /dev/null; then
        detect_project_root
        project_available=true
    fi

    echo "Initialize start-issue configuration:"
    if [[ "$project_available" == "true" ]]; then
        echo "  1) Project config ($PROJECT_ROOT/.start-issue)"
        echo "  2) User config ($HOME/.config/start-issue)"
        if ! read -r -p "Choice [1/2]: " choice; then
            die "No init scope selected. Use --project or --user."
        fi

        case "$choice" in
            1|p|P|project|Project)
                INIT_SCOPE="project"
                ;;
            2|u|U|user|User)
                INIT_SCOPE="user"
                ;;
            *)
                die "Invalid init scope: $choice. Use --project or --user."
                ;;
        esac
    else
        echo "  1) User config ($HOME/.config/start-issue)"
        if ! read -r -p "Choice [1]: " choice; then
            die "No init scope selected. Use --user outside a git repository."
        fi

        case "$choice" in
            ""|1|u|U|user|User)
                INIT_SCOPE="user"
                ;;
            *)
                die "Project config requires a git repository. Use --user outside a git repository."
                ;;
        esac
    fi
}

write_init_file() {
    local path="$1"
    local content="$2"
    local label="$3"

    if [[ -e "$path" && "$INIT_FORCE" != "true" ]]; then
        log_warn "$label already exists, keeping: $path"
        return
    fi

    if [[ "$DRY_RUN" == "true" ]]; then
        echo "   [DRY-RUN] Would write $label: $path"
        return
    fi

    mkdir -p "$(dirname "$path")"
    printf "%s\n" "$content" > "$path"
    log_success "   Wrote $label: $path"
}

write_init_model_file() {
    local path="$1"

    if [[ -n "$MODEL" ]]; then
        write_init_file "$path" "$MODEL" "model config"
        return
    fi

    if [[ -e "$path" && "$INIT_FORCE" != "true" ]]; then
        log_warn "model config already exists, keeping: $path"
        return
    fi

    if [[ "$DRY_RUN" == "true" ]]; then
        if [[ -e "$path" ]]; then
            echo "   [DRY-RUN] Would remove model config: $path"
        else
            echo "   [DRY-RUN] No model config to write (built-in default: unset)"
        fi
        return
    fi

    if [[ -e "$path" ]]; then
        rm -f "$path"
        log_success "   Removed model config: $path"
    fi
}

run_config_init() {
    if [[ -z "$INIT_SCOPE" ]]; then
        select_init_scope
    fi

    case "$INIT_SCOPE" in
        project)
            check_git_repo
            detect_project_root
            ;;
        user) ;;
        *)
            die "Invalid init scope: $INIT_SCOPE. Use --project or --user."
            ;;
    esac

    local target_dir=""
    local scope_label=""

    if [[ "$INIT_SCOPE" == "project" ]]; then
        target_dir="$PROJECT_ROOT/.start-issue"
        scope_label="project config"
    else
        target_dir="$HOME/.config/start-issue"
        scope_label="user config"
    fi

    resolve_init_agent "$target_dir/agent"
    resolve_init_model "$target_dir/model"
    validate_model_selection_support "launch"
    resolve_init_prompt_template

    echo "Scope: $scope_label"
    echo "Directory: $target_dir"
    echo "Agent: $AGENT"
    echo "Agent source: $AGENT_SOURCE"
    echo "Model: ${MODEL:-<unset>}"
    echo "Model source: $MODEL_SOURCE"
    echo "Prompt source: $PROMPT_SOURCE"
    echo ""

    write_init_file "$target_dir/agent" "$AGENT" "agent config"
    write_init_model_file "$target_dir/model"
    write_init_file "$target_dir/prompt.md" "$PROMPT_TEMPLATE" "prompt template"
}
