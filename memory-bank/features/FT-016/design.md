---
title: "FT-016: Design"
doc_kind: feature
doc_function: canonical
purpose: "Selected design for the opt-in real-Codex human-gate E2E suite."
derived_from:
  - brief.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_016_scope
  - ft_016_acceptance_criteria
  - implementation_sequence
---

# FT-016: Design

## C4 Applicability

| C4 ID | Decision | Trigger / reason | Artifact |
| --- | --- | --- | --- |
| `C4-00` | not required | A local shell test script adds no runtime boundary. | none |

## Selected Solution

- `SOL-01` Use one opt-in Bash runner with `done` and `human-gate` scenarios.
- `SOL-02` Run the source executable from an independently selected target repository and create a unique temporary worktree parent.

## Accepted Local Decisions

- `SD-01` Preserve logs and run state on failure or explicit `START_ISSUE_E2E_KEEP=1`; clean disposable resources after success.
- `SD-02` Keep the suite outside CI and `make test`; the operator explicitly authorizes live execution with `START_ISSUE_E2E=1`.

## Contracts

| Contract ID | Input / Output | Semantics / Constraints |
| --- | --- | --- |
| `CTR-01` | `START_ISSUE_E2E`, issue, optional project dir | Authorization, target issue, and target repo must be explicit or safely defaulted. |
| `CTR-02` | Codex events and last message | Require `thread.started`, all three state files, and the expected terminal status. |

## Failure Modes

- `FM-01` Fake Codex is found on `PATH`; fail before issue work starts.
- `FM-02` Codex, `gh`, issue access, or terminal status is unavailable; fail with the preserved log path.
