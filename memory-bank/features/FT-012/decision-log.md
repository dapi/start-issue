---
title: "FT-012: Decision Log"
doc_kind: feature
doc_function: decision_log
purpose: "Resolved local decisions and conflict handling for FT-012. Stores only decisions grounded in the current feature package and in-scope artifacts."
derived_from:
  - https://github.com/dapi/start-issue/issues/12
  - feature.md
  - solution.md
  - ../../../scripts/lib/start_issue/agent.sh
  - ../../../scripts/lib/start_issue/output.sh
  - ../../../scripts/lib/start_issue/init.sh
status: active
audience: humans_and_agents
---

# FT-012: Decision Log

## Decision Entries

| Decision ID | Date | Status | Topic | Decision |
| --- | --- | --- | --- | --- |
| `DL-01` | 2026-05-24 | accepted | Model precedence | Add optional model resolution as an independent config axis with precedence `CLI -> project -> user -> environment -> built-in unset`, while leaving existing agent precedence unchanged. |
| `DL-02` | 2026-05-24 | accepted | Adapter validation | Centralize explicit model support checks in the agent adapter boundary so unsupported combinations fail clearly instead of silently ignoring the model. |
| `DL-03` | 2026-05-24 | accepted | Unset model compatibility | Treat model as optional: launch commands omit model flags when unset, `agent=none` never consumes model, and existing non-launch adapter defaults remain unchanged until the user configures an explicit model. |

## FPF-Closed Questions

### `FPF-01`: What backward-compatible meaning should "no model configured" have?

#### Why this mattered

The issue requires explicit model selection and also requires backward compatibility when no model is configured. Current code already differs by operation: launch commands omit model flags, while Claude non-launch operations in `scripts/lib/start_issue/agent.sh` explicitly use `--model haiku`.

#### Available facts

- Issue #12 defines model precedence with built-in/default behavior "omit model and let the selected agent CLI decide".
- The same issue explicitly says launch commands should preserve current behavior and omit model arguments when no model is configured.
- Current code in `scripts/lib/start_issue/agent.sh` hardcodes `claude --model haiku` for prompt improvement and AI branch-name generation.
- Current launch commands in `scripts/lib/start_issue/agent.sh` do not pass a model argument.

#### FPF reasoning summary

- Bounded contexts: config resolution, launch commands, and non-launch adapter operations are distinct contexts and should not be conflated.
- Evidence discipline: only launch backward compatibility is stated as an explicit acceptance condition; the current non-launch adapter default is a concrete in-repo fact.
- Decision rule: prefer the smallest change that satisfies explicit acceptance criteria without inventing unsupported behavior.

#### Resolution

Use one optional resolved model value across the adapter boundary, but define backward compatibility per operation:

- Launch commands omit model flags when no model is configured.
- `agent=none` never passes a model to a launch command.
- Existing non-launch adapter defaults remain unchanged until an explicit model is configured, at which point the adapter must either use it or fail clearly if unsupported.

#### Conflict handled

This resolves the tension between the issue's general "built-in unset" language and the repository's current Claude internal-call behavior. The conflict is handled conservatively in favor of explicit acceptance criteria and present code facts, not assumptions about undocumented CLI defaults.

### `FPF-02`: Where should unsupported agent/model combinations be rejected?

#### Why this mattered

Issue #12 requires a clear validation error for unsupported model selection, but the current workflow has multiple entry points: regular execution, dry-run, missing-issue display, and `init`.

#### Available facts

- The current architecture already centralizes agent-specific operations in `scripts/lib/start_issue/agent.sh`.
- `start-issue init` already validates and writes agent config and is explicitly in scope for model config.
- Dry-run output is part of the acceptance criteria and must reflect the real launch command shape.

#### FPF reasoning summary

- Role-method-work alignment: adapter support knowledge belongs with the adapter boundary, not scattered across CLI/help/init.
- Trust boundary: a config value should be validated before a command is shown or persisted as if it were supported.

#### Resolution

Perform explicit model-support validation through the adapter boundary before rendering launch commands and before persisting an agent/model pair through `init`.

#### Conflict handled

No cross-document conflict remained after assigning support knowledge to the existing adapter boundary and making `init` consume the same validation contract.
