---
title: "FT-018: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution plan for the agent CLI launch compatibility fix."
derived_from:
  - brief.md
  - design.md
  - ../../engineering/testing-policy.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_018_scope
  - ft_018_selected_design
  - ft_018_acceptance_criteria
---

# FT-018: Implementation Plan

## Grounding / Support References

| Path / module | Current role | Why relevant | Reuse / mirror |
| --- | --- | --- | --- |
| `cmd/start-issue/main.go` | Go adapter and process execution | Emits launch/helper commands | Update Kimi args and cwd only |
| `cmd/start-issue/main_test.go` | Go adapter/cwd regression tests | Verifies command shape and process cwd | Extend Kimi expectations |
| `cmd/start-issue/parity_integration_test.go` | Bash-v1 vs Go observable parity | Detects output drift | Keep parity contract aligned |
| `cmd/start-issue/testdata/bash-v1/scripts/lib/start_issue/{agent,output}.sh` | Baseline fixture | Must represent the corrected public contract | Mirror Kimi cwd/args |
| `README.md`, `README.ru.md`, `docs/agent-examples*`, `doc/spec.md` | User contract | Prevents docs from reintroducing the bad flag | Update examples and mapping |

## Test Strategy

| Test surface | Canonical refs | Planned automated coverage | Required command |
| --- | --- | --- | --- |
| Kimi command and helper args | `REQ-02`–`REQ-04`, `SC-01` | Exact args omit `--work-dir`/`--yolo`; helper runs with root cwd | `go test ./cmd/start-issue` |
| Agent cwd behavior | `REQ-02`, `FM-02` | Fake Kimi records cwd; launch must use worktree | `go test ./cmd/start-issue` |
| Baseline parity and docs/index | `REQ-01`, `REQ-05`, `CHK-02` | Updated fixture matches Go and memory-bank links remain valid | `make test` |

## Open Questions / Ambiguities

None after checking the installed `kimi --help` and current Kimi Code CLI reference.

## Environment Contract

| Area | Contract | Failure symptom |
| --- | --- | --- |
| Kimi CLI | Current prompt interface supports `-p/--prompt` and `--model`; no `--work-dir` | Old launch fails with unknown option |
| Tests | Agent binaries are fakes or dry-run only | Network or real agent invocation appears |
| Repository gate | `make test` is required before handoff | Unverified adapter or documentation drift |

## Preconditions

- `PRE-01` `brief.md` and `design.md` are active and define the Kimi cwd contract.
- `PRE-02` Existing non-Kimi parity cases remain unchanged.

## Workstreams

- `WS-01` Update Go Kimi adapter, helper cwd execution, and focused tests (`REQ-02`–`REQ-04`).
- `WS-02` Update Bash parity fixture and all command examples (`REQ-05`).
- `WS-03` Run focused and repository verification (`REQ-01`, `CHK-01`, `CHK-02`).

## Execution Order

1. `STEP-01` Update adapter command/cwd mapping and tests; verify with `go test ./cmd/start-issue`.
2. `STEP-02` Update the parity fixture and docs/spec; verify no product or docs reference emits `--work-dir`.
3. `STEP-03` Run `make test`, inspect the diff for unrelated adapter changes, and record evidence for `CHK-01`/`CHK-02`.

## Stop Conditions / Fallback

- `STOP-01` If parity exposes a non-Kimi regression, stop and restore only the unrelated drift; do not broaden the Kimi adapter.
- `STOP-02` If the installed Kimi CLI contradicts the documented interface, stop before adding version probing and escalate as a new design decision.
