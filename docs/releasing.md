# Releasing

This module is distributed as a normal Go module from GitHub. There is no separate package registry publish step.

## First release

1. Ensure `main` is clean and CI is green.
2. Run the local release checks:

   ```sh
   make pre-push
   ```

3. Create and push the first semantic version tag:

   ```sh
   git tag v0.1.0
   git push origin v0.1.0
   ```

4. Publish the draft GitHub release for `v0.1.0`.
5. Confirm the module resolves:

   ```sh
   go list -m github.com/uffejaeger/intercom-go@v0.1.0
   ```

## Release notes

Release Drafter updates a draft GitHub release whenever PRs are merged to `main`. It groups changes by labels and defaults unlabeled PRs to patch-level changes.

Use these labels to control release notes and version bumps:

- `breaking`, `breaking-change`, or `major` for breaking changes.
- `feature`, `enhancement`, or `minor` for new public SDK behavior.
- `bug`, `fix`, or `patch` for fixes.
- `documentation`, `docs`, `chore`, `dependencies`, or `maintenance` for patch-level non-feature changes.
- `skip-changelog` to omit a PR from release notes.

## Versioning

Before `v1.0.0`, minor versions may include public API adjustments as the SDK stabilizes. Patch versions should remain backwards-compatible fixes.
