# Missing items React pilot design

## Scope

Migrate `/admin/missing` to the React shell and implement real unavailable-item
JSON APIs. Reuse the foundation shell, dual-theme tokens, alerts, and confirm
dialog. Do not migrate other admin pages.

## Backend

Missing items are rows with `unavailable = 1` in:

- `titles` (title metadata)
- `ids` (entry metadata)

Storage methods:

- `ListMissingTitles` / `ListMissingEntries`
- `DeleteMissingTitle` / `DeleteMissingEntry`
- `DeleteAllMissingTitles` / `DeleteAllMissingEntries`

API shapes (compatible with the old browser client field names):

- `GET /api/admin/titles/missing` → `{ success, titles: [{id, path}] }`
- `GET /api/admin/entries/missing` → `{ success, entries: [{id, path}] }`
- `DELETE /api/admin/titles/missing/{tid}`
- `DELETE /api/admin/entries/missing/{eid}`
- `DELETE /api/admin/titles/missing`
- `DELETE /api/admin/entries/missing`

All endpoints remain admin-authenticated under BaseURL.

## Frontend

- `pageId: "missing-items"` mounts `MissingItemsPage`.
- Load both lists, render table, support single delete and bulk delete confirm.
- Loading / empty / error states via shell primitives.
- Go `handleMissingItems` renders `react-shell` instead of `missing-items.tmpl`.

## Validation

```bash
npm run typecheck
npm run build
cd go && go test ./internal/storage ./internal/server
make check
make test
make build
```
