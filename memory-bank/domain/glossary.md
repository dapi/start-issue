---
title: Domain Glossary
doc_kind: domain
doc_function: canonical
purpose: "Ubiquitous language for start-issue workflows, configuration, worktrees, agents, releases, and memory-bank."
derived_from:
  - ../dna/governance.md
  - ../product/context.md
  - ../../README.md
  - ../../doc/spec.md
status: active
audience: humans_and_agents
canonical_for:
  - ubiquitous_language
  - domain_terms
---

# Domain Glossary

## Terms

| Term | Meaning | Context | Do not confuse with |
| --- | --- | --- | --- |
| `issue input` | GitHub issue number or full GitHub issue URL passed to the CLI | Normal issue-start workflow | GitHub issue body |
| `repository context` | Resolved `owner/repo`, git root, origin remote, and base branch | Issue fetch and worktree planning | Local worktree path |
| `base branch` | Branch used as the starting point for the issue worktree branch | Worktree creation | Current branch unless selected as fallback |
| `branch name` | Generated or AI-produced git branch for the issue | Worktree lifecycle | Worktree directory name, though they are related |
| `worktree` | Git worktree dedicated to one issue branch | Main product output | The original repository checkout |
| `worktree parent directory` | Directory under which issue worktrees are created | `--worktree-dir`, `START_ISSUE_WORKTREE_DIR` | Specific worktree path |
| `agent` | Selected coding CLI adapter: `claude`, `codex`, `kimi`, `pi`, or `none` | Config and launch | Model |
| `model` | Optional model string passed to adapters that support explicit model args | Config and launch | Agent |
| `agent adapter` | Internal boundary that validates agent support and builds agent-specific commands | `scripts/lib/start_issue/agent.sh` | Public agent CLI implementation |
| `prompt template` | Text with supported placeholders rendered into an agent prompt | Prompt resolution and launch | Final rendered prompt |
| `prompt source` | Where the active prompt came from: CLI, env, project config, user config, or built-in default | Config output and dry-run | Prompt location |
| `project config` | `.start-issue/*` files under the git root | Repository-local defaults | User config |
| `user config` | `~/.config/start-issue/*` files | User defaults and first-run setup | Project config |
| `setup` | User-level onboarding flow for `~/.config/start-issue` | `start-issue setup`, `--setup`, first-run gate | `init` |
| `init` | Manual initializer for project or user config | `start-issue init` | `setup` |
| `prompt proposal` | Reviewable improved prompt file written by `--improve-prompt` | Prompt improvement | Active prompt template |
| `release asset` | Built single-file `start-issue` binary uploaded to GitHub Releases | Install and update | Source modules |
| `running executable` | The exact `start-issue` path invoked by the user | Self-update | Repository source script |
| `human-gate run` | Codex batch execution with persisted state and final status parsing | `--human-gate` | Normal interactive Codex launch |
| `memory-bank` | Governed project documentation and process layer adapted from `dapi/memory-bank` | Agent/project context | Runtime state under `.start-issue` |

## Naming Rules

- Use `agent` for the selected CLI adapter and `model` for the optional model
  argument.
- Use `setup` only for friendly user-level onboarding; use `init` for explicit
  config initialization.
- Use `worktree` for git worktrees only; use `workspace` only in prose when the
  branch/worktree/agent setup is meant together.
- Use `human-gate` only for the Codex batch/resume workflow currently grounded
  in `codex exec` and `codex resume`.

## Ambiguous Terms

| Term | Allowed meaning | Forbidden / overloaded meaning | Replacement |
| --- | --- | --- | --- |
| `default agent` | Effective fallback after CLI, project config, user config, and env are resolved | Hardcoded claim that always ignores user/project settings | `built-in default` or `resolved agent` |
| `prompt` | Rendered final text only when explicitly stated | Any of prompt template, prompt source, and proposal file | `prompt template`, `prompt source`, `prompt proposal` |
| `update` | Self-update from latest GitHub Release | Rebuild from local checkout | `build`, `install`, or `make install` |

## Source Documents

- [README.md](../../README.md)
- [doc/spec.md](../../doc/spec.md)
