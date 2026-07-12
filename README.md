# devsentry

Devsentry is a CLI to inspect Docker resources and basic environment status.

## Installation

Release binaries (macOS arm64/amd64) are published on GitHub Releases.

Local build:

```bash
go build -o devsentry .
```

Install locally (optional):

```bash
sudo mv devsentry /usr/local/bin/
```

Local run (development):

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
- `devsentry clean` — preview cleanup of dangling images, stopped containers, and build cache.
- `devsentry clean --images|--containers|--cache` — preview selected cleanup scopes.
- `devsentry clean --all` — preview removal of stopped containers, build cache, and all images unused afterward.
- `devsentry clean --all --apply` — execute the aggressive cleanup previewed by `--all`.
- `devsentry status` — show local environment status.

`clean` never removes resources unless `--apply` is combined with an explicit scope. Multiple scopes can be selected, for example:

```bash
devsentry clean --images --cache --apply
```

The default image cleanup only removes dangling images. `--all` additionally
removes tagged images that are not used by a running container. Because stopped
containers are removed first in this mode, images referenced only by those
containers are also eligible. Images used by running containers are always
preserved.
