# Public SDK Coverage

This SDK wraps the pinned Intercom API `2.16` OpenAPI spec with public root-package services while keeping generated code internal.

## Current status

- The generated client exposes all response-returning operations in the pinned 2.16 specification.
- Public SDK services cover those operations through idiomatic wrappers.
- `DataEvents.List` is the known audit exception: it intentionally uses `Client.NewRequest` and `Client.Do` instead of the generated `LisDataEventsWithResponse` helper so the SDK can provide explicit identifier validation and query encoding.
- `TestGeneratedOperationsAreAccountedFor` is the adopted offline contract check for public wrapper coverage. It parses the generated `ClientWithResponsesInterface` and root-package SDK code, then fails if a generated operation is neither wrapped nor listed as an intentional exception.
- `make coverage` currently passes at the required `99.9%`.
- `make generate-check` currently passes, so committed generated code is in sync with the pinned spec.

## Public services

- `AIContent`
- `Admins`
- `Articles`
- `Audiences`
- `AwayStatusReasons`
- `Brands`
- `Calls`
- `Collections`
- `Companies`
- `Content`
- `Contacts`
- `Conversations`
- `ConversationAttributes`
- `CustomObjects`
- `DataAttributes`
- `DataEvents`
- `DataConnectors`
- `Emails`
- `Fin`
- `HelpCenters`
- `HelpCenterRedirects`
- `InternalArticles`
- `Messages`
- `Macros`
- `News`
- `Notes`
- `OfficeHours`
- `PhoneSwitches`
- `Segments`
- `SubscriptionTypes`
- `Tags`
- `Teams`
- `Tickets`
- `Visitors`
- `WhatsApp`
- `Workspace`

## Audit notes

Some public methods intentionally use `WithBodyWithResponse` generated helpers or lower-level requests instead of the simpler generated helpers. This preserves public request types, fixes awkward upstream operation names, and avoids leaking generated union details where a clearer SDK API is practical.

Spec example fixture tests are intentionally deferred. The pinned Intercom spec is useful for generated code and coverage audits, but broad example-based response tests would duplicate generated-client behavior and likely be brittle. Representative offline SDK integration tests cover SDK request/response behavior without depending on live Intercom state.
