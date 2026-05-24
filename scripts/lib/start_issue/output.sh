# shellcheck shell=bash disable=SC2153
model_display_value() {
    if [[ -n "$MODEL" ]]; then
        printf "%s" "$MODEL"
    else
        printf "%s" "<unset>"
    fi
}

show_help() {
    show_version
    cat << 'EOF'

Start working on a GitHub issue with git worktree and a configurable agent

Usage: start-issue <issue-url-or-number> [options]
       start-issue init [options]

Arguments:
  <issue-url-or-number>    GitHub issue URL or issue number
                           Examples: 123, https://github.com/owner/repo/issues/123
  init                     Create default start-issue configuration

Options:
  --repo, -r <owner/repo>    Repository (default: detected from git remote)
  --base, -b <branch>        Base branch (default: main or master)
  --worktree-dir, -w <dir>   Directory for worktrees
                             Default: START_ISSUE_WORKTREE_DIR or ~/worktrees
  --flat                     Use flat worktree structure (replace / with - in path)
  --agent <name>             Agent to launch: claude, codex, kimi, pi, none
                             With init: default agent to write
  --model <name>             Model to use for the selected agent
                             With init: default model config to write
  --no-agent                 Only create worktree, do not start an agent session
  --no-claude                Compatibility alias for --no-agent
  --prompt <text>            Prompt template for the launched agent
  --prompt-file <path>       Prompt template file for the launched agent
  --improve-prompt           Ask the selected agent to improve the selected
                             prompt template and write a reviewable proposal
  --prompt-output-file <path>
                             Output path for --improve-prompt proposal
  --no-init                  Skip init.sh execution
  --command, -c <cmd>        Compatibility: initial command for Claude default launch
  --ai                       Use the selected agent for branch name generation
                             Default: fast bash heuristics
  --project                  With init: write .start-issue config in this repo
  --user                     With init: write config in ~/.config/start-issue
  --force                    With init: overwrite existing config files
  --dry-run                  Show what would be done without executing
  --version, -v              Show version
  --help, -h                 Show this help

Agent selection precedence:
  CLI --agent / --no-agent
  .start-issue/agent in the git root
  ~/.config/start-issue/agent
  START_ISSUE_AGENT
  built-in default: claude

Model selection precedence:
  CLI --model
  .start-issue/model in the git root
  ~/.config/start-issue/model
  START_ISSUE_MODEL
  built-in default: unset (agent CLI decides)

Prompt template precedence:
  CLI --prompt-file / --prompt
  .start-issue/prompt.md in the git root
  ~/.config/start-issue/prompt.md
  START_ISSUE_PROMPT_FILE / START_ISSUE_PROMPT
  built-in default

Prompt improvement:
  --improve-prompt uses the selected agent to generate a complete improved
  prompt template proposal. It does not overwrite the active prompt template.
  File-backed prompts write next to the source as *.improved.md by default.
  Built-in and inline prompts write to .start-issue/prompt.improved.md by
  default. Use --prompt-output-file to choose another proposal path.

Prompt variables:
  {ISSUE_URL}, {ISSUE_NUMBER}, {ISSUE_TITLE}, {ISSUE_BODY}, {ISSUE_LABELS},
  {REPO}, {BRANCH_NAME}, {WORKTREE_PATH}, {BASE_BRANCH}

Environment variables:
  START_ISSUE_AGENT
  START_ISSUE_MODEL
  START_ISSUE_PROMPT
  START_ISSUE_PROMPT_FILE
  START_ISSUE_WORKTREE_DIR
  START_ISSUE_DUMP_PROMPT

Examples:
  start-issue 123
  start-issue https://github.com/owner/repo/issues/123
  start-issue 123 --repo owner/repo --base develop
  start-issue 123 --agent codex
  start-issue 123 --agent codex --model gpt-5.2
  start-issue 123 --agent claude --model sonnet
  start-issue 123 --agent kimi --prompt-file .start-issue/prompt.md
  start-issue 123 --no-agent              # Only create worktree
  start-issue 123 --command "/debug"      # Claude command prefix
  start-issue 123 --flat                  # Flat worktree path
  start-issue 123 --dry-run
  start-issue init
  start-issue init --project --agent codex --model gpt-5.2
  start-issue init --project --agent codex
  start-issue init --user --force
EOF
}

