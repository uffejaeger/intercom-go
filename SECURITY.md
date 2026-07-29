# Security Policy

## Reporting a vulnerability

Do not disclose a suspected vulnerability in a public issue, discussion, pull
request, or social post.

Use [GitHub private vulnerability reporting](https://github.com/uffejaeger/intercom-go/security/advisories/new).
Include the affected SDK version, impact, reproduction details, and any known
mitigation. Remove Intercom access tokens, customer data, and other secrets from
the report.

The maintainer aims to acknowledge a report within three business days and
provide an initial assessment or status update within seven business days.
These are best-effort targets, not a service-level agreement. Please allow time
to investigate and prepare a coordinated fix before publishing details.

If private reporting is unavailable, do not fall back to a public issue. Open a
non-sensitive issue that asks the maintainer to verify the private reporting
configuration, without including vulnerability details.

## Supported versions

Security fixes are released for the latest published minor release series. A
fix may also be backported when an older series is still widely used and the
backport is low risk, but backports are not guaranteed before v1.0.0.

Users should run a supported Go toolchain, keep this module and its transitive
dependencies current, and review release notes before upgrading. See
[`docs/compatibility.md`](docs/compatibility.md) for the Go and SDK compatibility
policy.

## Disclosure and remediation

Validated reports are handled in a private GitHub security advisory. The
maintainer will coordinate the patch, advisory, CVE request when appropriate,
and disclosure timing with the reporter. Security fixes may bypass the normal
deprecation window when preserving vulnerable behavior would put users at risk.
