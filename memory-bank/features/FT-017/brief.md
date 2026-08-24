---
title: "FT-017: Codex human-gate delivery permissions"
doc_kind: feature
doc_function: canonical
purpose: "Canonical brief for making Codex human-gate capabilities explicit and supporting opt-in end-to-end Git delivery."
derived_from:
  - ../../flows/feature-flow.md
  - ../../product/context.md
  - ../../engineering/testing-policy.md
  - ../FT-015/feature.md
  - https://github.com/dapi/start-issue/issues/37
status: active
delivery_status: in_progress
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - solution_space
---

# FT-017: Codex human-gate delivery permissions

## What

### Problem

FT-015 established a resumable Codex batch flow, but its runtime contract is
limited to `workspace-write`. That mode can edit the prepared worktree, yet it
does not reliably provide network access or Git metadata writes. A run can
therefore complete implementation and tests but fail before reading complete
GitHub context, committing, pushing, or creating a pull request.

The current help does not distinguish working-tree automation from full Git
delivery. Capability failures can consequently look like task-level
`HUMAN_GATE` decisions even though the real blocker is the launcher policy.

The obsolete argument-order failure described in issue #37 has already been
removed on `master`; this feature must preserve compatibility with the current
supported Codex CLI while closing the remaining capability-contract gap.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Human-gate capability contract visibility | One implicit `workspace-write` command | Every run reports either restricted or full-delivery permissions | Dry-run/help assertions |
| `MET-02` | Full Git delivery reachability | GitHub/network/Git writes are not guaranteed | An explicitly authorized mode can edit, test, commit, push, and create/update a PR | Deterministic command tests plus opt-in live E2E evidence |
| `MET-03` | Supported Codex command compatibility | Compatibility can drift with CLI option placement | Generated commands are accepted by the supported Codex CLI contract | Go command-shape tests and real-Codex smoke validation |

### Scope

- `REQ-01` Preserve a restricted human-gate mode for working-tree-only
  automation and make that restricted completion boundary explicit in output
  and help.
- `REQ-02` Add an explicit opt-in full-delivery mode whose declared contract
  includes GitHub reads, network operations, Git metadata writes, push, and
  pull-request creation/update.
- `REQ-03` Resolve the human-gate permission mode from a documented CLI option,
  environment variable, and safe built-in default, and show the winning value
  in dry-run output.
- `REQ-04` Generate a Codex command compatible with the supported CLI syntax,
  including correct placement of global permission options and existing model,
  worktree, JSONL, and last-message arguments.
- `REQ-05` Keep FT-015 state artifacts, final-status parsing, explicit-thread
  resume behavior, and exit-code contract unchanged unless a capability error
  occurs before the batch run.
- `REQ-06` Explain authentication, network access, Git write access, the risk of
  full-delivery mode, restricted-mode limitations, and troubleshooting in
  dedicated human-gate help and project documentation.
- `REQ-07` Add deterministic automated coverage for permission precedence,
  validation, command construction, dry-run visibility, default compatibility,
  and existing `DONE`/`HUMAN_GATE` behavior.
- `REQ-08` Provide an explicit opt-in real-Codex E2E procedure for validating
  full delivery without adding live GitHub writes to `make test` or CI.

### Non-Scope

- `NS-01` Do not generalize human-gate mode to Claude, Kimi, Pi, or `agent=none`.
- `NS-02` Do not provision GitHub credentials, modify Codex user configuration,
  or store secrets in project files.
- `NS-03` Do not authorize production changes, destructive Git operations, or
  other product/security decisions merely because full-delivery mode is active;
  the prompt's `HUMAN_GATE` rules remain authoritative.
- `NS-04` Do not redesign thread capture, state storage, status parsing, or
  interactive resume semantics from FT-015.
- `NS-05` Do not couple this feature to the Go runtime migration or another CLI
  rewrite.

### Constraints / Assumptions

- `ASM-01` Issue #37 reproduces the rejected argument order with Codex CLI
  `0.144.6`. Local parser validation confirms both selected command forms on
  Codex CLI `0.145.0`; live full-delivery behavior still requires the explicit
  `CHK-03` approval gate. The repository does not pin an installed version.
- `ASM-02` Full delivery requires independently configured GitHub
  authentication and repository authorization; `start-issue` can select a
  launcher policy but cannot grant those external capabilities.
- `CON-01` Restricted mode remains the built-in default so upgrading the CLI
  does not silently broaden command execution privileges.
- `CON-02` Full-delivery mode must require an explicit user choice and must be
  visible before Codex starts.
- `CON-03` Secrets and tokens must remain outside CLI arguments, tracked config,
  logs, JSONL events, and state artifacts.
- `CON-04` Live delivery verification performs external GitHub writes and is
  therefore manual, explicit-opt-in, and outside normal local/CI checks.

## Design Requirement Decision

| Decision | Reason | Downstream owner |
| --- | --- | --- |
| `Design required: yes` | The feature changes CLI/environment contracts and a security/trust boundary, and requires explicit trade-offs, command contracts, failure modes, and rollout rules. | `design.md` |

## Verify

### Exit Criteria

- `EC-01` Restricted mode remains the default, is clearly reported, and does
  not claim commit/push/PR completion.
- `EC-02` Full-delivery mode is accepted only through explicit opt-in and
  produces the documented supported-Codex command contract.
