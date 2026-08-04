---
title: "FT-019: Design"
doc_kind: feature
doc_function: canonical
purpose: "Solution-space contract for deterministic built-binary E2E execution in a temporary local sandbox."
derived_from:
  - brief.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_019_scope
  - ft_019_acceptance_criteria
  - implementation_sequence
---

# FT-019: Design

## C4 Applicability

| C4 ID | Decision | Trigger / reason | Artifact |
| --- | --- | --- | --- |
| `C4-00` | `not required` | This is a test harness and CI job within the existing CLI container; it creates no deployed runtime boundary. | none |

## Selected Solution

- `SOL-01` Build the Go binary first and invoke it as a subprocess from `test/e2e/sandbox.sh`.
- `SOL-02` Create a temporary local git repository with a commit and origin remote, allowing the product's real git/worktree code to run.
- `SOL-03` Put fake `gh` and Kimi executables first on PATH. The fake `gh` serves issue JSON and auth status; the fake Kimi records cwd and arguments.
- `SOL-04` Run two scenarios: a real worktree/init/Kimi launch and a no-agent dry-run.
- `SOL-05` Add `make e2e-sandbox` and a dedicated CI job. Keep real-Codex E2E opt-in and separate.

## Alternatives Considered

| Alternative ID | Option | Why not selected |
| --- | --- | --- |
| `ALT-01` | Use the real GitHub fixture and agent in CI | Requires credentials/network and is nondeterministic. |
| `ALT-02` | Test only Go functions | Does not verify the built executable and process/filesystem boundaries. |
| `ALT-03` | Fake git as well | Would not exercise actual worktree creation and reuse behavior. |

## Accepted Local Decisions

- `SD-01` The sandbox has no external network dependency; every command that could access a service is controlled by PATH fakes.
- `SD-02` The temporary root is created by `mktemp` and removed by an EXIT trap, including failure paths.
- `SD-03` Assertions are made against explicit artifacts and command logs, not broad output snapshots.

## Contracts

| Contract ID | Input / Output | Producer / Consumer | Semantics / Constraints |
| --- | --- | --- | --- |
| `CTR-01` | Built binary + isolated env → exit status/artifacts | E2E script / start-issue | Non-zero means CI failure; no credentials are needed. |
| `CTR-02` | Fake Kimi log → cwd and rendered args | fake CLI / E2E assertions | cwd equals the created worktree and model/prompt are present. |

## Invariants

- `INV-01` The sandbox never invokes a real `gh`, Kimi, Codex, Claude, or Pi binary.
- `INV-02` The sandbox never writes outside its unique temporary root except the tested binary's normal process execution.
- `INV-03` Dry-run does not create its requested worktree directory.

## Failure Modes

- `FM-01` PATH fake is not executable or is shadowed; the script fails before claiming PASS.
- `FM-02` Cleanup is skipped after an assertion failure; the EXIT trap still removes the unique root.
- `FM-03` A product change bypasses the worktree cwd or prompt rendering; explicit log assertions fail.

## Traceability

| Requirement ID | Solution refs | Contracts / invariants | Failure refs |
| --- | --- | --- | --- |
| `REQ-01`–`REQ-02` | `SOL-01`–`SOL-03` | `CTR-01`, `INV-01`, `INV-02` | `FM-01`, `FM-02` |
| `REQ-03` | `SOL-02`–`SOL-04` | `CTR-02`, `INV-03` | `FM-03` |
| `REQ-04`–`REQ-05` | `SOL-04`, `SOL-05` | `CTR-01`, `INV-03` | `FM-02` |
