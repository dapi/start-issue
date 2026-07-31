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

`start-issue` is distributed as platform-specific Go binaries through GitHub
Releases. There is no server deployment.

## Release Flow

1. Add user-facing changes under `## [Unreleased]` in `CHANGELOG.md`.
2. From a clean worktree, run the required checks and create the SemVer tag.
   The tag is the source of the release version:

```bash
make test
make build
git tag vX.Y.Z
```

3. Publish with:

```bash
git push origin master --follow-tags
```

4. GitHub Actions reruns the test suite and uses GoReleaser to publish the
   binaries and checksum manifest.

## Release Assets

Each release uploads one binary for every supported target:

- `start-issue-linux-amd64`
- `start-issue-linux-arm64`
- `start-issue-darwin-amd64`
- `start-issue-darwin-arm64`
- `start-issue-windows-amd64.exe`
- `checksums.txt`
- `start-issue` and `start-issue.sha256` during the v1-to-v2 transition only

The installer and self-update workflow download the release asset and verify the
checksum before install. The two legacy-named assets are a verified POSIX
migration bridge: v1.13.2 and older updaters install it, then it resolves,
verifies, and replaces itself with the matching v2 platform binary.

## Version Source

```bash
make print-version
```

`make build` and `make install` derive the embedded version from `git describe`
using the nearest SemVer tag. A tagged release therefore reports the tag
version, while a checkout between tags reports its describe suffix. GoReleaser
embeds the pushed release tag in published binaries.

## Release Verification

Required before publishing:

```bash
make test
make build
```

Release-specific checks:

- `make print-version` matches the intended tag version when run at that tag.
- `CHANGELOG.md` has the new version/date section.
- All five platform assets, `checksums.txt`, and the v1 migration bridge assets
  `start-issue`/`start-issue.sha256` are present in GitHub Releases after CI.
- `start-issue update` can resolve the latest release and no-op/install
  correctly.

## Rollback

Rollback unit is the GitHub Release/tag plus the locally installed executable.
If a release is bad:

1. stop promoting the release;
2. publish a fixed patch release rather than mutating installed user machines;
3. if absolutely necessary, remove or mark the GitHub Release with a clear note;
4. document the problem in `CHANGELOG.md` or a follow-up issue.
