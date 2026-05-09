---
title: "FT-009: Solution"
doc_kind: feature
doc_function: canonical
purpose: "Canonical solution document for FT-009. Defines the selected modular architecture without redefining feature scope or acceptance criteria."
derived_from:
  - feature.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_009_scope
  - ft_009_acceptance_criteria
  - ft_009_delivery_status
  - detailed_current_system_inventory
  - implementation_sequence
---

# FT-009: Solution

## Selected Design

- `SOL-01` Keep `scripts/start-issue` as the stable CLI entrypoint and turn it into a thin bootstrap/orchestration layer.
- `SOL-02` Move cohesive responsibilities into focused shell modules such as `cli`, `config`, `github`, `branch`, `worktree`, `agent`, and `output`.
- `SOL-03` Make the workflow explicit as a pipeline with named stages and shared state passed through controlled functions.
- `SOL-04` Normalize agent-specific behavior behind a single internal adapter contract.
- `SOL-05` Introduce output/planning seams that can later support structured output and additional lifecycle commands without redesigning the whole script.
- `SOL-06` Document operational signals that would mean Bash has become the wrong implementation core and a targeted Python migration should be planned.

## Requirement Mapping

| Requirement ID | Solution / architecture refs | Notes |
| --- | --- | --- |
| `REQ-01` | `SOL-01`, `SOL-02`, `C4-L3-01` | Modular layout behind the same entrypoint. |
| `REQ-02` | `SOL-03`, `C4-L3-02` | Pipeline becomes explicit instead of implicit control flow. |
| `REQ-03` | `SOL-04`, `CTR-01` | Agent behavior is centralized by contract. |
| `REQ-04` | `SOL-03`, `SOL-05`, `CTR-02` | Planning and output boundaries support future growth. |
| `REQ-05` | `SOL-06` | Migration threshold is documented as part of the architecture package. |
| `REQ-06` | `SOL-01`, `SOL-02`, `SOL-03`, `SOL-04` | User-facing behavior stays stable. |

## To-Be C4 Model

| Level | Include? | Model refs | Selection rationale |
| --- | --- | --- | --- |
| System Context (L1) | no | `none` | The issue changes internal architecture, not external actors or boundaries. |
| Container (L2) | no | `none` | Runtime deployment shape is unchanged. |
| Component (L3) | yes | `C4-L3-01`, `C4-L3-02`, `C4-L3-03` | The change is primarily about internal shell components and their relationships. |

| C4 ref | Level | Element / relationship | To-be architecture statement | Related refs |
| --- | --- | --- | --- | --- |
| `C4-L3-01` | Component | CLI entrypoint to modules | `scripts/start-issue` bootstraps environment, loads modules, and invokes the pipeline instead of holding all business logic. | `SOL-01`, `SOL-02` |
| `C4-L3-02` | Component | Orchestration pipeline | A named pipeline coordinates parse, config, issue fetch, planning, execution, and agent launch. | `SOL-03`, `CTR-02` |
| `C4-L3-03` | Component | Agent adapter boundary | Agent operations are exposed through a consistent adapter contract instead of scattered conditionals. | `SOL-04`, `CTR-01` |

## Target Architecture

### Architecture Invariants

- The CLI contract, prompt precedence, and ordinary workflow behavior remain unchanged from the user's perspective.
- Module boundaries reduce coupling; they do not add a second hidden workflow.
- Agent-specific behavior has one source of truth.

### Target Shape

| Layer / responsibility | To-be role | Boundary / non-owner | Related refs |
| --- | --- | --- | --- |
| Entry script | Parse bootstrap environment, source modules, invoke the top-level pipeline. | Does not own detailed feature logic. | `SOL-01` |
| CLI / config modules | Parse arguments and resolve config sources into normalized state. | Do not fetch GitHub data or launch agents. | `SOL-02`, `SOL-03` |
| GitHub / planning modules | Fetch issue metadata and derive branch/worktree plan. | Do not own output formatting or agent launch. | `SOL-02`, `SOL-03` |
| Execution modules | Create/reuse worktree, run init, and coordinate side effects. | Do not parse CLI or infer agent-specific behavior. | `SOL-02`, `SOL-03` |
| Agent module | Validate adapter support, build launch commands, generate branch names, and improve prompts. | Does not own orchestration decisions outside the adapter contract. | `SOL-04`, `CTR-01` |
| Output / reporting module | Render logs, plan summaries, and future structured output shapes from normalized state. | Does not own core workflow decisions. | `SOL-05`, `CTR-02` |

