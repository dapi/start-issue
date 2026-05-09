---
title: "FT-004: Solution"
doc_kind: feature
doc_function: canonical
purpose: "Canonical solution document for FT-004. Defines the selected prompt-improvement proposal workflow without redefining feature scope or acceptance criteria."
derived_from:
  - feature.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_004_scope
  - ft_004_acceptance_criteria
  - ft_004_delivery_status
  - detailed_current_system_inventory
  - implementation_sequence
---

# FT-004: Solution

## Selected Design

- `SOL-01` Add `--improve-prompt` as a mode that reuses existing prompt precedence to identify the active prompt template.
- `SOL-02` Generate a complete improved prompt template through the selected agent, using the current issue and active template as context.
- `SOL-03` Write the generated template to a reviewable proposal file instead of mutating the active prompt.
- `SOL-04` Exit prompt-improvement mode before any worktree-start side effects.

## Requirement Mapping

| Requirement ID | Solution / architecture refs | Notes |
| --- | --- | --- |
| `REQ-01` | `SOL-01`, `C4-L3-01` | Existing prompt resolution remains the source of truth. |
| `REQ-02` | `SOL-02`, `CTR-01` | The selected agent returns a complete prompt template, not a diff. |
| `REQ-03` | `SOL-03`, `CTR-02` | Proposal output defaults are deterministic and non-destructive. |
| `REQ-04` | `SOL-04`, `C4-L3-02` | Improvement mode returns before Zellij/worktree/agent launch. |
| `REQ-05` | `SOL-01`, `SOL-02`, `SOL-03`, `SOL-04` | Docs and tests cover the changed contract. |

## To-Be C4 Model

| Level | Include? | Model refs | Selection rationale |
| --- | --- | --- | --- |
| System Context (L1) | no | `none` | No external actors, systems, or system boundary change. |
| Container (L2) | no | `none` | The feature changes one shell script and documentation, not runtime/deployment containers. |
| Component (L3) | yes | `C4-L3-01`, `C4-L3-02` | Prompt resolution and mode orchestration gain new responsibilities. |

| C4 ref | Level | Element / relationship | To-be architecture statement | Related refs |
| --- | --- | --- | --- | --- |
| `C4-L3-01` | Component | `scripts/start-issue` prompt resolution | Prompt improvement consumes the same active prompt template selected by normal precedence. | `SOL-01`, `CTR-01` |
| `C4-L3-02` | Component | `scripts/start-issue` workflow control | Prompt improvement is an early-exit mode after issue fetch and before side-effectful startup steps. | `SOL-04`, `CTR-02` |

## Target Architecture

### Architecture Invariants

- Prompt template precedence remains unchanged.
- Normal start workflow remains unchanged when `--improve-prompt` is absent.
- Generated prompt improvements are reviewable artifacts and are not auto-applied.

### Target Shape

| Layer / responsibility | To-be role | Boundary / non-owner | Related refs |
| --- | --- | --- | --- |
| Prompt template selection | Select exactly one active prompt template by existing precedence. | Does not improve or write templates. | `SOL-01` |
| Prompt improvement request | Build an agent request from issue metadata, prompt source, supported placeholders, and current template. | Does not render placeholders or create worktrees. | `SOL-02`, `CTR-01` |
| Proposal writing | Write the complete improved prompt template to a proposal path. | Does not overwrite existing proposal or active prompt. | `SOL-03`, `CTR-02` |
| Workflow control | Stop after proposal generation. | Does not rename Zellij, generate branch, create worktree, run init, or launch agent. | `SOL-04` |

## Accepted Local Decisions

- `SD-01` Use `--improve-prompt` as an explicit mode because prompt improvement is separate from starting a worktree.
- `SD-02` Use `--prompt-output-file` to let users choose the proposal path without adding an apply/overwrite mode.
- `SD-03` For file-backed prompts, default proposal path is adjacent to the source as `*.improved.md`; for built-in or inline prompts, default proposal path is `.start-issue/prompt.improved.md`.

## Change Surface

| Ref | Surface | Type | Why it changes |
| --- | --- | --- | --- |
| `SOL-01` - `SOL-04` | `scripts/start-issue` | code | Add prompt-improvement mode, agent request, proposal path handling, and early exit. |
| `SOL-01` - `SOL-04` | `test/start_issue.bats`, `test/helpers/fake-bin/*` | test | Cover proposal writing, built-in prompt output path, dry-run, and invalid agent. |
| `SOL-01` - `SOL-04` | `README.md`, `README.ru.md`, `doc/spec.md` | doc | Document the new mode and non-destructive proposal contract. |

## Internal Flow

1. `SOL-01` Resolve agent, prompt template, repo, base branch, and issue as usual.
2. `SOL-04` If `--improve-prompt` is active, validate that the agent is not `none`.
3. `SOL-02` Fetch issue metadata and build the improvement request.
4. `SOL-02` Ask the selected agent to return only the complete improved prompt template.
5. `SOL-03` Write the result to the proposal path and exit before startup side effects.

## Contracts

| Contract ID | Related refs | Input / Output | Producer / Consumer | Notes |
| --- | --- | --- | --- | --- |
| `CTR-01` | `SOL-02`, `REQ-02` | Input: active prompt template plus issue metadata; output: complete improved prompt template. | selected agent / `start-issue` | Agent commentary, diffs, and code fences are not part of the requested output. |
| `CTR-02` | `SOL-03`, `REQ-03` | Input: generated template; output: proposal file. | `start-issue` / human reviewer | Existing proposal files are not overwritten. |
| `CTR-03` | `SOL-04`, `REQ-04` | Input: `--improve-prompt`; output: no worktree-start side effects. | CLI user / `start-issue` | Zellij rename, branch generation, worktree, init, and launch are skipped. |

## Failure Modes

- `FM-01` The selected agent returns empty output or fails; `start-issue` fails without writing a proposal.
- `FM-02` The proposal path already exists; `start-issue` fails without overwriting it.

## Rollout / Backout

- `RB-01` Roll out by shipping the explicit `--improve-prompt` mode; normal starts remain unchanged.
- `RB-02` Back out by not using `--improve-prompt` or by deleting the proposal file before accepting it.

## ADR Dependencies

None.
