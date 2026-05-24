---
title: "FT-013: Decision Log"
doc_kind: feature
doc_function: decision_log
purpose: "Feature-local decisions for FT-013. Records how issue ambiguities and document conflicts were resolved."
derived_from:
  - feature.md
  - solution.md
  - ../../../install.sh
  - ../../../README.md
  - ../../../doc/spec.md
status: active
audience: humans_and_agents
---

# FT-013: Decision Log

## Decisions

### `DL-01` Release-backed installer contract is the update baseline

- Date: 2026-05-24
- Status: accepted
- Context:
  Issue #13 requires the update source to be the latest GitHub Release rather than the local checkout.
  The current repository already documents and implements a release-backed install path in `README.md`, `README.ru.md`, and `install.sh`.
- Decision:
  The self-update workflow reuses the same release-backed contract as `install.sh`: release asset plus checksum, verification before install, and installation into the project's standard binary target shape.
- Evidence:
  `install.sh` downloads `releases/latest/download/start-issue` and `start-issue.sha256`, verifies the checksum, and installs with mode `0755`.
  `README.md` documents that install path as the primary "latest release" installation flow.
- Consequence:
  Update behavior stays consistent with the documented installation story and avoids inventing a second release-install mechanism.

### `DL-02` The running executable path is the safe default update target

- Date: 2026-05-24
- Status: accepted
- Context:
  Issue #13 explicitly notes that multiple `start-issue` executables may exist in `PATH` and says the safest default is to update the resolved executable path for the current invocation.
- Decision:
  The update workflow detects the executable path of the running `start-issue` process and updates that path rather than guessing a different installation target.
- Evidence:
  The issue's design notes define this as the safest default.
  This matches the feature requirement to detect the currently installed version from the executable the user is running.
- Consequence:
  The command updates the installation the user actually invoked and avoids silently modifying another `PATH` entry.

### `DL-03` Update mode must not depend on git repository state

- Date: 2026-05-24
- Status: accepted
- Context:
  Issue #13 requires the command to work outside a git repository.
  The normal issue-starting workflow currently depends on git repo discovery, remote detection, and worktree planning.
- Decision:
  Update mode is treated as a separate top-level workflow that bypasses issue parsing, repo detection, base-branch detection, and worktree orchestration.
- Evidence:
  The issue says the update command should work outside a git repository.
  Current `doc/spec.md` shows the ordinary workflow requires git repository context for issue starts.
- Consequence:
  Update mode can run from any directory and does not create accidental coupling to the issue-starting path.

### `DL-04` Version comparison uses normalized tags, not raw surface strings

- Date: 2026-05-24
- Status: accepted
- Context:
  The current CLI exposes a bare numeric version such as `1.12.0` in `scripts/start-issue`.
  Issue #13 refers to GitHub release names with a `v` prefix such as `v1.11.1`.
  Raw string comparison would therefore misclassify equivalent versions.
- Decision:
  The update workflow compares normalized version strings by stripping one optional leading `v` from both the installed version and the latest release tag before equality and ordering checks.
- Evidence:
  `scripts/start-issue` currently defines `VERSION=\"1.12.0\"`.
  Issue #13 states the latest published release is `v1.11.1`.
- Consequence:
  The no-op path reflects semantic version equality rather than formatting differences, and tests must cover the normalization rule.

### `DL-05` Release lookup and asset download are separate contracts

- Date: 2026-05-24
- Status: accepted
- Context:
  Issue #13 requires the command to resolve the latest published release from GitHub Releases and also to install the release artifact.
  `install.sh` already defines how assets and checksum files are downloaded, but it does not need to compare versions first.
- Decision:
  FT-013 separates release metadata lookup from asset download:
  lookup resolves the canonical latest release tag and asset/checksum URLs;
  installer helpers then download, verify, and install those resolved assets.
- Evidence:
  The issue requires both latest-release resolution and version comparison.
  `install.sh` already provides the download and verification contract for the release asset and checksum.
- Consequence:
  The design keeps one source of truth for "what is latest" and a separate source of truth for "how to install it," reducing drift and simplifying tests.

### `DL-06` Self-update must not downgrade a newer local executable

- Date: 2026-05-24
- Status: accepted
- Context:
  The repository currently contains `VERSION="1.12.0"` in `scripts/start-issue`.
  Issue #13 states that on 2026-05-24 the latest published release is `v1.11.1`.
  A self-update command that blindly installed "latest published release" would therefore downgrade a newer local executable.
- Decision:
  If the running executable version is newer than the latest published GitHub release, the update workflow exits successfully with a no-op status and leaves the executable unchanged.
- Evidence:
  `scripts/start-issue` currently exposes version `1.12.0`.
  Issue #13 states that the latest published release is `v1.11.1`.
- Consequence:
  Self-update remains an upgrade-only workflow and does not clobber a newer local or pre-release installation.

## Conflict Resolution

- Resolved conflict: "standard install location used by the project" versus "update the resolved executable path for the current invocation."
  Conflicting sources:
  issue #13 requirements emphasize the standard install location;
  issue #13 design notes emphasize the running executable path as the safest default when multiple installations exist.
  Resolution:
  interpret the release-backed installer contract as the standard install behavior and the running executable path as the concrete target for self-update.
  Why this is consistent:
  it preserves the existing installation method while making the self-update target explicit and safe.

- Resolved conflict: bare CLI version strings versus `v`-prefixed release tags.
  Conflicting sources:
  `scripts/start-issue` exposes a bare numeric `VERSION`;
  issue #13 and GitHub Releases use `v`-prefixed release tags.
  Resolution:
  compare normalized versions rather than raw strings.
  Why this is consistent:
  it preserves the current CLI output surface while matching the release system's tagging convention.
