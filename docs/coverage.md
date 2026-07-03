# Public SDK Coverage

This SDK wraps the pinned Intercom API `2.15` OpenAPI spec with public root-package services while keeping generated code internal.

## Current status

- The generated client exposes 161 response-returning operations.
- Public SDK services cover those operations through idiomatic wrappers.
- `DataEvents.List` is the known audit exception: it intentionally uses `Client.NewRequest` and `Client.Do` instead of the generated `LisDataEventsWithResponse` helper so the SDK can provide explicit identifier validation and query encoding.
- `TestGeneratedOperationsAreAccountedFor` is the adopted offline contract check for public wrapper coverage. It parses the generated `ClientWithResponsesInterface` and root-package SDK code, then fails if a generated operation is neither wrapped nor listed as an intentional exception.
- `make coverage` currently passes at the required `99.9%`.
- `make generate-check` currently passes, so committed generated code is in sync with the pinned spec.

## Public services

- `AIContent`
- `Admins`
- `Articles`
- `AwayStatusReasons`
- `Brands`
- `Calls`
- `Collections`
- `Companies`
- `Contacts`
- `Conversations`
- `CustomObjects`
- `DataAttributes`
- `DataEvents`
- `Emails`
- `Fin`
- `HelpCenters`
- `InternalArticles`
- `Messages`
- `News`
- `Notes`
- `PhoneSwitches`
- `Segments`
- `SubscriptionTypes`
- `Tags`
- `Teams`
- `Tickets`
- `Visitors`
- `Workspace`

## Audit notes

Some public methods intentionally use `WithBodyWithResponse` generated helpers or lower-level requests instead of the simpler generated helpers. This preserves public request types, fixes awkward upstream operation names, and avoids leaking generated union details where a clearer SDK API is practical.

Spec example fixture tests are intentionally deferred. The pinned Intercom spec is useful for generated code and coverage audits, but broad example-based response tests would duplicate generated-client behavior and likely be brittle. Representative offline SDK integration tests cover SDK request/response behavior without depending on live Intercom state.
