---
title: Domain States
doc_kind: domain
doc_function: canonical
purpose: "Lifecycle states and allowed transitions for start-issue workflows."
derived_from:
  - ../dna/governance.md
  - model.md
  - rules.md
status: active
audience: humans_and_agents
canonical_for:
  - domain_states
  - state_transitions
---

# Domain States

## State Machines

| State Machine | Concept | Owner | Notes |
| --- | --- | --- | --- |
| `SM-01` | Issue-start workflow | `pipeline.sh` | Normal path from parse to agent launch/manual next steps |
| `SM-02` | Worktree lifecycle | `worktree.sh` | Plan before side effects |
| `SM-03` | Configuration setup | `init.sh` / `config.sh` | `setup` and `init` have distinct scopes |
| `SM-04` | Self-update workflow | `update.sh` | Independent from git repo and issue workflow |
| `SM-05` | Codex batch run | Go launcher | Uses persisted run state and may end at a human gate |

## States

| State | Meaning | Entry condition | Exit condition | Terminal |
| --- | --- | --- | --- | --- |
| `parsed` | CLI input is normalized | Argument parsing succeeds | Config resolution starts | no |
| `config_resolved` | Agent/model/prompt/worktree settings have effective values and sources | Config precedence evaluated | Issue or mode-specific workflow continues | no |
| `issue_fetched` | GitHub issue metadata is available | `gh api` returns issue data | Branch/worktree planning starts | no |
| `worktree_planned` | Branch name and target path are known | Worktree planner validates branch/path situation | Create/reuse/delete/exit decision executes | no |
| `worktree_ready` | Target worktree is safe to use | Worktree creation/reuse succeeds | Optional init and prompt rendering run | no |
| `agent_launched` | Selected agent command replaces the process | Agent is not `none` and launch succeeds | External agent owns session | yes |
| `manual_next_steps` | Worktree is ready without agent launch | Agent is `none` | User continues manually | yes |
| `update_noop` | Running executable is current or newer | Version comparison says no update | Command exits `0` | yes |
| `update_installed` | New release asset installed into running executable path | Download and checksum succeed | Command exits `0` | yes |
| `human_gate_done` | Codex returned `STATUS: DONE` | Final status parsed from last message | Command exits `0` | yes |
| `human_gate_resume` | Codex returned `STATUS: HUMAN_GATE` and resume opened | Thread id captured and resume command succeeds | Interactive Codex resumes | yes |

## Transitions

| Transition ID | From | To | Trigger | Preconditions | Forbidden when |
| --- | --- | --- | --- | --- | --- |
| `TR-01` | `parsed` | `config_resolved` | Normal issue workflow | Valid mode and config inputs | Mutually exclusive prompt sources conflict |
| `TR-02` | `config_resolved` | `issue_fetched` | Issue workflow continues | Issue input is present and dependencies are available | Mode is `setup`, `init`, or `update` |
| `TR-03` | `issue_fetched` | `worktree_planned` | Worktree planner runs | Branch name can be generated | Branch/path state is unreadable |
| `TR-04` | `worktree_planned` | `worktree_ready` | User or planner selects create/reuse/delete path | Path validation passes | Existing path belongs to another branch |
| `TR-05` | `worktree_ready` | `agent_launched` | Agent launch | Agent is supported and not `none` | `--batch` with non-Codex |
| `TR-06` | `worktree_ready` | `manual_next_steps` | No-agent mode | Agent is `none` | none |
| `TR-07` | `config_resolved` | `update_noop` / `update_installed` | Update mode | Release metadata and running executable version known | Checksum/download/install fails |
| `TR-08` | `worktree_ready` | `human_gate_done` / `human_gate_resume` | Codex batch final status parsed | Agent is Codex and thread id/status are valid | Missing or unknown final status |

## State Invariants

- `SI-01` Side effects that assume a valid worktree cannot run before
  `worktree_ready`.
- `SI-02` Update states do not require git repository context.
- `SI-03` Batch terminal verdict comes only from saved `last-message.txt`.

## Implementation Notes

Technical shell variables can have more granular internal states. This document
tracks only workflow states that matter to product behavior and feature design.
