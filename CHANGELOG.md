# Changelog

## [1.0.0] - 2026-10-15

### Added

- Added stable Docker list and top filters.
- Added age and label filters for cleanup planning.
- Added versioned JSON envelopes and schemas.
- Added stable automation exit codes and `--quiet`.
- Added release smoke tests for macOS `amd64` and `arm64` artifacts.
- Added Homebrew Cask publishing through the project tap.

### Changed

- Froze the macOS CLI command and flag contract.
- Installer validates release version before atomic replacement.
- Cleanup previews expose the same filters used by apply operations.

### Migration

- Automation should consume `--json` and `CLI_COMPATIBILITY.md`.
- Linux and Windows remain unsupported.

## [0.5.0] - 2026-10-15

### Added

- Added a simplified Docker interface with `docker`, `docker --list`, and `docker --top` modes.
- Added mandatory colored human-readable terminal output.
- Added health indicators and usage bars to `status`.
- Added JSON output for automation.
- Added installer and uninstaller validation to CI.

### Changed

- Reduced the public command surface to `devsentry`, `docker`, `clean`, and `version`.
- Kept detailed status and Docker commands available as advanced commands.
- Improved cleanup, Docker summaries, top resource listings, and container tables.
- Release checks now run tests, vet, and GoReleaser validation before publishing.

### Fixed

- Kept cleanup previews aligned with the exact apply command.
- Handled macOS environments with no listening ports.
- Added bounded execution time for macOS status collection.
