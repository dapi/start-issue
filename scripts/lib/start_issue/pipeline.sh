# shellcheck shell=bash disable=SC2153
start_agent_session() {
    if [[ "$AGENT" == "none" ]]; then
        print_manual_next_steps
        return
    fi

    if [[ "$HUMAN_GATE_MODE" == "true" ]]; then
        run_codex_human_gate_session
        return $?
    fi

    log_info "🚀 Starting $AGENT agent session..."

    if [[ "$DRY_RUN" == "true" ]]; then
        print_dry_run_launch_command
        return
    fi

    print_session_header
    build_launch_command

    if [[ -n "$LAUNCH_CWD" ]]; then
        cd "$LAUNCH_CWD" || exit
    fi

    exec "${LAUNCH_CMD[@]}"
}

handle_missing_issue_mode() {
    if [[ "$IMPROVE_PROMPT" == "true" ]]; then
        die "--improve-prompt requires <issue-url-or-number>. Example: start-issue 123 --improve-prompt"
    fi

    detect_project_root_if_available
    resolve_agent
    resolve_model
    resolve_prompt_template
    show_missing_issue_summary
    echo ""
    show_missing_issue_help
    echo ""
    show_current_configuration
    exit 1
}

run_start_issue_pipeline() {
    check_core_dependencies
    check_git_repo
    detect_project_root
    parse_issue_input "$ISSUE_INPUT"
    detect_repo_from_remote
    detect_base_branch
    resolve_agent
    resolve_model
    validate_human_gate_mode
    check_selected_agent_dependency
    resolve_prompt_template
    validate_prompt_improvement_mode

    print_selected_configuration
    fetch_issue

    if [[ "$IMPROVE_PROMPT" == "true" ]]; then
        improve_prompt_template
        return
    fi

    rename_zellij_tab
    generate_branch_name
    create_worktree
    run_init_script
    render_prompt_template
    start_agent_session
}

run_update_mode() {
    run_update_pipeline
}
