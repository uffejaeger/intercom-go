# Support

`intercom-go` is a community-maintained SDK. Support is best effort and does not
include a response-time or resolution-time service-level agreement.

## Where to ask

- Open a GitHub issue for a reproducible SDK bug, missing public wrapper, or
  focused feature request.
- Use Intercom's official support channels for account access, billing,
  workspace configuration, API availability, or questions about Intercom's
  service behavior.
- Use [GitHub private vulnerability reporting](SECURITY.md) for suspected
  security issues. Never include vulnerability details in a public issue.

Before opening an SDK issue, search existing issues and test with the latest
released SDK version and a supported Go toolchain. Include the SDK and Go
versions, the Intercom API region, a minimal reproduction, expected and actual
behavior, and a sanitized request ID when available. Never include access
tokens, authorization headers, or customer data.

## Maintenance expectations

Maintainers prioritize security reports, regressions in released behavior,
upstream API compatibility, and broadly useful SDK improvements. Feature
requests may remain open until there is clear demand and a stable API design.
Questions that require access to an Intercom workspace or customer data may be
redirected to Intercom support.

The release, compatibility, and deprecation commitments are documented in
[`docs/compatibility.md`](docs/compatibility.md).
