#!/usr/bin/env bash

set -euo pipefail

REPO="${START_ISSUE_REPOSITORY:-dapi/start-issue}"
PREFIX="${PREFIX:-$HOME/.local}"
BINDIR="${BINDIR:-$PREFIX/bin}"
TARGET="${TARGET:-$BINDIR/start-issue}"
ASSET_URL="${START_ISSUE_ASSET_URL:-https://github.com/$REPO/releases/latest/download/start-issue}"
CHECKSUM_URL="${START_ISSUE_CHECKSUM_URL:-https://github.com/$REPO/releases/latest/download/start-issue.sha256}"
INSTALL_TMPDIR=""
DEBUG=0

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

cleanup() {
    if [[ -n "$INSTALL_TMPDIR" ]]; then
        rm -rf -- "$INSTALL_TMPDIR"
    fi
}

fetch() {
    local url="$1"
    local output="$2"

    debug "Fetching $url -> $output"

    if command -v curl >/dev/null 2>&1; then
        if [[ "$DEBUG" -eq 1 ]]; then
            curl -fL -v "$url" -o "$output"
        else
            curl -fsSL "$url" -o "$output"
        fi
        return
    fi

    if command -v wget >/dev/null 2>&1; then
        if [[ "$DEBUG" -eq 1 ]]; then
            wget -O "$output" "$url"
        else
            wget -qO "$output" "$url"
        fi
        return
    fi

    die "Neither curl nor wget is installed."
}

sha256_file() {
    local path="$1"

    if command -v sha256sum >/dev/null 2>&1; then
        debug "Using sha256sum for checksum verification"
        sha256sum "$path" | awk '{ print $1 }'
        return
    fi

    if command -v shasum >/dev/null 2>&1; then
        debug "Using shasum for checksum verification"
        shasum -a 256 "$path" | awk '{ print $1 }'
        return
    fi

    if command -v openssl >/dev/null 2>&1; then
        debug "Using openssl for checksum verification"
        openssl dgst -sha256 "$path" | awk '{ print $NF }'
        return
    fi

    die "No SHA-256 tool found. Install sha256sum, shasum, or openssl."
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
    local tmpfile
    local checksum_file
    local expected_checksum
    local actual_checksum

    parse_args "$@"

    if [[ "$DEBUG" -eq 1 ]]; then
        PS4='+ install.sh:${LINENO}: '
        set -x
        debug "Repository: $REPO"
        debug "Install target: $TARGET"
        debug "Asset URL: $ASSET_URL"
        debug "Checksum URL: $CHECKSUM_URL"
    fi

    INSTALL_TMPDIR="$(mktemp -d)"
    trap cleanup EXIT

    tmpfile="$INSTALL_TMPDIR/start-issue"
    checksum_file="$INSTALL_TMPDIR/start-issue.sha256"

    log "Downloading latest release from $REPO"
    fetch "$ASSET_URL" "$tmpfile"
    fetch "$CHECKSUM_URL" "$checksum_file"

    debug "Verifying checksum"
    expected_checksum="$(awk '{ print $1; exit }' "$checksum_file")"
    actual_checksum="$(sha256_file "$tmpfile")"

    if [[ -z "$expected_checksum" ]]; then
        die "Downloaded checksum file is empty."
    fi

    if [[ "$expected_checksum" != "$actual_checksum" ]]; then
        die "Checksum verification failed."
    fi

    debug "Installing binary into $TARGET"
    mkdir -p "$BINDIR"
    install -m 0755 "$tmpfile" "$TARGET"

    log "Installed: $TARGET"
    log "Version: $("$TARGET" --version)"
}

main "$@"
