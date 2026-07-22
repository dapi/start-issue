---
title: "FT-017: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution plan for FT-017. Sequences discovery, parity, Go migration, and distribution verification without redefining canonical feature or solution facts."
derived_from:
  - brief.md
  - design.md
  - decision-log.md
  - ../../../memory-bank/engineering/testing-policy.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_016_scope
  - ft_016_selected_design
  - ft_016_acceptance_criteria
  - ft_016_blocker_state
---

# FT-017: Implementation Plan

## Goal

Deliver the Go `start-issue` runtime only after the Bash baseline and Go implementation satisfy the canonical parity and distribution evidence contract in `brief.md`.

## Grounding / Support References

| Document | Role in this plan | Facts reused | Conflict action |
| --- | --- | --- | --- |
| `brief.md` | Canonical problem/verify owner | `REQ-*`, `SC-*`, `CHK-*`, `EVID-*` | Update `brief.md` first |
| `design.md` | Canonical solution owner | `SOL-*`, `C4-02`, `SD-*`, `CTR-*`, `INV-*`, `FM-*`, `RB-*` | Update `design.md` first |
| `decision-log.md` | Decision reference | `DL-01`, `DL-02` | Update the log and design before the plan |

## Current State / Reference Points

| Path / module | Current role | Why relevant | Reuse / mirror |
| --- | --- | --- | --- |
| `scripts/start-issue` | Bash executable entrypoint and version owner | Baseline CLI behavior and bundled-script marker | Black-box baseline only; do not change during `RB-01` |
| `scripts/lib/start_issue/{cli,config,github,worktree,agent,init,update,release,output,pipeline}.sh` | Existing responsibility boundaries | Grounding for `SOL-01` and parity case inventory | Mirror responsibility, not shell syntax |
| `test/start_issue.bats`, `test/fixtures/issue-1.json`, `test/helpers/fake-bin/` | Deterministic regression environment | Grounding for `SOL-03` | Reuse fixtures/fakes; add case-specific assertions rather than a second fake ecosystem |
| `Makefile`, `install.sh`, `.github/workflows/{ci,release}.yml` | Build/install/release paths | Change surface for `SOL-05`/`SOL-06` | Replace only at `RB-02`; preserve existing validation until then |
| `dapi/port-selector` `.goreleaser.yml` and `install.sh` | Referenced release pattern | `DL-01` release asset/checksum selection | Adapt names and verification; do not copy unrelated Homebrew policy |

## Test Strategy

| Test surface | Canonical refs | Existing coverage | Planned automated coverage | Required local suites / commands | Required CI suites / jobs | Manual-only gap / justification | Approval ref |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Baseline/Go CLI behavior | `REQ-01`–`REQ-04`, `SC-01`, `CTR-01` | Bats and fake CLIs cover current shell behavior | Fixture-driven parity cases plus Go unit/integration tests | `go test ./...`; parity command; `make test` | Go test/parity job | none | none |
| Build and static quality | `REQ-01`, `REQ-05`, `CTR-04` | shell syntax/shellcheck | `go vet ./...`, `gofmt` check, Go build with version injection | `go vet ./...`; `gofmt -l .`; build command | CI Go quality/build job | none | none |
| POSIX install/update | `REQ-05`, `SC-03`, `CTR-03` | macOS/Ubuntu install-script CI and shell test | asset selection, checksum verification, installed `--version`, updater replacement/failure tests | `make test`; targeted install/update tests | Ubuntu and macOS install matrix | Live published-release download only after tagged release; local assets/mock release endpoint cover deterministic paths | none |
| Windows release/update | `REQ-05`, `SC-03`, `CTR-03`, `SD-04`, `SD-06` | No Windows baseline runtime | GoReleaser target-manifest check, Windows binary `--version`, and deterministic assertion that `update` prints manual asset instruction | `goreleaser check`; Windows cross-build | Windows CI job | Manual PATH installation is the accepted delivery contract; it must be documented and the binary must execute in Windows CI | none |

## Open Questions / Ambiguities

None. `DEC-01` was resolved as `DL-01`; implementation discoveries that change scope, design, or evidence must be promoted to their canonical owner before work continues.

## Environment Contract