show_current_configuration() {
    echo "Current configuration:"
    echo "  Agent: $AGENT ($AGENT_SOURCE)"
    echo "  Model: $(model_display_value) ($MODEL_SOURCE)"
    echo "  Prompt source: $PROMPT_SOURCE"
    echo "  Prompt location: $PROMPT_LOCATION"
    print_agent_model_file_locations "  "
    echo "  Worktree dir: $WORKTREE_DIR ($WORKTREE_DIR_SOURCE)"
    print_prompt_file_locations "  "
}

show_missing_issue_summary() {
    echo "Error: missing issue URL or issue number"
    echo ""
    echo "Run \`start-issue --help\` for full usage and prompt variables."
}

show_missing_issue_help() {
    show_version
    cat << 'EOF'

Start working on a GitHub issue with git worktree and a configurable agent

Usage: start-issue <issue-url-or-number> [options]
       start-issue init [options]

Examples:
  start-issue 123
  start-issue https://github.com/owner/repo/issues/123
  start-issue 123 --agent codex
  start-issue init --project

Prompt variables:
  {ISSUE_URL}, {ISSUE_NUMBER}, {ISSUE_TITLE}, {ISSUE_BODY}, {ISSUE_LABELS},
  {REPO}, {BRANCH_NAME}, {WORKTREE_PATH}, {BASE_BRANCH}
EOF
}

print_agent_model_file_locations() {
    local indent="${1:-}"

    echo "${indent}Default agent/model files:"
    echo "${indent}  Project agent: $PROJECT_ROOT/.start-issue/agent"
    echo "${indent}  Project model: $PROJECT_ROOT/.start-issue/model"
    echo "${indent}  User agent: $HOME/.config/start-issue/agent"
    echo "${indent}  User model: $HOME/.config/start-issue/model"
}

print_prompt_file_locations() {
    local indent="${1:-}"

    echo "${indent}Default prompt files:"
    echo "${indent}  Project: $PROJECT_ROOT/.start-issue/prompt.md"
    echo "${indent}  User: $HOME/.config/start-issue/prompt.md"
}

