# Devsentry Roadmap

This roadmap keeps the project macOS-focused and prioritizes a small, stable
command surface. Each cycle is executed through a branch, small Conventional
Commits, a pull request, and a tag-driven release.

## Operating Model

1. Start from an up-to-date `main` with `./scripts/start-cycle.sh`.
2. Implement one behavior at a time.
3. Use Conventional Commits:
   - `feat:` user-visible functionality.
   - `fix:` behavior correction.
   - `refactor:` internal change without behavior change.
   - `test:` test-only change.
   - `ci:` automation change.
   - `chore:` maintenance.
4. Run `./scripts/verify.sh` before opening the PR.
5. Open a PR against `main` and wait for CI.
6. Merge only after review and successful checks.
7. Update `CHANGELOG.md` on `main` through a PR before tagging.
8. Run `./scripts/release.sh vX.Y.Z --dry-run`.
9. Run `./scripts/release.sh vX.Y.Z` to push the tag.
10. Verify the GitHub Release artifacts and test the installer.

## Cycle 0.5.1: Stabilization

### Goal

Validate the `v0.5.0` user experience in real macOS environments without
adding new public commands.

### Features and fixes

- Add hermetic command-dispatch tests for `docker`, `docker --list`, and
  `docker --top`.
- Add installer smoke coverage for both supported architectures through local
  release fixtures.
- Verify ANSI output, JSON output, and error output separately.
- Improve errors for missing Docker, stopped Docker, invalid flags, and failed
  cleanup operations.
- Confirm cleanup previews and apply commands remain identical.

### Suggested commits

```text
test: cover docker mode dispatch
test: add release installer smoke coverage
fix: improve docker command errors
fix: preserve cleanup preview safety
```

### Acceptance criteria

- `./scripts/verify.sh` passes on macOS and CI.
- Every public Docker mode has a parser or dispatch test.
- JSON output contains no ANSI escape sequences.
- No cleanup operation mutates Docker state without `--apply`.

### Release

- Version: `v0.5.1`.
- Update `CHANGELOG.md` with fixes only.
- Merge PR, run the dry-run release check, push the tag, and verify both
  macOS artifacts.

## Cycle 0.6.0: Docker Workflows

### Goal

Make common Docker maintenance tasks faster without increasing the number of
top-level commands.

### Features and implementation stages

1. Add container filters to `docker --list` for name, image, and status.
2. Add image and volume filters to `docker --top`.
3. Add cleanup age filters with explicit preview output.
4. Add label-based cleanup exclusions.
5. Add a clear confirmation summary for destructive `clean --apply` flows.
6. Keep all advanced behavior behind flags on existing commands.

### Suggested commits

```text
feat: add docker list filters
test: cover docker list filtering
feat: add docker top filters
feat: add age-based cleanup planning
fix: protect excluded docker resources
test: cover filtered cleanup plans
```

### Acceptance criteria

- Filters are represented in dry-run output.
- A filtered preview and its apply command target the same resources.
- Running-container images remain protected.
- Invalid filter expressions fail before Docker mutation.
- Existing invocations keep their current behavior.

### Release

- Version: `v0.6.0`.
- Include new flags and examples in `CHANGELOG.md`.
- Release only after integration tests cover every destructive path.

## Cycle 0.7.0: Automation Contract

### Goal

Make Devsentry dependable in scripts and CI while preserving readable output
for humans.

### Features and implementation stages

1. Define and version the JSON schema for Docker, cleanup, and status output.
2. Add stable exit codes for usage, Docker unavailable, timeout, and mutation
   failures.
3. Add a quiet mode for scripts while preserving errors on stderr.
4. Add fixture tests for supported Docker CLI output variants.
5. Add a compatibility matrix for Docker CLI versions used in CI.

### Suggested commits

```text
feat: define stable json output schema
feat: add stable cli exit codes
feat: add quiet automation mode
test: cover docker output compatibility fixtures
ci: validate automation contract
```

### Acceptance criteria

- JSON schemas are tested against golden fixtures.
- Human output and JSON output use the same underlying data.
- Automation errors never contain terminal color codes.
- Exit codes are documented and tested.
- Backward-compatible fields are not removed within the minor series.

### Release

- Version: `v0.7.0`.
- Publish migration notes for any JSON or exit-code additions.

## Cycle 1.0.0: Stability Contract

### Goal

Declare the macOS CLI interface stable.

### Required work

- Freeze the public command and flag surface.
- Define semantic compatibility rules for human and JSON output.
- Review installer, uninstaller, checksums, and rollback procedures.
- Require release notes and migration guidance for breaking changes.
- Add end-to-end smoke tests for the complete release path.
- Confirm no known destructive cleanup ambiguity remains.

### Suggested commits

```text
test: add end-to-end release smoke tests
ci: enforce stable release checks
docs: define cli compatibility policy
chore: prepare v1.0.0 release
```

### Acceptance criteria

- The complete branch, PR, merge, tag, release, and installer flow is repeatable.
- The public CLI contract is explicitly versioned.
- All supported release artifacts install and report the expected version.
- No planned breaking command changes remain for the release.

## Release Procedure

### Feature branch

```bash
./scripts/start-cycle.sh feat docker-list-filters
```

### Commit and verify

```bash
./scripts/verify.sh
git push -u origin feat/docker-list-filters
```

### Pull request

```bash
  --title "feat: add Docker list filters"
```

Wait for CI, review the PR, and merge it through GitHub. Do not tag a release
from a feature branch.

### Release

After the merge, update the version section in `CHANGELOG.md` through a PR and
merge it into `main`. Then run:

```bash
./scripts/release.sh v0.5.1 --dry-run
./scripts/release.sh v0.5.1
gh run list --workflow Release --limit 1
```

The tag starts the existing GitHub Actions release workflow, which tests,
validates GoReleaser, and publishes macOS `amd64` and `arm64` artifacts.
