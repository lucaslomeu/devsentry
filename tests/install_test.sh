#!/bin/sh

set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
TMP_DIR=$(mktemp -d 2>/dev/null || mktemp -d -t devsentry-install-test)
trap 'rm -rf "$TMP_DIR"' EXIT HUP INT TERM

RELEASE_DIR="$TMP_DIR/release"
PAYLOAD_DIR="$TMP_DIR/payload"
INSTALL_DIR="$TMP_DIR/bin"
ARCHIVE="devsentry_0.2.0_darwin_arm64.tar.gz"

mkdir -p "$RELEASE_DIR" "$PAYLOAD_DIR"

cat >"$PAYLOAD_DIR/devsentry" <<'EOF'
#!/bin/sh
if [ "${1:-}" = "version" ]; then
	printf 'devsentry version: 0.2.0\n'
	exit 0
fi
printf 'fake devsentry\n'
EOF
chmod +x "$PAYLOAD_DIR/devsentry"

tar -czf "$RELEASE_DIR/$ARCHIVE" -C "$PAYLOAD_DIR" devsentry

if command -v shasum >/dev/null 2>&1; then
	CHECKSUM=$(shasum -a 256 "$RELEASE_DIR/$ARCHIVE" | awk '{ print $1 }')
else
	CHECKSUM=$(sha256sum "$RELEASE_DIR/$ARCHIVE" | awk '{ print $1 }')
fi
printf '%s  %s\n' "$CHECKSUM" "$ARCHIVE" >"$RELEASE_DIR/checksums.txt"

DEVSENTRY_VERSION=v0.2.0 \
	DEVSENTRY_INSTALL_DIR="$INSTALL_DIR" \
	DEVSENTRY_RELEASE_BASE_URL="file://$RELEASE_DIR" \
	DEVSENTRY_TEST_OS=Darwin \
	DEVSENTRY_TEST_ARCH=arm64 \
	sh "$ROOT_DIR/install.sh" >"$TMP_DIR/install.log"

[ -x "$INSTALL_DIR/devsentry" ] || {
	printf 'installed binary is missing or not executable\n' >&2
	exit 1
}

VERSION_OUTPUT=$("$INSTALL_DIR/devsentry" version)
[ "$VERSION_OUTPUT" = "devsentry version: 0.2.0" ] || {
	printf 'unexpected installed version: %s\n' "$VERSION_OUTPUT" >&2
	exit 1
}

if DEVSENTRY_TEST_OS=Linux sh "$ROOT_DIR/install.sh" >"$TMP_DIR/linux.log" 2>&1; then
	printf 'installer should reject non-macOS systems\n' >&2
	exit 1
fi

printf '%s  %s\n' "invalid-checksum" "$ARCHIVE" >"$RELEASE_DIR/checksums.txt"
if DEVSENTRY_VERSION=v0.2.0 \
	DEVSENTRY_INSTALL_DIR="$TMP_DIR/invalid-bin" \
	DEVSENTRY_RELEASE_BASE_URL="file://$RELEASE_DIR" \
	DEVSENTRY_TEST_OS=Darwin \
	DEVSENTRY_TEST_ARCH=arm64 \
	sh "$ROOT_DIR/install.sh" >"$TMP_DIR/checksum.log" 2>&1; then
	printf 'installer should reject an invalid checksum\n' >&2
	exit 1
fi

printf 'install tests passed\n'
