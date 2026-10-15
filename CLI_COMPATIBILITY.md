# CLI Compatibility

## Scope

DevSentry v1.0.0 supports macOS `amd64` and `arm64`. Docker commands require a
working Docker CLI and daemon.

The stable command surface is:

- `devsentry`
- `devsentry version`
- `devsentry --version`
- `devsentry docker`
- `devsentry docker --list`
- `devsentry docker --top`
- `devsentry clean`
- `devsentry status`

Hidden aliases `ls`, `summary`, and `top` remain supported for compatibility.

## Human Output

Human output is for interactive use. ANSI colors may be present. Scripts must
use `--json`, not parse human-readable text.

Command meaning, cleanup safety, and documented headings remain stable within
the v1 major series. Additional informational lines may be added.

## JSON Output

Commands supporting `--json` return an envelope with:

```json
{
  "schema_version": 1,
  "command": "docker.list",
  "data": {}
}
```

Schemas live under `schemas/v1/`. Existing fields and types are not removed or
renamed within v1. New fields are additive. JSON output never contains ANSI
escape sequences.

## Exit Codes

| Code | Meaning |
| --- | --- |
| `0` | Success |
| `1` | Unexpected error |
| `2` | Usage or validation error |
| `3` | Docker unavailable |
| `4` | Operation timeout |
| `5` | Docker mutation failure |

Errors are written to `stderr`.

## Quiet Mode

`--quiet` suppresses successful human-readable output. JSON output remains
enabled when explicitly requested. Errors remain on `stderr`.

## Compatibility Changes

Removing commands, flags, JSON fields, or changing exit-code meaning requires a
new major version. Breaking changes require migration notes and a release-note
entry.
