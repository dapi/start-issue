#!/usr/bin/env bash

set -euo pipefail

REPO="${START_ISSUE_REPOSITORY:-dapi/start-issue}"
PREFIX="${PREFIX:-$HOME/.local}"
BINDIR="${BINDIR:-$PREFIX/bin}"
TARGET="${TARGET:-$BINDIR/start-issue}"
ASSET_URL="${START_ISSUE_ASSET_URL:-https://github.com/$REPO/releases/latest/download/start-issue}"
CHECKSUM_URL="${START_ISSUE_CHECKSUM_URL:-https://github.com/$REPO/releases/latest/download/start-issue.sha256}"
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

# This script is deliberately self-contained: the documented installation
# command pipes it to Bash, where no repository directory or sibling modules
# are available.
release_fetch() {
    local url="$1"
    local output="$2"

    if command -v curl >/dev/null 2>&1; then
        if [[ "${RELEASE_FETCH_VERBOSE:-0}" == "1" ]]; then
            curl -fL -v "$url" -o "$output"
        else
            curl -fsSL "$url" -o "$output"
        fi
        return
    fi

    if command -v wget >/dev/null 2>&1; then
        if [[ "${RELEASE_FETCH_VERBOSE:-0}" == "1" ]]; then
            wget -O "$output" "$url"
        else
            wget -qO "$output" "$url"
        fi
        return
    fi

    die "Neither curl nor wget is installed."
}

release_sha256_file() {
    local path="$1"

    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum "$path" | awk '{ print $1 }'
        return
    fi

    if command -v shasum >/dev/null 2>&1; then
        shasum -a 256 "$path" | awk '{ print $1 }'
        return
    fi

    if command -v openssl >/dev/null 2>&1; then
        openssl dgst -sha256 "$path" | awk '{ print $NF }'
        return
    fi

    die "No SHA-256 tool found. Install sha256sum, shasum, or openssl."
}

release_install_verified_asset() {
    local asset_url="$1"
    local checksum_url="$2"
    local target_path="$3"
    local tmpdir
    local tmpfile
    local checksum_file
    local expected_checksum
    local actual_checksum
    local cleanup_cmd

    tmpdir="$(mktemp -d)"
    printf -v cleanup_cmd 'rm -rf %q' "$tmpdir"
    # shellcheck disable=SC2064
    trap "$cleanup_cmd" RETURN

    tmpfile="$tmpdir/start-issue"
    checksum_file="$tmpdir/start-issue.sha256"

    if declare -F debug >/dev/null 2>&1; then
        debug "Fetching $asset_url -> $tmpfile"
    fi
    release_fetch "$asset_url" "$tmpfile" || die "Failed to download release asset: $asset_url"
    if declare -F debug >/dev/null 2>&1; then
        debug "Fetching $checksum_url -> $checksum_file"
    fi
    release_fetch "$checksum_url" "$checksum_file" || die "Failed to download release checksum: $checksum_url"

    if declare -F debug >/dev/null 2>&1; then
        debug "Verifying checksum"
    fi
    expected_checksum="$(awk '{ print $1; exit }' "$checksum_file")"
    actual_checksum="$(release_sha256_file "$tmpfile")"

    if [[ -z "$expected_checksum" ]]; then
        die "Downloaded checksum file is empty."
    fi

    if [[ "$expected_checksum" != "$actual_checksum" ]]; then
        die "Checksum verification failed."
    fi

    if declare -F debug >/dev/null 2>&1; then
        debug "Installing binary into $target_path"
    fi
    mkdir -p "$(dirname "$target_path")"
    install -m 0755 "$tmpfile" "$target_path" || die "Failed to install updated release to $target_path"
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
