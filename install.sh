#!/usr/bin/env bash

set -euo pipefail

REPO="${START_ISSUE_REPOSITORY:-dapi/start-issue}"
PREFIX="${PREFIX:-$HOME/.local}"
BINDIR="${BINDIR:-$PREFIX/bin}"
TARGET="${TARGET:-$BINDIR/start-issue}"
ASSET_URL="https://github.com/$REPO/releases/latest/download/start-issue"
CHECKSUM_URL="https://github.com/$REPO/releases/latest/download/start-issue.sha256"

log() {
    printf '%s\n' "$1"
}

die() {
    printf 'Error: %s\n' "$1" >&2
    exit 1
}

fetch() {
    local url="$1"
    local output="$2"

    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$url" -o "$output"
        return
    fi

    if command -v wget >/dev/null 2>&1; then
        wget -qO "$output" "$url"
        return
    fi

    die "Neither curl nor wget is installed."
}

sha256_file() {
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

main() {
    local tmpdir
    local tmpfile
    local checksum_file
    local expected_checksum
    local actual_checksum

    tmpdir="$(mktemp -d)"
    trap 'rm -rf "$tmpdir"' EXIT

    tmpfile="$tmpdir/start-issue"
    checksum_file="$tmpdir/start-issue.sha256"

    log "Downloading latest release from $REPO"
    fetch "$ASSET_URL" "$tmpfile"
    fetch "$CHECKSUM_URL" "$checksum_file"

    expected_checksum="$(awk '{ print $1; exit }' "$checksum_file")"
    actual_checksum="$(sha256_file "$tmpfile")"

    if [[ -z "$expected_checksum" ]]; then
        die "Downloaded checksum file is empty."
    fi

    if [[ "$expected_checksum" != "$actual_checksum" ]]; then
        die "Checksum verification failed."
    fi

    mkdir -p "$BINDIR"
    install -m 0755 "$tmpfile" "$TARGET"

    log "Installed: $TARGET"
    log "Version: $("$TARGET" --version)"
}

main "$@"
