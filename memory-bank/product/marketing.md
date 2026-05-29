---
title: Marketing And Positioning
doc_kind: product
doc_function: canonical
purpose: "Positioning and messaging for start-issue as an open-source developer CLI."
derived_from:
  - ../dna/governance.md
  - context.md
  - customers.md
status: active
audience: humans_and_agents
canonical_for:
  - product_positioning
  - product_messaging
  - go_to_market_context
---

# Marketing And Positioning

## Positioning

| Audience | Current alternative | Product difference | Proof |
| --- | --- | --- | --- |
| `SEG-01` | Manual `gh` lookup, branch naming, worktree creation, prompt copy/paste | One command creates a consistent workspace and launches the selected agent | README workflow and tests |
| `SEG-03` | Manual asset download or reinstall from source | Release-backed installer and self-update with checksum verification | Release docs and update tests |

## Messaging

- `MSG-01` "Turn a GitHub issue into a branch, worktree, and coding-agent
  session."
- `MSG-02` "Keep agent choice explicit: Claude, Codex, Kimi, Pi, or manual
  next steps."
- `MSG-03` "Use dry-run and visible config sources to know what will happen
  before worktree or agent side effects."

## Channels

| Channel | Audience | Goal | Constraint | Owner |
| --- | --- | --- | --- | --- |
| GitHub README | Developers | Install, usage, trust | Must match shipped CLI behavior | Maintainer |
| GitHub Releases | Users with installed binary | Distribution and update source | Tag/version/checksum must align | Maintainer / CI |

## Competitive Alternatives

- `ALT-01` Manual shell workflow using `gh`, `git checkout`, and `git worktree`.
- `ALT-02` Agent-specific launch scripts that do not share config precedence.
- `ALT-03` Full project management tools, which are intentionally out of scope.

## Launch Constraints

- `LC-01` README, spec, help text, tests, and release assets must agree before a
  user-facing release.
- `LC-02` Do not claim support for an agent capability unless the adapter and
  tests prove it.
