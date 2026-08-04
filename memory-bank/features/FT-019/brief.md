---
title: "FT-019: CI sandbox E2E coverage"
doc_kind: feature
doc_function: canonical
purpose: "Canonical brief for running deterministic end-to-end smoke scenarios against the built Go CLI in CI/CD."
derived_from:
  - ../../flows/feature-flow.md
  - ../../engineering/testing-policy.md
  - ../../ops/development.md
status: active
delivery_status: in_progress
audience: humans_and_agents
must_not_define:
  - implementation_sequence
---

# FT-019: CI sandbox E2E coverage

## What

### Problem

The existing real-Codex E2E requires authenticated GitHub, a private fixture
repository, and an interactive external agent, so it cannot be a reliable CI
gate. Unit and parity tests do not exercise the built executable against a real
git worktree process boundary.

### Outcome

CI runs a network-free sandbox E2E against the built Go binary and verifies the
complete local issue-start path with controlled external command fakes.

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Built-binary workflow coverage | Unit/parity tests plus manual real-agent E2E | At least one deterministic CI E2E job covers worktree/init/agent and dry-run paths | CI job output |

## Scope

- `REQ-01` Run the actual built `start-issue` binary in an isolated temporary sandbox.
- `REQ-02` Use a real local git repository/worktree and fake `gh` and agent CLIs; do not require network, credentials, or real agents.
- `REQ-03` Cover an executing Kimi launch with rendered issue prompt, model, cwd, and `init.sh`.
- `REQ-04` Cover the dry-run/no-agent path and prove it creates no worktree.
- `REQ-05` Expose the sandbox E2E through Make and run it in CI/CD.

## Non-Scope

- `NS-01` Do not replace the existing manual real-Codex human-gate E2E.
- `NS-02` Do not claim that fake agents prove vendor CLI compatibility or model behavior.
- `NS-03` Do not use GitHub API, authenticated secrets, Docker, or persistent shared state.

## Constraints / Assumptions

- `ASM-01` CI provides Go, Bash, git, and standard POSIX utilities.
- `CON-01` The test must clean its own uniquely-created temporary directory on success and failure.

## Design Requirement Decision

| Decision | Reason | Downstream owner |
| --- | --- | --- |
| `Design required: yes` | The feature introduces a CI test boundary, fake external processes, cleanup guarantees, and a new workflow gate. | `design.md` |

## Verify

### Exit Criteria

- `EC-01` Sandbox E2E runs locally against a built binary without network or secrets.
- `EC-02` The executing scenario proves git worktree creation, `init.sh`, rendered prompt, model, and Kimi cwd.
- `EC-03` The dry-run scenario proves no worktree directory is created.
- `EC-04` CI invokes the same Make target and reports a failing exit code on assertion failure.

### Traceability matrix

| Requirement ID | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- |
| `REQ-01` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-02` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-03` | `EC-02`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-04` | `EC-03`, `SC-02` | `CHK-01` | `EVID-01` |
| `REQ-05` | `EC-04` | `CHK-02` | `EVID-02` |

### Acceptance Scenarios

- `SC-01` Given an isolated git fixture and fake external CLIs, when the built binary starts issue #42 with Kimi, then the worktree, init marker, rendered prompt, model, and cwd are correct.
- `SC-02` Given the same fixture, when the built binary runs `--dry-run --agent none`, then it prints the plan and does not create the requested worktree directory.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`–`EC-03`, `SC-01`, `SC-02` | `make e2e-sandbox` | Both sandbox scenarios pass | `artifacts/ft-019/verify/sandbox-e2e/` |
| `CHK-02` | `EC-04` | Inspect CI `sandbox-e2e` job | The job runs the Make target and fails on non-zero status | `artifacts/ft-019/verify/ci/` |

### Evidence

- `EVID-01` Local sandbox E2E output.
- `EVID-02` CI job output and workflow definition.
