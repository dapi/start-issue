---
title: "FT-014: First-run setup onboarding for user config"
doc_kind: feature
doc_function: index
purpose: "Bootstrap-safe navigation for the FT-014 feature package. Read feature.md first, then solution.md, decision-log.md, and implementation-plan.md."
derived_from:
  - feature.md
status: active
audience: humans_and_agents
---

# FT-014: First-run setup onboarding for user config

## About

This feature package tracks issue #25: adding a dedicated `setup` onboarding flow for user-level configuration in `~/.config/start-issue` and a one-time first-run prompt to initialize that config directory.

## Annotated Index

- [feature.md](feature.md)
  Read for the canonical problem statement, scope, constraints, acceptance scenarios, and verification contract.

- [solution.md](solution.md)
  Read for the selected onboarding architecture, command-entry equivalence, and compatibility rules with the existing `init` flow.

- [decision-log.md](decision-log.md)
  Read for feature-local decisions that close the issue's ambiguous points and document conflict resolution.

- [implementation-plan.md](implementation-plan.md)
  Read for execution order, touchpoints, risks, and planned verification.
