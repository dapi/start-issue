# shellcheck shell=bash disable=SC2034
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

release_normalize_version() {
    local version="${1:-}"
    version="${version#v}"
    printf "%s" "$version"
}

release_compare_versions() {
    local left
    local right
    local i
    local max_parts
    local left_part
    local right_part
    local IFS=.

    left="$(release_normalize_version "${1:-}")"
    right="$(release_normalize_version "${2:-}")"

    read -r -a left_parts <<< "$left"
    read -r -a right_parts <<< "$right"

    max_parts="${#left_parts[@]}"
    if [[ ${#right_parts[@]} -gt $max_parts ]]; then
        max_parts="${#right_parts[@]}"
    fi

    for ((i = 0; i < max_parts; i++)); do
        left_part="${left_parts[i]:-0}"
        right_part="${right_parts[i]:-0}"

        if ((10#$left_part > 10#$right_part)); then
            printf "1"
            return
        fi

        if ((10#$left_part < 10#$right_part)); then
            printf -- "-1"
            return
        fi
    done

    printf "0"
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
