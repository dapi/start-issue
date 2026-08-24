---
title: Stages And Non-Local Environments
doc_kind: ops
doc_function: canonical
purpose: "Non-local environment policy for start-issue."
derived_from:
  - ../dna/governance.md
  - release.md
status: active
audience: humans_and_agents
---

# Stages And Non-Local Environments

`start-issue` has no hosted staging or production environment. Non-local state is
limited to GitHub and external CLIs.

## Environment Inventory

| Environment | Purpose | Access path | Notes |
| --- | --- | --- | --- |
| GitHub repository | Issues, CI, releases, tags | `gh`, git remote, GitHub UI | Auth required for issue/update/release operations |
| GitHub Releases | Distribution source for installer and update | `gh api`, release URLs | Publishes platform binaries and `checksums.txt` |
| User machine | Installed executable and user config | local filesystem | Update installs into running executable path |

## Common Operations

```bash
gh issue view <number> --repo dapi/start-issue
gh api repos/dapi/start-issue/releases/latest
make test
make build
```

Publishing releases and pushing tags require explicit maintainer intent.

## Credentials And Access

- GitHub access is owned by the user's `gh` authentication.
- Agent credentials are owned by the selected agent CLI.
- No project secrets should be stored in this repository.

## Version And Health Checks

```bash
start-issue --version
make print-version
gh api repos/dapi/start-issue/releases/latest --jq .tag_name
```

## Logs And Observability

There is no centralized runtime observability. Diagnostics come from:

- command output;
- Go test logs;
- GitHub Actions logs;
- batch state files under
  `<worktree>/.start-issue/runs/<timestamp>/`.

## Test Data And Smoke Targets

Use temporary git repositories and fake helper binaries in tests. Do not depend
on live GitHub issues for deterministic regression coverage.
