---
title: "FT-013: Decision Log"
doc_kind: feature
doc_function: decision_log
purpose: "Feature-local decisions and conflict resolutions for the issue #35/#13 self-update slice."
derived_from:
  - brief.md
  - design.md
  - ../../../cmd/start-issue/main.go
  - ../../../README.md
  - ../../../doc/spec.md
status: active
audience: humans_and_agents
---

# FT-013: Decision Log

## Decisions

### `DL-01` Issue #35 is reconciled to FT-013

- Date: 2026-08-04
- Status: accepted
- Context: Issue #35 requests a release-backed self-update command for the installed CLI. The feature inventory identifies it as a duplicate of the existing issue #13 self-update slice.
- Decision: FT-013 remains the single feature package; issue #35 is added as provenance and no second package is created.
- Evidence: GitHub issue #35, `memory-bank/features/missing.md`, and the existing FT-013 package.
- Consequence: Scope, design, plan, and verification have one canonical owner.

### `DL-02` Current Go runtime is the implementation baseline

- Date: 2026-08-04
- Status: accepted
- Context: Legacy FT-013 documents referenced `scripts/start-issue` and Bash tests, but the current repository entrypoint is `cmd/start-issue/main.go` and the current test surfaces are Go and parity tests.
- Decision: Active feature documents use the current Go implementation and tests as grounding; legacy `feature.md` and `solution.md` are archived redirects.
- Evidence: `cmd/start-issue/main.go`, `cmd/start-issue/main_test.go`, `cmd/start-issue/parity_integration_test.go`, `Makefile`.
- Consequence: Plan steps and evidence no longer claim unstarted Bash work.

### `DL-03` Release metadata and platform assets are the update contract

- Date: 2026-08-04
- Status: accepted
- Context: Issue #35 requires the latest GitHub Release, a binary, and SHA-256 verification. The current Go CLI resolves release metadata with `gh`, selects the current-platform asset, and downloads `checksums.txt`.
- Decision: Use `gh api repos/<repository>/releases/latest`, the current-platform binary asset, and `checksums.txt`; keep the documented repository override available.
- Evidence: `updateMode`, `githubRelease.assetURLs`, `releaseAssetName`, `validChecksum`, README self-update section, and `doc/spec.md`.
- Consequence: Verification covers both metadata resolution and asset integrity without reintroducing the legacy shell installer as a runtime dependency.

### `DL-04` The running executable is the update target

- Date: 2026-08-04
- Status: accepted
- Context: Issue #35 requires updating the executable invoked by the user. Multiple installations may exist, and the current implementation resolves `os.Executable()` then evaluates symlinks.
- Decision: Update the resolved executable path for the current process; never guess another PATH entry or silently fall back to the default install path.
- Evidence: `runningExecutablePath`, `SD-01`, and issue #35 user experience/acceptance text.
- Consequence: The command updates the installation actually invoked.

### `DL-05` Update is upgrade-only

- Date: 2026-08-04
- Status: accepted
- Context: The requirement says current or newer installations must exit successfully without downloading; unconditional “latest release” replacement could downgrade a newer binary.
- Decision: Compare semantic versions; equal or newer is a successful no-op.
- Evidence: `compareVersions`, `updateMode`, `SC-03`, and `SC-04`.
- Consequence: Self-update never replaces a newer installed version.

### `DL-06` Replacement requires checksum and staged-version verification

- Date: 2026-08-04
- Status: accepted
- Context: A checksum verifies bytes, while a release asset can still be mislabeled or incompatible. Replacement must not destroy the working binary on a failed update.
- Decision: Download to memory, verify `checksums.txt`, stage beside the target with executable mode, verify staged `--version` against the release tag, then rename atomically.
- Evidence: `validChecksum`, `stageBinary`, `verifyStagedBinary`, `installVerifiedUpdate`, and focused tests.
- Consequence: checksum, version, permission, or rename failures leave the existing target untouched.

### `DL-07` Update mode is independent of repository context

- Date: 2026-08-04
- Status: accepted
- Context: Issue #35 explicitly requires operation outside a git repository; ordinary issue-starting flow has separate repository/worktree prerequisites.
- Decision: Dispatch update mode before ordinary repo/worktree orchestration and require only the update-specific external boundary.
- Evidence: `main`, `runMode`, `updateMode`, `SC-06`, and `TestRunModeUpdateDoesNotRequireHomeDirectory`.
- Consequence: update can run outside git while ordinary issue-start behavior remains isolated.

## Conflict Resolution

### Legacy FT-013 documents vs current Go implementation

The archived `feature.md` / `solution.md` described Bash paths and issue #13
only. They conflicted with the current Go change surface and issue #35
provenance. `DL-02` resolves this by making `brief.md` and `design.md` active
owners and retaining the old files only as migration redirects.

### Generic checksum wording vs current release manifest

Issue #35 says “its SHA-256 checksum”; the repository's current release contract
uses a platform binary plus `checksums.txt`. `DL-03` makes that existing
contract explicit without changing issue scope.
