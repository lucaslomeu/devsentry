#!/bin/sh

set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
TMP_DIR=$(mktemp -d 2>/dev/null || mktemp -d -t devsentry-install-test)
trap 'rm -rf "$TMP_DIR"' EXIT HUP INT TERM

RELEASE_DIR="$TMP_DIR/release"
PAYLOAD_DIR="$TMP_DIR/payload"
INSTALL_DIR="$TMP_DIR/bin"

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

for ARCH in arm64 amd64; do
	ARCHIVE="devsentry_0.2.0_darwin_$ARCH.tar.gz"
	tar -czf "$RELEASE_DIR/$ARCHIVE" -C "$PAYLOAD_DIR" devsentry

	if command -v shasum >/dev/null 2>&1; then
		CHECKSUM=$(shasum -a 256 "$RELEASE_DIR/$ARCHIVE" | awk '{ print $1 }')
	else
		CHECKSUM=$(sha256sum "$RELEASE_DIR/$ARCHIVE" | awk '{ print $1 }')
	fi
	printf '%s  %s\n' "$CHECKSUM" "$ARCHIVE" >>"$RELEASE_DIR/checksums.txt"

	ARCH_INSTALL_DIR="$INSTALL_DIR/$ARCH"
	DEVSENTRY_VERSION=v0.2.0 \
		DEVSENTRY_INSTALL_DIR="$ARCH_INSTALL_DIR" \
		DEVSENTRY_RELEASE_BASE_URL="file://$RELEASE_DIR" \
		DEVSENTRY_TEST_OS=Darwin \
		DEVSENTRY_TEST_ARCH="$ARCH" \
		sh "$ROOT_DIR/install.sh" >"$TMP_DIR/install-$ARCH.log"

	[ -x "$ARCH_INSTALL_DIR/devsentry" ] || {
		printf 'installed binary is missing or not executable for %s\n' "$ARCH" >&2
		exit 1
	}

	VERSION_OUTPUT=$("$ARCH_INSTALL_DIR/devsentry" version)
	[ "$VERSION_OUTPUT" = "devsentry version: 0.2.0" ] || {
		printf 'unexpected installed version for %s: %s\n' "$ARCH" "$VERSION_OUTPUT" >&2
		exit 1
	}
done

if DEVSENTRY_TEST_OS=Linux sh "$ROOT_DIR/install.sh" >"$TMP_DIR/linux.log" 2>&1; then
	printf 'installer should reject non-macOS systems\n' >&2
	exit 1
fi

ARCHIVE="devsentry_0.2.0_darwin_arm64.tar.gz"
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

MISMATCH_RELEASE_DIR="$TMP_DIR/mismatch-release"
MISMATCH_PAYLOAD_DIR="$TMP_DIR/mismatch-payload"
MISMATCH_INSTALL_DIR="$TMP_DIR/mismatch-bin"
mkdir -p "$MISMATCH_RELEASE_DIR" "$MISMATCH_PAYLOAD_DIR" "$MISMATCH_INSTALL_DIR"
cp "$PAYLOAD_DIR/devsentry" "$MISMATCH_PAYLOAD_DIR/devsentry"
printf '#!/bin/sh\nprintf old-version\n' >"$MISMATCH_INSTALL_DIR/devsentry"
chmod +x "$MISMATCH_INSTALL_DIR/devsentry"
MISMATCH_ARCHIVE="devsentry_0.3.0_darwin_arm64.tar.gz"
tar -czf "$MISMATCH_RELEASE_DIR/$MISMATCH_ARCHIVE" -C "$MISMATCH_PAYLOAD_DIR" devsentry
if command -v shasum >/dev/null 2>&1; then
	MISMATCH_CHECKSUM=$(shasum -a 256 "$MISMATCH_RELEASE_DIR/$MISMATCH_ARCHIVE" | awk '{ print $1 }')
else
	MISMATCH_CHECKSUM=$(sha256sum "$MISMATCH_RELEASE_DIR/$MISMATCH_ARCHIVE" | awk '{ print $1 }')
fi
printf '%s  %s\n' "$MISMATCH_CHECKSUM" "$MISMATCH_ARCHIVE" >"$MISMATCH_RELEASE_DIR/checksums.txt"

if DEVSENTRY_VERSION=v0.3.0 \
	DEVSENTRY_INSTALL_DIR="$MISMATCH_INSTALL_DIR" \
	DEVSENTRY_RELEASE_BASE_URL="file://$MISMATCH_RELEASE_DIR" \
	DEVSENTRY_TEST_OS=Darwin \
	DEVSENTRY_TEST_ARCH=arm64 \
	sh "$ROOT_DIR/install.sh" >"$TMP_DIR/mismatch.log" 2>&1; then
	printf 'installer should reject a version mismatch\n' >&2
	exit 1
fi

[ "$("$MISMATCH_INSTALL_DIR/devsentry" version)" = "old-version" ] || {
	printf 'version mismatch replaced existing binary\n' >&2
	exit 1
}

printf 'install tests passed\n'
