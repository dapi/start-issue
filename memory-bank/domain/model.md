---
title: Domain Model
doc_kind: domain
doc_function: canonical
purpose: "Conceptual model for start-issue workflows and ownership boundaries."
derived_from:
  - ../dna/governance.md
  - glossary.md
  - ../product/context.md
status: active
audience: humans_and_agents
canonical_for:
  - domain_model
  - domain_concepts
---

# Domain Model

## Concepts

| Concept | Kind | Owns / Represents | Key relationships | Notes |
| --- | --- | --- | --- | --- |
| `IssueInput` | value object | User-provided issue number or URL | Resolves into `IssueContext` | Number input needs `RepositoryContext` |
| `IssueContext` | value object | Issue URL, number, title, body, labels | Feeds branch naming and prompt rendering | Fetched through `gh` |
| `RepositoryContext` | value object | Git root, `owner/repo`, origin remote, base branch | Required by issue fetch and worktree planning | Not required by setup/update |
| `Configuration` | aggregate | Resolved agent, model, prompt, worktree directory, sources | Read from CLI, config files, env, defaults | Must expose effective source |
| `ConfigSource` | value object | One point in precedence chain | Belongs to `Configuration` | CLI beats project, user, env, default |
| `PromptTemplate` | value object | Template text plus source/location | Renders into `RenderedPrompt` | Unknown placeholders remain unchanged |
| `WorktreePlan` | policy/value object | Branch name, worktree path, reuse/delete/create decision | Executed by worktree lifecycle | Planning should precede side effects |
| `AgentAdapter` | policy/boundary | Agent-specific validation and command construction | Uses `Configuration`, `RenderedPrompt`, worktree path | Owns Codex/Claude/Kimi/Pi differences |
| `ReleaseMetadata` | value object | Latest release tag and asset URLs | Used by self-update | Comes from GitHub Releases |
| `InstalledExecutable` | entity | Running binary path and current version | Updated in place by self-update | Not necessarily repo checkout |
| `BatchRun` | entity | Codex batch state directory, events, last message, thread id | Belongs to one worktree run | Final status comes from `last-message.txt`; `HUMAN_GATE` is one outcome |
| `FeaturePackage` | documentation aggregate | Governed docs for one delivery unit | Uses memory-bank flows and stable IDs | Existing packages may use legacy layout |

## Relationship Map

- `IssueInput` plus `RepositoryContext` produces `IssueContext`.
- `Configuration` resolves before issue fetch and governs prompt, model, agent,
  and worktree directory.
- `IssueContext` and `Configuration` render a `PromptTemplate` into a
  `RenderedPrompt`.
- `WorktreePlan` must be validated before side effects such as `init.sh` or
  agent launch.
- `AgentAdapter` receives a rendered prompt and worktree path; it must not own
  generic CLI parsing or config precedence.
- `ReleaseMetadata` and `InstalledExecutable` define update behavior without
  requiring repository context.

## Concept Ownership

| Concept | Canonical owner | Allowed writers | Allowed readers | Notes |
| --- | --- | --- | --- | --- |
| `Configuration` | Go configuration helpers and `doc/spec.md` | CLI/config code and docs updates | Pipeline, output, adapters | Help and dry-run must stay aligned |
| `WorktreePlan` | Go worktree helpers and feature docs | Worktree lifecycle changes | Pipeline, tests | Reuse must be exact and safe |
| `AgentAdapter` | Go agent adapter helpers | Agent feature work | Pipeline, branch naming, prompt improvement | Keep adapter-specific logic centralized |
| `ReleaseMetadata` | Go release/update helpers | Release/update features | Installer, update workflow | Latest source is GitHub Releases |
| `FeaturePackage` | `memory-bank/flows/feature-flow.md` | Agents and maintainers | Future feature work | Legacy packages are grandfathered |

## Model Boundaries

- `MB-01` GitHub issues are external source data, not entities owned by this
  project.
- `MB-02` Agent CLIs are external executables; `start-issue` owns adapter
  command construction, not agent internals.
- `MB-03` Git worktrees are local git state; `start-issue` plans and creates
  them but does not become a git porcelain replacement.
- `MB-04` Memory-bank feature packages are documentation/process artifacts, not
  runtime configuration.

## Related Documents

- Business rules: [rules.md](rules.md)
- State transitions: [states.md](states.md)
- Context boundaries: [context-map.md](context-map.md)
