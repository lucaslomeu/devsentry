#!/bin/sh

set -eu

usage() {
	printf 'Usage: %s vX.Y.Z [--dry-run]\n' "$0" >&2
	exit 2
}

[ "$#" -ge 1 ] && [ "$#" -le 2 ] || usage
VERSION=$1
DRY_RUN=false
if [ "${2:-}" = "--dry-run" ]; then
	DRY_RUN=true
elif [ "$#" -eq 2 ]; then
	usage
fi

case "$VERSION" in
	v[0-9]*.[0-9]*.[0-9]*) ;;
	*)
		printf 'release: version must match vX.Y.Z\n' >&2
		exit 2
		;;
esac

ROOT_DIR=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
cd "$ROOT_DIR"

[ "$(git branch --show-current)" = "main" ] || {
	printf 'release: must run from main\n' >&2
	exit 1
}

[ -z "$(git status --porcelain)" ] || {
	printf 'release: working tree must be clean\n' >&2
	exit 1
}

git fetch origin main --tags
[ "$(git rev-parse HEAD)" = "$(git rev-parse origin/main)" ] || {
	printf 'release: local main is not synchronized with origin/main\n' >&2
	exit 1
}

if git rev-parse "$VERSION" >/dev/null 2>&1; then
	printf 'release: tag already exists: %s\n' "$VERSION" >&2
	exit 1
fi

CHANGELOG_VERSION=${VERSION#v}
grep -F "## [$CHANGELOG_VERSION]" CHANGELOG.md >/dev/null 2>&1 || {
	printf 'release: missing CHANGELOG.md heading for %s\n' "$CHANGELOG_VERSION" >&2
	exit 1
}

./scripts/verify.sh

if $DRY_RUN; then
	printf '\nDry run passed. Would create and push %s.\n' "$VERSION"
	exit 0
fi

git tag -a "$VERSION" -m "Release $VERSION"
git push origin "$VERSION"
