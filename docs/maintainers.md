# Maintainer Security Baseline

This checklist records the repository settings that protect releases in
addition to checks committed in the repository. Maintainers should review it
quarterly, before v1.0.0, and after changing GitHub plans or repository
ownership.

Verified for `uffejaeger/intercom-go` on 2026-07-29:

- `main` requires a pull request and an up-to-date successful `test` check.
- Branch protection applies to administrators.
- Force pushes and branch deletion are disabled for `main`.
- Dependency graph, Dependabot alerts, and Dependabot security updates are
  enabled.
- Secret scanning and push protection are enabled.
- Code scanning default setup covers Go and GitHub Actions on a weekly schedule.
- GitHub private vulnerability reporting is enabled.
- Open Dependabot, code-scanning, and secret-scanning alerts have been reviewed.

When reviewing the baseline:

1. Confirm the required `test` check still maps to
   [`.github/workflows/test.yml`](../.github/workflows/test.yml).
2. Confirm code scanning, secret scanning, push protection, the dependency
   graph, and Dependabot security updates remain enabled.
3. Review all open security alerts and document or remediate accepted risk.
4. Confirm GitHub presents the private vulnerability-reporting form to a
   non-administrator account. Do not submit a test report.
5. Confirm releases are created from protected `main` commits and use signed
   GitHub-generated source archives or other release artifacts with recorded
   provenance when binary artifacts are introduced.