| Area | Contract | Used by | Failure symptom |
| --- | --- | --- | --- |
| Go toolchain | Use pinned Go `1.21` in `go.mod`, `mise.toml`, and CI/release jobs. | `STEP-01`–`STEP-06` | Inconsistent build/test behavior across local and CI environments |
| Test | `make test` remains the repository gate and is extended to invoke the required Go and parity suites. | `STEP-02`, `STEP-06` | A green partial suite lacks feature acceptance evidence |
| External command fakes | Deterministic parity runs place existing fake `git`/`gh`/agent surfaces first on `PATH` and isolate temp worktrees. | `STEP-02`–`STEP-05` | A test calls real network/agent tools or leaks filesystem state |
| Release | GoReleaser v2 has tag metadata and GitHub token access in tag CI; local work uses check/snapshot only. | `STEP-06` | An unverified local release is mistaken for publish evidence |

## Preconditions

| Precondition ID | Canonical ref | Required state | Used by steps | Blocks start |
| --- | --- | --- | --- | --- |
| `PRE-01` | `DL-01`, `DL-02`, `SD-04`, `SD-06` | Target matrix, Go 1.21, and Windows manual-install/update scope accepted | `STEP-01`, `STEP-06` | yes |
| `PRE-02` | `INV-01`, `CTR-01` | Bash baseline and deterministic fixtures remain runnable | `STEP-02`–`STEP-05` | yes |

## Workstreams

| Workstream | Implements | Result | Owner | Dependencies |
| --- | --- | --- | --- | --- |
| `WS-01` | `REQ-01`, `REQ-04`, `SOL-03` | Frozen case inventory and executable parity harness | agent | `PRE-02` |
| `WS-02` | `REQ-01`–`REQ-03`, `SOL-01`, `SOL-02` | Go CLI behavior ported in dependency order | agent | `WS-01` checkpoints |
| `WS-03` | `REQ-05`, `SOL-05`, `SOL-06` | Multi-platform build, POSIX installer/updater, Windows manual-update output, docs, and CI/release configuration | agent | `PRE-01`, parity checkpoint |
| `WS-04` | `REQ-01`–`REQ-05`, `RB-02`, `RB-03` | Cutover evidence and release-ready verification | agent | `WS-02`, `WS-03` |

## Approval Gates

No human approval gates are required. Publishing remains protected by the automated `CHK-01`–`CHK-03` evidence gates and the repository's normal tag/release permissions.

## Execution Order

| Step ID | Actor | Implements | Goal | Touchpoints | Artifact | Verifies | Evidence IDs | Check procedure | Blocked by | Needs approval | Escalate if |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | agent | `REQ-01`, `SOL-01`, `SOL-05` | Scaffold module, command entrypoint, pinned Go toolchain, and GoReleaser config without changing default runtime | `go.mod`, `cmd/`, tooling, `.goreleaser.yml` | Buildable parallel command and release manifest | `CHK-02` | `EVID-02` | Go build/vet/format and GoReleaser config check | `PRE-01` | none | Platform targets diverge from `DL-01` |
| `STEP-02` | agent | `REQ-04`, `SOL-03`, `CTR-01` | Convert existing critical deterministic behaviors into named baseline-vs-Go parity cases before porting their implementation | `test/`, fixtures, fakes, Go test packages | Parity inventory and runner | `CHK-01` | `EVID-01` | Run baseline and placeholder/implemented Go cases under same fake environment | `PRE-02` | none | A case needs real network or undefined observable expectations |
| `STEP-03` | agent | `REQ-02`, `REQ-03`, `SOL-01`, `SOL-02` | Port pure CLI/config/prompt behavior and invalid-input/help cases | Go CLI/config packages, parity cases | Matching pure behavior | `CHK-01` | `EVID-01` | Go tests plus parity cases | `STEP-02` | none | Mismatch changes canonical CLI contract |
| `STEP-04` | agent | `REQ-02`, `REQ-03`, `SOL-01`, `SOL-02` | Port repository detection, issue retrieval, branch/worktree planning/reuse, and init behavior | Go repository/worktree/GitHub packages, fakes | Matching orchestration behavior | `CHK-01` | `EVID-01` | Isolated worktree parity cases | `STEP-02` | none | Unsafe worktree reuse violates `PCON-02` |
| `STEP-05` | agent | `REQ-02`, `REQ-03`, `SOL-01`, `SOL-02` | Port agent adapters, prompt improvement, human-gate, launch commands, and update logic using external processes | Go agent/update packages, fakes | Matching launch/update behavior | `CHK-01` | `EVID-01` | Adapter/update parity and failure cases | `STEP-03`, `STEP-04` | none | A native client or unexplained difference appears necessary |
| `STEP-06` | agent | `REQ-05`, `SOL-04`–`SOL-06`, `CTR-03`, `CTR-04` | Integrate Go into Makefile, installer/updater, CI/release, and docs; cut over only after parity is green | `Makefile`, `install.sh`, workflows, README/spec | Five-asset release-ready path and Go-default distribution | `CHK-02`, `CHK-03` | `EVID-02`, `EVID-03` | Full `make test`, GoReleaser snapshot/check, platform CI matrix | `STEP-01`–`STEP-05` | none | Any required check fails or asset/checksum selection is ambiguous |

