---
title: Domain Events
doc_kind: domain
doc_function: canonical
purpose: "Business-significant events emitted or observed by start-issue workflows."
derived_from:
  - ../dna/governance.md
  - model.md
  - rules.md
status: active
audience: humans_and_agents
canonical_for:
  - domain_events
  - business_events
---

# Domain Events

## Events

| Event ID | Event | Meaning | Producer | Consumers | Minimal facts |
| --- | --- | --- | --- | --- | --- |
| `DE-01` | `IssueContextResolved` | The issue and repository facts needed for work can be used | GitHub/repo workflow | Worktree planner, prompt renderer | repo, issue number, URL, title, labels |
| `DE-02` | `ConfigurationResolved` | Effective agent/model/prompt/worktree settings are known | Config workflow | Output, pipeline, adapters | values and sources |
| `DE-03` | `WorktreeReady` | It is safe to run init/prompt/agent inside the target worktree | Worktree lifecycle | Pipeline, agent launch | branch name and path |
| `DE-04` | `PromptProposalWritten` | A reviewable improved prompt proposal exists | Prompt improvement workflow | User/maintainer | output path and source prompt |
| `DE-05` | `ReleaseUpdateInstalled` | Running executable has been replaced by verified release asset | Update workflow | CLI user | old version, new version, executable path |
| `DE-06` | `HumanGateReached` | Codex batch run requires a human decision | Batch workflow | CLI user, resume command | thread id, state directory, final message |
| `DE-07` | `MemoryBankAuditPassed` | Governed docs are reachable and links are valid | `check_memory_bank_index.py` | Maintainer/agent | scope, entrypoint, max depth |

## Event Rules

- Events describe facts that have happened, not commands to perform.
- Technical logs are not domain events unless they carry a workflow verdict.
- If a new event changes allowed workflow transitions, update [states.md](states.md).
- If an event crosses a context boundary, update [context-map.md](context-map.md).

## Delivery Semantics

This project does not currently publish runtime domain events. The events above
are conceptual workflow facts used for documentation, tests, and feature design.
For Codex batch mode, JSONL `thread.started` is an external technical event
observed by the workflow, while `HumanGateReached` is the local workflow verdict.
