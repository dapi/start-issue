---
title: Product Metrics
doc_kind: product
doc_function: canonical
purpose: "Product success metrics and guardrails for start-issue."
derived_from:
  - ../dna/governance.md
  - context.md
status: active
audience: humans_and_agents
canonical_for:
  - product_metrics
  - success_measurement
---

# Product Metrics

## North Star

| Metric ID | Metric | Why it matters | Current baseline | Target | Review cadence |
| --- | --- | --- | --- | --- | --- |
| `NSM-01` | Successful issue-start completion for supported workflows | It captures the core product value | Covered by Bats regression scenarios | No known regression in supported flows | Each release |

## Product Metrics

| Metric ID | Metric | Owner | Baseline | Target | Measurement method | Source |
| --- | --- | --- | --- | --- | --- | --- |
| `MET-01` | Local verification pass rate | Maintainer | `make test` is canonical | Pass before release and handoff | Local command and CI | Makefile / GitHub Actions |
| `MET-02` | Config visibility | Maintainer | Missing issue and dry-run paths print effective config | Every config source change updates output tests | Bats assertions | `test/start_issue.bats` |
| `MET-03` | Release artifact integrity | Maintainer | Release asset plus `.sha256` | Installer/update verify checksum | Release workflow and tests | GitHub Releases |
| `MET-04` | Memory-bank navigation health | Maintainer/agent | New audit introduced | `scripts/check_memory_bank_index.py --max-depth 4` passes | Local audit in `make test` | memory-bank audit |

## Guardrails

| Guardrail ID | Metric | Why it must not regress | Threshold | Response |
| --- | --- | --- | --- | --- |
| `GR-01` | Public CLI compatibility | Existing scripts and users depend on flags/commands | Any behavior change without README/spec/test update | Stop and update contract or revert |
| `GR-02` | Worktree safety | Wrong-path reuse risks corrupting unrelated work | Any reuse without exact branch/path proof | Add regression test and fix before release |
| `GR-03` | Prompt safety | Silent prompt overwrite would damage future sessions | Any overwrite without explicit user intent | Fail fast and preserve proposal workflow |

## Instrumentation Constraints

- `ICON-01` This project currently uses repository evidence, tests, and release
  checks rather than telemetry.
- `ICON-02` Product metrics should be verified through deterministic local/CI
  checks where possible.

## Metric Change Policy

- New feature-level metrics stay in the feature package until they become shared
  product outcomes.
- If a metric affects release readiness, update this file and
  [engineering/testing-policy.md](../engineering/testing-policy.md) together.
