#!/bin/sh

set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
cd "$ROOT_DIR"

command -v go >/dev/null 2>&1 || {
	printf 'verify: go is required\n' >&2
	exit 1
}

printf '%s\n' '== Shell syntax =='
sh -n install.sh uninstall.sh tests/install_test.sh tests/uninstall_test.sh

printf '%s\n' '== Go formatting =='
FORMATTED=$(git ls-files '*.go' | xargs gofmt -l)
if [ -n "$FORMATTED" ]; then
	printf 'verify: unformatted Go files:\n%s\n' "$FORMATTED" >&2
	exit 1
fi

printf '%s\n' '== Installer tests =='
sh tests/install_test.sh
sh tests/uninstall_test.sh

printf '%s\n' '== Vet =='
go vet ./...

printf '%s\n' '== Tests =='
go test -race ./...

printf '%s\n' '== Build =='
go build ./...

printf '%s\n' 'Verification passed.'
