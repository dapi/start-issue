---
title: Autonomy Boundaries
doc_kind: engineering
doc_function: canonical
purpose: "Agent autonomy, supervision checkpoints, and escalation triggers for start-issue work."
derived_from:
  - ../dna/governance.md
  - ../product/context.md
  - testing-policy.md
canonical_for:
  - agent_autonomy_rules
  - escalation_triggers
  - supervision_checkpoints
status: active
audience: humans_and_agents
---

# Autonomy Boundaries

## Autopilot

Agents may do these without asking when they are in scope:

- edit Go modules, tests, and memory-bank docs;
- run local checks including `make test`;
- add focused Go tests for changed behavior;
- create or update feature packages;
- fix memory-bank link/index issues found by the audit;
- update docs/spec/help when directly required by a behavior change.

## Supervision Checkpoints

Proceed, but surface the plan/result clearly:

- public CLI contract changes;
- release/update workflow changes;
- prompt contract changes;
- new agent adapter behavior;
- migration from legacy feature package layout to new feature-flow layout;
- broad refactors across several Go packages.

## Escalation

Stop and ask before:

- deleting user worktrees or local branches outside an explicit tested workflow;
- changing release tags, publishing releases, or pushing to `master`;
- adding a new runtime language/core rewrite;
- generalizing Codex batch behavior to other agents without evidence;
- changing security/sandbox approval defaults for agent launch;
- resolving contradictory product requirements by guessing.

## Escalation Rule

If failures repeat after two to three attempts and the remaining problem appears
to be ambiguous requirements, missing external credentials, live GitHub state, or
an ungrounded product decision, stop and report the facts plus options.
