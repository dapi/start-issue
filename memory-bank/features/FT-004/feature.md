---
title: "FT-004: Prompt Improvement Workflow"
doc_kind: feature
doc_function: canonical
purpose: "Canonical feature document for adding a reviewable prompt-template improvement workflow to start-issue. Owns only problem space and the verification contract."
derived_from:
  - https://github.com/dapi/start-issue/issues/4
  - ../../../doc/spec.md
  - ../../../README.md
status: active
delivery_status: in_progress
audience: humans_and_agents
must_not_define:
  - selected_design
  - implementation_sequence
---

# FT-004: Prompt Improvement Workflow

## What

### Problem

`start-issue` can use a prompt template from CLI, project config, user config, environment, or built-in defaults. There is no guided mechanism to improve the prompt template that is actually used to start development. A user must manually locate the active prompt, decide how to improve it, and edit the file directly.

### Outcome

Users can ask `start-issue` to generate a reviewable improved prompt-template proposal for the active prompt source, without silently overwriting the prompt that future agent sessions use.

### Scope

- `REQ-01` `start-issue` supports a prompt-improvement mode that resolves the active prompt template through existing prompt precedence.
- `REQ-02` Prompt-improvement mode uses the selected agent and current issue context to produce a complete improved prompt template proposal.
- `REQ-03` Prompt-improvement mode writes a proposal file and never overwrites the active prompt template silently.
- `REQ-04` Prompt-improvement mode exits before Zellij rename, branch generation, worktree creation, init, and agent launch.
- `REQ-05` README, Russian README, spec, and Bats coverage describe and verify the behavior.

### Non-Scope

- `NS-01` Do not auto-apply the generated prompt improvement.
- `NS-02` Do not add additive prompt fragments or a second prompt-composition mechanism.
- `NS-03` Do not change normal prompt rendering, prompt precedence, branch naming, worktree creation, or agent launch behavior.

### Constraints

- `CON-01` Existing `start-issue` behavior must remain unchanged when `--improve-prompt` is not used.

## Verify

### Exit Criteria

- `EC-01` A project prompt can be improved into a proposal file, a built-in prompt can be improved into a project proposal file, dry-run does not write a proposal, and `--agent none` is rejected for prompt improvement.

### Acceptance Scenarios

- `SC-01` Given `.start-issue/prompt.md`, when `start-issue 1 --agent codex --improve-prompt --no-init` runs, the command fetches the issue, writes `.start-issue/prompt.improved.md`, and does not create a worktree.

### Traceability Matrix

| Requirement ID | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- |
| `REQ-01` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-02` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-03` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-04` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-05` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |

### Checks

| Check ID | Covers | How to check | Expected |
| --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `SC-01` | `bash -n scripts/start-issue && shellcheck scripts/start-issue && mise exec -- bats test` | Syntax, static checks, and Bats suite pass. |

### Test Matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | Local terminal output from the verification commands and CI output for PR #7. |

### Evidence

- `EVID-01` Verification command output showing the updated test suite passes.

### Evidence Contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | Local and CI command output | implementer / CI | Terminal output and GitHub Actions job | `CHK-01` |
