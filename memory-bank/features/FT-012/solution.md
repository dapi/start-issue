---
title: "FT-012: Solution"
doc_kind: feature
doc_function: canonical
purpose: "Canonical solution document for FT-012. Defines the selected config-resolution and adapter-validation design without redefining feature scope or acceptance criteria."
derived_from:
  - feature.md
  - decision-log.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_012_scope
  - ft_012_acceptance_criteria
  - ft_012_delivery_status
  - implementation_sequence
---

# FT-012: Solution

## Selected Design

- `SOL-01` Resolve model as a first-class optional config value parallel to the existing resolved agent, including a tracked `MODEL_SOURCE`.
- `SOL-02` Extend CLI/help/missing-issue/config-display flows so agent and model are shown together, including their config files and environment variables.
- `SOL-03` Extend `init` so project/user config can write `agent` and `model` side by side without changing prompt-file ownership.
- `SOL-04` Extend the agent adapter contract so each adapter declares whether explicit model selection is supported for each relevant operation and how the model flag is rendered.
- `SOL-05` Fail fast on unsupported agent/model combinations instead of silently dropping `--model` or config-driven model values.
- `SOL-06` Keep model optional: launch commands omit model arguments when model is unset, `agent=none` never consumes model, and existing non-launch adapter defaults remain unchanged unless an explicit model is configured.
- `SOL-07` Update README, Russian README, spec, examples, and test fixtures together so agent/model behavior is documented as one coherent user contract.

## Requirement Mapping

| Requirement ID | Solution / architecture refs | Notes |
| --- | --- | --- |
| `REQ-01` | `SOL-01`, `CTR-01` | Model resolution mirrors agent resolution as a separate config axis. |
| `REQ-02` | `SOL-01`, `SD-01` | Agent precedence stays intact while model precedence is added independently. |
| `REQ-03` | `SOL-02`, `CTR-03` | Runtime-facing output shows resolved values, and help text documents the config surfaces. |
| `REQ-04` | `SOL-03`, `CTR-04` | Init writes both config files at the selected scope. |
| `REQ-05` | `SOL-04`, `SOL-05`, `CTR-02` | Adapter capability and validation are centralized. |
| `REQ-06` | `SOL-06`, `SD-03` | Unset model preserves backward-compatible launch behavior. |
| `REQ-07` | `SOL-07` | Docs and tests move together. |

## To-Be C4 Model

| Level | Include? | Model refs | Selection rationale |
| --- | --- | --- | --- |
| System Context (L1) | no | `none` | External actors and repos do not change. |
| Container (L2) | no | `none` | Runtime/container shape is unchanged. |
| Component (L3) | yes | `C4-L3-01`, `C4-L3-02`, `C4-L3-03` | The change affects config resolution, init, and agent-adapter boundaries. |

| C4 ref | Level | Element / relationship | To-be architecture statement | Related refs |
| --- | --- | --- | --- | --- |
| `C4-L3-01` | Component | CLI/config resolution | CLI and config modules resolve `AGENT` and optional `MODEL` with independent source tracking. | `SOL-01`, `CTR-01` |
| `C4-L3-02` | Component | Init/config writer | Init writes agent/model config at the selected scope and reports both to the user. | `SOL-03`, `CTR-04` |
| `C4-L3-03` | Component | Agent adapter boundary | Adapter validation and command-building decide whether and how model is consumed per operation. | `SOL-04`, `SOL-05`, `SOL-06`, `CTR-02` |

## Target Architecture

### Architecture Invariants

- Agent precedence remains unchanged.
- Model is optional and never silently ignored once configured.
- `agent=none` is still a no-launch path.
- Docs, help text, and tests describe the same precedence and validation contract.

### Target Shape

| Layer / responsibility | To-be role | Boundary / non-owner | Related refs |
| --- | --- | --- | --- |
| CLI/config modules | Parse `--model`, resolve model files/env/default, and track `MODEL_SOURCE`. | Do not decide adapter-specific command syntax. | `SOL-01`, `CTR-01` |
| Output/help flows | Show resolved agent/model state and where both are configured. | Do not re-implement precedence logic. | `SOL-02`, `CTR-03` |
| Init flow | Write `agent` and `model` files at project or user scope. | Does not own adapter command-building. | `SOL-03`, `CTR-04` |
| Agent module | Validate model support, build model-aware commands, and preserve no-model behavior when unset. | Does not change generic config precedence. | `SOL-04`, `SOL-05`, `SOL-06`, `CTR-02` |
| Docs/tests | Describe and verify one combined agent/model contract. | Do not introduce behavior not present in the implementation. | `SOL-07` |

