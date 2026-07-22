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

- Follow idiomatic Go style and run `gofmt` before committing.
- Keep changes scoped to the requested behavior and touched module boundary.
- Add comments only for non-obvious why/boundary conditions.
- Prefer explicit user-facing errors over surprising implicit fallback.
- Keep docs, help, spec, and tests aligned for public behavior changes.

## Go Rules

- Keep adapter-specific command syntax in dedicated Go helpers.
- Keep config precedence and prompt rendering separate from worktree side effects.
- Return wrapped errors at external process and filesystem boundaries.

## Tooling Contract

- Formatter: `gofmt`.
- Linter: `go vet` through `make test`.
- Test runner: Go tests through `make test`.
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
