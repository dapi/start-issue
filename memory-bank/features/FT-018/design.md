---
title: "FT-018: Design"
doc_kind: feature
doc_function: canonical
purpose: "Solution-space contract for agent-specific launch arguments and working-directory handling."
derived_from:
  - brief.md
  - ../../../doc/spec.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_018_scope
  - ft_018_acceptance_criteria
  - implementation_sequence
---

# FT-018: Design

## Design Pack

| Artifact | Role | Owns |
| --- | --- | --- |
| `design.md` | Feature-local solution owner | `SOL-*`, `C4-*`, `SD-*`, `CTR-*`, `INV-*`, `FM-*` |

## C4 Applicability

| C4 ID | Decision | Trigger / reason | Artifact |
| --- | --- | --- | --- |
| `C4-00` | `not required` | Local adapter command mapping inside the existing CLI container; no runtime boundary changes. | none |

## Selected Solution

- `SOL-01` Keep each agent adapter responsible for its own arguments and cwd policy.
- `SOL-02` For Kimi, run `kimi [--model MODEL] -p PROMPT` with `cmd.Dir = WORKTREE_PATH`; prompt mode provides automatic permission handling, so `--yolo` is omitted.
- `SOL-03` For Kimi helper calls, run the same command with `cmd.Dir = REPOSITORY_ROOT`; do not encode cwd as a removed CLI option.
- `SOL-04` Keep Codex's `--cd`, Claude/Pi cwd behavior, and `none` manual output unchanged except for documentation alignment.

## Alternatives Considered

| Alternative ID | Option | Why not selected |
| --- | --- | --- |
| `ALT-01` | Keep `--work-dir` and require an older Kimi CLI | Fails with the installed/current CLI and makes start-issue unusable for Kimi users. |
| `ALT-02` | Add runtime version probing and two Kimi syntaxes | Adds fragile branching and is outside the requested compatibility fix. |
| `ALT-03` | Use `--add-dir` while keeping the caller cwd | Does not make the worktree the primary Kimi workspace. |

## Accepted Local Decisions

- `SD-01` Process cwd is the canonical worktree transport for Kimi.
- `SD-02` Kimi prompt mode omits `--yolo` because the current CLI rejects that combination and auto-approves prompt-mode tool calls.

## Contracts

| Contract ID | Input / Output | Producer / Consumer | Semantics / Constraints |
| --- | --- | --- | --- |
| `CTR-01` | agent, model, worktree, prompt → process command + cwd | adapter / selected agent CLI | Kimi has no path flag in the supported CLI; cwd must equal the requested directory. |

## Invariants

- `INV-01` No Kimi command emitted by the product contains `--work-dir`.
- `INV-02` No Kimi prompt command emits the incompatible `--yolo` and `-p` combination.
- `INV-03` The selected worktree remains the process cwd for Kimi launch and repository-root cwd for helper calls.

## Failure Modes

- `FM-01` A future Kimi CLI changes its flags; the adapter tests and explicit docs must be updated together rather than silently probing alternatives.
- `FM-02` Kimi is launched with the caller cwd due to a missed `cmd.Dir`; the cwd test must fail.

## Traceability

| Requirement ID | Solution refs | Contracts / invariants | Failure refs |
| --- | --- | --- | --- |
| `REQ-01` | `SOL-01`, `SOL-04` | `CTR-01` | `FM-01` |
| `REQ-02`–`REQ-04` | `SOL-02`, `SOL-03` | `CTR-01`, `INV-01`–`INV-03` | `FM-01`, `FM-02` |
| `REQ-05` | `SOL-01` | `INV-01` | `FM-01` |
