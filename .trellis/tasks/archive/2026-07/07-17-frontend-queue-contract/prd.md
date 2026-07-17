# Frontend queue contract repair

## Closure

Closed before implementation because the download feature is disabled. The
queue UI/API contract remains unrepaired and should be planned again if the
download feature is re-enabled.

## Goal

Restore a truthful, tested download-manager UI against the queue API implemented
by the Go server.

## Confirmed Facts

- The page currently calls `/api/admin/mangadex/queue` and opens a WebSocket at
  that path, but neither route exists.
- The server exposes `GET /api/admin/queue` and
  `POST /api/admin/queue/{action}` for `delete` and `retry`.
- The server returns jobs in a `data` envelope; the page expects `jobs` and
  lower-case job fields.
- The page sends an item ID as a query parameter, while the handler reads a JSON
  request body.
- Pause/resume and queue push notifications have no backend implementation.

## Requirements

- Define one JSON contract for list, item delete, item retry, delete-completed,
  and retry-failed behavior.
- Update the page to use the current non-MangaDex route and request format.
- Remove pause/resume and WebSocket behavior unless implemented end to end.
- Preserve BaseURL mounting, authentication, admin authorization, loading
  states, and actionable error feedback.
- Use semantic buttons for queue actions and retain localized accessible labels.
- Add focused server contract tests and frontend behavior coverage appropriate to
  the chosen test runner.

## Acceptance Criteria

- [ ] Queue listing renders current Go API data and all documented status values.
- [ ] Deleting or retrying one job sends the documented JSON body and refreshes
      the list after success.
- [ ] Delete-completed and retry-failed perform explicit bulk operations with
      tested semantics.
- [ ] The page contains no MangaDex queue route, unsupported pause/resume action,
      or unsupported WebSocket connection.
- [ ] Invalid actions, malformed bodies, missing jobs, and queue failures return
      consistent non-success responses without presenting false success in UI.
- [ ] Behavior works when Mango is mounted below a non-root BaseURL.

## Dependencies

- This is the first implementation milestone and has no dependency on the other
  frontend child tasks.
- It may add testable selectors or focused frontend tests, but the shared browser
  runner belongs to `07-17-frontend-browser-smoke`.

## Out of Scope

- Queue pause/resume and real-time push delivery.
- General dependency upgrades or page redesign.
