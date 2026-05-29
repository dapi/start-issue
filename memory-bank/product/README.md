---
title: Product Documentation Index
doc_kind: product
doc_function: index
purpose: "Navigation for start-issue product context, users, outcomes, positioning, and roadmap."
derived_from:
  - ../dna/governance.md
  - context.md
status: active
audience: humans_and_agents
---

# Product Documentation Index

`memory-bank/product/` describes why `start-issue` exists, who uses it, what
outcomes matter, and which product constraints guide feature work.

Product documents do not own domain rules, architecture, acceptance criteria, or
implementation sequence.

## Product Boundaries

| Layer | Owns | Does not own |
| --- | --- | --- |
| `product/` | Developer/user value, core workflows, outcomes, positioning, roadmap themes | Shell module boundaries, test implementation, feature execution steps |
| `domain/` | Shared workflow terms, concepts, rules, states, events, context map | Product priority and messaging |
| `features/` | One delivery unit's requirements, solution, plan, and evidence | Project-wide product background |

## Annotated Index

- [Product Context](context.md) - project-wide context, core workflows,
  top-level outcomes, constraints, and sources.
- [Vision](vision.md) - product promise, strategic bets, experience principles,
  non-goals, and decision rules.
- [Customers](customers.md) - user segments, actors, jobs to be done, assumptions,
  and things not to assume.
- [Metrics](metrics.md) - product metrics, release/readiness guardrails, and
  measurement policy.
- [Marketing](marketing.md) - open-source CLI positioning, messaging, channels,
  alternatives, and launch constraints.
- [Roadmap](roadmap.md) - product themes and horizons without replacing feature
  packages.
