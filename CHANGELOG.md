# Changelog

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
