---
title: "FT-013: Solution"
doc_kind: feature
doc_function: canonical
purpose: "Canonical solution document for FT-013. Defines the selected self-update design without redefining feature scope or acceptance criteria."
derived_from:
  - feature.md
  - decision-log.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_013_scope
  - ft_013_acceptance_criteria
  - ft_013_delivery_status
  - implementation_sequence
---

# FT-013: Solution

## Selected Design

- `SOL-01` Add a dedicated update mode that can be entered either by the `update` subcommand or the `--update` flag and normalize both forms into one workflow.
- `SOL-02` Resolve the latest published release from GitHub Release assets for `dapi/start-issue`, not from the current checkout or branch.
- `SOL-03` Determine the installed version and target path from the currently running executable, so the command updates the same installation the user invoked.
- `SOL-03A` Normalize version strings before comparison so equivalent forms such as `1.11.1` and `v1.11.1` are treated as the same release.
- `SOL-03B` Treat an installed version newer than the latest published release as a successful no-op rather than a downgrade trigger.
- `SOL-04` Reuse the project's existing release-backed install contract: download the release asset plus checksum, verify integrity, and install with executable permissions preserved.
- `SOL-05` Keep the update workflow independent of git-repository detection and ordinary issue-starting orchestration.
- `SOL-06` Keep the ordinary `start-issue ISSUE` path behaviorally unchanged.
- `SOL-07` Document the command forms, update states, and installer relationship consistently across `README.md`, `README.ru.md`, and `doc/spec.md`.

## Requirement Mapping

| Requirement ID | Solution / architecture refs | Notes |
| --- | --- | --- |
| `REQ-01` | `SOL-01`, `CTR-01` | Both entry forms converge into one mode. |
| `REQ-02` | `SOL-02`, `CTR-02`, `DL-05` | Latest release is resolved from GitHub Releases. |
| `REQ-03` | `SOL-03`, `SOL-03A`, `DL-02`, `DL-04` | The running executable defines both current version source and safe update target. |
| `REQ-04` | `SOL-02`, `SOL-03`, `SOL-03A`, `SOL-03B`, `CTR-03`, `DL-04`, `DL-06` | Normalized version comparison drives the no-op success path. |
| `REQ-05` | `SOL-04`, `DL-01`, `DL-02` | Update reuses the release-backed installer contract and preserves mode. |
| `REQ-06` | `SOL-02`, `SOL-04`, `CTR-02`, `CTR-03` | Lookup/download/checksum/install failures stay explicit. |
| `REQ-07` | `SOL-05`, `SOL-06` | Update mode is isolated from the issue workflow. |
| `REQ-08` | `SOL-07` | Docs are part of the delivery surface. |

## To-Be Flow

1. Parse CLI input and normalize `update` and `--update` into update mode.
2. Resolve the executable path and current installed version from the running `start-issue`.
3. Resolve the latest published release metadata from GitHub Releases for `dapi/start-issue`, including the canonical release tag and asset URLs.
4. Normalize the installed version and release tag into the same comparison form before deciding whether an update is needed.
5. If already current, print a success message and exit `0`.
6. If the installed version is newer than the latest published release, print a success no-op message and exit `0` without downgrading.
7. If a newer release exists, download the release asset and checksum, verify integrity, install to the current executable path with executable permissions preserved, and print the resulting version.
8. If any release lookup, download, checksum, or install step fails, print a clear actionable error and exit non-zero.

## Contracts

| Contract ID | Related refs | Input / Output | Producer / Consumer | Notes |
| --- | --- | --- | --- | --- |
| `CTR-01` | `SOL-01`, `REQ-01` | Input: raw CLI args; output: normalized update-mode state. | CLI parser / orchestration | `update` and `--update` are equivalent entry points. |
| `CTR-02` | `SOL-02`, `SOL-04`, `REQ-02`, `REQ-05`, `REQ-06` | Input: repository id `dapi/start-issue`; output: latest release tag plus asset and checksum URLs. | update workflow / GitHub lookup helper | Lookup owns release metadata resolution only. |
| `CTR-03` | `SOL-03`, `SOL-03A`, `SOL-03B`, `REQ-03`, `REQ-04`, `REQ-06` | Input: running executable path plus latest release tag; output: current version, normalized comparison result, and install/no-op decision. | update workflow / user-facing output | The running executable is the source of truth for the local installation being updated. |
| `CTR-04` | `SOL-04`, `REQ-05`, `REQ-06` | Input: resolved release asset and checksum URLs plus target executable path; output: verified installed executable or explicit install failure. | installer helpers / update workflow | Download, checksum verification, and install remain aligned with `install.sh`. |

