#!/bin/sh
set -eu

VERSION=${1:-}
MODE=${2:-}

if ! printf '%s\n' "$VERSION" | awk '$0 ~ /^v[0-9]+\.[0-9]+\.[0-9]+$/ { valid = 1 } END { exit valid ? 0 : 1 }'; then
	printf 'usage: %s vX.Y.Z [--dry-run]\n' "$0" >&2
	exit 2
fi

if [ -n "$MODE" ] && [ "$MODE" != "--dry-run" ]; then
	printf 'usage: %s vX.Y.Z [--dry-run]\n' "$0" >&2
	exit 2
fi

[ "$(git branch --show-current)" = "main" ] || {
	printf 'release must run from main\n' >&2
	exit 1
}

git diff --quiet

VERSION_NUMBER=${VERSION#v}
grep -F "## [$VERSION_NUMBER]" CHANGELOG.md >/dev/null || {
	printf 'missing changelog entry for %s\n' "$VERSION" >&2
	exit 1
}

if git rev-parse "$VERSION" >/dev/null 2>&1; then
	printf 'tag already exists: %s\n' "$VERSION" >&2
	exit 1
fi

./scripts/verify.sh

[ "$MODE" = "--dry-run" ] && exit 0

git tag -a "$VERSION" -m "Release $VERSION"
git push origin "$VERSION"
