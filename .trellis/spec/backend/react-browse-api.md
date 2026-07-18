# React Browse API Contracts

## 1. Scope / Trigger

Apply this contract when changing the React home, library, title-detail APIs or
their title/entry metadata mutations. These routes coexist with legacy API
consumers, so additive React DTOs must not change old envelope semantics.

## 2. Signatures

```text
GET  /api/home
GET  /api/library?show_hidden=1
GET  /api/book/{tid}
PUT  /api/admin/display_name/{tid}       { name, eid? }
PUT  /api/admin/sort_title/{tid}         { sort_name, eid? }
POST /api/admin/upload/cover?tid=&eid?   multipart file
```

The path-encoded `PUT /api/admin/display_name/{tid}/{name}` remains registered
only for legacy coexistence.

## 3. Contracts

- React read DTOs use top-level `title`, `titles`, `entries`, `parents`, and
  `is_admin` fields. `progress` is a percentage from 0 through 100 and entry
  `page` is the persisted page value.
- All cover URLs are already BaseURL-aware. Clients must not prepend BaseURL to
  response URLs, but must use `baseUrl()` for API and navigation paths.
- `GET /api/library` keeps `data` as an alias of the title array and each title
  keeps `display_name` for old consumers.
- `GET /api/book/{tid}` keeps `data.entries[].progress` as the legacy persisted
  page number. The new top-level `entries[].progress` is the percentage.
- Display names persist in each title directory's `info.json`. Title names use
  `display_name`; direct entries use `entry_display_name[fileName]`. Unknown
  `info.json` keys must survive every write.
- Title and entry sort names persist through storage `sort_title` fields.

## 4. Validation & Error Matrix

| Condition | Result |
|---|---|
| Missing title | `404` JSON `Title not found` |
| Missing direct entry on metadata write | `404` JSON `Entry not found` |
| Empty/whitespace display name | `400` JSON error |
| Invalid JSON body | `400` JSON error |
| Non-admin metadata mutation | admin middleware returns `403` |
| Admin requests `show_hidden=1` | hidden titles included and marked |
| Non-admin requests `show_hidden=1` | parameter ignored |
| Metadata file/storage write fails | non-2xx JSON error; memory is not updated first |

## 5. Good/Base/Bad Cases

- Good: update an entry display name, then `GET /api/book/{tid}` immediately
  returns it and a fresh scan/cache load returns the same value.
- Base: missing `info.json` yields filesystem names and an empty metadata map.
- Bad: reuse the React percentage DTO as legacy `data.entries`; a saved page 2
  becomes `66.7` and breaks old reader/title consumers.

## 6. Tests Required

- Library unit tests assert unknown `info.json` keys survive and title/entry
  display names apply during fresh construction and cache hydration.
- Authenticated route tests assert page IDs for `/`, `/library`, and
  `/book/{tid}`, hidden filtering, admin boundaries, mutation validation, and
  mutation-to-read round trips.
- BaseURL tests assert mounted API and shell paths under `/mango/` and reject
  unprefixed application paths.
- Frontend typecheck/build must consume the shared `BrowseTitle` and
  `BrowseEntry` types rather than page-local casts.

## 7. Wrong vs Correct

Wrong: replace both response locations with the React percentage object.

```go
sendJSON(w, map[string]any{"entries": entries, "data": map[string]any{"entries": entries}})
```

Correct: expose the typed React DTO at the top level and explicitly project the
legacy page-number contract inside `data`.

```go
sendJSON(w, map[string]any{
    "entries": entries,
    "data": map[string]any{"entries": legacyEntries},
})
```
