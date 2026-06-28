# Public SDK Coverage

This document tracks which Intercom API areas are exposed through the public `intercom` package.

## Implemented

- Admins
- AI content
- Articles
- Away status reasons
- Brands
- Calls
- Collections
- Companies
- Contacts
- Conversations
- Fin
- Help centers
- Internal articles
- Messages
- News
- Notes
- Phone switches
- Segments
- Subscription types
- Tags
- Teams
- Tickets
- Visitors
- Workspace exports, jobs, and IP allowlist

## Partial

- Fin: response payloads with anonymous generated schemas are wrapped through local decode structs; the surface is usable but not yet polished.
- News: list endpoints are wrapped, but the generated response uses the generic paginated envelope rather than typed item lists.
- Tags: list/retrieve/delete are typed. The create/update/tagging endpoint is exposed through a generic JSON body wrapper because the upstream spec uses a multi-shape request union.
- Tickets: reply and ticket tag endpoints are now wrapped, but the request payloads are still intentionally loose because the upstream spec uses union request shapes.

## Pending

- Any remaining Intercom API areas not yet represented in the pinned OpenAPI spec normalization output should be added through the same public wrapper pattern.

## Notes

- Some Intercom endpoints use `oneOf` request bodies in the OpenAPI spec. The generated request types for those endpoints are not yet appropriate as-is for an idiomatic public SDK.
- For those endpoints, the remaining work is not just wiring methods into `Client`; it also includes designing stable request types that avoid leaking generator-specific union helpers into the public API.
