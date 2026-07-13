#!/bin/sh

set -eu

REPOSITORY="${DEVSENTRY_REPOSITORY:-lucaslomeu/devsentry}"
INSTALL_DIR="${DEVSENTRY_INSTALL_DIR:-/usr/local/bin}"
REQUESTED_VERSION="${DEVSENTRY_VERSION:-}"
OS_NAME="${DEVSENTRY_TEST_OS:-$(uname -s)}"
MACHINE_ARCH="${DEVSENTRY_TEST_ARCH:-$(uname -m)}"

fail() {
	printf 'devsentry installer: %s\n' "$1" >&2
	exit 1
}

if [ "$OS_NAME" != "Darwin" ]; then
	fail "only macOS is supported"
fi

case "$MACHINE_ARCH" in
	arm64 | aarch64)
		ARCH="arm64"
		;;
	x86_64 | amd64)
		ARCH="amd64"
		;;
	*)
		fail "unsupported architecture: $MACHINE_ARCH"
		;;
esac

command -v curl >/dev/null 2>&1 || fail "curl is required"
command -v tar >/dev/null 2>&1 || fail "tar is required"
command -v install >/dev/null 2>&1 || fail "install is required"

if [ -n "$REQUESTED_VERSION" ]; then
	case "$REQUESTED_VERSION" in
		v*)
			TAG="$REQUESTED_VERSION"
			;;
		*)
			TAG="v$REQUESTED_VERSION"
			;;
	esac
else
	LATEST_URL=$(curl -fsSL -o /dev/null -w '%{url_effective}' "https://github.com/$REPOSITORY/releases/latest") ||
		fail "could not find the latest release"
	TAG=${LATEST_URL##*/}
fi

case "$TAG" in
	v?*) ;;
	*) fail "invalid release tag: $TAG" ;;
esac

VERSION=${TAG#v}
ARCHIVE="devsentry_${VERSION}_darwin_${ARCH}.tar.gz"
if [ -n "${DEVSENTRY_RELEASE_BASE_URL:-}" ]; then
	RELEASE_BASE_URL=${DEVSENTRY_RELEASE_BASE_URL%/}
else
	RELEASE_BASE_URL="https://github.com/$REPOSITORY/releases/download/$TAG"
fi

TMP_DIR=$(mktemp -d 2>/dev/null || mktemp -d -t devsentry)
trap 'rm -rf "$TMP_DIR"' EXIT HUP INT TERM

printf 'Downloading devsentry %s for macOS %s...\n' "$TAG" "$ARCH"
curl -fsSL "$RELEASE_BASE_URL/$ARCHIVE" -o "$TMP_DIR/$ARCHIVE" ||
	fail "could not download $ARCHIVE"
curl -fsSL "$RELEASE_BASE_URL/checksums.txt" -o "$TMP_DIR/checksums.txt" ||
	fail "could not download checksums.txt"

EXPECTED_CHECKSUM=$(awk -v file="$ARCHIVE" '$2 == file { print $1; exit }' "$TMP_DIR/checksums.txt")
[ -n "$EXPECTED_CHECKSUM" ] || fail "checksum entry not found for $ARCHIVE"

if command -v shasum >/dev/null 2>&1; then
	ACTUAL_CHECKSUM=$(shasum -a 256 "$TMP_DIR/$ARCHIVE" | awk '{ print $1 }')
elif command -v sha256sum >/dev/null 2>&1; then
	ACTUAL_CHECKSUM=$(sha256sum "$TMP_DIR/$ARCHIVE" | awk '{ print $1 }')
else
	fail "shasum or sha256sum is required"
fi

[ "$EXPECTED_CHECKSUM" = "$ACTUAL_CHECKSUM" ] || fail "checksum verification failed"

tar -xzf "$TMP_DIR/$ARCHIVE" -C "$TMP_DIR" || fail "could not extract $ARCHIVE"
[ -f "$TMP_DIR/devsentry" ] || fail "release archive does not contain the devsentry binary"

if mkdir -p "$INSTALL_DIR" 2>/dev/null && [ -w "$INSTALL_DIR" ]; then
	install -m 0755 "$TMP_DIR/devsentry" "$INSTALL_DIR/devsentry"
elif command -v sudo >/dev/null 2>&1; then
	printf 'Administrator permission is required to install in %s.\n' "$INSTALL_DIR"
	sudo mkdir -p "$INSTALL_DIR"
	sudo install -m 0755 "$TMP_DIR/devsentry" "$INSTALL_DIR/devsentry"
else
	fail "cannot write to $INSTALL_DIR; set DEVSENTRY_INSTALL_DIR to a writable directory"
fi

"$INSTALL_DIR/devsentry" version >/dev/null 2>&1 || fail "the installed binary could not be executed"

printf 'devsentry %s installed in %s/devsentry\n' "$TAG" "$INSTALL_DIR"
case ":${PATH:-}:" in
	*":$INSTALL_DIR:"*)
		printf 'Run: devsentry --help\n'
		;;
	*)
		printf 'Run: %s/devsentry --help\n' "$INSTALL_DIR"
		printf 'Add %s to PATH to run devsentry from anywhere.\n' "$INSTALL_DIR"
		;;
esac
