---
title: "FT-017: Go parity-first migration"
doc_kind: feature
doc_function: index
purpose: "Navigation for issue #34. Read the canonical brief, selected design, decision log, then the derived execution plan."
derived_from:
  - brief.md
status: active
audience: humans_and_agents
---

# FT-017: Go parity-first migration

## About

This package tracks [issue #34](https://github.com/dapi/start-issue/issues/34): rewrite the `start-issue` CLI in Go through a parity-first migration, without changing the user-facing workflow unintentionally.

## Annotated Index

- [brief.md](brief.md)
  Read first for the canonical problem, scope, blocker, and verify contract.

- [decision-log.md](decision-log.md)
  Read for the FPF-grounded, accepted release-distribution decision.

- [design.md](design.md)
  Read for the selected Go architecture, parity oracle, distribution contract, C2 view, and rollout/backout semantics.

- [implementation-plan.md](implementation-plan.md)
  Read for grounded execution sequencing, test strategy, checkpoints, and stop conditions.
