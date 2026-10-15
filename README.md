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

Install with Homebrew:

```bash
brew install --cask lucaslomeu/tap/devsentry
devsentry version
```

Upgrade or remove the Homebrew installation:

```bash
brew upgrade --cask devsentry
brew uninstall --cask devsentry
```

Install in a custom directory without `sudo`:

```bash
curl -fsSL https://raw.githubusercontent.com/lucaslomeu/devsentry/main/install.sh | \
  DEVSENTRY_INSTALL_DIR="$HOME/.local/bin" sh
```

Install a specific version:

```bash
curl -fsSL https://raw.githubusercontent.com/lucaslomeu/devsentry/main/install.sh | \
  DEVSENTRY_VERSION=v1.0.0 sh
```

Verify the installation:

```bash
devsentry version
devsentry --help
```

Uninstall from the default directory:

```bash
curl -fsSL https://raw.githubusercontent.com/lucaslomeu/devsentry/main/uninstall.sh | sh
```

For a custom directory, set `DEVSENTRY_INSTALL_DIR` in the same way as the installer.

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
- macOS for `status`.

## Commands

| Command | Description |
| --- | --- |
| `devsentry` | Show a welcome message. |
| `devsentry --version` | Show the CLI version. |
| `devsentry version` | Show the CLI version. |
| `devsentry docker` | Show a Docker overview. |
| `devsentry clean` | Preview cleanup of dangling images, stopped containers, and build cache. |

Docker details are available without memorizing subcommands:

```bash
devsentry docker --list
devsentry docker --top --limit 5
devsentry docker --top --all
devsentry docker --list --name api --status running
devsentry docker --top --image 'postgres:*' --volume data
```

`devsentry status` remains available as an advanced macOS diagnostic command.

Informational commands support `--json`: `docker`, `clean`, and `status`.

`clean` never removes resources unless `--apply` is combined with an explicit scope. Multiple scopes can be selected, for example:

```bash
devsentry clean --images --cache --apply
devsentry clean --all --older-than 168h --exclude-label keep=true
```

The default image cleanup only removes dangling images. `--all` additionally
removes tagged images that are not used by a running container. Because stopped
containers are removed first in this mode, images referenced only by those
containers are also eligible. Images used by running containers are always
preserved.

Terminal colors are always enabled for human-readable output. JSON output remains plain for automation.

Automation users can rely on versioned JSON envelopes, stable exit codes, and
`--quiet`. See [CLI compatibility](CLI_COMPATIBILITY.md) and
[v1 migration notes](MIGRATING_TO_V1.md).

## Release artifacts

v1.0.0 publishes macOS `amd64` and `arm64` archives with SHA-256 checksums.
Linux and Windows are not supported by this release.
