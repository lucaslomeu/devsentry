# devsentry

Devsentry is a CLI to inspect Docker resources and basic environment status.

## Installation

Local build:

```bash
go build ./...
```

Local run:

```bash
go run .
```

## Requirements

- Docker CLI for Docker-related commands.
- macOS for `status` (Linux/Windows support planned).

## Commands

- `devsentry` — show a welcome message.
- `devsentry version` — show CLI version.
- `devsentry docker ls` — list running containers.
- `devsentry docker summary` — show Docker totals.
- `devsentry docker top` — top images and volumes by size (`--all` includes anonymous volumes).
- `devsentry clean` — dry-run cleanup of dangling images, stopped containers, and build cache (`--apply` to execute).
- `devsentry status` — show local environment status.

Aliases:

- `devsentry ls`
- `devsentry summary`
