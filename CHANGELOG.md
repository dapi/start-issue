# Changelog

All notable changes to this project are documented here.

This project follows Semantic Versioning.

## [Unreleased]

### Added

- Added `setup` and `--setup` onboarding for user-level configuration in `~/.config/start-issue`.
- Added first-run onboarding that offers to initialize user configuration before the first ordinary issue workflow.
- Added Codex-only `--human-gate` batch mode with resumable `codex exec` / `codex resume` flow.
- Added dedicated `--human-gate-help` documentation for the Codex human-gate contract.
- Added this changelog.

### Changed

- Updated release preparation to require `CHANGELOG.md` entries and move `Unreleased` notes under the new release version and date.

### Fixed

- Fixed branch slug generation for non-Latin titles: leading bracketed workflow tags (e.g. `[brief]`) are now stripped, and Cyrillic titles are transliterated to a meaningful Latin slug instead of collapsing to the tag or `work`. The AI branch-name prompt now also transliterates non-English titles and ignores bracketed tags.

## [1.13.1] - 2026-05-24

### Changed

- Normalized configuration source output so selected config details are reported consistently.

## [1.13.0] - 2026-05-24

### Added

- Added explicit model selection via CLI, environment, and project/user configuration.
- Added a `curl | bash` installer for installing the latest published GitHub Release.
- Added installer debug mode for diagnosing slow or stuck downloads.
- Added self-update support through `start-issue update` and `start-issue --update`.

### Changed

- Hardened branch slug generation.
- Polished the local release workflow.
- Added CI coverage for the installer.

## [1.12.0] - 2026-05-24

### Added

- Added GitHub Release automation that builds and uploads the bundled `start-issue` asset and checksum.

### Changed

- Let environment prompt settings override project and user prompt files.
- Improved missing-issue help output.
- Included the base branch in the portable prompt.
- Preserved script permissions when bumping versions.

## [1.11.1] - 2026-05-18

### Changed

- Showed default prompt file locations in configuration output.

## [1.11.0] - 2026-05-09

### Added

- Added bundled single-file builds from the modular shell sources for installation and release assets.

### Changed

- Split the `start-issue` implementation into focused shell modules.
- Stabilized worktree lifecycle behavior and reuse checks.
- Improved missing-issue errors for prompt improvement workflows.

### Fixed

- Fixed module loading for installed `start-issue` scripts.
- Fixed symlinked module lookup.
- Avoided repo-local builds during install and stopped install when bundling fails.
- Fixed CI shellcheck coverage for the modular layout.

## [1.10.0] - 2026-05-09

### Added

- Extracted the initial `start-issue` workflow into this repository.
- Added support for configurable coding agents instead of a Claude-only workflow.
- Added project and user configuration initialization.
- Added prompt improvement workflow support.
- Added CI test coverage.

### Changed

- Moved the workflow and configuration documentation into the README.
- Moved the script specification into `doc/spec.md`.
- Improved missing-issue output to show selected configuration.
