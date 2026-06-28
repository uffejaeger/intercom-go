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

## Pending

- Any remaining Intercom API areas not yet represented in the pinned OpenAPI spec normalization output should be added through the same public wrapper pattern.

## Notes

- Some Intercom endpoints use `oneOf` request bodies in the OpenAPI spec. The public SDK wraps those endpoints with package-owned request types so callers do not need generator-specific union helpers.
