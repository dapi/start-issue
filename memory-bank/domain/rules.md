---
title: Domain Rules
doc_kind: domain
doc_function: canonical
purpose: "Rules and invariants that every start-issue implementation path must preserve."
derived_from:
  - ../dna/governance.md
  - model.md
  - ../../doc/spec.md
status: active
audience: humans_and_agents
canonical_for:
  - domain_rules
  - domain_invariants
---

# Domain Rules

## Invariants

| Rule ID | Rule | Applies to | Why it exists | Source |
| --- | --- | --- | --- | --- |
| `DR-01` | CLI config has highest precedence, then project config, user config, env, built-in default | `Configuration` | User intent must be predictable | `doc/spec.md` |
| `DR-02` | `--prompt` and `--prompt-file` are mutually exclusive; same for env prompt sources | `PromptTemplate` | Avoid ambiguous prompt source | `doc/spec.md` |
| `DR-03` | Prompt improvement never overwrites the active prompt silently | `PromptTemplate` / `prompt proposal` | Preserve reviewable changes | README/spec |
| `DR-04` | Existing worktree reuse requires exact branch/path validation before side effects | `WorktreePlan` | Avoid corrupting unrelated local work | FT-008 |
| `DR-05` | `setup` writes only user config, while project config belongs to `init --project` | `setup`, `init` | Keep onboarding safe outside git repos | FT-014 |
| `DR-06` | `update` uses latest GitHub Release and updates the running executable path | `InstalledExecutable` | Avoid updating the wrong binary | FT-013 |
| `DR-07` | `--human-gate` is Codex-only until another adapter has an equivalent grounded resume contract | `HumanGateRun` | Avoid false support claims | FT-015 |
| `DR-08` | Agent-specific behavior must stay behind the adapter boundary | `AgentAdapter` | Prevent scattered per-agent branching | FT-009 / spec |

## Policies

| Policy ID | Policy | Input | Output / Verdict | Owner |
| --- | --- | --- | --- | --- |
| `POL-01` | Branch type selection | Issue labels | `feature/`, `fix/`, `docs/`, `refactor/`, `test/`, `chore/`, or `hotfix/` prefix | Go worktree helper |
| `POL-02` | Model support | Resolved agent and optional model | Adapter command includes model or fails if unsupported | Go agent adapter |
| `POL-03` | Release comparison | Installed version and latest tag | no-op, update, or fail | Go update helper |
| `POL-04` | First-run gate | Missing `~/.config/start-issue` on ordinary launch | Run setup or create marker directory, then continue | Go orchestration/config helpers |

## Cross-Context Rules

- `XDR-01` Changes to public CLI behavior must update help text, README,
  Russian README/spec when relevant, and Go test coverage together.
- `XDR-02` Changes to release behavior must update release docs, release/update
  tests, and install/update code together.
- `XDR-03` Changes to memory-bank structure must keep
  `scripts/check_memory_bank_index.py --max-depth 4` green.

## Rule Change Policy

- Shared rules move here only when they apply beyond one feature package.
- Feature-local rules stay in the feature package until they become stable
  project rules.
- Architecture-level decisions that outlive one feature should become ADRs in
  [adr/](../adr/README.md).
