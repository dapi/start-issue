---
title: "FT-016: Real Codex human-gate E2E suite"
doc_kind: feature
doc_function: canonical
purpose: "Canonical problem and verification contract for an opt-in real-Codex human-gate smoke suite."
derived_from:
  - ../../flows/feature-flow.md
  - ../../engineering/testing-policy.md
status: active
delivery_status: in_progress
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - solution_space
---

# FT-016: Real Codex human-gate E2E suite

## What

### Problem

The deterministic Bats suite uses a fake Codex executable and cannot validate compatibility with a locally installed real Codex CLI.

### Outcome

| Metric ID | Metric | Target | Measurement method |
| --- | --- | --- | --- |
| `MET-01` | Operator can run real-Codex smoke validation | One documented opt-in command per terminal state | Script output and saved state artifacts |

### Scope

- `REQ-01` Provide an explicit opt-in local suite for a real Codex `STATUS: DONE` run.
- `REQ-02` Provide a manually completable `STATUS: HUMAN_GATE` resume scenario.
- `REQ-03` Verify state artifacts and reject accidental fake-Codex execution.

### Non-Scope

- `NS-01` Do not run real Codex sessions in CI or `make test`.
- `NS-02` Do not alter the human-gate runtime behavior.

### Constraints / Assumptions

- `ASM-01` The operator has authenticated `gh` and a current `codex`; the private fixture repository owns the control issue.
- `CON-01` A real agent session can have side effects, so explicit opt-in and an isolated worktree parent are required.

## Design Requirement Decision

| Decision | Reason | Downstream owner |
| --- | --- | --- |
| `Design required: yes` | The suite introduces an operator-facing environment and safety contract. | `design.md` |

## Verify

### Exit Criteria

- `EC-01` The `done` scenario validates a real `thread.started` event, state files, and `STATUS: DONE`.
- `EC-02` The `human-gate` scenario validates the reported resume command and `STATUS: HUMAN_GATE` after the operator exits resume.

### Traceability matrix

| Requirement ID | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- |
| `REQ-01`, `REQ-03` | `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-02`, `REQ-03` | `SC-02` | `CHK-02` | `EVID-02` |

### Acceptance Scenarios

- `SC-01` With explicit authorization and a real Codex CLI, the `done` scenario finishes successfully and reports preserved artifacts.
- `SC-02` With the same prerequisites, the `human-gate` scenario opens resume and validates its state after the operator exits.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `SC-01` | `START_ISSUE_E2E=1 make e2e-human-gate` | `PASS` plus a state path | temporary fixture clone |
| `CHK-02` | `SC-02` | `START_ISSUE_E2E=1 test/e2e/human-gate.sh --scenario human-gate` | `PASS` after resume exits | temporary fixture clone |

### Evidence

- `EVID-01` Preserved `e2e.log`, events, last message, and thread-id from `CHK-01`.
- `EVID-02` The same artifacts plus reported resume command from `CHK-02`.
