---
title: Git Workflow
doc_kind: engineering
doc_function: convention
purpose: "Git, branch, worktree, commit, and PR expectations for start-issue."
derived_from:
  - ../dna/governance.md
  - ../../README.md
status: active
audience: humans_and_agents
---

# Git Workflow

## Default Branch

The default branch is `master`.

## Branches And Worktrees

`start-issue` normally creates issue branches in git worktrees under
`~/worktrees` unless overridden by `--worktree-dir` or
`START_ISSUE_WORKTREE_DIR`.

Generated branch format:

```text
{type}/issue-{number}-{kebab-case-title}
```

Do not manually reuse a worktree path for another branch without cleaning up git
worktree metadata.

## Commits

- Use concise present-tense messages.
- Release commits are created by release scripts as `Release vX.Y.Z`.
- Do not include unrelated generated churn in a feature commit.

## Pull Requests

Before PR handoff, report:

- changed behavior/docs;
- verification command and result;
- any manual-only gap or residual risk.

For memory-bank changes, include the index audit result.

## Release Tags

Release tags use SemVer with a `v` prefix, for example `v1.12.0`. Pushing the
tag with `git push origin master --follow-tags` publishes the release through
GitHub Actions.
