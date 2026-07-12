# Even out routes coverage (admin/reader/opds)

## Goal

Re-enable and register Go routes that already have handlers but were commented
out, and add missing Crystal-critical paths (especially POST /api/login).

## Done (this session)

- [x] POST `/api/login` (JSON) — unauthenticated, cookie + session_id
- [x] Re-enable admin plugin/queue/subscription API routes
- [x] Re-enable `/admin/downloads`, `/admin/subscriptions`, `/download/plugins`
- [x] Reader/OPDS/main core paths already registered

## Out of Scope

- Deep handler behavior parity for every plugin endpoint
- Full Crystal mangadex path naming (Go uses /plugin and /queue)

## Acceptance

- [x] Routes registered; build/test green
