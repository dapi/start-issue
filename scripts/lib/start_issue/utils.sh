# shellcheck shell=bash disable=SC2034
log_info() {
    echo -e "${BLUE}$1${NC}"
}

log_success() {
    echo -e "${GREEN}$1${NC}"
}

log_warn() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}" >&2
}

die() {
    log_error "$1"
    exit 1
}

show_version() {
    echo "start-issue v$VERSION"
}

shell_join() {
    local out=""
    local quoted
    local arg

    for arg in "$@"; do
        printf -v quoted "%q" "$arg"
        if [[ -n "$out" ]]; then
            out+=" "
        fi
        out+="$quoted"
    done

    printf "%s" "$out"
}

trim() {
    local value="$1"
    value="${value#"${value%%[![:space:]]*}"}"
    value="${value%"${value##*[![:space:]]}"}"
    printf "%s" "$value"
}

absolute_path() {
    local path="$1"

    if [[ "$path" == /* ]]; then
        printf "%s" "$path"
    else
        printf "%s/%s" "$(pwd)" "$path"
    fi
}

read_first_config_value() {
    local file="$1"
    awk '
        {
            sub(/#.*/, "")
            gsub(/^[[:space:]]+|[[:space:]]+$/, "")
        }
        NF {
            print
            exit
        }
    ' "$file"
}

prompt_preview() {
    local preview="$PROMPT_TEMPLATE"

    preview="${preview//$'\n'/ }"
    preview=$(printf "%s" "$preview" | sed 's/[[:space:]][[:space:]]*/ /g')
    preview=$(trim "$preview")

    if [[ ${#preview} -gt 140 ]]; then
        preview="${preview:0:137}..."
    fi

    printf "%s" "$preview"
}

check_core_dependencies() {
    if ! command -v git &> /dev/null; then
        die "git not found. Please install git."
    fi

    if ! command -v gh &> /dev/null; then
        die "gh CLI not found. Install: https://cli.github.com"
    fi

    if ! gh auth status &> /dev/null; then
        die "gh not authenticated. Run: gh auth login"
    fi

    if ! command -v jq &> /dev/null; then
        die "jq not found. Please install jq."
    fi
}

check_git_repo() {
    if ! git rev-parse --git-dir &> /dev/null; then
        die "Not in a git repository"
    fi
}

detect_project_root() {
    PROJECT_ROOT=$(git rev-parse --show-toplevel 2>/dev/null || pwd)
}

detect_project_root_if_available() {
    PROJECT_ROOT="$(pwd)"

    if command -v git &> /dev/null && git rev-parse --git-dir &> /dev/null; then
        detect_project_root
    fi
}
