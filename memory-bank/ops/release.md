---
title: Release And Deployment
doc_kind: ops
doc_function: canonical
purpose: "Release, versioning, changelog, build, GitHub Release, install, and self-update process for start-issue."
derived_from:
  - ../dna/governance.md
  - ../product/metrics.md
  - ../../README.md
  - ../../doc/spec.md
status: active
audience: humans_and_agents
---

# Release And Deployment

`start-issue` is distributed as a single-file Bash executable through GitHub
Releases. There is no server deployment.

## Release Flow

1. Add user-facing changes under `## [Unreleased]` in `CHANGELOG.md`.
2. Run one release prep command:

```bash
make release-patch
make release-minor
make release-major
```

3. The release script bumps `VERSION`, updates changelog entries, runs
   `make test`, runs `make build`, creates `Release vX.Y.Z`, and creates an
   annotated tag.
4. Publish with:

```bash
git push origin master --follow-tags
```

5. GitHub Actions builds the release asset and checksum.

## Release Assets

Each release uploads:

- `start-issue`
- `start-issue.sha256`

The installer and self-update workflow download the release asset and verify the
checksum before install.

## Release Commands

```bash
make print-version
make bump-patch
make bump-minor
make bump-major
make release-patch
make release-minor
make release-major
```

Do not run release commands with a dirty worktree unless the release script
explicitly supports the current state.

## Release Verification

Required before publishing:

```bash
make test
make build
```

Release-specific checks:

- `VERSION` in `scripts/start-issue` matches the tag without the `v` prefix.
- `CHANGELOG.md` has the new version/date section.
- Release asset and checksum are present in GitHub Releases after CI.
- `start-issue update` can resolve the latest release and no-op/install
  correctly.

## Rollback

Rollback unit is the GitHub Release/tag plus the locally installed executable.
If a release is bad:

1. stop promoting the release;
2. publish a fixed patch release rather than mutating installed user machines;
3. if absolutely necessary, remove or mark the GitHub Release with a clear note;
4. document the problem in `CHANGELOG.md` or a follow-up issue.
