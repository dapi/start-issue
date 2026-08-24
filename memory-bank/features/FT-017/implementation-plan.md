---
title: "FT-017: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution plan for explicit restricted and full-delivery Codex human-gate permission modes."
derived_from:
  - brief.md
  - design.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_017_scope
  - ft_017_selected_design
  - ft_017_acceptance_criteria
  - ft_017_blocker_state
---

# FT-017: Implementation Plan

## Current goal

Implement the accepted FT-017 permission-mode contract while preserving the
existing FT-015 batch, state, status, and resume behavior.

Deterministic implementation, documentation, command-shape coverage, and local
Codex CLI `0.145.0` parser validation are complete. `STEP-06` remains pending
because the real full-delivery run requires explicit `AG-01` authorization and
creates retained fixture GitHub state.

## Grounding / Support References

| Document | Role in this plan | Facts reused | Conflict action |
| --- | --- | --- | --- |
| `brief.md` | Canonical problem / verify owner | `REQ-*`, `SC-*`, `NEG-*`, `CHK-*`, `EVID-*` | Update `brief.md` first |
| `design.md` | Canonical solution owner | `SOL-*`, `C4-01`, `SD-*`, `CTR-*`, `INV-*`, `FM-*`, `RB-*` | Update `design.md` first |
| `../FT-015/feature.md` and `solution.md` | Existing human-gate contract | State, status, thread, resume, and exit semantics | Preserve FT-015 ownership; update FT-017 if extension assumptions change |
| `../FT-016/brief.md` and `design.md` | Existing real-Codex E2E boundary | Explicit opt-in, fake-Codex rejection, retained artifacts | Extend rather than duplicate the runner contract |

## Current State / Reference Points

| Path / module | Current role | Why relevant | Reuse / mirror |
| --- | --- | --- | --- |
| `cmd/start-issue/main.go` | Go CLI parser, config resolution, launcher, help, and human-gate state | Owns the new option, command mapping, and diagnostics | Extend existing options and array-based `exec.Cmd` construction |
| `cmd/start-issue/main_test.go` | Deterministic Go regression suite | Existing tests cover human-gate command, DONE, HUMAN_GATE, and errors | Add precedence, invalid value, full-delivery, and order assertions |
| `cmd/start-issue/parity_integration_test.go` | Go/Bash observable parity coverage | Protects unaffected legacy behavior during the Go implementation | Keep non-human-gate parity cases green |
| `test/e2e/human-gate.sh` | Opt-in real-Codex smoke runner | Closest existing live verification surface | Add a separately guarded full-delivery scenario only after approval |
| `README.md`, `README.ru.md`, `doc/human-gate-permissions*.md`, `doc/spec.md` | Public guides and canonical behavior docs | Must match help and command behavior | Update together with output assertions |

## Test Strategy

| Test surface | Canonical refs | Existing coverage | Planned automated coverage | Required local suites / commands | Required CI suites / jobs | Manual-only gap / justification | Manual-only approval ref |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Permission resolution and validation | `REQ-03`, `SC-01`, `SC-02`, `NEG-01`, `CTR-01`, `CTR-02` | None | CLI beats env; env beats default; invalid value fails before fetch/mutation | `make test` | Existing test job | none | none |
| Codex command construction | `REQ-01`, `REQ-02`, `REQ-04`, `SC-03`, `CTR-03`, `CTR-04`, `INV-02`, `INV-04` | Restricted dry-run and fake Codex execution | Assert semantic mode mapping, global-before-`exec` order, model coexistence, and no raw interpolation | `make test` | Existing test job | Real installed-Codex parser behavior is external | `AG-01` for live run |
| FT-015 state/resume regression | `REQ-05`, `SC-04`, `INV-05` | DONE, HUMAN_GATE, missing status/thread | Run existing scenarios under default restricted and one full-delivery fake path | `make test` | Existing test job | none | none |
| Help/docs contract | `REQ-01`, `REQ-06`, `SC-05`, `NEG-02`, `CTR-05` | Dedicated human-gate help assertions | Assert default, full-delivery warning, prerequisites, limitations, and troubleshooting | `make test`; documentation review | Existing test job | Prose consistency review is manual | reviewer approval in PR |
| End-to-end Git delivery | `REQ-02`, `REQ-08`, `SC-06`, `NEG-03`, `SOL-06`, `RB-03` | Real Codex terminal-state smoke only | Keep syntax/static coverage automated; add guarded scenario entrypoint | `make test` plus explicitly approved E2E | Excluded from CI | Requires real Codex, credentials, network, Git writes, push, and PR creation | `AG-01` |

## Open Questions / Ambiguities

| Open Question ID | Question | Why unresolved | Blocks | Default action / escalation owner |
| --- | --- | --- | --- | --- |
| `OQ-01` | Which future Codex versions remain compatible after the issue baseline? | The external CLI has no repository-owned stability guarantee; local parser validation covers `0.145.0` only. | Does not block deterministic implementation; blocks claims about future versions | Treat the approved live executable as the acceptance baseline and update adapter/docs together on command-shape failure. |
| `OQ-02` | Which isolated fixture repository/issue should receive the live full-delivery PR? | Live target selection is operator-owned and may change. | `STEP-06` only | Require explicit target and approval through `AG-01`; never infer from global focus or an unrelated repo. |

