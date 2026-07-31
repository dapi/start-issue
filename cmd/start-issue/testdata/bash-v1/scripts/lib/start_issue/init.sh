# shellcheck shell=bash disable=SC2034
USER_CONFIG_DIR="$HOME/.config/start-issue"
SETUP_AGENT_FILE_VALUE=""
SETUP_SAVE_PROMPT=false

confirm_yes_default() {
    local prompt="$1"
    local reply=""

    printf "%s" "$prompt"
    if ! read -r reply; then
        die "No response received."
    fi

    case "$reply" in
        ""|y|Y|yes|YES|Yes)
            return 0
            ;;
        n|N|no|NO|No)
            return 1
            ;;
        *)
            die "Invalid response: $reply. Use y or n."
            ;;
    esac
}

ensure_directory_exists() {
    local path="$1"
    local label="$2"

    if [[ "$DRY_RUN" == "true" ]]; then
        echo "   [DRY-RUN] Would create $label: $path"
        return
    fi

    mkdir -p "$path"
}

write_setup_file() {
    local path="$1"
    local content="$2"
    local label="$3"

    if [[ "$DRY_RUN" == "true" ]]; then
        echo "   [DRY-RUN] Would write $label: $path"
        return
    fi

    mkdir -p "$(dirname "$path")"
    printf "%s\n" "$content" > "$path"
    log_success "   Wrote $label: $path"
}

remove_setup_file_if_present() {
    local path="$1"
    local label="$2"

    if [[ ! -e "$path" ]]; then
        if [[ "$DRY_RUN" == "true" ]]; then
            echo "   [DRY-RUN] No $label to remove: $path"
        fi
        return
    fi

    if [[ "$DRY_RUN" == "true" ]]; then
        echo "   [DRY-RUN] Would remove $label: $path"
        return
    fi

    rm -f "$path"
    log_success "   Removed $label: $path"
}

select_setup_agent() {
    local choice=""

    echo "Select default agent:"
    echo "1) claude"
    echo "2) codex"
    echo "3) kimi"
    echo "4) pi"
    echo "5) skip"
    echo ""

    printf "Choice: "
    if ! read -r choice; then
        die "No setup agent selected."
    fi

    case "$choice" in
        1|claude|Claude)
            AGENT="claude"
            AGENT_SOURCE="setup selection"
            SETUP_AGENT_FILE_VALUE="claude"
            ;;
        2|codex|Codex)
            AGENT="codex"
            AGENT_SOURCE="setup selection"
            SETUP_AGENT_FILE_VALUE="codex"
            ;;
        3|kimi|Kimi)
            AGENT="kimi"
            AGENT_SOURCE="setup selection"
            SETUP_AGENT_FILE_VALUE="kimi"
            ;;
        4|pi|Pi)
            AGENT="pi"
            AGENT_SOURCE="setup selection"
            SETUP_AGENT_FILE_VALUE="pi"
            ;;
        5|skip|Skip|"")
            AGENT="claude"
            AGENT_SOURCE="built-in default"
            SETUP_AGENT_FILE_VALUE=""
            ;;
        *)
            die "Invalid setup choice: $choice. Use 1-5."
            ;;
    esac
}

resolve_setup_prompt_template() {
    if [[ "$AGENT" == "claude" ]]; then
        PROMPT_TEMPLATE=$(default_claude_prompt_template)
        PROMPT_SOURCE="built-in Claude command"
    else
        PROMPT_TEMPLATE=$(default_portable_prompt_template)
        PROMPT_SOURCE="built-in portable prompt"
    fi
}

show_setup_prompt_preview() {
    echo "Default prompt:"
    echo ""
    printf "%s\n" "$PROMPT_TEMPLATE"
    echo ""
}

run_setup_flow() {
    local target_dir="$USER_CONFIG_DIR"
    local agent_file="$target_dir/agent"
    local prompt_file="$target_dir/prompt.md"

    ensure_directory_exists "$target_dir" "user config directory"
    select_setup_agent
    resolve_setup_prompt_template
    show_setup_prompt_preview

    if confirm_yes_default "Save this prompt to $prompt_file? [Y/n] "; then
        SETUP_SAVE_PROMPT=true
    else
        SETUP_SAVE_PROMPT=false
    fi

    echo "Directory: $target_dir"
    echo "Agent: ${SETUP_AGENT_FILE_VALUE:-<unset>}"
    echo "Prompt source: $PROMPT_SOURCE"
    echo ""

    if [[ -n "$SETUP_AGENT_FILE_VALUE" ]]; then
        write_setup_file "$agent_file" "$SETUP_AGENT_FILE_VALUE" "agent config"
    else
        remove_setup_file_if_present "$agent_file" "agent config"
    fi

    if [[ "$SETUP_SAVE_PROMPT" == "true" ]]; then
        write_setup_file "$prompt_file" "$PROMPT_TEMPLATE" "prompt template"
    else
        remove_setup_file_if_present "$prompt_file" "prompt template"
    fi
}

materialize_first_run_marker() {
    ensure_directory_exists "$USER_CONFIG_DIR" "user config directory"
}

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

run_setup_mode() {
    run_setup_flow
}
