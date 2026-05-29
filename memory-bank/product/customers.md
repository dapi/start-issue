---
title: Customers And Users
doc_kind: product
doc_function: canonical
purpose: "Canonical user segments, jobs to be done, pains, and assumptions for start-issue."
derived_from:
  - ../dna/governance.md
  - context.md
status: active
audience: humans_and_agents
canonical_for:
  - product_customers
  - user_segments
  - jobs_to_be_done
---

# Customers And Users

## Segments

| Segment ID | Segment | Job To Be Done | Current Pain | Success Signal | Evidence |
| --- | --- | --- | --- | --- | --- |
| `SEG-01` | Maintainer/developer of this repository | Start work on a GitHub issue in an isolated worktree with the right agent | Repeated manual setup and inconsistent launch commands | One command prepares the expected workspace | README workflow and feature history |
| `SEG-02` | Coding agent session | Receive enough context and repository instructions to work safely | Missing repo context, wrong prompt, unclear checks | Prompt contains issue/worktree/base branch and repo instructions are discoverable | Prompt template and AGENTS.md |
| `SEG-03` | Release maintainer | Publish and update a single-file CLI safely | Manual version/changelog/tag/release steps can drift | `make release-*`, checksums, and update tests pass | Release docs and scripts |

## Users And Actors

| Actor ID | Actor | Uses product how | Decision power | Notes |
| --- | --- | --- | --- | --- |
| `ACT-01` | CLI user | Runs `start-issue ISSUE`, `setup`, `init`, `update`, or `--dry-run` | Chooses flags, agents, prompts, reuse/delete options | Primary human actor |
| `ACT-02` | Selected agent CLI | Receives rendered prompt and worktree context | Executes code changes within its own behavior model | Agent-specific details stay in adapters |
| `ACT-03` | GitHub CLI/session | Supplies issue and release metadata | External dependency | Requires authenticated `gh` for issue/update flows |
| `ACT-04` | Release workflow | Builds and uploads release assets from tags | Automation actor | Must verify version/tag alignment |

## Research Inputs

- Existing feature packages under [features/](../features/README.md).
- Current public docs in [README.md](../../README.md) and [doc/spec.md](../../doc/spec.md).
- Test suite behavior in `test/start_issue.bats`.

## Assumptions

- `ASM-01` Most users work inside git repositories with GitHub remotes.
- `ASM-02` Users value predictable local workflow more than broad project
  management features.
- `ASM-03` Agent CLIs remain separately installed tools; `start-issue` should
  not vendor or manage them.

## Must Not Assume

- `NA-01` Do not assume every user wants an agent launched; `--agent none` is a
  first-class mode.
- `NA-02` Do not assume all agents support Codex-style resumable batch sessions.
- `NA-03` Do not assume user config should be silently written beyond explicit
  setup/init and the documented first-run marker directory.
