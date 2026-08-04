---
title: "FT-013: Self-update command for installed CLI"
doc_kind: feature
doc_function: canonical
purpose: "Canonical problem-space and verification contract for issue #35, reconciled with the already delivered issue #13 slice."
derived_from:
  - https://github.com/dapi/start-issue/issues/35
  - https://github.com/dapi/start-issue/issues/13
  - ../../flows/feature-flow.md
  - ../../product/context.md
  - ../../engineering/testing-policy.md
  - ../../../README.md
  - ../../../README.ru.md
  - ../../../doc/spec.md
status: active
delivery_status: done
audience: humans_and_agents
must_not_define:
  - selected_design
  - implementation_sequence
---

# FT-013: Self-update command for installed CLI

## What

### Problem

An installed `start-issue` binary needs a release-backed way to update itself
without requiring a source checkout. Issue #35 restates the self-update slice
originally tracked by issue #13; the existing FT-013 package is therefore the
single feature owner.

### Outcome

Users can run `start-issue update` or `start-issue --update` to compare the
running installation with the latest `dapi/start-issue` GitHub Release and
upgrade that installation when a newer compatible release exists.

## Scope

- `REQ-01` Support `start-issue update` and `start-issue --update` as equivalent entry points.
- `REQ-02` Resolve the latest published release for `dapi/start-issue` through GitHub Release metadata.
- `REQ-03` Use the executable resolved for the current invocation as the update target and version source.
- `REQ-04` Treat an equal or newer installed version as a successful no-op without downgrading.
- `REQ-05` For an available update, download the current-platform binary and checksum manifest, verify both, and install with executable permissions.
- `REQ-06` Fail with actionable errors for missing access/tooling, release lookup, missing assets, download, checksum, staged-version, or installation failures.
- `REQ-07` Work outside a git repository and leave the ordinary issue-start workflow unchanged.
- `REQ-08` Document both entry points, update states, prerequisites, and failure behavior in the public README/spec surfaces.

## Non-Scope

- `NS-01` No automatic update during ordinary issue-start commands.
- `NS-02` No arbitrary release selection or downgrade command.
- `NS-03` No package-manager integration and no change to local-source `make install` semantics.
- `NS-04` No second feature package for issue #35; issue #35 is reconciled to FT-013.

## Constraints / Assumptions

- `CON-01` The release source and repository are `dapi/start-issue`, unless the existing documented repository override is intentionally used for tests or compatible deployments.
- `CON-02` The update path must not require repository, worktree, issue, or agent context.
- `CON-03` Installation must be staged and verified before replacement so a failed update does not replace the existing executable.
- `CON-04` Platform-specific release assets and `checksums.txt` are the current release contract documented by the repository.

## Design Requirement Decision

| Decision | Reason | Downstream owner |
| --- | --- | --- |
| `Design required: yes` | The feature changes CLI mode parsing, external GitHub Release interaction, executable replacement, checksum verification, and failure handling. | `design.md` |

## Verify

### Exit Criteria

- `EC-01` Both command forms enter the same update workflow.
- `EC-02` Current, newer, and update-available versions produce the specified no-op or update result.
- `EC-03` A successful update replaces only the invoked executable after checksum and staged-version verification, preserving executable permissions.
- `EC-04` Lookup, download, checksum, staged-version, permission, and unsupported-platform failures are actionable and non-zero where applicable.
- `EC-05` The workflow works outside git and existing issue-start behavior remains covered.
- `EC-06` README/spec documentation matches the implemented Go CLI contract.

### Acceptance Scenarios

- `SC-01` Older installed binary + `update` installs the latest current-platform release.
- `SC-02` Older installed binary + `--update` has the same result as `update`.
- `SC-03` Equal versions exit `0` and do not download or replace the executable.
- `SC-04` Installed version newer than latest exits `0` and does not downgrade.
- `SC-05` Lookup, asset, checksum, staged-version, or install failure exits with an actionable error and leaves the old executable intact.
- `SC-06` Update from outside a git repository does not invoke repo/worktree discovery.
- `SC-07` Existing issue-start invocation remains behaviorally compatible.

### Checks and Evidence

| Check ID | Covers | Procedure | Evidence |
| --- | --- | --- | --- |
| `CHK-01` | `REQ-01`–`REQ-07`, `SC-01`–`SC-07` | `go test ./...` | `EVID-01` |
| `CHK-02` | `REQ-08`, `EC-06` | `make test` and `git diff --check` | `EVID-02` |

`EVID-01` is focused Go/parity test output covering release lookup, version
ordering, checksum/staged verification, executable replacement, no-op paths,
failure paths, and outside-git execution. `EVID-02` is repository-level check
output and documentation/index validation.

### Traceability Matrix

| Requirement | Acceptance | Checks |
| --- | --- | --- |
| `REQ-01` | `SC-01`, `SC-02` | `CHK-01` |
| `REQ-02`–`REQ-06` | `SC-01`–`SC-06` | `CHK-01` |
| `REQ-07` | `SC-06`, `SC-07` | `CHK-01` |
| `REQ-08` | `EC-06` | `CHK-02` |
