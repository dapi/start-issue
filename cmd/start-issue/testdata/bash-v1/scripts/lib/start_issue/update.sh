# shellcheck shell=bash disable=SC2034
UPDATE_REPO="${START_ISSUE_REPOSITORY:-dapi/start-issue}"
UPDATE_MODE=false
LATEST_RELEASE_JSON=""
LATEST_RELEASE_TAG=""
LATEST_RELEASE_ASSET_URL=""
LATEST_RELEASE_CHECKSUM_URL=""
CURRENT_INSTALL_VERSION=""
CURRENT_INSTALL_VERSION_NORMALIZED=""
LATEST_RELEASE_VERSION_NORMALIZED=""

check_update_dependencies() {
    if ! command -v gh &> /dev/null; then
        die "gh CLI not found. Install: https://cli.github.com"
    fi

    if ! gh auth status &> /dev/null; then
        die "gh not authenticated. Run: gh auth login"
    fi

    if ! command -v jq &> /dev/null; then
        die "jq not found. Please install jq."
    fi

    if ! command -v install &> /dev/null; then
        die "install command not found."
    fi
}

resolve_current_installation() {
    CURRENT_INSTALL_VERSION="$VERSION"
    CURRENT_INSTALL_VERSION_NORMALIZED="$(release_normalize_version "$CURRENT_INSTALL_VERSION")"
}

fetch_latest_release_metadata() {
    log_info "🔍 Resolving latest release for $UPDATE_REPO..."

    LATEST_RELEASE_JSON="$(gh api "repos/$UPDATE_REPO/releases/latest" 2>/dev/null)" || \
        die "Failed to resolve the latest GitHub release for $UPDATE_REPO."

    LATEST_RELEASE_TAG="$(printf "%s" "$LATEST_RELEASE_JSON" | jq -r '.tag_name // empty')"
    LATEST_RELEASE_ASSET_URL="$(printf "%s" "$LATEST_RELEASE_JSON" | jq -r '.assets[] | select(.name == "start-issue") | .browser_download_url' | head -n 1)"
    LATEST_RELEASE_CHECKSUM_URL="$(printf "%s" "$LATEST_RELEASE_JSON" | jq -r '.assets[] | select(.name == "start-issue.sha256") | .browser_download_url' | head -n 1)"

    if [[ -z "$LATEST_RELEASE_TAG" ]]; then
        die "Latest release metadata for $UPDATE_REPO did not include a tag name."
    fi

    if [[ -z "$LATEST_RELEASE_ASSET_URL" ]]; then
        die "Latest release $LATEST_RELEASE_TAG does not include a start-issue asset."
    fi

    if [[ -z "$LATEST_RELEASE_CHECKSUM_URL" ]]; then
        die "Latest release $LATEST_RELEASE_TAG does not include a start-issue.sha256 asset."
    fi

    LATEST_RELEASE_VERSION_NORMALIZED="$(release_normalize_version "$LATEST_RELEASE_TAG")"
}

print_update_status() {
    echo "Executable: $SCRIPT_PATH"
    echo "Installed version: v$CURRENT_INSTALL_VERSION_NORMALIZED"
    echo "Latest release: $LATEST_RELEASE_TAG"
}

run_update_pipeline() {
    local comparison
    local installed_version_output

    check_update_dependencies
    resolve_current_installation
    fetch_latest_release_metadata

    echo ""
    print_update_status

    comparison="$(release_compare_versions "$CURRENT_INSTALL_VERSION_NORMALIZED" "$LATEST_RELEASE_VERSION_NORMALIZED")"

    if [[ "$comparison" == "0" ]]; then
        log_success "✅ start-issue is already up to date."
        return
    fi

    if [[ "$comparison" == "1" ]]; then
        log_success "✅ Installed version is newer than the latest published release. No update needed."
        return
    fi

    if [[ "$DRY_RUN" == "true" ]]; then
        echo "   [DRY-RUN] Would download: $LATEST_RELEASE_ASSET_URL"
        echo "   [DRY-RUN] Would verify with: $LATEST_RELEASE_CHECKSUM_URL"
        echo "   [DRY-RUN] Would install to: $SCRIPT_PATH"
        return
    fi

    log_info "📥 Downloading and installing $LATEST_RELEASE_TAG..."
    release_install_verified_asset "$LATEST_RELEASE_ASSET_URL" "$LATEST_RELEASE_CHECKSUM_URL" "$SCRIPT_PATH"

    installed_version_output="$("$SCRIPT_PATH" --version 2>/dev/null)" || \
        die "Updated executable installed, but version verification failed."

    log_success "✅ Updated start-issue at: $SCRIPT_PATH"
    echo "Version: $installed_version_output"
}
