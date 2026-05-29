---
title: Product Vision
doc_kind: product
doc_function: canonical
purpose: "Long-term product direction, experience principles, and non-goals for start-issue."
derived_from:
  - ../dna/governance.md
  - context.md
status: active
audience: humans_and_agents
canonical_for:
  - product_vision
  - product_strategy_principles
---

# Product Vision

## Product Promise

`start-issue` should make starting issue work boring and reliable. A developer
should be able to point at a GitHub issue and get the same workspace shape every
time: clear branch name, isolated worktree, known prompt source, and a launch
path for the chosen agent.

The tool should stay small enough to understand locally. When complexity grows,
the project should first clarify workflow boundaries and tests, not hide
behavior behind implicit magic.

## Strategic Bets

| Bet ID | Bet | Why now | Evidence | Review cadence |
| --- | --- | --- | --- | --- |
| `BET-01` | Agent-neutral start workflow | Developers switch between Claude, Codex, Kimi, Pi, and manual mode | Current adapter support in README/spec | Revisit when adding an agent |
| `BET-02` | Reviewable prompt/config evolution | Prompt templates materially affect agent output | `--improve-prompt`, setup/init flows | Revisit when prompt placeholders change |
| `BET-03` | Resumable batch work for Codex | Some issue work can run unattended until a real human gate | `--human-gate` feature | Revisit when other agents expose equivalent resume contracts |

## Experience Principles

- `XP-01` Show effective decisions: agent, model, prompt source, prompt
  location, worktree path, and launch command should be inspectable.
- `XP-02` Prefer explicit failure over silent fallback when user intent would be
  misrepresented, especially for agent support and worktree reuse.
- `XP-03` Keep destructive or irreversible actions behind clear user choice.
- `XP-04` Keep normal issue start fast; heavier documentation/process belongs
  in memory-bank for medium and large work, not in every small fix.

## Product Non-Goals

- `PNG-01` Do not become a general GitHub issue manager.
- `PNG-02` Do not own agent internals beyond launch, non-interactive helpers,
  and prompt contracts.
- `PNG-03` Do not replace code review, PR creation policy, or merge policy.
- `PNG-04` Do not add a daemon, server, database, or UI without an explicit
  product and architecture decision.

## Decision Rules

- If a change improves the common issue-start path but makes config precedence
  harder to predict, preserve predictability first.
- If an agent feature is only grounded for one CLI, implement it as
  adapter-specific and reject unsupported agents clearly.
- If a workflow starts needing persistent lifecycle state beyond local files in
  the worktree, create a feature design and reevaluate the Bash boundary.

## Source Documents

- [Product context](context.md)
- [Specification](../../doc/spec.md)