## Internal Flow

1. `SOL-01` CLI/config stages resolve `AGENT`, `AGENT_SOURCE`, optional `MODEL`, and `MODEL_SOURCE`.
2. `SOL-02` Help and missing-issue flows show the effective agent/model config and where each file lives.
3. `SOL-03` Init resolves agent/model defaults for the selected scope and writes both files.
4. `SOL-04` Agent validation checks whether the selected operation supports explicit model selection for the resolved agent.
5. `SOL-05` Unsupported combinations fail before execution.
6. `SOL-06` Launch and other adapter operations consume the explicit model only when configured and supported.

## Contracts

| Contract ID | Related refs | Input / Output | Producer / Consumer | Notes |
| --- | --- | --- | --- | --- |
| `CTR-01` | `SOL-01`, `REQ-01`, `REQ-02` | Input: CLI args, config files, env; output: `AGENT`, `AGENT_SOURCE`, optional `MODEL`, `MODEL_SOURCE`. | cli/config modules / pipeline | Model resolution mirrors agent resolution but keeps its own precedence chain and built-in unset state. |
| `CTR-02` | `SOL-04`, `SOL-05`, `SOL-06`, `REQ-05`, `REQ-06` | Input: agent id, optional model, operation id; output: validation result and command shape. | agent module / pipeline | Central place for support matrix and explicit errors. |
| `CTR-03` | `SOL-02`, `REQ-03`, `REQ-07` | Input: resolved config state; output: help text and runtime summaries. | output/help modules / user | Keeps all user-visible config reporting aligned. |
| `CTR-04` | `SOL-03`, `REQ-04` | Input: init scope plus selected config values; output: written `agent` and `model` files. | init flow / filesystem | Existing prompt-config behavior remains separate. |

## Accepted Local Decisions

- `SD-01` Model precedence mirrors agent precedence structurally but uses its own files and environment variable so existing agent behavior is untouched.
- `SD-02` Adapter support for explicit model selection is validated centrally; unsupported combinations return a clear error rather than degraded execution.
- `SD-03` An unset model preserves current behavior: launch commands omit model flags, while non-launch adapter operations keep existing adapter defaults until an explicit model is provided.
- `SD-04` The feature documents support requirements for all existing agents without inventing vendor-specific model catalogs.

## Change Surface

| Ref | Surface | Type | Why it changes |
| --- | --- | --- | --- |
| `SOL-01` - `SOL-03` | `scripts/lib/start_issue/cli.sh`, `config.sh`, `init.sh`, `pipeline.sh`, `output.sh` | code | Resolve, display, and persist model config beside agent config. |
| `SOL-04` - `SOL-06` | `scripts/lib/start_issue/agent.sh` | code | Centralize model-aware validation and command construction. |
| `SOL-04` - `SOL-07` | `test/start_issue.bats`, `test/helpers/fake-bin/*` | test | Cover precedence, dry-run command shapes, init writes, and unsupported combinations. |
| `SOL-02` - `SOL-07` | `README.md`, `README.ru.md`, `doc/spec.md`, `docs/agent-examples.md`, `docs/agent-examples.ru.md`, `memory-bank/features/FT-012/*` | doc | Document one coherent agent/model user contract. |

## Failure Modes

- `FM-01` Model precedence disagrees between help text, runtime behavior, and init output.
- `FM-02` An adapter accepts `--model` but renders the wrong flag or places it in the wrong command position.
- `FM-03` Unsupported model selection is silently ignored for one agent path.
- `FM-04` No-model compatibility regresses because launch or internal adapter flows always inject a model flag.

## Rollout / Backout

- `RB-01` Roll out by landing model resolution, adapter validation, docs, and tests together.
- `RB-02` Back out by removing model-specific resolution and docs if compatibility cannot be preserved, while keeping existing agent-selection behavior intact.

## ADR Dependencies

See [decision-log.md](decision-log.md).