print_session_header() {
    local term_width
    local min_width=60
    local display_path="${WORKTREE_PATH/#$HOME/\~}"
    local line1="Agent: $AGENT"
    local line2="Branch: $BRANCH_NAME"
    local line3="Issue: #$ISSUE_NUMBER - $ISSUE_TITLE"
    local line4="Path: $display_path"
    local max_content_width
    local h_line=""
    local i

    term_width=$(tput cols 2>/dev/null || echo 80)
    [[ $term_width -lt $min_width ]] && term_width=$min_width

    max_content_width=$((term_width - 4))
    [[ ${#line2} -gt $max_content_width ]] && line2="${line2:0:$((max_content_width - 3))}..."
    [[ ${#line3} -gt $max_content_width ]] && line3="${line3:0:$((max_content_width - 3))}..."
    [[ ${#line4} -gt $max_content_width ]] && line4="${line4:0:$((max_content_width - 3))}..."

    for ((i = 0; i < term_width - 2; i++)); do
        h_line+="─"
    done

    pad_line() {
        local text="$1"
        local padding=$((term_width - 4 - ${#text}))
        printf "│ %s%*s │\n" "$text" "$padding" ""
    }

    echo ""
    printf "╭%s╮\n" "$h_line"
    pad_line "$line1"
    printf "├%s┤\n" "$h_line"
    pad_line "$line2"
    pad_line "$line3"
    pad_line "$line4"
    printf "╰%s╯\n" "$h_line"
    echo ""
}

print_manual_next_steps() {
    log_success "✅ Worktree ready at: $WORKTREE_PATH"
    echo ""
    echo "Selected agent: none ($AGENT_SOURCE)"
    echo "Resolved model: $(model_display_value) ($MODEL_SOURCE)"
    echo "Prompt source: $PROMPT_SOURCE"
    echo "To start working:"
    echo "  cd $(shell_join "$WORKTREE_PATH")"
    echo ""
    echo "Suggested agent commands:"
    echo "  claude"
    echo "  codex --cd $(shell_join "$WORKTREE_PATH")"
    echo "  kimi --work-dir $(shell_join "$WORKTREE_PATH")"
    echo "  pi"
}

print_dry_run_launch_command() {
    local cmd=""

    build_launch_command

    echo "   Agent: $AGENT ($AGENT_SOURCE)"
    echo "   Model: $(model_display_value) ($MODEL_SOURCE)"
    echo "   Prompt source: $PROMPT_SOURCE"
    echo "   Prompt length: ${#AGENT_PROMPT} chars"

    if [[ ${#AGENT_PROMPT} -gt 4000 && "${START_ISSUE_DUMP_PROMPT:-}" != "1" ]]; then
        echo "   Prompt omitted from command display because it is large."
        echo "   Set START_ISSUE_DUMP_PROMPT=1 to print the full rendered prompt."
        case "$AGENT" in
            claude)
                if [[ -n "$MODEL" ]]; then
                    cmd=$(shell_join claude --model "$MODEL" --dangerously-skip-permissions "<rendered prompt: ${#AGENT_PROMPT} chars>")
                else
                    cmd=$(shell_join claude --dangerously-skip-permissions "<rendered prompt: ${#AGENT_PROMPT} chars>")
                fi
                ;;
            codex)
                if [[ -n "$MODEL" ]]; then
                    cmd=$(shell_join codex --model "$MODEL" --cd "$WORKTREE_PATH" --dangerously-bypass-approvals-and-sandbox "<rendered prompt: ${#AGENT_PROMPT} chars>")
                else
                    cmd=$(shell_join codex --cd "$WORKTREE_PATH" --dangerously-bypass-approvals-and-sandbox "<rendered prompt: ${#AGENT_PROMPT} chars>")
                fi
                ;;
            kimi)
                if [[ -n "$MODEL" ]]; then
                    cmd=$(shell_join kimi --model "$MODEL" --work-dir "$WORKTREE_PATH" --yolo -p "<rendered prompt: ${#AGENT_PROMPT} chars>")
                else
                    cmd=$(shell_join kimi --work-dir "$WORKTREE_PATH" --yolo -p "<rendered prompt: ${#AGENT_PROMPT} chars>")
                fi
                ;;
            pi)
                if [[ -n "$MODEL" ]]; then
                    cmd=$(shell_join pi --model "$MODEL" "<rendered prompt: ${#AGENT_PROMPT} chars>")
                else
                    cmd=$(shell_join pi "<rendered prompt: ${#AGENT_PROMPT} chars>")
                fi
                ;;
            *)
                cmd=""
                ;;
        esac
    else
        cmd=$(shell_join "${LAUNCH_CMD[@]}")
    fi

    if [[ -n "$LAUNCH_CWD" ]]; then
        echo "   [DRY-RUN] Would run: cd $(shell_join "$LAUNCH_CWD") && $cmd"
    else
        echo "   [DRY-RUN] Would run: $cmd"
    fi
}

print_selected_configuration() {
    echo "Agent: $AGENT ($AGENT_SOURCE)"
    echo "Model: $(model_display_value) ($MODEL_SOURCE)"
    echo "Worktree directory: $WORKTREE_DIR ($WORKTREE_DIR_SOURCE)"
    echo "Prompt source: $PROMPT_SOURCE"
    echo "Prompt location: $PROMPT_LOCATION"
    print_agent_model_file_locations
    print_prompt_file_locations
    echo ""
}
