#!/bin/sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
cd "$ROOT_DIR"

die() {
	printf 'verify: %s\n' "$1" >&2
	exit 1
}

command -v go >/dev/null 2>&1 || die "go is required"
command -v goreleaser >/dev/null 2>&1 || die "goreleaser is required"

GO_FILES=$(git ls-files --cached --others --exclude-standard '*.go')
if [ -n "$GO_FILES" ]; then
	OUT=$(printf '%s\n' "$GO_FILES" | xargs gofmt -l)
	[ -z "$OUT" ] || die "Go files are not formatted:\n$OUT"
fi

go vet ./...
go test ./...

sh -n install.sh uninstall.sh tests/*.sh scripts/*.sh
sh tests/install_test.sh
sh tests/uninstall_test.sh

goreleaser check
goreleaser release --snapshot --clean
DEVSENTRY_DIST_DIR="$ROOT_DIR/dist" tests/release_smoke_test.sh

printf 'verification passed\n'
