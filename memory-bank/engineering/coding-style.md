---
title: Coding Style
doc_kind: engineering
doc_function: convention
purpose: "Coding and documentation style conventions for start-issue."
derived_from:
  - ../dna/governance.md
  - architecture.md
status: active
audience: humans_and_agents
---

# Coding Style

## General Rules

- Follow the surrounding shell style before introducing a new pattern.
- Keep changes scoped to the requested behavior and touched module boundary.
- Add comments only for non-obvious why/boundary conditions.
- Prefer explicit user-facing errors over surprising implicit fallback.
- Keep docs, help, spec, and tests aligned for public behavior changes.

## Bash Rules

- Use `set -euo pipefail` semantics where already established by the script.
- Quote variable expansions unless the local pattern deliberately relies on word
  splitting.
- Keep adapter-specific command syntax in `agent.sh`.
- Keep config precedence and prompt rendering in `config.sh`.
- Keep top-level ordering in `pipeline.sh`.
- Do not grow `scripts/start-issue` with feature logic; source modules should
  remain the implementation home.

## Tooling Contract

- Formatter: none configured; preserve existing formatting.
- Linter: `shellcheck` through `make test`.
- Test runner: Bats through `make test`.
- Documentation audit: `scripts/check_memory_bank_index.py --max-depth 4`.

## Documentation Style

- Use Markdown with YAML frontmatter for governed `memory-bank` docs.
- For feature work, prefer stable IDs (`REQ-*`, `SC-*`, `CHK-*`, `EVID-*`,
  `SOL-*`, `STEP-*`) over prose-only tracking.
- When a document is adapted from the template but not yet project-specific,
  state that explicitly instead of leaving placeholder examples.

## Change Discipline

- Do not rewrite unrelated modules for style while implementing a feature.
- If a feature changes a shared domain rule, update
  [domain/rules.md](../domain/rules.md).
- If a feature changes architecture boundaries, update this directory or create
  an ADR.
