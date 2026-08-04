---
title: "FT-017: Codex human-gate delivery permissions"
doc_kind: feature
doc_function: index
purpose: "Navigation for the Codex human-gate permission and full-delivery contract feature package."
derived_from:
  - ../../dna/governance.md
  - brief.md
  - design.md
  - implementation-plan.md
status: active
audience: humans_and_agents
---

# FT-017: Codex human-gate delivery permissions

## About

This package tracks GitHub issue #37. It extends the FT-015 batch/resume flow
with an explicit capability contract for restricted work and opt-in end-to-end
Git delivery.

## Annotated index

- [brief.md](brief.md)
  Canonical problem space, scope, acceptance scenarios, checks, and evidence
  contract.
- [design.md](design.md)
  Selected permission-mode design, Codex command contract, trust boundary,
  failure modes, and rollout/backout rules.
- [implementation-plan.md](implementation-plan.md)
  Grounded execution sequence, test strategy, checkpoints, and approval gate
  for live GitHub-writing verification.

- [decision-log.md](decision-log.md)
  FPF decisions, evidence provenance, and the remaining live-verification gate.
