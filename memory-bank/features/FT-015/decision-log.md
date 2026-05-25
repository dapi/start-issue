---
title: "FT-015: Decision Log"
doc_kind: feature
doc_function: decision_log
purpose: "Resolved local decisions and conflict handling for FT-015. Stores only decisions grounded in the current feature package and in-scope artifacts."
derived_from:
  - https://github.com/dapi/start-issue/issues/26
  - feature.md
  - solution.md
  - ../../../scripts/lib/start_issue/agent.sh
  - ../../../scripts/lib/start_issue/cli.sh
  - ../../../scripts/lib/start_issue/output.sh
  - ../../../scripts/lib/start_issue/pipeline.sh
  - ../../../test/helpers/fake-bin/codex
status: active
audience: humans_and_agents
---

# FT-015: Decision Log

## Decision Entries

| Decision ID | Date | Status | Topic | Decision |
| --- | --- | --- | --- | --- |
| `DL-01` | 2026-05-24 | accepted | Agent scope | `--human-gate` is Codex-only in FT-015 and must fail clearly for any resolved non-Codex agent. |
| `DL-02` | 2026-05-24 | accepted | Dedicated help surface | Dedicated human-gate documentation is exposed through `--human-gate-help`, while normal `--help` mentions the mode briefly. |
| `DL-03` | 2026-05-24 | accepted | Run-state artifacts | Human-gate runs store `events.jsonl`, `last-message.txt`, and `thread-id` under `<worktree>/.start-issue/runs/<timestamp>/` and keep those artifacts for diagnostics and resume fidelity. |
| `DL-04` | 2026-05-24 | accepted | Final-status parser | The saved last-message artifact is authoritative; only `STATUS: DONE` and `STATUS: HUMAN_GATE` are recognized terminal statuses. |
| `DL-05` | 2026-05-24 | accepted | Resume failure contract | If `HUMAN_GATE` was reached but interactive resume could not be opened, the command exits `2` after printing the exact resume command and captured thread id. |

## FPF-Closed Questions

### `FPF-01`: Should `--human-gate` be rejected or ignored for non-Codex agents?

#### Why this mattered

Issue #26 explicitly proposes a Codex-oriented workflow and says the option should be rejected or clearly ignored for non-Codex agents unless a generic implementation is intentionally supported later. The document set needed one explicit contract so feature, solution, plan, and future tests would not drift.

#### Available facts

- Issue #26 defines the workflow in terms of `codex exec`, `thread.started`, `codex resume --include-non-interactive`, and Codex-specific final-status handling.
- Current `scripts/lib/start_issue/agent.sh` already centralizes agent-specific launch and non-interactive behavior and uses `codex exec` only for Codex paths.
- Current repository facts do not show equivalent resumable session or JSON event contracts for Claude, Kimi, or Pi.

#### FPF reasoning summary

- Bounded contexts: generic issue workflow and Codex session-resume behavior are different semantic frames and should not be collapsed.
- Propose -> analyze -> test: the smallest plausible extension is Codex-only; testing it against repo facts shows only Codex has grounded resumable-session evidence in the issue and current code patterns.
- Surface precision restoration: the flag name could imply broader support than the evidence allows, so the contract needs an explicit boundary.

#### Resolution

Reject `--human-gate` for any resolved non-Codex agent in FT-015.

#### Conflict handled

This resolves the issue's "rejected or clearly ignored" branch in favor of rejection because the current document set has evidence for one safe bounded context only: Codex. Ignoring the flag would look successful while silently skipping the requested behavior.

### `FPF-02`: What dedicated-help surface fits the current CLI without inventing a second subcommand family?

#### Why this mattered

Issue #26 requires dedicated help but gives multiple example entrypoints. The feature package needed one concrete user-facing surface.

#### Available facts

- Current `scripts/lib/start_issue/cli.sh` supports two positional command families today: `init` and `update`.
- Current `scripts/lib/start_issue/output.sh` owns a single normal help surface plus missing-issue help.
- Current positional parsing treats unknown bare words as either `ISSUE_INPUT` or an error, so a help-only positional token would expand that grammar.

#### FPF reasoning summary

- Bounded contexts: "workflow invocation" and "documentation surface" are related but not identical; dedicated help does not need to become a new workflow family.
- Propose -> analyze -> test: a flag-based help surface is the smallest new claim; when tested against current parser facts, it fits the established flag-help model with less parsing drift than a new positional help token.
- Surface precision restoration: use a surface that says "this is help for the mode" directly, instead of overloading issue-position parsing.

#### Resolution

Use `--human-gate-help` as the dedicated help entrypoint and mention `--human-gate` briefly in normal `--help`.

#### Conflict handled

This resolves the ambiguity between `start-issue human-gate --help` and `start-issue --human-gate-help` by choosing the surface that is best aligned with the current CLI grammar and current help ownership.

### `FPF-03`: Where should resumable run artifacts live, and should they be kept after the batch run?

#### Why this mattered

Issue #26 requires a predictable state directory and also requires later interactive resume by exact thread id. The docs needed one artifact contract across solution, plan, and future tests.

#### Available facts

- Issue #26 proposes `.start-issue/runs/<timestamp>/events.jsonl`, `last-message.txt`, and `thread-id` under the created worktree.
- Current issue workflow already centers work product inside the prepared worktree, including prompt rendering and optional `init.sh`.
- Current code has no separate global session registry owned by `start-issue`.

#### FPF reasoning summary

- Bounded contexts: issue work product and its human-gate execution evidence belong to the same local worktree context.
- Propose -> analyze -> test: co-locating artifacts under the worktree is the simplest hypothesis; it matches the issue proposal and does not require inventing a second storage scope.
- Trust and evidence discipline: keeping the artifacts after the run preserves the evidence needed to diagnose unknown statuses or failed resume handoffs.

#### Resolution

Store `events.jsonl`, `last-message.txt`, and `thread-id` under `<worktree>/.start-issue/runs/<timestamp>/` and keep them by default.

#### Conflict handled

No cross-document conflict remained after aligning the state contract with the issue's proposed location and the repository's existing worktree-centered workflow.

## Conflict Resolution

- Resolved conflict: dedicated help is required, but the current CLI grammar has only `init` and `update` as positional subcommand families.
  Conflicting sources:
  issue #26 allows either a positional help surface or a flag help surface;
  current `cli.sh` and `output.sh` are structured around flag-driven help and issue positional parsing.
  Resolution:
  choose `--human-gate-help` and keep human-gate help in the existing help-output ownership model.
  Why this is consistent:
  it adds the required documentation surface without redefining the CLI grammar more broadly than the feature needs.

- Resolved conflict: the flag could look generic, but only Codex-specific resumable evidence is currently grounded.
  Conflicting sources:
  the flag name `--human-gate` sounds generic;
  issue #26 and current code facts only ground Codex-specific batch and resume mechanics.
  Resolution:
  make FT-015 explicitly Codex-only and reject other agents.
  Why this is consistent:
  it matches the documented evidence boundary and avoids a misleading pseudo-generic contract.
