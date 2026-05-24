# Rust Parity Checklist

This document defines the current Bash behavior that should be treated as the migration contract for the first Rust implementation.

The intent is parity-first migration, not a product redesign. Any intentional behavior changes should be documented explicitly and reviewed as deltas from this checklist.

## Behavior surfaces to preserve

- CLI parsing and mode selection
  - issue mode vs `init` vs `update`
  - required option values
  - invalid option and invalid issue input failures
- Config precedence
  - agent: CLI, project file, user file, environment, built-in default
  - model: CLI, project file, user file, environment, built-in default
  - prompt: CLI, project file, user file, environment, built-in default
  - conflict handling when both inline and file-backed prompt inputs are set
- Prompt rendering
  - placeholder substitution
  - unknown placeholders preserved as-is
  - large prompt handling in dry-run output
  - prompt improvement proposal path rules
- Repository and issue resolution
  - parse full GitHub issue URLs and numeric issue IDs
  - derive repo from SSH and HTTPS remotes
  - derive base branch from `origin/HEAD`, then current branch fallback
- Branch and worktree planning
  - fast branch naming heuristic
  - optional AI branch naming with fallback on invalid output
  - flat and nested worktree path modes
  - branch reuse detection
  - worktree path reuse validation
  - interactive conflict resolution for reuse, versioned rename, and recreate
- Agent integration
  - supported agent set and validation errors
  - launch command construction
  - explicit model propagation only where supported
  - `--no-agent` manual next-step output
- Non-core integrations
  - optional `zellij-tab-status` rename behavior
  - update/install flows

## Verification sources

- Executable spec: `test/start_issue.bats`
- User-facing contract: `README.md` and `README.ru.md`
- Current orchestration entrypoint: `scripts/lib/start_issue/pipeline.sh`

## Migration rule

Until cutover, new Rust behavior should either:

1. match the Bash behavior covered by tests, or
2. document an intentional difference and update the parity checklist plus tests accordingly.
