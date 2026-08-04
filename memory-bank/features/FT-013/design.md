---
title: "FT-013: Design"
doc_kind: feature
doc_function: canonical
purpose: "Canonical solution-space contract for the release-backed self-update command."
derived_from:
  - brief.md
  - decision-log.md
  - ../../../cmd/start-issue/main.go
  - ../../../cmd/start-issue/main_test.go
status: active
audience: humans_and_agents
must_not_define:
  - ft_013_scope
  - ft_013_acceptance_criteria
  - implementation_sequence
---

# FT-013: Design

## C4 Applicability

| C4 ID | Decision | Reason | Artifact |
| --- | --- | --- | --- |
| `C4-01` | `C1` | The CLI introduces a user-to-CLI interaction with GitHub Release metadata and release assets across a trust boundary. | This design's flow and contracts |

### `C4-01` System Context Artifact

| Element | Direction / boundary | Responsibility |
| --- | --- | --- |
| Developer | invokes `start-issue update` or `--update` | Starts an explicit update and observes status/errors |
| `start-issue` CLI | outbound authenticated `gh` call and HTTPS asset download | Resolves, compares, verifies, and installs the release |
| GitHub Releases / `gh` | release metadata and asset bytes enter the CLI trust boundary | Supplies latest tag, platform binary, and checksum manifest |
| Invoked executable path | local filesystem write after staged verification | Receives only the verified replacement binary |

## Selected Solution

- `SOL-01` Normalize `update` and `--update` in the existing Go parser and dispatch them to one top-level update mode.
- `SOL-02` Resolve the latest release through `gh api repos/<repository>/releases/latest`; use the current-platform binary asset and `checksums.txt`.
- `SOL-03` Compare the running version with the release tag using semantic ordering, including optional `v` normalization and source-build suffix handling.
- `SOL-04` Resolve `os.Executable()`, follow symlinks, stage the downloaded binary beside the target, verify its `--version`, then atomically rename it into place with mode `0755`.
- `SOL-05` Keep update mode before repository/worktree orchestration; Windows follows the documented manual path.

## Alternatives Considered

| Alternative | Why not selected |
| --- | --- |
| `ALT-01` Invoke `install.sh` from the CLI | It creates a second process contract and cannot own the running executable path as precisely as the Go implementation. |
| `ALT-02` Update the default `~/.local/bin/start-issue` path | It can modify a different installation when multiple binaries or symlinks exist. |
| `ALT-03` Download without staged verification | A checksum alone does not prove the downloaded binary reports the requested release tag. |

## Accepted Local Decisions

- `SD-01` The resolved running executable is the concrete update target; symlinks are evaluated before replacement.
- `SD-02` Equal or newer versions are successful no-ops, so self-update never downgrades.
- `SD-03` `gh` is the external release/auth boundary; release data and binary downloads remain separately testable.

## Contracts

| Contract | Input / Output | Semantics |
| --- | --- | --- |
| `CTR-01` | CLI args → update mode | `update` and `--update` are equivalent and mutually exclusive with issue input. |
| `CTR-02` | repository + platform → release tag and asset URLs | Latest release metadata must contain the platform binary and `checksums.txt`. |
| `CTR-03` | installed version + release tag → compare decision | Equal/newer means exit `0` without download or replacement; older means update. |
| `CTR-04` | verified binary + target path → replaced executable | Stage, verify `--version`, preserve executable mode, then rename; failures leave the target untouched. |

## Invariants

- `INV-01` Update mode does not require git repository discovery or issue context.
- `INV-02` A checksum or staged-version failure cannot replace the target executable.
- `INV-03` The ordinary issue-start path is not routed through update mode.

## Failure Modes

- `FM-01` Missing `gh` or authentication prevents release lookup with remediation in the error.
- `FM-02` Missing platform asset/checksum or failed download stops before replacement.
- `FM-03` Checksum or staged-version mismatch leaves the existing executable intact.
- `FM-04` Permission or rename failure is returned rather than falling back to another installation.

## Traceability

| Requirement | Solution | Contract / invariant | Failure |
| --- | --- | --- | --- |
| `REQ-01` | `SOL-01` | `CTR-01` | `INV-03` |
| `REQ-02`–`REQ-04` | `SOL-02`, `SOL-03` | `CTR-02`, `CTR-03` | `FM-01`, `FM-02` |
| `REQ-05`–`REQ-06` | `SOL-04` | `CTR-04`, `INV-02` | `FM-02`–`FM-04` |
| `REQ-07` | `SOL-05` | `INV-01`, `INV-03` | `FM-01` |
| `REQ-08` | existing public docs/spec | `CTR-01`–`CTR-04` | `FM-*` |
