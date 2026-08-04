---
title: "FT-013: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution and verification plan for the reconciled issue #35 self-update slice."
derived_from:
  - brief.md
  - design.md
  - decision-log.md
  - ../../engineering/testing-policy.md
status: archived
audience: humans_and_agents
must_not_define:
  - ft_013_scope
  - ft_013_selected_design
  - ft_013_acceptance_criteria
---

# FT-013: Implementation Plan

## Grounding / Support References

| Path | Current role | Relevance |
| --- | --- | --- |
| `cmd/start-issue/main.go` | Go CLI, parser, update workflow, release/install helpers | Current implementation and change surface |
| `cmd/start-issue/main_test.go` | Focused unit/integration tests | Version, release, failure, staging, and update behavior |
| `cmd/start-issue/parity_integration_test.go` | Bash-v1/Go parity and installer checks | Regression and release-contract evidence |
| `README.md`, `README.ru.md`, `doc/spec.md` | Public CLI/spec contract | User-facing update entry points and prerequisites |
| `Makefile`, `.github/workflows/*` | Local/CI verification | Required repository gates |

## Test Strategy

| Surface | Canonical refs | Automated coverage | Required verification |
| --- | --- | --- | --- |
| Entry points and dispatch | `REQ-01`, `SC-01`, `SC-02` | Go parser/run-mode tests | `go test ./...` |
| Release metadata/assets | `REQ-02`, `REQ-05`, `SC-01`, `SC-05` | Fake `gh`, HTTP fixtures, asset selection tests | `go test ./...` |
| Version ordering/no downgrade | `REQ-03`, `REQ-04`, `SC-03`, `SC-04` | Semantic-version and no-op tests | `go test ./...` |
| Safe replacement | `REQ-05`, `REQ-06`, `SC-05` | Checksum, staged-version, mode, rename tests | `go test ./...` |
| Outside-git and ordinary workflow | `REQ-07`, `SC-06`, `SC-07` | Mode isolation and parity tests | `go test ./...`, `make test` |
| Documentation/index integrity | `REQ-08` | Memory-bank and whitespace checks | `make test`, `git diff --check` |

No manual-only gap is currently required; live GitHub access is represented by
deterministic fakes in automated tests.

## Open Questions / Ambiguities

None blocking. The issue's generic “asset and SHA-256 checksum” wording is
resolved by the existing repository release contract: current-platform binary
asset plus `checksums.txt`, as recorded in `DL-03`.

## Environment Contract

| Area | Contract | Failure symptom |
| --- | --- | --- |
| Build/test | Go toolchain, git, Python, and repository Make checks are available | Required local gate cannot run |
| Release boundary | Tests use fake `gh` and local HTTP fixtures; no live network is required | Test becomes nondeterministic or needs credentials |
| Filesystem | Tests use temporary targets and must not modify the installed user binary | Unsafe test setup or cleanup failure |

## Preconditions

| ID | Canonical ref | Required state |
| --- | --- | --- |
| `PRE-01` | `brief.md`, `design.md`, `DL-07` | Reconciled scope and current Go design are active |
| `PRE-02` | `REQ-07`, `INV-03` | Existing issue-start regression tests remain the baseline |

## Workstreams

- `WS-01` Verify parser and isolated update-mode behavior (`REQ-01`, `REQ-07`).
- `WS-02` Verify release lookup, semantic comparison, checksums, staged version, and replacement (`REQ-02`–`REQ-06`).
- `WS-03` Align public docs and memory-bank navigation (`REQ-08`).

## Approval Gates

None. The plan only changes repository documentation and uses temporary test
targets; it does not authorize replacing a user's installed binary during this
review pass.

## Work Order

| Step | Implements | Artifact | Verifies | Evidence |
| --- | --- | --- | --- | --- |
| `STEP-01` | `REQ-01`, `REQ-07` | Focused parser/mode test results | `CHK-01` | `EVID-01` |
| `STEP-02` | `REQ-02`–`REQ-06` | Release/update test results | `CHK-01` | `EVID-01` |
| `STEP-03` | `REQ-08` | README/spec and feature-pack docs | `CHK-02` | `EVID-02` |
| `STEP-04` | `EC-01`–`EC-06` | Final local verification output | `CHK-01`, `CHK-02` | `EVID-01`, `EVID-02` |

## Checkpoints

- `CP-01`: focused Go/parity tests cover all update branches and preserve the ordinary workflow.
- `CP-02`: `make test` passes and canonical docs have no stale active legacy owner.

## Execution Risks / Stop Conditions

- `ER-01`: A documentation claim diverges from the Go implementation; stop and update the canonical owner before proceeding.
- `ER-02`: A test requires live GitHub or an installed user binary; stop and replace it with a deterministic fixture.
- `STOP-01`: If current implementation and issue acceptance materially conflict, stop at the owner document and reopen the decision log rather than expanding scope.

## Plan-local Evidence

- `EVID-01` `cmd/start-issue/main_test.go` and
  `cmd/start-issue/parity_integration_test.go`; `make test` passed on 2026-08-04.
- `EVID-02` `README.md`, `README.ru.md`, `doc/spec.md`, and the successful
  memory-bank link audit and `git diff --check` within `make test` on
  2026-08-04.
- `EVID-03` Review-improve cycle results stored in
  `feature-review-report.md`.

## Acceptance Result

All canonical checks passed, the current Go runtime and public documentation
satisfy `SC-01` through `SC-07`, the simplify review found no required runtime
change, and `brief.md` records the delivery as done.
