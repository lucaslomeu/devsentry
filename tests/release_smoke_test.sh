#!/bin/sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
DIST_DIR=${DEVSENTRY_DIST_DIR:-$ROOT_DIR/dist}
TMP_DIR=$(mktemp -d 2>/dev/null || mktemp -d -t devsentry-release-smoke)
trap 'rm -rf "$TMP_DIR"' EXIT HUP INT TERM

fail() {
	printf 'release smoke: %s\n' "$1" >&2
	exit 1
}

[ -f "$DIST_DIR/checksums.txt" ] || fail "checksums.txt is missing"

for ARCH in arm64 amd64; do
	if [ -n "${DEVSENTRY_EXPECTED_VERSION:-}" ]; then
		ARCHIVE="devsentry_${DEVSENTRY_EXPECTED_VERSION}_darwin_${ARCH}.tar.gz"
		ARCHIVE_PATH="$DIST_DIR/$ARCHIVE"
	else
		ARCHIVE_PATH=""
		for CANDIDATE in "$DIST_DIR"/devsentry_*_darwin_"$ARCH".tar.gz; do
			if [ -f "$CANDIDATE" ]; then
				ARCHIVE_PATH="$CANDIDATE"
				break
			fi
		done
		ARCHIVE=${ARCHIVE_PATH##*/}
	fi
	[ -f "$ARCHIVE_PATH" ] || fail "missing archive: $ARCHIVE"

	EXPECTED=$(awk -v file="$ARCHIVE" '$2 == file { print $1; exit }' "$DIST_DIR/checksums.txt")
	[ -n "$EXPECTED" ] || fail "missing checksum: $ARCHIVE"

	if command -v shasum >/dev/null 2>&1; then
		ACTUAL=$(shasum -a 256 "$ARCHIVE_PATH" | awk '{ print $1 }')
	else
		ACTUAL=$(sha256sum "$ARCHIVE_PATH" | awk '{ print $1 }')
	fi
	[ "$EXPECTED" = "$ACTUAL" ] || fail "checksum mismatch: $ARCHIVE"

	EXTRACT_DIR="$TMP_DIR/$ARCH"
	mkdir -p "$EXTRACT_DIR"
	tar -xzf "$ARCHIVE_PATH" -C "$EXTRACT_DIR"
	[ -x "$EXTRACT_DIR/devsentry" ] || fail "archive binary is missing or not executable: $ARCHIVE"

	if [ -n "${DEVSENTRY_EXPECTED_VERSION:-}" ] && [ "$(uname -s)" = "Darwin" ]; then
		if [ "$ARCH" = "$(uname -m | sed 's/x86_64/amd64/; s/arm64/arm64/')" ]; then
			EXPECTED_OUTPUT="devsentry version: $DEVSENTRY_EXPECTED_VERSION"
			ACTUAL_OUTPUT=$("$EXTRACT_DIR/devsentry" version)
			[ "$ACTUAL_OUTPUT" = "$EXPECTED_OUTPUT" ] || fail "unexpected version in $ARCH archive"
		fi
	fi
done

printf 'release smoke tests passed\n'
