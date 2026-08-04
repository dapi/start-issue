---
title: "FT-018: Agent CLI launch compatibility"
doc_kind: feature
doc_function: canonical
purpose: "Canonical brief for keeping start-issue agent launches compatible with the installed agent CLIs."
derived_from:
  - ../../flows/feature-flow.md
  - ../../product/context.md
  - ../../domain/glossary.md
  - ../../engineering/testing-policy.md
  - ../../../doc/spec.md
status: active
delivery_status: in_progress
audience: humans_and_agents
must_not_define:
  - implementation_sequence
---

# FT-018: Agent CLI launch compatibility

## What

### Problem

The Kimi Code CLI currently installed in the user's environment does not
accept the legacy `--work-dir` option. Its current prompt mode also rejects
`--yolo` together with `--prompt`. `start-issue` therefore fails before Kimi
can receive the issue prompt, despite Codex continuing to work.

### Outcome

Agent launch commands use the supported interface of each selected agent and
always execute from the intended issue worktree.

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Kimi launch success | Fails with unknown option | Deterministic launch command is accepted by the current Kimi CLI | Go tests and local `kimi --help` contract check |

## Scope

- `REQ-01` Keep the Codex, Claude, Pi, and `none` launch contracts unchanged.
- `REQ-02` Launch Kimi from the resolved worktree cwd without `--work-dir`.
- `REQ-03` Remove incompatible Kimi `--yolo` usage from prompt mode and keep model forwarding.
- `REQ-04` Keep helper operations (branch naming and prompt improvement) in the repository cwd for Kimi.
- `REQ-05` Update parity fixtures, documentation, and manual next steps together with the adapter behavior.

## Non-Scope

- `NS-01` Do not add version probing or support multiple incompatible Kimi command syntaxes in one launch.
- `NS-02` Do not change the public agent names, config precedence, worktree lifecycle, or Codex human-gate behavior.

## Constraints / Assumptions

- `ASM-01` Current Kimi Code CLI help exposes `-p/--prompt`, `--model`, `--yolo`, and `--add-dir`, but not `--work-dir`.
- `CON-01` The worktree path must be supplied through process cwd for agents whose CLI has no path option.

## Design Requirement Decision

| Decision | Reason | Downstream owner |
| --- | --- | --- |
| `Design required: yes` | This changes an external CLI integration contract and requires explicit command/cwd mapping. | `design.md` |

## Verify

### Exit Criteria

- `EC-01` Kimi dry-run output contains `cd <worktree> && kimi` and no `--work-dir` or incompatible `--yolo -p` combination.
- `EC-02` Kimi helper and launch tests preserve model/prompt forwarding and execute in the requested cwd.
- `EC-03` Existing non-Kimi adapter contracts and parity checks remain green.

### Traceability matrix

| Requirement ID | Problem refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `CON-01` | `EC-03`, `SC-02` | `CHK-01` | `EVID-01` |
| `REQ-02` | `ASM-01`, `CON-01` | `EC-01`, `SC-01` | `CHK-01`, `CHK-02` | `EVID-01`, `EVID-02` |
| `REQ-03` | `ASM-01` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-04` | `CON-01` | `EC-02`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-05` | `CON-01` | `EC-03` | `CHK-02` | `EVID-02` |

### Acceptance Scenarios

- `SC-01` Given `--agent kimi --dry-run`, when start-issue renders its launch, then Kimi runs from the worktree and receives `-p` plus the optional model without unsupported flags.
- `SC-02` Given another supported agent, when its launch is rendered, then its existing adapter-specific command remains unchanged.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`–`EC-03`, `SC-01`, `SC-02` | `go test ./cmd/start-issue` | All adapter and cwd tests pass | `artifacts/ft-018/verify/go-test/` |
| `CHK-02` | `EC-03` | `make test` | Repository checks and parity suite pass | `artifacts/ft-018/verify/make-test/` |

### Evidence

- `EVID-01` Go adapter and cwd test output.
- `EVID-02` Full `make test` output.
