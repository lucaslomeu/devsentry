#!/bin/sh

set -eu

usage() {
	printf 'Usage: %s <feat|fix|refactor|chore> <short-name>\n' "$0" >&2
	exit 2
}

[ "$#" -eq 2 ] || usage
TYPE=$1
NAME=$2

case "$TYPE" in
	feat | fix | refactor | chore) ;;
	*) usage ;;
esac

case "$NAME" in
	*[!a-z0-9-]* | "")
		printf 'start-cycle: short-name must use lowercase letters, numbers, and hyphens\n' >&2
		exit 2
		;;
esac

[ -z "$(git status --porcelain)" ] || {
	printf 'start-cycle: working tree must be clean\n' >&2
	exit 1
}

git switch main
git pull --ff-only origin main

BRANCH="$TYPE/$NAME"
if git show-ref --verify --quiet "refs/heads/$BRANCH"; then
	printf 'start-cycle: branch already exists: %s\n' "$BRANCH" >&2
	exit 1
fi

git switch -c "$BRANCH"
printf '\nCreated %s\n' "$BRANCH"
