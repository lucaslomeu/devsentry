# Migrating to v1.0.0

## Installation

The curl installer remains available for pinned release archives. Package-manager
distribution will be configured after v1.0.0.

## Automation

Use `--json` instead of parsing colored human output. JSON now includes
`schema_version` and `command` fields. Validate consumers against the schemas in
`schemas/v1/`.

Use exit codes `2` through `5` to distinguish usage errors, Docker availability,
timeouts, and mutation failures.

Use `--quiet` when successful output is not needed. Errors continue on stderr.

## Cleanup

Cleanup remains preview-only by default. Resource deletion requires `--apply`
and an explicit cleanup scope. New age and label filters are reflected in the
preview and apply command.

## Platform

v1.0.0 supports macOS `amd64` and `arm64`. Linux and Windows are outside this
release.
