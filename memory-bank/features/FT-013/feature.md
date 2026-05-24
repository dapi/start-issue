---
title: "FT-013: Self-update from latest GitHub release"
doc_kind: feature
doc_function: canonical
purpose: "Canonical feature document for adding a release-backed self-update workflow to start-issue. Owns only the problem space and verification contract."
derived_from:
  - https://github.com/dapi/start-issue/issues/13
  - ../../../README.md
  - ../../../README.ru.md
  - ../../../doc/spec.md
  - ../../../install.sh
status: active
delivery_status: in_progress
audience: humans_and_agents
must_not_define:
  - selected_design
  - implementation_sequence
---

# FT-013: Self-update from latest GitHub release

## What

### Problem

`start-issue` can currently be installed from the latest GitHub Release through `install.sh` or the documented manual download flow, but the CLI has no built-in way to upgrade an existing installation to the latest published release. Users must detect new releases and reinstall on their own.

As of 2026-05-24, issue #13 states that the latest published release is `v1.11.1` published on 2026-05-18, and the new workflow must resolve updates from GitHub Releases rather than from the local checkout state.

### Outcome

`start-issue` exposes a first-class self-update workflow through both `start-issue update` and `start-issue --update`, compares the running installation against the latest published GitHub release, installs an update when needed, and reports the result clearly.

### Scope

- `REQ-01` Support both `start-issue update` and `start-issue --update` as equivalent entry points.
- `REQ-02` Resolve the latest published release from GitHub Releases for `dapi/start-issue`, not from the local checkout state.
- `REQ-03` Detect the currently installed version from the executable path of the running `start-issue` invocation.
- `REQ-04` If the installed version already matches the latest release, exit successfully with a clear "already up to date" message.
- `REQ-05` If a newer release exists, download and install that release using the project's release-backed installation contract while preserving executable permissions.
- `REQ-06` Fail with a clear error message when release lookup, download, checksum verification, or installation fails.
- `REQ-07` Keep the existing `start-issue ISSUE` workflow unchanged.
- `REQ-08` Document the new update workflow in `README.md`, `README.ru.md`, and `doc/spec.md`, including both command-entry forms and expected output states.

### Non-Scope

- `NS-01` Do not add auto-update on ordinary issue-starting commands.
- `NS-02` Do not change how `make install` builds from the local checkout.
- `NS-03` Do not add support for selecting arbitrary target releases; this issue covers only updating to the latest published release.
- `NS-04` Do not broaden the feature into a package-manager integration beyond the project's existing release-backed install path.

### Constraints

- `CON-01` The update source must be a published GitHub Release asset or release-backed installer path.
- `CON-02` The update workflow must work outside a git repository.
- `CON-03` Multiple `start-issue` executables may exist in `PATH`; the feature must define a safe default target without guessing about other installations.
- `CON-04` The normal `start-issue ISSUE` path must remain behaviorally unchanged.

## Verify

### Exit Criteria

- `EC-01` `start-issue update` and `start-issue --update` both enter the same update workflow.
- `EC-02` The workflow compares the running executable version against the latest GitHub Release and either reports "already up to date" or installs the newer release.
- `EC-03` The workflow updates the intended local executable path, preserves executable permissions, and fails clearly on lookup, download, checksum, or install failures.
- `EC-04` Documentation and spec describe the new command forms, install/update behavior, and output states without contradicting the implementation.
- `EC-05` Existing issue-starting behavior remains intact and automated coverage passes.

### Acceptance Scenarios

- `SC-01` Given `start-issue` is installed at an older version, when the user runs `start-issue update`, then the command installs the latest GitHub release and `start-issue --version` reports that release version.
- `SC-02` Given `start-issue` is installed at an older version, when the user runs `start-issue --update`, then the command performs the same update behavior as the subcommand form.
- `SC-03` Given the installed version already matches the latest release, when the user runs either update form, then the command exits `0` and reports that no update was necessary.
- `SC-04` Given release lookup, download, checksum verification, or installation fails, when the user runs either update form, then the command exits non-zero with an actionable error message.
- `SC-05` Given the user runs `start-issue 123`, then the existing issue workflow behaves exactly as before.
- `SC-06` Given the user runs an update command outside a git repository, then the command still resolves the running executable path and performs update checks without repo discovery.
- `SC-07` Given the installed version and the latest release tag differ only by an optional leading `v`, when the user runs either update form, then the command treats them as the same version and exits successfully without reinstalling.
- `SC-08` Given the installed version is newer than the latest published GitHub release, when the user runs either update form, then the command exits `0`, reports that no update is needed, and does not downgrade the executable.

### Traceability Matrix

| Requirement ID | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- |
| `REQ-01` | `EC-01`, `SC-01`, `SC-02` | `CHK-01`, `CHK-02` | `EVID-01` |
| `REQ-02` | `EC-02`, `SC-01`, `SC-02`, `SC-03`, `SC-04`, `SC-07`, `SC-08` | `CHK-01`, `CHK-02` | `EVID-01` |
| `REQ-03` | `EC-02`, `EC-03`, `SC-01`, `SC-03`, `SC-06`, `SC-07`, `SC-08` | `CHK-01`, `CHK-02` | `EVID-01` |
| `REQ-04` | `EC-02`, `SC-03`, `SC-07`, `SC-08` | `CHK-02` | `EVID-01` |
| `REQ-05` | `EC-02`, `EC-03`, `SC-01`, `SC-02` | `CHK-01`, `CHK-02` | `EVID-01` |
| `REQ-06` | `EC-03`, `SC-04` | `CHK-01`, `CHK-02` | `EVID-01` |
| `REQ-07` | `EC-05`, `SC-05` | `CHK-02` | `EVID-01` |
| `REQ-08` | `EC-04` | `CHK-03` | `EVID-01` |

### Checks

| Check ID | Covers | How to check | Expected |
| --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `EC-02`, `EC-03` | `bash -n scripts/start-issue && shellcheck install.sh scripts/start-issue scripts/lib/start_issue/*.sh` | CLI parsing, update helpers, and installer-related code remain syntactically valid and shellcheck-clean. |
| `CHK-02` | `EC-01`, `EC-02`, `EC-03`, `EC-05`, `SC-01` - `SC-08` | `mise exec -- bats test` | Automated coverage proves both update entry points, version normalization, no-downgrade behavior, success/no-op/failure paths, outside-git support, and unchanged issue workflow. |
| `CHK-03` | `EC-04` | `git diff --check` | Documentation/spec edits are internally consistent and whitespace-clean. |

### Test Matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | Local terminal output from syntax and shellcheck runs. |
| `CHK-02` | `EVID-01` | Local terminal output from Bats runs and CI output after branch update. |
| `CHK-03` | `EVID-01` | Local terminal output from `git diff --check`. |

### Evidence

- `EVID-01` Verification command output showing update workflow coverage, version normalization, no-downgrade and no-op behavior, failure handling, and unchanged existing workflow.

### Evidence Contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | Local and CI command output | implementer / CI | Terminal output and GitHub Actions job | `CHK-01`, `CHK-02`, `CHK-03` |
