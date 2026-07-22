---
title: "FT-017: Decision Log"
doc_kind: feature-support
doc_function: reference
purpose: "Records FPF analysis and accepted local decisions for FT-017. It does not own feature scope, selected design, acceptance criteria, or execution sequence."
derived_from:
  - brief.md
  - ../../../.github/workflows/ci.yml
  - ../../../.github/workflows/release.yml
  - ../../../install.sh
status: active
audience: humans_and_agents
must_not_define:
  - ft_016_scope
  - ft_016_selected_design
  - ft_016_acceptance_criteria
  - implementation_sequence
---

# FT-017: Decision Log

## Purpose and Ownership

This log records why `DEC-01` remains open. The canonical owner of the blocker and the verify contract is [brief.md](brief.md). A selected solution belongs in a future `design.md`, not here.

## DL-01 — Multi-platform Go release distribution contract

**Status:** accepted on 2026-07-22 by feature requester.

### FPF framing

- **Bounded context:** distribution is separate from CLI-semantic parity. It owns the relationship among a compiled artifact, release assets, installer/update selection, and the platform on which a user executes the artifact.
- **Evidence boundary:** facts below come only from the current repository and issue #34. The issue requests a Go binary but defines neither supported OS/architecture targets nor asset-selection rules.
- **Decision criterion:** provide Go releases for the requester-selected operating systems with explicit platform assets, verifiable integrity, and no inferred reduction of platform support.

### Available facts

1. `install.sh` downloads one fixed asset named `start-issue` and one fixed checksum file named `start-issue.sha256`.
2. `.github/workflows/release.yml` builds the current sole release asset on `ubuntu-latest`.
3. `.github/workflows/ci.yml` verifies installation on both `ubuntu-latest` and `macos-latest`.
4. A Bash release artifact is portable across those CI operating systems; a Go executable is platform-specific.
5. Issue #34 requires Go to become the primary distribution artifact and requires installation/release workflows to publish it successfully, but does not state the intended OS/architecture matrix or compatibility policy.

### Decision

| Area | Accepted contract |
| --- | --- |
| Target matrix | `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, and `windows/amd64`. The operating systems are requester-selected; the architecture set follows the explicit `dapi/port-selector` release pattern. |
| Build/release | Use GoReleaser v2 with `CGO_ENABLED=0`, one statically built executable per target, `start-issue-<os>-<arch>` asset names, and a SHA-256 `checksums.txt` manifest. |
| POSIX install | Adapt the referenced install-script strategy: detect `uname -s`/`uname -m`, select the matching asset, download it, verify its checksum from `checksums.txt`, and install it under the public name `start-issue`. |
| Windows delivery | Publish `start-issue-windows-amd64.exe` as a first-class release asset and document manual download/PATH installation. The existing POSIX shell installer is not a Windows installer. |
| Cutover | No separate human release-approval gate. The normal tag-triggered release proceeds only after `CHK-01` through `CHK-03` are green. |

### Resolution rationale

The requester directly chose macOS, Linux, and Windows and delegated release-strategy selection to this feature. The selected GoReleaser layout and target architecture set are grounded in the referenced `dapi/port-selector` repository: its `.goreleaser.yml` uses the exact five targets, `CGO_ENABLED=0`, binary-format archives, and `checksums.txt`; its installer performs POSIX OS/architecture detection. The decision preserves explicit asset integrity while avoiding a false claim that the POSIX installer supports Windows.

### Rejected alternatives

- A single cross-platform `start-issue` Go asset is rejected: compiled Go executables are platform-specific.
- A narrower target matrix is rejected: the requester selected all three operating systems and the referenced strategy supplies the matching explicit matrix.
- A release approval gate is rejected: the requester explicitly said it is unnecessary; automated evidence gates remain mandatory.

## DL-02 — Go toolchain and Windows update boundary

**Status:** accepted on 2026-07-22 by feature owner under delegated release-strategy choice.

### FPF framing and facts

- The toolchain is an execution-environment contract, not a user-facing CLI capability; it must be deterministic in local, CI, and release paths.
- The referenced `dapi/port-selector` release pattern pins `go 1.21` in `go.mod` and GitHub Actions. This repository currently has no Go toolchain contract.
- A POSIX process can replace its executable through the existing install/update style; Windows generally locks a running executable. The referenced release strategy documents a Windows binary download rather than a shell installer.

### Decision

1. Pin Go `1.21` in `go.mod`, `mise.toml`, and CI/release setup for this migration.
2. The initial Windows contract is binary release plus manual installation and manual update: `start-issue update` on Windows must not try to overwrite its running `.exe`; it returns a clear instruction naming the matching release asset. POSIX retains verified automatic install/update behavior.

### Rationale and risk control

Go 1.21 is the explicit baseline in the chosen release reference and yields a single reproducible toolchain contract. The Windows manual-update behavior avoids an unsafe or undeclared helper-process design. It is a documented platform-specific delivery difference, not a hidden parity exception, because the Bash baseline has no Windows runtime contract.
