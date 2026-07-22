---
title: Testing Policy
doc_kind: engineering
doc_function: canonical
purpose: "Testing policy for start-issue: required Go checks, release checks, memory-bank audit, and manual-only exceptions."
derived_from:
  - ../dna/governance.md
  - ../flows/feature-flow.md
  - ../../README.md
  - ../../doc/spec.md
status: active
canonical_for:
  - repository_testing_policy
  - feature_test_case_inventory_rules
  - automated_test_requirements
  - sufficient_test_coverage_definition
  - manual_only_verification_exceptions
  - simplify_review_discipline
  - verification_context_separation
must_not_define:
  - feature_acceptance_criteria
  - feature_scope
audience: humans_and_agents
---

# Testing Policy

## Canonical Local Command

```bash
make test
```

`make test` runs:

1. `gofmt` cleanliness for `cmd/`
2. `go vet ./...`
3. `go test ./...`
4. `python3 scripts/check_memory_bank_index.py --max-depth 4`
5. `git diff --check`

## Test Stack

- Formatting: `gofmt`
- Static analysis: `go vet`
- Behavior/regression tests: Go tests under `cmd/` and future Go packages
- Opt-in real-agent E2E smoke tests: scripts under `test/e2e/`, run manually and never in CI
- Memory-bank navigation: `scripts/check_memory_bank_index.py`
- Whitespace/conflict-marker check: `git diff --check`

## Core Rules

- Any deterministic behavior change needs automated Go test coverage.
- Any public CLI contract change must update help/output assertions and docs.
- Any config precedence change must test the winning source and displayed source.
- Any worktree lifecycle change must test safe reuse/reject/delete behavior.
- Any release/update change must test version comparison, latest lookup failure,
  checksum/install path, or no-op behavior as appropriate.
- Any memory-bank structure change must keep the index audit green.

## Ownership Split

- Feature `brief.md` owns acceptance scenarios and evidence IDs for new
  feature-flow packages.
- Legacy `feature.md` packages own their existing `SC-*`, `CHK-*`, and `EVID-*`
  until migrated.
- `implementation-plan.md` owns concrete test commands and sequencing.
- Go tests own executable regression behavior.

## Sufficient Coverage

Coverage is sufficient when the changed behavior is exercised at the CLI level
or at the closest practical shell helper boundary, and failure behavior is
covered when it affects user trust or data safety.

Line coverage is not a target. Scenario coverage matters more:

- normal path;
- precedence or adapter-specific variation;
- failure path;
- no unintended side effects in dry-run/setup/update modes.

## Manual-Only Exceptions

Manual-only verification is acceptable only for:

- real external agent CLI behavior that cannot be faked deterministically;
- live GitHub/network behavior beyond mocked/fake helper coverage;
- visual/manual review of long help text or docs when no stable assertion is
  useful.

Real-agent E2E scripts must require an explicit opt-in environment variable,
avoid fake agent binaries, preserve diagnostic artifacts, and stay outside
`make test` and CI.

For each manual-only gap, record the reason and the manual procedure in the
feature plan or final handoff.

## Simplify Review

After tests pass, review for shell complexity:

- avoid scattered agent-specific branching;
- prefer small functions with explicit inputs over implicit global mutation;
- avoid abstraction unless it removes real duplication or clarifies a boundary;
- keep user-facing output stable and direct.

## Verification Context Separation

1. Functional verification: run relevant tests or `make test`.
2. Simplify review: inspect the changed shell/docs for unnecessary complexity.
3. Acceptance: map results back to `SC-*`/`CHK-*` or the user request.
