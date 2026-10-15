#!/bin/sh
set -eu

NAME=${1:-}

if [ -z "$NAME" ]; then
	printf 'usage: %s <branch-name>\n' "$0" >&2
	exit 2
fi

git diff --quiet
git diff --cached --quiet
git switch main
git pull --ff-only origin main
git switch -c "$NAME"
