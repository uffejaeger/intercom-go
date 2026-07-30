# Maintainer Security Baseline

This checklist records the repository settings that protect releases in
addition to checks committed in the repository. Maintainers should review it
quarterly, before v1.0.0, and after changing GitHub plans or repository
ownership.

Verified for `uffejaeger/intercom-go` on 2026-07-29:

- `main` requires a pull request and an up-to-date successful `test` check.
- The required `test` check is bound to the GitHub Actions app.
- Branch protection applies to administrators.
- Force pushes and branch deletion are disabled for `main`.
- Pull request conversations must be resolved before merge, and workflows
  cannot approve pull requests.
- Third-party Actions are pinned to full commit SHAs, and GitHub rejects
  workflow changes that use unpinned Actions.
- Release tags matching `v*` cannot be updated or deleted.
- Immutable releases are enabled for releases created after the setting was
  turned on.
- Dependency graph, Dependabot alerts, and Dependabot security updates are
  enabled.
- Secret scanning and push protection are enabled.
- Code scanning default setup covers Go and GitHub Actions on a weekly schedule.
- GitHub private vulnerability reporting is enabled.
- Open Dependabot, code-scanning, and secret-scanning alerts have been reviewed.

When reviewing the baseline:

1. Confirm the required `test` check still maps to
   [`.github/workflows/test.yml`](../.github/workflows/test.yml) and is bound to
   the GitHub Actions app.
2. Confirm all `uses:` references remain pinned to full commit SHAs and
   checkout steps do not persist Git credentials.
3. Confirm immutable releases and the active `v*` tag-protection ruleset remain
   enabled.
4. Confirm code scanning, secret scanning, push protection, the dependency
   graph, and Dependabot security updates remain enabled.
5. Review all open security alerts and document or remediate accepted risk.
6. Confirm GitHub presents the private vulnerability-reporting form to a
   non-administrator account. Do not submit a test report.
7. Confirm releases are created from protected `main` commits and use signed
   GitHub-generated source archives or other release artifacts with recorded
   provenance when binary artifacts are introduced.
