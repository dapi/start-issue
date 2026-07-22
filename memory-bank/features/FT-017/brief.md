---
title: "FT-017: Rewrite start-issue CLI in Go with parity-first migration"
doc_kind: feature
doc_function: canonical
purpose: "Canonical problem-space brief for issue #34. Defines the parity-first Go migration scope and verification contract without selecting the Go implementation or release design."
derived_from:
  - ../../flows/feature-flow.md
  - ../../product/context.md
  - ../../engineering/testing-policy.md
  - ../../../README.md
  - ../../../doc/spec.md
  - https://github.com/dapi/start-issue/issues/34
status: active
delivery_status: in_progress
audience: humans_and_agents
must_not_define:
  - solution_space
  - implementation_sequence
  - release_asset_matrix
---

# FT-017: Rewrite start-issue CLI in Go with parity-first migration

## What

### Problem

The current public `start-issue` executable is a Bash entrypoint assembled from shell modules. Issue #34 requires a Go implementation and distribution artifact while keeping the current CLI contract and workflow guarantees unless a difference is intentional and documented. The legacy runtime and Bats stack must not remain as production or test dependencies after cutover.

### Outcome

The selected Go implementation can replace the current default `start-issue` runtime only after deterministic compatibility checks demonstrate parity for the critical workflows named in issue #34 and any intentional differences are explicitly documented.

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Critical workflow parity | Current Bash CLI and its Bats fixtures | Go is made default only after every approved parity case passes or has a documented intentional difference | Repeatable baseline-vs-Go parity suite and its report |
| `MET-02` | Verification integration | `make test` runs shell checks and Bats | `make test` includes the relevant Go and compatibility suites | Local command and CI job output |

### Scope

- `REQ-01` Create the Go module and make `cmd/start-issue` the sole public CLI entrypoint.
- `REQ-02` Preserve the current public CLI contract and workflow semantics for argument parsing, configuration precedence, prompt rendering, repository detection, branch/worktree lifecycle, and supported-agent launch adapters, except for differences explicitly recorded as intentional.
- `REQ-03` Continue to invoke `git`, `gh`, and supported agent CLIs as external commands during the initial migration; this feature does not replace them with native protocol or API clients.
- `REQ-04` Establish deterministic Go tests for help, invalid input, configuration precedence, dry-run output, worktree planning/reuse, and agent launch-command generation.
- `REQ-05` Replace build, install, release, documentation, and `make test` with Go-native paths and platform-specific binaries.

### Non-Scope

- `NS-01` Do not redesign product workflows, configuration precedence, prompts, or supported agent behavior as part of the runtime migration.
- `NS-02` Do not replace `git`, `gh`, or agent CLIs with native Go API/protocol integrations in this migration.
- `NS-04` Do not replace `git`, `gh`, or agent CLIs with native Go API/protocol integrations, even where doing so might simplify platform-specific packaging.

### Constraints / Assumptions

- `ASM-01` The legacy scripts and fixtures are migration references only; the delivered runtime and test stack are Go-only.
- `ASM-02` The current CI uses `mise` and runs integration tests on Ubuntu; the install-script job runs on Ubuntu and macOS.
- `CON-01` The public executable name is `start-issue`; existing `install.sh` downloads a fixed asset and validates its checksum. The selected distribution contract must keep installation integrity verification.
- `CON-02` The current release workflow builds its only release artifact on Ubuntu, while the installed artifact is verified on both Ubuntu and macOS by CI. `DL-01` supersedes this single-asset layout for the Go release.
- `CON-03` Product constraint `PCON-01` requires the public CLI contract to stay stable unless this feature changes it explicitly with synchronized docs and tests.
- No unresolved blocking decisions remain. The accepted release-distribution decision is `DL-01` in `decision-log.md`.

## Design Requirement Decision

| Decision | Reason | Downstream owner |
| --- | --- | --- |
| `Design required: yes` | The feature changes the implementation/runtime and release artifact boundary, requires a parity oracle, and has explicit cutover/backout semantics. | `design.md` |

## Verify

### Exit Criteria

- `EC-01` The Go CLI is the default runtime and its deterministic tests cover all in-scope critical workflows.
- `EC-02` The Go implementation continues to use external `git`, `gh`, and supported-agent CLIs in the first migration phase.
- `EC-03` `make test` and CI run the required Go and compatibility suites after cutover work is introduced.
- `EC-04` Build, install, update, release, and documentation follow the human-approved distribution contract and publish the Go binary successfully.

### Traceability Matrix

| Requirement ID | Problem refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `ASM-01`, `CON-03`, `DL-01` | `EC-01` | `CHK-01`, `CHK-02` | `EVID-01`, `EVID-02` |
| `REQ-02` | `ASM-01`, `CON-03` | `EC-01` | `CHK-01` | `EVID-01` |
| `REQ-03` | issue #34 | `EC-02` | `CHK-01` | `EVID-01` |
| `REQ-04` | `ASM-01` | `EC-01` | `CHK-01` | `EVID-01` |
| `REQ-05` | `CON-01`, `CON-02`, `DL-01` | `EC-03`, `EC-04` | `CHK-02`, `CHK-03` | `EVID-02`, `EVID-03` |

### Acceptance Scenarios

- `SC-01` Given deterministic Go test fixtures for help, invalid input, configuration precedence, dry-run output, worktree planning/reuse, or an agent launch command, when the Go CLI runs in the fake environment, then it produces the documented result.
- `SC-02` Given a build/install/release change, when it is merged, then the Go binary remains the default `start-issue` runtime.
- `SC-03` Given a user installs or updates `start-issue` after cutover, when the approved target platform is used, then build, release asset selection, checksum verification, installation, and `--version` complete according to the approved distribution contract.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `EC-02`, `SC-01` | Run the implemented baseline-vs-Go deterministic parity suite | All in-scope cases match, or each deviation links to an approved intentional-difference record | `artifacts/ft-016/verify/parity/` |
| `CHK-02` | `EC-01`, `EC-03`, `SC-02` | Run `make test` after Go and parity integration | Required local checks, Go tests, and compatibility suite pass | `artifacts/ft-016/verify/make-test/` |
| `CHK-03` | `EC-04`, `SC-03` | Run the approved release/install verification matrix in CI | Linux/macOS automatically select the correct artifact, verify checksum, and report a version; Windows binary execution and its documented manual install/update instruction pass | `artifacts/ft-016/verify/distribution/` |

### Test Matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | `artifacts/ft-016/verify/parity/` |
| `CHK-02` | `EVID-02` | `artifacts/ft-016/verify/make-test/` |
| `CHK-03` | `EVID-03` | `artifacts/ft-016/verify/distribution/` |

### Evidence

- `EVID-01` Machine-readable or reviewable parity report for the baseline and Go runs, including any linked intentional differences.
- `EVID-02` Local and CI output for `make test` after Go and compatibility-suite integration.
- `EVID-03` CI evidence for the approved build/install/release platform matrix, asset selection, checksum verification, and version output.

### Evidence Contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | Parity-suite report and fixtures/results | parity test runner | `artifacts/ft-016/verify/parity/` | `CHK-01` |
| `EVID-02` | `make test` logs and CI job link | implementer / CI | `artifacts/ft-016/verify/make-test/` | `CHK-02` |
| `EVID-03` | Platform-matrix install/release logs and published asset manifest | CI / release workflow | `artifacts/ft-016/verify/distribution/` | `CHK-03` |
