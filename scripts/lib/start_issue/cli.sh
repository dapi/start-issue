# shellcheck shell=bash disable=SC2034
require_value() {
    local option="$1"
    local value="${2:-}"

    if [[ -z "$value" || "$value" == -* ]]; then
        die "$option requires a value."
    fi
}

parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --repo|-r)
                require_value "$1" "${2:-}"
                REPO="$2"
                shift 2
                ;;
            --base|-b)
                require_value "$1" "${2:-}"
                BASE_BRANCH="$2"
                shift 2
                ;;
            --worktree-dir|-w)
                require_value "$1" "${2:-}"
                WORKTREE_DIR="$2"
                WORKTREE_DIR_SOURCE="CLI"
                shift 2
                ;;
            --agent)
                require_value "$1" "${2:-}"
                AGENT_CLI="$2"
                shift 2
                ;;
            --no-agent|--no-claude)
                AGENT_CLI="none"
                shift
                ;;
            --prompt-file)
                require_value "$1" "${2:-}"
                PROMPT_FILE_CLI="$2"
                shift 2
                ;;
            --prompt)
                require_value "$1" "${2:-}"
                PROMPT_INLINE_CLI="$2"
                shift 2
                ;;
            --improve-prompt)
                IMPROVE_PROMPT=true
                shift
                ;;
            --prompt-output-file)
                require_value "$1" "${2:-}"
                PROMPT_IMPROVEMENT_OUTPUT_FILE="$2"
                shift 2
                ;;
            --no-init)
                NO_INIT=true
                shift
                ;;
            --flat)
                FLAT_WORKTREE=true
                shift
                ;;
            --command|-c)
                require_value "$1" "${2:-}"
                INITIAL_COMMAND="$2"
                shift 2
                ;;
            --ai)
                FAST_MODE=false
                shift
                ;;
            --project)
                if [[ "$INIT_SCOPE" == "user" ]]; then
                    die "Use either --project or --user, not both."
                fi
                INIT_SCOPE="project"
                shift
                ;;
            --user)
                if [[ "$INIT_SCOPE" == "project" ]]; then
                    die "Use either --project or --user, not both."
                fi
                INIT_SCOPE="user"
                shift
                ;;
            --force)
                INIT_FORCE=true
                shift
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --version|-v)
                show_version
                exit 0
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            -*)
                die "Unknown option: $1. Use --help for usage."
                ;;
            *)
                if [[ "$1" == "init" && -z "$ISSUE_INPUT" && "$INIT_CONFIG" == "false" ]]; then
                    INIT_CONFIG=true
                elif [[ "$INIT_CONFIG" == "true" ]]; then
                    die "Unexpected argument for init: $1"
                elif [[ -z "$ISSUE_INPUT" ]]; then
                    ISSUE_INPUT="$1"
                else
                    die "Unexpected argument: $1"
                fi
                shift
                ;;
        esac
    done

    if [[ "$INIT_CONFIG" == "true" ]]; then
        return
    fi

    if [[ -n "$INIT_SCOPE" || "$INIT_FORCE" == "true" ]]; then
        die "--project, --user, and --force are only valid with init."
    fi

    if [[ -z "$ISSUE_INPUT" ]]; then
        MISSING_ISSUE=true
    fi
}
