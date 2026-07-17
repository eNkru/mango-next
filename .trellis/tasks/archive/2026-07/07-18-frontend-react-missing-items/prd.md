# Missing items React pilot

## Goal

Migrate the admin missing-items page to React and prove the foundation with a
complete, API-driven admin flow.

## Confirmed Facts

- Route exists at `/admin/missing` and currently renders `missing-items.tmpl`.
- Browser code expects list/delete APIs under `/api/admin/titles/missing` and
  `/api/admin/entries/missing`.
- Current Go handlers return empty success stubs, so the pilot requires real
  backend behavior as part of the page migration.
- Existing UI needs loading, empty, list, single delete, bulk delete with
  confirmation, and error feedback.
- Parent strategy is route-level coexistence: only this route switches to React.

## Requirements

- Depends on `07-18-frontend-react-foundation`.
- Replace the `/admin/missing` template page with the Go HTML shell + React page.
- Implement real JSON contracts for listing and deleting missing titles/entries.
- Keep admin authorization and BaseURL behavior.
- Render dual-theme-compatible table/list UI with loading, empty, error, delete,
  and bulk-delete confirmation states.
- Use TypeScript types for API requests/responses.
- Preserve unmigrated admin pages on Go templates.
- Add focused Go API tests and frontend verification appropriate to the chosen
  runner.

## Acceptance Criteria

- [x] `/admin/missing` is served by the React shell, not `missing-items.tmpl`.
- [x] Missing titles and entries load from real backend data, not empty stubs.
- [x] Single-item delete and bulk delete refresh the list after success.
- [x] Errors and empty states are shown without false success.
- [x] Works under root and non-root BaseURL with admin auth enforced.
- [x] Unrelated admin pages remain on Go templates.
- [x] Route-level rollback to the old template remains possible without removing
      the React foundation.

## Dependencies

- Parent: `07-17-frontend-react-vite`
- Requires: `07-18-frontend-react-foundation`

## Out of Scope

- Migrating other admin pages.
- Download-manager cleanup.
- Reader/library migration.
- Full SPA client-side routing across admin.
