# devsentry

Devsentry is a CLI to inspect Docker resources and basic environment status.

## Installation

Install the latest release on macOS:

```bash
curl -fsSL https://raw.githubusercontent.com/lucaslomeu/devsentry/main/install.sh | sh
```

The installer detects Apple Silicon or Intel, verifies the release checksum, and
installs `devsentry` in `/usr/local/bin`. Administrator permission may be
requested when that directory is not writable.

Install in a custom directory without `sudo`:

```bash
curl -fsSL https://raw.githubusercontent.com/lucaslomeu/devsentry/main/install.sh | \
  DEVSENTRY_INSTALL_DIR="$HOME/.local/bin" sh
```

Install a specific version:

```bash
curl -fsSL https://raw.githubusercontent.com/lucaslomeu/devsentry/main/install.sh | \
  DEVSENTRY_VERSION=v0.3.0 sh
```

Verify the installation:

```bash
devsentry version
devsentry --help
```

Release binaries can also be downloaded manually from
[GitHub Releases](https://github.com/lucaslomeu/devsentry/releases).

## Development

Build locally:

```bash
go build -o devsentry .
```

Run without installing:

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
