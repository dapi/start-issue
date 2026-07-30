---
title: "FT-016: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution plan for the real-Codex human-gate E2E suite."
derived_from:
  - brief.md
  - design.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_016_scope
  - ft_016_selected_design
  - ft_016_acceptance_criteria
---

# FT-016: Implementation Plan

## Current State / Reference Points

| Path | Role |
| --- | --- |
| `scripts/lib/start_issue/agent.sh` | Runtime state and terminal-status contract. |
| `test/helpers/fake-bin/codex` | Deterministic coverage that the new suite must avoid. |
| `Makefile` | Local test command entrypoints. |

## Test Strategy

| Surface | Coverage | Local command | CI |
| --- | --- | --- | --- |
| Runner syntax and integration | `bash -n`, shellcheck, documented target | `make test` | existing test job |
| Real Codex done/resume | Manual opt-in `SC-01`, `SC-02` | `CHK-01`, `CHK-02` | excluded by `NS-01` |

## Work Order

| Step ID | Implements | Goal | Verifies |
| --- | --- | --- | --- |
| `STEP-01` | `REQ-01` - `REQ-03`, `SOL-01` - `SOL-02` | Add the guarded E2E runner and Make target. | Script help and static checks. |
| `STEP-02` | `REQ-01` - `REQ-03`, `SD-01` - `SD-02` | Document prerequisites, commands, and retained artifacts. | README review and `make test`. |
| `STEP-03` | `SC-01`, `SC-02` | Offer manual acceptance commands. | `CHK-01`, `CHK-02`. |

## Stop Conditions / Fallback

| Stop ID | Trigger | Safe fallback |
| --- | --- | --- |
| `STOP-01` | Missing credentials, unavailable issue, or a live Codex failure | Keep deterministic Bats coverage as the release gate and inspect the preserved E2E log. |