## Environment Contract

| Area | Contract | Used by | Failure symptom |
| --- | --- | --- | --- |
| setup | Go, Bash, Git, and the current Go source tree; deterministic tests use fakes | `STEP-01` - `STEP-05` | `make test` dependency or fixture failure |
| supported Codex | Issue failure baseline is `0.144.6`; local parser validation covers `0.145.0`; the approved live executable must complete the recorded full-delivery flow | `STEP-03`, `STEP-06` | Parser rejection or no `thread.started` event |
| deterministic test | `make test` is canonical and must not use network or real agent binaries | `CHK-01`, `STEP-02` - `STEP-05` | External side effects or nondeterministic test failures |
| live access | Explicit opt-in, authenticated `gh`, real Codex, authorized fixture repo/issue, network, and permission to push/create a PR | `STEP-06` | Missing auth, push rejection, absent PR, or no terminal status |
| secrets | Credentials remain in existing authenticated tools/environment and never enter tracked files or command output | all steps | Token-like data appears in diff, logs, or state artifacts |

## Preconditions

| Precondition ID | Canonical ref | Required state | Used by steps | Blocks start |
| --- | --- | --- | --- | --- |
| `PRE-01` | `CON-01`, `SOL-01`, `INV-01` | Restricted remains the accepted default | `STEP-01` - `STEP-05` | yes |
| `PRE-02` | `ASM-01`, `SOL-03`, `CTR-03`, `CTR-04` | Supported command grammar is recorded and locally inspectable | `STEP-03`, `STEP-06` | yes |
| `PRE-03` | `CON-04`, `SD-04`, `RB-03` | Explicit operator approval and isolated target exist | `STEP-06` | no; blocks live acceptance only |

## Workstreams

| Workstream | Implements | Result | Owner | Dependencies |
| --- | --- | --- | --- | --- |
| `WS-1` | `REQ-03`, `SOL-01`, `SD-02`, `CTR-01`, `CTR-02` | Validated permission resolution and early errors | agent | `PRE-01` |
| `WS-2` | `REQ-01`, `REQ-02`, `REQ-04`, `SOL-02` - `SOL-05`, `CTR-03` - `CTR-05` | Correct commands and visible capability status | agent | `WS-1`, `PRE-02` |
| `WS-3` | `REQ-06`, `REQ-07`, `SC-01` - `SC-05` | Automated regression coverage and aligned docs | agent | `WS-1`, `WS-2` |
| `WS-4` | `REQ-08`, `SOL-06`, `SC-06`, `RB-03` | Guarded full-delivery E2E evidence | human + agent | `WS-2`, `WS-3`, `PRE-03`, `AG-01` |

## Approval Gates

| Approval Gate ID | Trigger | Applies to | Why approval is required | Approver / evidence |
| --- | --- | --- | --- | --- |
| `AG-01` | Running a real full-delivery session that can commit, push, and create/update a PR | `STEP-06`, `WS-4`, `CHK-03` | The run is unsandboxed and creates external GitHub state | User names/approves the fixture target; retained fixture directory, log, state artifacts, and PR URL record approval context |

## Work Order

| Step ID | Actor | Implements | Goal | Touchpoints | Artifact | Verifies | Evidence IDs | Check command / procedure | Blocked by | Needs approval | Escalate if |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | agent | `REQ-03`, `SOL-01`, `SD-02`, `CTR-01`, `CTR-02`, `FM-03` | Add CLI/environment/default resolution and fail-fast validation | `cmd/start-issue/main.go`, `main_test.go` | Resolved mode and source | `CHK-01`, `NEG-01` | `EVID-01` | Focused Go tests, then `make test` | `PRE-01` | none | Validation occurs after fetch or mutation |
| `STEP-02` | agent | `REQ-07`, `INV-01`, `INV-02` | Extend the fake Codex process and Go tests before changing launcher behavior | `cmd/start-issue/main_test.go` | Red/green command contract tests | `CHK-01`, `SC-01` - `SC-04` | `EVID-01` | `go test ./cmd/start-issue` | `STEP-01` | none | Fake cannot distinguish global and exec arguments |
| `STEP-03` | agent | `REQ-01`, `REQ-02`, `REQ-04`, `SOL-02`, `SOL-03`, `CTR-03`, `CTR-04`, `INV-04` | Build validated restricted/full-delivery commands in supported order | `cmd/start-issue/main.go` | Array-based Codex command mapping | `CHK-01`, `SC-02`, `SC-03` | `EVID-01` | Focused Go tests; inspect `--dry-run` command | `STEP-02`, `PRE-02` | none | Supported Codex rejects generated grammar |
| `STEP-04` | agent | `REQ-01`, `REQ-06`, `SOL-04`, `CTR-05`, `FM-02`, `FM-04` | Add capability output, warning, dedicated help, and public docs | `cmd/start-issue/main.go`, README files, practical guides, spec | Consistent operator contract | `CHK-02`, `SC-05`, `NEG-02` | `EVID-02` | Help assertions and documentation review | `STEP-03` | none | Docs imply permission equals credentials or product authorization |
| `STEP-05` | agent | `REQ-05`, `REQ-07`, `SOL-05`, `INV-05`, `RB-01`, `RB-02` | Run full deterministic regression and simplify review | All changed runtime/tests/docs | Green local suite and complexity verdict | `CHK-01`, `CHK-02`, `SC-04` | `EVID-01`, `EVID-02`, `EVID-09` | `make test`; inspect diff for unnecessary branches/abstractions | `STEP-01` - `STEP-04` | none | FT-015 state/resume behavior changes |
| `STEP-06` | human + agent | `REQ-08`, `SOL-06`, `SD-04`, `SC-06`, `NEG-03`, `RB-03` | Extend/run isolated live full-delivery verification and retain evidence | E2E runner and approved fixture repo/issue | E2E log, state artifacts, commit and PR URL | `CHK-03`, `EC-06` | `EVID-03` | Follow canonical cmux caller-tab procedure and poll to terminal PASS/failure | `STEP-05`, `PRE-03`, `OQ-02` | `AG-01` | Target/auth/caller context is missing, or any unexpected external scope appears |

