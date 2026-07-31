---
title: Engineering Documentation Index
doc_kind: engineering
doc_function: index
purpose: "Navigation for start-issue engineering rules: architecture, tests, style, autonomy, git workflow, and CLI UI policy."
derived_from:
  - ../dna/governance.md
  - architecture.md
  - testing-policy.md
status: active
audience: humans_and_agents
---

# Engineering Documentation Index

`memory-bank/engineering/` contains the implementation rules for the Go CLI
and its documentation/test workflow.

- [Engineering Architecture Patterns](architecture.md) - package boundaries,
  adapter boundary, config ownership, failure handling, and process boundary.
- [Frontend Engineering](frontend.md) - current CLI UI surfaces and the policy
  for any future non-CLI UI.
- [Testing Policy](testing-policy.md) - canonical local checks, Go coverage,
  memory-bank audit, sufficient coverage, and manual-only exceptions.
- [Autonomy Boundaries](autonomy-boundaries.md) - what agents may do
  autonomously, what needs supervision, and when to escalate.
- [Coding Style](coding-style.md) - Go/documentation style, tooling, and
  change discipline.
- [Git Workflow](git-workflow.md) - default branch, worktrees, commits, PR
  handoff, and release tags.
- [ADR](../adr/README.md) - location for accepted architecture decisions.
