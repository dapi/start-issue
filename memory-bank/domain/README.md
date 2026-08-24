---
title: Domain Documentation Index
doc_kind: domain
doc_function: index
purpose: "Navigation for start-issue domain language, workflow concepts, rules, states, events, and context boundaries."
derived_from:
  - ../dna/governance.md
  - glossary.md
status: active
audience: humans_and_agents
---

# Domain Documentation Index

`memory-bank/domain/` defines the stable language and workflow model for
`start-issue`: issue input, repository context, configuration, prompt templates,
worktrees, agent adapters, release assets, batch runs, human gates, and feature
packages.

Domain documents do not own product positioning, shell implementation sequence,
release commands, or feature acceptance criteria.

## Boundary With Product And Engineering

| Layer | Owns | Does not own |
| --- | --- | --- |
| `product/` | Why the CLI exists, who it serves, what outcomes matter | Detailed workflow invariants |
| `domain/` | Terms, concepts, rules, states, events, context boundaries | Bash module layout and command syntax details |
| `engineering/` | Code/module boundaries, testing policy, style, autonomy, git workflow | Product priority or domain truth |

## Annotated Index

- [Glossary](glossary.md) - canonical terms and ambiguous naming rules.
- [Domain Model](model.md) - conceptual model and ownership boundaries.
- [Domain Rules](rules.md) - invariants and policies that implementation paths
  must preserve.
- [States](states.md) - workflow state machines and allowed transitions.
- [Events](events.md) - business-significant workflow facts and event rules.
- [Context Map](context-map.md) - bounded contexts and upstream/downstream
  relationships.
