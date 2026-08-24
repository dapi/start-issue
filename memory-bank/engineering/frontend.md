---
title: Frontend Engineering
doc_kind: engineering
doc_function: canonical
purpose: "Frontend/UI policy for start-issue."
derived_from:
  - ../dna/governance.md
  - ../product/context.md
status: active
audience: humans_and_agents
---

# Frontend Engineering

`start-issue` currently has no web or graphical frontend. Its user interface is
the CLI: flags, prompts, help text, dry-run output, and agent launch commands.

## CLI UI Surfaces

- Normal help and missing-issue guidance.
- `--dry-run` output.
- `setup` and first-run onboarding prompts.
- `init` prompts and planned writes.
- `--batch-help` (`--human-gate-help` remains a compatibility alias).
- Release/update status messages.

## CLI UX Rules

- Keep output explicit about effective config and sources.
- Do not hide destructive choices behind defaults.
- Keep normal help concise; use dedicated help for complex modes such as
  `--batch-help`.
- When adding interactive prompts, support non-interactive test coverage through
  Go test input simulation.

## If A Frontend Is Added Later

Create a feature package with required `design.md` before adding any non-CLI UI.
That design must define the UI surface, state ownership, and how it maps to the
existing CLI/domain model.
