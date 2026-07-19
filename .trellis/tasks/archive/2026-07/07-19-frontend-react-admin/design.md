# React admin home design

## Boundaries

| Route | pageId | Chrome |
|---|---|---|
| `/admin` | `admin` | Shared `AppShell` (not immersive reader) |

Unmigrated: subscriptions, plugin download, download manager, OPDS stay Go
templates. This child does not migrate those pages and does not add home cards
to them.

Theme / UI style controls move to **AppShell** (all React pages). Admin home
focuses on ops cards: users, missing, scan, thumbnails.

## Data flow

```
GET /admin
  → renderReactShell pageId=admin (admin middleware)
  → React AdminPage
       → POST /api/admin/scan
       → GET  /api/admin/scan_progress   (poll while running)
       → POST /api/admin/generate_thumbnails
       → GET  /api/admin/thumbnail_progress  (poll while running)
       → links: baseUrl('admin/user'), baseUrl('admin/missing')
AppShell
  → localStorage theme + ui-style → apply html class markers
```

## Scan job contract (new)

Mirror thumbnail job shape so the UI can reuse one polling pattern.

### Start

`POST /api/admin/scan` (admin)

- If a scan is already running: return `200` with `{ success: true, running: true }`
  (or `409` — prefer soft success + `running: true` for simpler client).
- Else start `Library.Scan()` in a background goroutine; return
  `{ success: true, running: true }` immediately.

### Progress

`GET /api/admin/scan_progress` (admin)

```json
{
  "success": true,
  "running": false,
  "progress": 0,
  "titles": 12,
  "milliseconds": 340,
  "error": ""
}
```

Rules:

- While running: `running: true`; `progress` may be `0` if no finer grain is
  available (scan is one critical section today — binary busy is enough).
- When idle after success: `running: false`, last result fields populated
  (`titles` = title count from last `ScanResult`, `milliseconds` = wall time).
- When idle after failure: `running: false`, `error` non-empty; metrics may be 0.
- Before any scan this process: `running: false`, metrics 0, empty error.

### Library status

Add scan job state next to thumbnails (e.g. `scanCtx` or small struct on
`Library`):

- `Start() bool` — claim job or reject duplicate
- `Finish(result, err, ms)`
- `Status() (running bool, titles int, ms int, err string, progress float64)`

`Library.Scan()` already holds `scanMu`; job start must not deadlock with
status reads. Prefer status under a separate mutex (like thumbnail context).

Keep existing `POST /api/admin/scan` path; change behavior from fire-and-forget
without status to status-tracked async (breaking only the undocumented empty
metrics legacy JS expected but never received).

## Thumbnail APIs (unchanged)

```text
POST /api/admin/generate_thumbnails  → { success: true }
GET  /api/admin/thumbnail_progress   → { success, progress, running }
```

React polls every ~1–2s while `running` (legacy used 5s; 1–2s is fine).

## Frontend architecture

```
frontend/src/
  pages/AdminPage.tsx              # ops cards + scan/thumb UI
  shell/AppShell.tsx               # + theme / ui-style selects
  lib/theme.ts                     # read/write localStorage + applyHtmlTheme()
  lib/i18n.tsx                     # admin + theme strings
```

### AppShell theme controls

- Keys: `theme` = `dark|light|system`, `ui-style` = `comic|flat`
- On change: write localStorage, recompute dark (system → matchMedia), set
  `html` classes exactly as `react-shell.tmpl` FOUC script:
  - comic: `comic-theme` [+ `comic-theme-dark`]
  - flat: `flat-theme` [+ `flat-theme-dark`]
- Remove the opposite style’s classes when switching.

### AdminPage cards

1. Link: user management → `/admin/user`
2. Link: missing entries → `/admin/missing` (optional badge via missing list
   length if cheap)
3. Action: scan library (busy + last result)
4. Action: generate thumbnails (busy + progress %)
5. No theme cards (shell owns them)
6. No subscriptions / plugin / downloads cards

Logout remains in AppShell.

## Compatibility

- Public URL `/admin` unchanged; admin middleware unchanged.
- Additive scan_progress API; scan POST stays same path with richer status.
- BaseURL via existing `baseUrl()` / `appPath`.
- Leave `admin.tmpl` + `admin.js` on disk for rollback.

## Rollback

Restore `handleAdmin` to `renderLayout(w, "admin", …)` and stop registering
`pageId: admin`. Scan job helpers can remain unused or be kept for API clients.

## Tradeoffs

| Choice | Why | Cost |
|---|---|---|
| Theme in AppShell | User request; available on all React pages | Touches shared shell |
| Scan progress API | User-selected UX; matches thumbnails | New library job state + tests |
| Only React deep links on home | Avoid half-migrated nav | Ops still use bookmarks for other admin pages |
