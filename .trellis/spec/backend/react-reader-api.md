# React Reader API Contracts

## 1. Scope / Trigger

Apply when changing the React immersive reader route, reader bootstrap API, page
image serving for reading, or progress writes used by the reader. Browse
title-detail only deep-links into `/reader/...`; image bytes stay on the page
endpoint.

## 2. Signatures

```text
GET  /reader/{title}/{entry}              → 302 BaseURL-aware .../1
GET  /reader/{title}/{entry}/{page}       → React shell pageId=reader boot {tid,eid,page,isAdmin}
GET  /api/reader/{tid}/{eid}              → bootstrap JSON (auth)
GET  /api/page/{tid}/{eid}/{page}         → raw image bytes (1-based page)
PUT  /api/progress/{tid}/{page}?eid=      → { success: true }
GET  /api/dimensions/{tid}/{eid}          → legacy dimensions helper (still used elsewhere)
```

## 3. Contracts

### Bootstrap success (`GET /api/reader/{tid}/{eid}`)

```json
{
  "success": true,
  "data": {
    "title": { "id": "...", "name": "..." },
    "entry": { "id": "...", "name": "...", "pages": 120, "progress": 12 },
    "dimensions": [{ "width": 1000, "height": 1500 }],
    "entries": [{ "id": "...", "name": "...", "pages": 100, "progress": 0 }],
    "exit_url": "/book/{tid}",
    "next_entry_url": "/reader/{tid}/{nextEid}/1",
    "previous_entry_url": "/reader/{tid}/{prevEid}/1"
  }
}
```

- `exit_url`, `next_entry_url`, `previous_entry_url` are BaseURL-aware absolute
  path strings (`appPath`), or empty when absent.
- `dimensions.length` is the image list source of truth; if it differs from
  `entry.PageCount()`, bootstrap sets `entry.pages = len(dimensions)`.
- Sibling `entries` follow parent title order for entry jump.
- Progress is the persisted one-based page (or 0), not a percentage.

### Page index rule (critical)

Runtime `GET /api/page/{tid}/{eid}/{page}` and `Entry.ReadPage` are **1-based**
(legacy `reader.js` used `i+1`). Do **not** convert public URL pages to
0-based when calling `/api/page`. Public URL pages are also 1-based.

Design docs that say “zero-based page API” are outdated relative to the live
Go archive/dir readers.

### Shell boot

```json
{ "pageId": "reader", "pageName": "reader", "tid": "...", "eid": "...", "page": 1, "baseUrl": "/", "isAdmin": false }
```

Missing/corrupt entries still render the shell; the React page shows error
state from bootstrap JSON failures.

## 4. Validation & Error Matrix

| Condition | Result |
|---|---|
| Missing title/entry | `404` JSON `success:false` |
| Entry has zero pages / empty dimensions | `400` JSON `entry has no pages` |
| Dimensions read failure | `500` JSON dimensions error |
| Invalid progress page | `400` on `PUT /api/progress` |
| Unauthenticated API | auth middleware |

## 5. Good/Base/Bad Cases

- Good: bootstrap returns `dimensions.length === entry.pages`, BaseURL-prefixed
  exit/next/prev, React opens on boot `page`.
- Base: single-entry title → empty `next_entry_url` / `previous_entry_url`.
- Bad: treat `/api/page` as 0-based and request page `0` or `page-1` → blank or
  out-of-range errors.

## 6. Tests Required

- `reader_api_test.go`: bootstrap success fields, 404 missing title, React shell
  `pageId` for `/reader/.../{page}`, no-page redirect to `.../1`.
- Do not delete `reader.tmpl` / `reader.js` until browser smoke is accepted.

## 7. Wrong vs Correct

#### Wrong

```go
// Assume zero-based page endpoint when building image URLs
url := fmt.Sprintf("api/page/%s/%s/%d", tid, eid, page-1)
```

#### Correct

```go
// Match ReadPage / legacy reader.js (1-based)
url := fmt.Sprintf("api/page/%s/%s/%d", tid, eid, page)
```

#### Wrong

```go
s.renderPage(w, "views/reader", data) // new work path
```

#### Correct

```go
s.renderReactShell(w, "reader", "reader", map[string]any{
  "isAdmin": GetIsAdmin(r), "tid": titleID, "eid": entryID, "page": page,
})
```