- `EC-03` Invalid permission values and incompatible command construction fail
  before a batch session with actionable diagnostics.
- `EC-04` Existing FT-015 `DONE`, `HUMAN_GATE`, state-file, thread-resume, and
  exit-code behavior remains covered and unchanged.
- `EC-05` Help and project docs describe authentication, network/Git access,
  security risk, mode selection, and troubleshooting consistently.
- `EC-06` An explicitly approved live E2E can demonstrate end-to-end Git
  delivery and preserve auditable artifacts without becoming a CI dependency.

### Traceability matrix

| Requirement ID | Problem refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `CON-01` | `EC-01`, `SC-01` | `CHK-01`, `CHK-02` | `EVID-01`, `EVID-02` |
| `REQ-02` | `ASM-02`, `CON-02` | `EC-02`, `SC-02` | `CHK-01`, `CHK-03` | `EVID-01`, `EVID-03` |
| `REQ-03` | `CON-01`, `CON-02` | `EC-01`, `EC-02`, `SC-01`, `SC-02`, `NEG-01` | `CHK-01` | `EVID-01` |
| `REQ-04` | `ASM-01` | `EC-02`, `EC-03`, `SC-03` | `CHK-01`, `CHK-03` | `EVID-01`, `EVID-03` |
| `REQ-05` | `NS-04` | `EC-04`, `SC-04` | `CHK-01` | `EVID-01` |
| `REQ-06` | `ASM-02`, `CON-03` | `EC-05`, `SC-05` | `CHK-02` | `EVID-02` |
| `REQ-07` | `CON-01` | `EC-01` - `EC-04`, `SC-01` - `SC-04`, `NEG-01` | `CHK-01` | `EVID-01` |
| `REQ-08` | `CON-04` | `EC-06`, `SC-06` | `CHK-03` | `EVID-03` |

### Acceptance Scenarios

- `SC-01` Given no permission override, when a user inspects or runs
  human-gate mode, then restricted mode is selected and its working-tree-only
  completion boundary is visible.
- `SC-02` Given explicit full-delivery selection, when the human-gate command is
  built, then the mode is visibly reported and the command permits the declared
  GitHub/network/Git delivery workflow.
- `SC-03` Given the supported Codex CLI, when either permission mode builds the
  batch command, then global permission options occur in a supported position
  and batch-only output arguments remain attached to `exec`.
- `SC-04` Given either permission mode and a valid Codex response, when the
  final status is `DONE` or `HUMAN_GATE`, then FT-015 state and resume behavior
  remains unchanged.
- `SC-05` Given dedicated help or project documentation, when an operator plans
  a run, then prerequisites, limitations, risks, and recovery steps are clear
  without reading source code.
- `SC-06` Given explicit operator authorization, isolated fixture resources,
  and valid credentials, when the full-delivery E2E runs, then it records
  successful commit/push/PR delivery and retained diagnostic artifacts.

### Negative / Edge Scenarios

- `NEG-01` Given an unknown permission mode, when configuration is resolved,
  then `start-issue` exits before issue fetching, worktree mutation, or Codex
  launch and names the accepted values.
- `NEG-02` Given restricted mode and a task requiring Git delivery, when the
  operator reads help or run output, then the tool does not represent the
  restricted run as capable of completing commit/push/PR delivery.
- `NEG-03` Given full-delivery mode without valid GitHub credentials or remote
  authorization, when the live workflow reaches delivery, then it fails with a
  capability diagnostic and does not relabel the failure as a product decision.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01` - `EC-04`, `SC-01` - `SC-04`, `NEG-01` | `make test` | Go formatting/vet/tests, memory-bank audit, and deterministic human-gate coverage pass for both modes and FT-015 regressions. | Local terminal/CI test output |
| `CHK-02` | `EC-01`, `EC-05`, `SC-01`, `SC-05`, `NEG-02` | Review `--help`, `--human-gate-help`, README files, practical guides, and spec alongside Go output assertions | All surfaces state the same default, opt-in, capability, risk, and troubleshooting contract. | Review diff and Go test output |
| `CHK-03` | `EC-02`, `EC-03`, `EC-06`, `SC-02`, `SC-03`, `SC-06`, `NEG-03` | With explicit approval, run the real-Codex full-delivery E2E procedure from FT-017's plan | Supported Codex accepts the command and the isolated fixture records commit, push, PR, terminal status, and retained artifacts. | Retained E2E artifact directory and fixture PR URL |

### Test matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | Local terminal output and CI job |
| `CHK-02` | `EVID-02` | Review diff and deterministic help assertions |
| `CHK-03` | `EVID-03` | Retained opt-in E2E artifact directory and fixture PR URL |

### Evidence

- `EVID-01` Automated verification output covering permission resolution,
  command construction, errors, and FT-015 regression behavior.
- `EVID-02` Documentation/help review evidence showing one consistent operator
  contract.
- `EVID-03` Explicitly approved live-E2E evidence for full Git delivery.

### Evidence contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | Local and CI test output | implementer / CI | Terminal output and GitHub Actions job | `CHK-01` |
| `EVID-02` | Documentation diff plus help assertions | implementer / reviewer | Changed docs and Go test output | `CHK-02` |
| `EVID-03` | Live-E2E log, state files, commit/PR identifiers | approved operator | Retained E2E artifact path printed by runner | `CHK-03` |
