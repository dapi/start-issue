---
title: Missing feature packages inventory
doc_kind: feature
doc_function: index
purpose: "Inventory of GitHub issue work that is not represented by a dedicated feature package, including duplicate and superseded work."
derived_from:
  - ../dna/governance.md
  - ../product/context.md
  - ../../README.md
  - https://github.com/dapi/start-issue/issues
status: active
audience: humans_and_agents
---

# Missing Feature Packages

This inventory compares the GitHub issue history with the instantiated
packages under this directory. It is a navigation and gap document, not a
replacement for a feature `brief.md`.

## Current gaps

| Issue | Current state | Memory-bank state | Gap / recommended action |
| --- | --- | --- | --- |
| [#35](https://github.com/dapi/start-issue/issues/35) | Open; reconciled to the existing delivery slice | FT-013 is the single owner and now references #35 as provenance; current Go implementation and tests cover the contract. | Keep one package; close or update the GitHub issue separately after acceptance verification. |

## Closed issues without a dedicated package

These issues are implemented or closed in GitHub, but their requirements are
not independently navigable from `memory-bank/features/`:

| Issues | Missing coverage |
| --- | --- |
| [#1](https://github.com/dapi/start-issue/issues/1) | Agent-agnostic configurable portable prompt. |
| [#2](https://github.com/dapi/start-issue/issues/2) | No-issue configuration/status summary showing selected agent and prompt. |
| [#3](https://github.com/dapi/start-issue/issues/3) | `init` project/user configuration initialization with defaults. |
| [#30](https://github.com/dapi/start-issue/issues/30) | Meaningful transliterated branch slugs and removal of bracketed title tags. |
| [#32](https://github.com/dapi/start-issue/issues/32) | Explicit terminal status before waiting for input or handing off to an agent. |

Create packages for these only if their behavior is expected to remain a
separately maintained product contract. Otherwise, add their issue links to
the relevant canonical package or to a historical decision record.

## Not feature work: migrations

Issue [#34](https://github.com/dapi/start-issue/issues/34) is a Go rewrite and
distribution/runtime migration, not a product feature. It should be tracked
as an architecture/migration package (and, where needed, an ADR) under
`memory-bank/adr/` or a dedicated migration area. `FT-016` must remain reserved
for the real Codex E2E suite; the issue branch's attempted reuse of that ID
should not be copied into the feature catalog.

## Superseded migration issues

Issues [#15](https://github.com/dapi/start-issue/issues/15) and
[#17](https://github.com/dapi/start-issue/issues/17)–[#23](https://github.com/dapi/start-issue/issues/23)
describe a Rust migration sequence, but the current repository has no Rust
runtime or Rust feature packages. The newer [#34](https://github.com/dapi/start-issue/issues/34)
reframes the migration in Go. These issues should be represented by one
historical/supersession note or by the migration/ADR for #34, not by seven
duplicate feature packages.

## Already represented

The following issue work has a corresponding package:

- #4 → [FT-004](FT-004/README.md)
- #8 → [FT-008](FT-008/README.md)
- #9 → [FT-009](FT-009/README.md)
- #12 → [FT-012](FT-012/README.md)
- #13 → [FT-013](FT-013/README.md)
- #25 → [FT-014](FT-014/README.md)
- #26 → [FT-015](FT-015/README.md)
- Real Codex E2E validation → [FT-016](FT-016/README.md)
- #37 → [FT-017](FT-017/README.md)

The package `delivery_status` fields still describe several packages as
`in_progress`; issue closure alone is not treated as evidence that a feature
package has passed its documented verification contract.
