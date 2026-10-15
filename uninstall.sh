#!/bin/sh

set -eu

INSTALL_DIR="${DEVSENTRY_INSTALL_DIR:-/usr/local/bin}"
TARGET="$INSTALL_DIR/devsentry"

if [ ! -e "$TARGET" ] && [ ! -L "$TARGET" ]; then
	printf 'devsentry is not installed in %s\n' "$INSTALL_DIR"
	exit 0
fi

if [ -w "$INSTALL_DIR" ]; then
	rm -f "$TARGET"
elif command -v sudo >/dev/null 2>&1; then
	printf 'Administrator permission is required to uninstall from %s.\n' "$INSTALL_DIR"
	sudo rm -f "$TARGET"
else
	printf 'cannot write to %s; set DEVSENTRY_INSTALL_DIR to a writable directory\n' "$INSTALL_DIR" >&2
	exit 1
fi

printf 'devsentry uninstalled from %s\n' "$TARGET"
