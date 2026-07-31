---
title: Product Context
doc_kind: product
doc_function: canonical
purpose: "Project-wide product context for start-issue: users, core workflows, outcomes, and product constraints."
derived_from:
  - ../dna/governance.md
  - ../../README.md
  - ../../doc/spec.md
status: active
audience: humans_and_agents
canonical_for:
  - project_product_context
  - product_problem_space
  - top_level_outcomes
must_not_define:
  - domain_model
  - domain_invariants
  - implementation_sequence
  - architecture_decision
---

# Product Context

`start-issue` is a developer CLI that turns a GitHub issue into a prepared
coding workspace: issue context, branch, git worktree, optional initialization,
and an agent session. The product exists to make the start of issue work
repeatable and low-friction for humans and coding agents.

The primary users are developers who work from GitHub issues and want a clean
branch/worktree per issue. A secondary user is the coding agent itself: prompts,
configuration, worktree paths, and launch commands must be predictable enough
for unattended or semi-attended workflows.

The product boundary is deliberately narrow. `start-issue` starts or resumes
issue work; it does not replace GitHub, the selected agent CLI, a full task
tracker, or the developer's review and merge process.

## Boundary With PRD And Domain

- `product/context.md` owns the stable project-wide "why" and product workflow.
- `prd/` is only needed for larger initiatives that are not yet one delivery
  slice.
- `domain/` owns the shared language: issue input, repository context, config
  source, prompt template, worktree, agent adapter, release asset, and run state.
- Feature packages own slice-specific requirements, acceptance checks, and
  implementation plans.

## Core Product Workflows

- `WF-01` Start issue work: resolve issue and config, create/reuse a branch and
  worktree, run optional `init.sh`, render a prompt, and launch the selected
  agent or print manual next steps.
- `WF-02` Configure defaults: write project or user config with `init`, or run
  user-level onboarding with `setup` and the first-run gate.
- `WF-03` Improve the active prompt: resolve the prompt source, fetch issue
  context, ask an agent for an improved proposal, and never overwrite the active
  prompt silently.
- `WF-04` Update the installed CLI: compare the running executable version with
  the latest GitHub Release, verify checksum, and install into the same path.
- `WF-05` Run Codex human-gate mode: execute Codex in batch mode, persist run
  state, exit on `STATUS: DONE`, and resume interactively on `STATUS:
  HUMAN_GATE`.

## Top-Level Outcomes

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Time from issue reference to prepared worktree | Manual setup | One command for common cases | Manual workflow review and regression tests |
| `MET-02` | Predictability of config and launch behavior | Historically implicit defaults | Effective agent, model, prompt source, and launch command are visible | CLI output, `--dry-run`, Go test coverage |
| `MET-03` | Release/install reliability | Manual install/update risk | Checksummed release asset and self-update path | Release workflow and installer/update tests |

## Product Constraints

- `PCON-01` Keep the public CLI contract stable unless a feature explicitly
  changes it and docs/tests/spec are updated together.
- `PCON-02` Worktree reuse must be safe: never continue in a path that is not
  proven to belong to the intended branch.
- `PCON-03` Prompt and config changes must be visible and reviewable; prompt
  improvement writes proposals instead of overwriting active templates.
- `PCON-04` Agent-specific behavior belongs behind the adapter boundary.
- `PCON-05` The Go CLI is the sole runtime. New lifecycle behavior belongs in
  focused Go helpers and must not reintroduce a second shell or Python runtime.

## Source Documents

- [README.md](../../README.md)
- [doc/spec.md](../../doc/spec.md)
- [Feature packages](../features/README.md)
