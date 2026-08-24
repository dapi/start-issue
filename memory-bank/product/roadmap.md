---
title: Product Roadmap
doc_kind: product
doc_function: canonical
purpose: "Product themes and horizons for start-issue without replacing feature packages."
derived_from:
  - ../dna/governance.md
  - context.md
  - vision.md
  - metrics.md
status: active
audience: humans_and_agents
canonical_for:
  - product_roadmap
  - product_themes
---

# Product Roadmap

This roadmap is a direction guide, not a backlog. Delivery units live in
[features/](../features/README.md), and large initiatives should use
[epics/](../epics/README.md) or [prd/](../prd/README.md) before being split
into feature packages.

## Horizons

| Horizon | Theme | Intended outcome | Candidate PRD / Feature | Dependency | Status |
| --- | --- | --- | --- | --- | --- |
| `now` | Memory-bank adoption | Agents have project-specific process, product, domain, engineering, and ops context | Current memory-bank work | Template source from `dapi/memory-bank` | active |
| `next` | New feature-flow adoption | New medium/large features use `brief.md -> optional design.md -> implementation-plan.md` | Future `FT-*` packages | AGENTS.md and flows docs | planned |
| `next` | Release confidence | Release prep, changelog, version, build, and update path stay coherent | Existing release scripts | CI and `make test` | active |
| `later` | Richer lifecycle commands | Possible `resume`, `list`, `cleanup`, or structured output | Unknown | Requires Go design and module-boundary review | idea |

## Roadmap Rules

- Do not turn this file into a list of issues.
- If a theme needs several delivery slices, create an epic or PRD first.
- If a theme changes public CLI behavior, docs/spec/help/tests must change
  together.
- If a theme changes domain terms or state, update [domain/](../domain/README.md)
  before or with the feature package.

## Open Bets

- `BET-01` Whether Codex batch/human-gate patterns should remain Codex-only or become
  a generic agent capability after other CLIs expose equivalent contracts.
- `BET-02` Whether future lifecycle complexity warrants extracting Go helper
  packages from the current command package.
