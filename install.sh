#!/usr/bin/env bash

set -euo pipefail

REPO="${START_ISSUE_REPOSITORY:-dapi/start-issue}"
PREFIX="${PREFIX:-$HOME/.local}"
BINDIR="${BINDIR:-$PREFIX/bin}"
TARGET="${TARGET:-$BINDIR/start-issue}"
ASSET_URL="${START_ISSUE_ASSET_URL:-https://github.com/$REPO/releases/latest/download/start-issue}"
CHECKSUM_URL="${START_ISSUE_CHECKSUM_URL:-https://github.com/$REPO/releases/latest/download/start-issue.sha256}"
DEBUG=0
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# shellcheck source=scripts/lib/start_issue/release.sh
source "$SCRIPT_DIR/scripts/lib/start_issue/release.sh"

log() {
    printf '%s\n' "$1"
}

debug() {
    if [[ "$DEBUG" -eq 1 ]]; then
        printf 'DEBUG: %s\n' "$1" >&2
    fi
}

die() {
    printf 'Error: %s\n' "$1" >&2
    exit 1
}

usage() {
    cat <<'EOF'
Usage: install.sh [--debug]

Options:
  --debug  Enable verbose installer diagnostics.
  --help   Show this help.
EOF
}

parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --debug)
                DEBUG=1
                ;;
            --help|-h)
                usage
                exit 0
                ;;
            *)
                die "Unknown argument: $1"
                ;;
        esac
        shift
    done
}

main() {
    parse_args "$@"

    if [[ "$DEBUG" -eq 1 ]]; then
        PS4='+ install.sh:${LINENO}: '
        set -x
        export RELEASE_FETCH_VERBOSE=1
        debug "Repository: $REPO"
        debug "Install target: $TARGET"
        debug "Asset URL: $ASSET_URL"
        debug "Checksum URL: $CHECKSUM_URL"
    fi

    log "Downloading latest release from $REPO"
    mkdir -p "$BINDIR"
    release_install_verified_asset "$ASSET_URL" "$CHECKSUM_URL" "$TARGET"

    log "Installed: $TARGET"
    log "Version: $("$TARGET" --version)"
}

main "$@"