## Internal Flow

1. `SOL-01` Entry script bootstraps shared paths and sources the module set.
2. `SOL-03` CLI and config stages normalize user input and effective settings.
3. `SOL-03` GitHub and planning stages fetch issue data and build the execution plan.
4. `SOL-04` Agent stage supplies agent-specific operations through one adapter contract.
5. `SOL-03` Execution stage applies the plan and returns normalized status/output data.
6. `SOL-05` Output stage renders user-facing text from that normalized state.

## Contracts

| Contract ID | Related refs | Input / Output | Producer / Consumer | Notes |
| --- | --- | --- | --- | --- |
| `CTR-01` | `SOL-04`, `REQ-03` | Input: agent id plus normalized request; output: validation result, launch command, branch-name proposal, or prompt-improvement proposal. | agent module / orchestration layer | Unsupported operations fail at the adapter boundary, not ad hoc in the main flow. |
| `CTR-02` | `SOL-03`, `SOL-05`, `REQ-04` | Input: normalized workflow state; output: plan data, side-effect results, and rendered output. | orchestration layer / output module | Enables future `--json` or `--plan-only` work without another structural rewrite. |
| `CTR-03` | `SOL-06`, `REQ-05` | Input: observed architecture pressure; output: migration recommendation criteria. | maintainers / future design work | Criteria are documented, not auto-enforced. |

## Accepted Local Decisions

- `SD-01` Use shell modules rather than a language rewrite to reduce change risk in this iteration.
- `SD-02` Keep the current CLI entrypoint path and user contract stable while moving logic behind sourced modules.
- `SD-03` Favor normalized state and contracts over introducing a large object model that would fight Bash ergonomics.
- `SD-04` Treat future `--json`, `resume`, `list`, and `cleanup` needs as design pressures that shape seams now but are not implemented yet.

## Change Surface

| Ref | Surface | Type | Why it changes |
| --- | --- | --- | --- |
| `SOL-01` - `SOL-05` | `scripts/start-issue` | code | Reduce it to bootstrap/orchestration responsibilities. |
| `SOL-02` - `SOL-05` | `scripts/lib/*.sh` or equivalent module paths | code | Introduce focused shell modules and shared contracts. |
| `SOL-01` - `SOL-06` | `test/start_issue.bats`, `test/helpers/fake-bin/*` | test | Keep behavior covered while internal structure changes. |
| `SOL-05` - `SOL-06` | `README.md`, `README.ru.md`, `doc/spec.md`, `memory-bank/features/FT-009/*` | doc | Document the modular architecture and migration threshold. |

## Failure Modes

- `FM-01` Sourced modules depend on uninitialized globals and change behavior relative to the old script.
- `FM-02` Adapter normalization misses an existing agent behavior path and silently regresses one agent.
- `FM-03` New module boundaries still leak formatting and control-flow concerns, reducing the value of the refactor.

## Rollout / Backout

- `RB-01` Roll out by landing the modular shell structure behind the unchanged CLI contract.
- `RB-02` Back out by collapsing the module split only if regression repair within the modular design is not feasible.

## Bash-to-Python Threshold

- If lifecycle commands start requiring richer subcommand semantics, persistent structured state, or complex data transforms that are awkward in shell, the next iteration should evaluate a Python core.
- If config grows beyond the current file-per-setting approach into nested or schema-heavy settings, a typed configuration model becomes a migration trigger.
- If tests must simulate increasingly stateful orchestration logic to keep shell safe, that maintenance cost is a migration signal rather than a reason to keep stretching Bash.

## ADR Dependencies

None.