## Target Architecture

### Architecture Invariants

- Update mode never requires git-repository discovery.
- Release-backed installation behavior stays aligned with `install.sh`.
- Version equality is semantic for this workflow and ignores an optional leading `v`.
- Update mode never downgrades an installation that is already ahead of the latest published release.
- Update mode changes only the invoked installation by default.
- Existing issue-starting behavior remains isolated from update logic.

### Target Shape

| Layer / responsibility | To-be role | Boundary / non-owner | Related refs |
| --- | --- | --- | --- |
| CLI parsing | Recognize `update` and `--update`, normalize them into update mode, and keep existing issue parsing intact. | Does not perform release I/O directly. | `SOL-01`, `SOL-06` |
| Update workflow | Resolve executable path, current version, latest release tag, normalize versions, compare, and drive install/no-op/failure output. | Does not require repo/base-branch/issue context. | `SOL-02`, `SOL-03`, `SOL-03A`, `SOL-05` |
| Installer helpers | Download release asset and checksum, verify integrity, and install with executable mode preserved. | Do not own CLI mode detection or version comparison. | `SOL-04` |
| Existing issue pipeline | Continue handling `start-issue ISSUE` unchanged. | Does not branch into update-specific behavior. | `SOL-05`, `SOL-06` |

## Accepted Local Decisions

- `SD-01` Use the current one-liner installer contract as the canonical update source, but implement update as an in-process CLI workflow rather than invoking `curl ... | bash`.
- `SD-02` Treat the resolved executable path for the current invocation as the update target, which is safer than guessing across multiple `PATH` entries.
- `SD-03` Prefer release-backed metadata plus checksummed assets over inferring "latest" from repository state or local `VERSION`.
- `SD-03A` Normalize versions by stripping a single leading `v` before comparison, because the current CLI version string and release tags need not use the same surface format.
- `SD-03B` If the installed version is already ahead of the latest published release, treat that state as a no-op success and leave the executable untouched.
- `SD-04` Keep update behavior available outside git repositories by avoiding repo-root, worktree, and issue-fetch prerequisites.

## Change Surface

| Ref | Surface | Type | Why it changes |
| --- | --- | --- | --- |
| `SOL-01` - `SOL-06` | `scripts/start-issue`, `scripts/lib/start_issue/*.sh` | code | Add update mode parsing and execution without regressing the ordinary workflow. |
| `SOL-04` | `install.sh` or shared installer logic | code | Keep release-backed install behavior aligned across install and update paths. |
| `SOL-01` - `SOL-06` | `test/start_issue.bats`, `test/helpers/fake-bin/gh` | test | Cover both update entry forms, no-op path, failure path, and outside-git behavior. |
| `SOL-07` | `README.md`, `README.ru.md`, `doc/spec.md`, `memory-bank/features/FT-013/*` | doc | Document the new workflow and its relation to existing install paths. |

## Failure Modes

- `FM-01` Update mode accidentally requires a git repository and fails outside one.
- `FM-02` Version comparison uses the wrong executable instance when multiple `start-issue` copies exist.
- `FM-03` Release lookup/download/checksum errors collapse into a generic failure that users cannot act on.
- `FM-04` The update path drifts from `install.sh` and creates two incompatible release-install contracts.
- `FM-05` Version comparison treats `1.11.1` and `v1.11.1` as different releases and triggers a false update.

## Rollout / Backout

- `RB-01` Roll out by landing the update mode behind explicit `update` / `--update` entry points and documenting it as part of the install story.
- `RB-02` Back out by reverting the update-mode patch set; the existing install methods remain intact.

## ADR Dependencies

None.
