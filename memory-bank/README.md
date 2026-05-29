---
title: Start Issue Memory Bank
doc_kind: project
doc_function: index
purpose: "Root navigation and provenance note for this project's memory-bank."
derived_from:
  - https://github.com/dapi/memory-bank/
status: active
audience: humans_and_agents
---

# Start Issue Memory Bank

## Source

This project's `memory-bank/` is adapted from the reusable template at
https://github.com/dapi/memory-bank/.

Checked upstream reference:

- repository: `https://github.com/dapi/memory-bank/`
- branch: `main`
- commit: `921a6344e9f74c643a59ef72410685011ac45dbe`

## Local Scope

The current local adoption includes the upstream template structure needed for
governed feature work: `dna/`, `flows/`, `features/`, `prompts/`, `product/`,
`domain/`, `engineering/`, `ops/`, `prd/`, `use-cases/`, `adr/`, and `epics/`.

The product, domain, engineering, and ops sections have been adapted to
`start-issue` and should be treated as the current project baseline. Template
flows and templates remain upstream-derived process references.

## Annotated Index

- [dna/README.md](dna/README.md)
  Governance core: SSoT, frontmatter, lifecycle, and cross-reference rules.
- [flows/README.md](flows/README.md)
  Feature/epic workflows, task routing, and governed templates.
- [prompts/README.md](prompts/README.md)
  Reusable prompt documents for issue review, feature-pack review, implementation, and PR finish.
- [product/README.md](product/README.md)
  Product context, vision, customers, metrics, marketing, and roadmap.
- [domain/README.md](domain/README.md)
  Domain glossary, model, rules, states, events, and context map.
- [engineering/README.md](engineering/README.md)
  Architecture, testing, coding, git workflow, frontend, and autonomy guidance.
- [ops/README.md](ops/README.md)
  Development, configuration, environments, release, and runbook guidance.
- [prd/README.md](prd/README.md)
  Location for instantiated Product Requirements Documents.
- [use-cases/README.md](use-cases/README.md)
  Location for stable project-level use cases.
- [adr/README.md](adr/README.md)
  Location for Architecture Decision Records.
- [epics/README.md](epics/README.md)
  Location for larger initiatives that decompose into feature packages.
- [features/README.md](features/README.md)
  Index and rules for instantiated feature packages.
- [features/FT-004/README.md](features/FT-004/README.md)
  Prompt improvement workflow.
- [features/FT-008/README.md](features/FT-008/README.md)
  Worktree lifecycle stabilization.
- [features/FT-009/README.md](features/FT-009/README.md)
  Modularized `start-issue` pipeline.
- [features/FT-012/README.md](features/FT-012/README.md)
  Explicit agent and model selection.
- [features/FT-013/README.md](features/FT-013/README.md)
  Self-update workflow from latest release.
- [features/FT-014/README.md](features/FT-014/README.md)
  Setup onboarding and first-run user config.
- [features/FT-015/README.md](features/FT-015/README.md)
  Codex batch human-gate mode with resume flow.