## Parallelizable Work

- `PAR-01` Documentation wording and fake-Codex fixture preparation can proceed
  in parallel after `STEP-01` stabilizes names and precedence.
- `PAR-02` Launcher code and command-order assertions share the same contract
  and should be reviewed together even if edited in parallel.
- `PAR-03` Live E2E work cannot begin before deterministic checks and explicit
  approval complete.

## Checkpoints

| Checkpoint ID | Refs | Condition | Evidence IDs |
| --- | --- | --- | --- |
| `CP-01` | `STEP-01`, `STEP-02`, `CTR-01`, `CTR-02`, `FM-03` | Precedence/default/error behavior is deterministic and side-effect free | `EVID-01` |
| `CP-02` | `STEP-03`, `CTR-03`, `CTR-04`, `FM-01` | Both commands match supported grammar and exact sandbox mapping | `EVID-01` |
| `CP-03` | `STEP-04`, `STEP-05`, `CTR-05`, `INV-05` | Docs/help/tests agree and FT-015 regression suite passes | `EVID-01`, `EVID-02`, `EVID-09` |
| `CP-04` | `STEP-06`, `RB-03`, `AG-01` | Approved live run records full Git delivery or a precise capability failure | `EVID-03` |

## Execution Risks

| Risk ID | Risk | Impact | Mitigation | Trigger |
| --- | --- | --- | --- | --- |
| `ER-01` | Codex option grammar changes again | Batch mode fails before thread creation | Centralize command mapping, assert order, document tested baseline, run opt-in real smoke | Parser rejection or missing `thread.started` |
| `ER-02` | Full-delivery mode is mistaken for product authorization | Agent can perform unintended high-impact work | High-signal warning, explicit opt-in, preserve prompt human-gate rules | Docs/output omit warning or prompt bypass is proposed |
| `ER-03` | Configuration surface expands unnecessarily | More precedence and persistence bugs | Keep one CLI option, one env variable, one safe default | Proposal adds raw passthrough or project/user persistence |
| `ER-04` | Live E2E writes to the wrong repository | Unwanted external branch/PR state | Require explicit approved target and canonical cmux caller context | Target or caller context is ambiguous |

## Stop Conditions / Fallback

| Stop ID | Related refs | Trigger | Immediate action | Safe fallback state |
| --- | --- | --- | --- | --- |
| `STOP-01` | `FM-01`, `RB-01` | Supported Codex rejects restricted command | Stop rollout and retain existing FT-015-compatible restricted behavior | No full-delivery exposure |
| `STOP-02` | `FM-02`, `RB-03`, `AG-01` | Live target/auth/capability is missing | Stop without guessing credentials or target; retain logs | Deterministic implementation complete, live acceptance pending |
| `STOP-03` | `FM-04`, `NS-03`, `SD-03` | Implementation treats full delivery as destructive/production authorization | Remove that behavior and return to prompt-gated policy | Explicit mode controls capability only |
| `STOP-04` | `INV-05`, `REQ-05` | Thread/state/resume regression appears | Back out permission changes until FT-015 behavior is restored | Existing restricted human-gate flow |

## Plan-local Evidence

| Evidence ID | Artifact | Producer | Path contract | Reused by checkpoints |
| --- | --- | --- | --- | --- |
| `EVID-09` | Simplify-review verdict for the final runtime diff | implementer / reviewer | Final handoff or PR review note | `CP-03` |

## Ready for Acceptance

- `CP-01` through `CP-03` have deterministic evidence.
- `CP-04` has explicitly approved live evidence; otherwise FT-017 remains short
  of `delivery_status: done`.
- `make test` and required CI jobs are green.
- No secrets, arbitrary argument interpolation, or hidden privilege defaults
  were introduced.
- Final acceptance uses `brief.md` `CHK-01` through `CHK-03` and their evidence
  contracts.
