#!/bin/sh

set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
TMP_DIR=$(mktemp -d 2>/dev/null || mktemp -d -t devsentry-uninstall-test)
trap 'rm -rf "$TMP_DIR"' EXIT HUP INT TERM

INSTALL_DIR="$TMP_DIR/bin"
mkdir -p "$INSTALL_DIR"
printf '#!/bin/sh\n' >"$INSTALL_DIR/devsentry"

DEVSENTRY_INSTALL_DIR="$INSTALL_DIR" sh "$ROOT_DIR/uninstall.sh" >"$TMP_DIR/uninstall.log"
[ ! -e "$INSTALL_DIR/devsentry" ] || {
	printf 'binary was not removed\n' >&2
	exit 1
}

DEVSENTRY_INSTALL_DIR="$INSTALL_DIR" sh "$ROOT_DIR/uninstall.sh" >"$TMP_DIR/missing.log"
printf 'uninstall tests passed\n'
