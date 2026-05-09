---
title: "FT-008: Solution"
doc_kind: feature
doc_function: canonical
purpose: "Canonical solution document for FT-008. Defines the selected lifecycle-planning design and reuse safety rules without redefining feature scope."
derived_from:
  - feature.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_008_scope
  - ft_008_acceptance_criteria
  - ft_008_delivery_status
  - implementation_sequence
---

# FT-008: Solution

## Selected Design

- `SOL-01` Add a real local `make test` target that mirrors the repo's verification stack without requiring `mise trust` for the common case.
- `SOL-02` Split worktree lifecycle into explicit planning helpers: branch-to-path derivation, exact worktree lookup, reuse validation, and destructive resolution choices.
- `SOL-03` Allow reuse only after proving that the selected path is a registered worktree for the requested branch.
- `SOL-04` Keep destructive actions explicit and scoped to the exact conflicting worktree/branch that was discovered.

## Requirement Mapping

| Requirement ID | Solution / architecture refs | Notes |
| --- | --- | --- |
| `REQ-01` | `SOL-01` | Local verification becomes discoverable and repeatable. |
| `REQ-02` | `SOL-02`, `CTR-01` | Exact branch matching removes prefix ambiguity. |
| `REQ-03` | `SOL-03`, `CTR-02` | Reuse must be proven before later side effects run. |
| `REQ-04` | `SOL-02`, `C4-L3-01` | Planner and side-effect phases become more explicit. |
| `REQ-05` | `SOL-01` - `SOL-04` | Tests cover the new lifecycle contract. |

## To-Be C4 Model

| Level | Include? | Model refs | Selection rationale |
| --- | --- | --- | --- |
| System Context (L1) | no | `none` | No actor or system boundary changes. |
| Container (L2) | no | `none` | This remains a single-script local tool. |
| Component (L3) | yes | `C4-L3-01`, `C4-L3-02` | The internal script responsibilities are being separated. |

| C4 ref | Level | Element / relationship | To-be architecture statement | Related refs |
| --- | --- | --- | --- | --- |
| `C4-L3-01` | Component | `scripts/start-issue` lifecycle planner | Branch lookup, worktree path derivation, and reuse validation are pure planning decisions performed before side-effectful steps. | `SOL-02`, `SOL-03` |
| `C4-L3-02` | Component | `scripts/start-issue` side-effect execution | Worktree removal/creation, init, and agent launch occur only after the planner yields a safe result. | `SOL-03`, `SOL-04` |

## Target Architecture

### Architecture Invariants

- The CLI flags and workflow sequence remain intact.
- Unsafe reuse does not continue into `init.sh` or agent launch.
- Conflict handling stays interactive when a human choice is required.

### Target Shape

| Layer / responsibility | To-be role | Boundary / non-owner | Related refs |
| --- | --- | --- | --- |
| Verification entrypoint | Provide a stable local command for syntax, static checks, diff checks, and Bats. | Does not define CI itself. | `SOL-01` |
| Lifecycle planning | Compute branch refs, expected paths, exact existing-worktree matches, and reuse eligibility. | Does not mutate git state. | `SOL-02`, `SOL-03` |
| Conflict resolution | Interpret the user choice for reuse, rename, or delete/recreate. | Does not launch init or agents. | `SOL-03`, `SOL-04` |
| Side effects | Remove/create worktrees and then continue startup. | Runs only after a validated plan exists. | `SOL-04` |

## Accepted Local Decisions

- `SD-01` Exact branch matching is based on the full `refs/heads/<branch>` value from `git worktree list --porcelain`.
- `SD-02` Path reuse is valid only when the path is a registered worktree for the current repository and the registered branch exactly matches the requested branch.
- `SD-03` `make test` uses direct local commands so the common developer loop does not depend on `mise trust`.

## Change Surface

| Ref | Surface | Type | Why it changes |
| --- | --- | --- | --- |
| `SOL-01` | `Makefile` | code | Add a real local test target. |
| `SOL-02` - `SOL-04` | `scripts/start-issue` | code | Separate planning helpers from worktree side effects and harden reuse validation. |
| `SOL-01` - `SOL-04` | `test/start_issue.bats` | test | Add lifecycle regression coverage for conflict/reuse/delete/flat/base fallback paths. |
| `SOL-01` - `SOL-04` | `memory-bank/features/FT-008/*` | doc | Track feature intent, solution, and execution state in feature-flow format. |

## Internal Flow

1. `SOL-02` Derive the expected branch name and planned worktree path.
2. `SOL-02` Query `git worktree list --porcelain` to find exact branch/path relationships.
3. `SOL-03` If reuse is requested, validate that the path is both registered and mapped to the requested branch.
4. `SOL-04` Only after a safe plan exists, run remove/create side effects and continue the normal startup flow.

## Contracts

| Contract ID | Related refs | Input / Output | Producer / Consumer | Notes |
| --- | --- | --- | --- | --- |
| `CTR-01` | `SOL-02`, `REQ-02` | Input: branch name; output: zero or one exact worktree path. | lifecycle planner / conflict resolution | Prefix matches are invalid. |
| `CTR-02` | `SOL-03`, `REQ-03` | Input: reuse choice and path; output: validated reusable worktree or failure. | lifecycle planner / startup flow | Unregistered or mismatched paths fail fast. |
| `CTR-03` | `SOL-04`, `REQ-04` | Input: destructive choice; output: exact cleanup target. | conflict resolution / side-effect executor | Cleanup stays scoped to the discovered conflicting path/branch. |

## Failure Modes

- `FM-01` A user chooses reuse for a path that is not a worktree in this repo; the command fails before later startup steps.
- `FM-02` A user chooses reuse for a path owned by another branch; the command fails before later startup steps.
- `FM-03` A conflicting branch exists without a reusable worktree; the user must choose rename or delete/recreate.

## Rollout / Backout

- `RB-01` Roll out by shipping the safer lifecycle planner and `make test` target without changing CLI flags.
- `RB-02` Back out by reverting the FT-008 patch set; no data migration is involved.

## ADR Dependencies

None.
