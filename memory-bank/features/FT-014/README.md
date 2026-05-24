---
title: "FT-014: Codex batch human-gate mode with resume flow"
doc_kind: feature
doc_function: index
purpose: "Bootstrap-safe navigation for the FT-014 feature package. Read feature.md first, then solution.md, decision-log.md, and implementation-plan.md."
derived_from:
  - feature.md
status: active
audience: humans_and_agents
---

# FT-014: Codex batch human-gate mode with resume flow

## About

This feature package tracks issue #26: adding a Codex-only batch mode that can finish unattended, but reopens the same Codex session interactively when the agent returns `STATUS: HUMAN_GATE`.

## Annotated Index

- [feature.md](feature.md)
  Read for the canonical problem statement, scope, constraints, acceptance scenarios, and verification contract.

- [solution.md](solution.md)
  Read for the selected Codex-only workflow, state-artifact contract, and CLI/help shape.

- [decision-log.md](decision-log.md)
  Read for FPF-closed ambiguities, including dedicated-help shape, Codex-only enforcement, and resumable state handling.

- [implementation-plan.md](implementation-plan.md)
  Read for execution order, touchpoints, risks, and planned verification.