## Parallelizable Work

- `PAR-01` `STEP-01` scaffolding and case-inventory design can proceed in parallel while the Bash runtime remains unchanged.
- `PAR-02` After `STEP-02`, pure config/prompt cases (`STEP-03`) may progress independently of repository/worktree cases (`STEP-04`).
- `PAR-03` Distribution configuration drafting may begin after `STEP-01`, but `RB-02` integration in `STEP-06` must wait for all parity checkpoints.

## Checkpoints

| Checkpoint ID | Refs | Condition | Evidence IDs |
| --- | --- | --- | --- |
| `CP-01` | `STEP-02`, `CTR-01`, `INV-02` | Every issue-required critical workflow has a named deterministic parity case and no unexplained normalization | `EVID-01` |
| `CP-02` | `STEP-03`–`STEP-05`, `SC-01`, `RB-02` | All implemented Go cases pass against the Bash baseline or carry an approved intentional difference | `EVID-01` |
| `CP-03` | `STEP-06`, `CHK-02`, `CHK-03`, `RB-03` | Go-default build/install/release path passes local and CI evidence for every `DL-01` target | `EVID-02`, `EVID-03` |

## Execution Risks

| Risk ID | Risk | Impact | Mitigation | Trigger |
| --- | --- | --- | --- | --- |
| `ER-01` | Baseline behavior is underspecified by existing tests | False parity claim | Expand cases before porting each surface; require observable side effects | A Go behavior cannot be compared deterministically |
| `ER-02` | Worktree or update path has destructive behavior | User repository/executable could be damaged | Isolated temp fixtures, explicit failure tests, and `STOP-01` | A test reaches a real worktree or install target |
| `ER-03` | Platform asset selection fails | Install/update failure on supported platform | Test mapping and checksums on every target; reject unknown mapping | Missing or mismatched target asset |

## Stop Conditions / Fallback

| Stop ID | Related refs | Trigger | Immediate action | Safe fallback state |
| --- | --- | --- | --- | --- |
| `STOP-01` | `INV-01`, `FM-01`, `FM-03` | Unexplained parity failure or unsafe side effect | Stop the affected port, preserve baseline evidence, and promote semantic questions to `brief.md`/`design.md` | Bash remains default runtime |
| `STOP-02` | `CTR-03`, `FM-02`, `FM-04`, `RB-03` | Target asset/checksum/install matrix fails | Do not tag/publish; correct release configuration and rerun matrix | Previous published release remains available |

## Plan-local Evidence

| Evidence ID | Artifact | Producer | Path contract | Reused by checkpoints |
| --- | --- | --- | --- | --- |
| `EVID-09` | Grounding inventory, parity-case manifest, and simplify-review verdict | implementer / reviewer | `artifacts/ft-016/plan/` or committed test metadata | `CP-01`, `CP-02` |

## Ready for Acceptance

- All workstreams are complete and `CP-01`–`CP-03` have evidence.
- `CHK-01`–`CHK-03` pass; any manual live-release check is documented as evidence, not a substitute for deterministic tests.
- A simplify review confirms that Go package boundaries clarify responsibilities and do not add abstraction unsupported by `SOL-*`/`CTR-*`/`INV-*`.
- Final acceptance follows `brief.md` `Verify`.
