---
title: "FT-019: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution plan for sandbox E2E coverage and CI integration."
derived_from:
  - brief.md
  - design.md
  - ../../engineering/testing-policy.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_019_scope
  - ft_019_selected_design
  - ft_019_acceptance_criteria
---

# FT-019: Implementation Plan

## Grounding / Support References

| Path / module | Current role | Why relevant | Reuse / mirror |
| --- | --- | --- | --- |
| `test/e2e/human-gate.sh` | Existing real-agent E2E | Establishes E2E script conventions and opt-in boundary | Keep separate because it needs secrets/interactive Codex |
| `Makefile` | Build/test entrypoint | Owns the local and CI target | Add sandbox target beside human-gate target |
| `.github/workflows/ci.yml` | CI checks | Runs Go build and tests | Add network-free sandbox job |
| `cmd/start-issue/main.go` | Product executable | Must be exercised as a built subprocess | Do not add test-only product hooks |

## Test Strategy

| Test surface | Canonical refs | Planned automated coverage | Required command |
| --- | --- | --- | --- |
| Built binary issue start | `REQ-01`–`REQ-03`, `SC-01` | Real local git/worktree, fake gh/Kimi, init marker and cwd/args assertions | `make e2e-sandbox` |
| Built binary dry-run | `REQ-04`, `SC-02` | Output and no-directory assertion | `make e2e-sandbox` |
| CI wiring | `REQ-05` | Dedicated workflow job invokes Make target | CI run |

## Environment Contract

| Area | Contract | Failure symptom |
| --- | --- | --- |
| Toolchain | Go binary is built by Make; Bash, git, and POSIX utilities are available | Missing executable or command failure |
| Network/secrets | No network, GitHub auth, or agent credentials are allowed | Any fake command miss or unexpected external call fails |
| Cleanup | Temporary fixture root is unique and EXIT-trapped | No retained sandbox state after run |

## Preconditions

- `PRE-01` Go build succeeds.
- `PRE-02` `git` is available in the CI runner.

## Workstreams

- `WS-01` Implement the sandbox fixture and assertions (`REQ-01`–`REQ-04`).
- `WS-02` Add Make and CI wiring (`REQ-05`).
- `WS-03` Document usage and verify all repository gates.

## Execution Order

1. `STEP-01` Add `test/e2e/sandbox.sh`; run it against a local build and verify both scenarios.
2. `STEP-02` Add `make e2e-sandbox` and CI job; verify the job has no secrets or network setup.
3. `STEP-03` Update README/ops docs and run `make e2e-sandbox` plus `make test`.

## Stop Conditions / Fallback

- `STOP-01` If the script needs a real external service, stop and keep the test out of CI; do not weaken the no-network contract.
- `STOP-02` If cleanup or isolation fails, stop before merging the CI job.
