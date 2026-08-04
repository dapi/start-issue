---
title: "FT-017: Decision Log"
doc_kind: feature-support
doc_function: reference
purpose: "Records FPF analysis and accepted local decisions for FT-017. It does not own feature scope, selected design, acceptance criteria, or execution sequence."
derived_from:
  - brief.md
  - design.md
  - implementation-plan.md
  - https://github.com/openai/codex/blob/main/codex-rs/exec/src/cli.rs
status: active
audience: humans_and_agents
must_not_define:
  - ft_017_scope
  - ft_017_selected_design
  - ft_017_acceptance_criteria
  - implementation_sequence
---

# FT-017: Decision Log

## Purpose and Ownership

This log records FPF decisions for FT-017. `brief.md` owns problem space and
acceptance; `design.md` owns the selected solution; `implementation-plan.md`
owns execution sequencing. This file records rationale and provenance only.

## DL-01 — Permission boundary and default

**Status:** accepted by FPF review on 2026-08-04.

### Facts

- Issue #37 states that `workspace-write` does not provide the network and Git
  metadata writes needed for normal GitHub delivery.
- Existing FT-015 behavior uses `workspace-write` and must remain compatible.
- Full delivery can create external GitHub state and must therefore be opt-in.

### Decision

Keep two semantic modes: `restricted` and `full-delivery`. `restricted` is the
built-in default. Full delivery requires an explicit CLI/environment selection
and a visible warning before Codex starts.

### FPF rationale

The launcher capability boundary is a separate bounded context from task-level
approval decisions. Least privilege is the selection criterion: absent an
explicit user choice, preserve the existing restricted behavior. The semantic
names keep the public contract independent from Codex's low-level flags.

## DL-02 — Full-delivery command mapping

**Status:** accepted by FPF review on 2026-08-04, pending live verification.

### Evidence

- The official Codex `exec` CLI source marks
  `dangerously_bypass_approvals_and_sandbox` as a global option for `exec`.
- The same source marks `model`, `json`, and `output-last-message` as global
  options compatible with the `exec` command.
- The local repository has no installed `codex` executable, so live parser
  validation cannot be performed in this worktree.

### Decision

Map `full-delivery` to:

```text
codex [--model MODEL] --dangerously-bypass-approvals-and-sandbox exec \
  --cd WORKTREE --json --output-last-message PATH -
```

Keep restricted mode on the existing `codex exec --cd WORKTREE
--sandbox workspace-write --json --output-last-message PATH -` path. Do not
use `--ask-for-approval` in the generated command because issue #37 identifies
that spelling/placement as the compatibility failure under investigation.

### Rationale and limits

The bypass mapping is the only documented current CLI mechanism found that
explicitly covers both approvals and sandboxing. It is intentionally treated
as a high-risk capability switch, not as authorization for destructive or
production actions. `CHK-03` and `AG-01` remain mandatory before acceptance.

## DL-03 — Current implementation grounding

**Status:** accepted by FPF review on 2026-08-04.

The feature package must target the current Go implementation under
`cmd/start-issue/`, its Go tests, `test/e2e/human-gate.sh`, `Makefile`, and the
README/spec documentation. The earlier references to `scripts/lib/start_issue`
and Bats were stale artifacts from the pre-Go implementation and have been
removed from the execution plan.

## Open evidence gate

The exact supported release/version matrix is not asserted locally. Before
`delivery_status: done`, the approved live Codex executable must accept both
command forms and a retained E2E artifact must prove the declared full-delivery
behavior. If that verification fails, reject `full-delivery` and keep the
restricted path as the safe fallback.
