# React admin home migration

## Goal

Migrate `/admin` (admin home) from the legacy Alpine/jQuery/UIkit page to the
existing React + Vite shell, so day-2 ops (library scan, thumbnail generation)
and admin navigation live in React with the same auth and BaseURL rules.

## Confirmed Facts

- Route: `GET /admin` → `handleAdmin` renders layout template `admin` today.
- Template cards (`go/web/views/admin.tmpl` + `go/web/public/js/admin.js`):
  1. Link → `/admin/user` (already React)
  2. Link → `/admin/missing` (already React; template supports `MissingCount`
     badge but `handleAdmin` does **not** populate `MissingCount` today)
  3. **Scan library** → `POST /api/admin/scan`
  4. **Generate thumbnails** → `POST /api/admin/generate_thumbnails` + poll
     `GET /api/admin/thumbnail_progress` every 5s
  5. **Theme** select: Dark / Light / System (`localStorage` via legacy helpers)
  6. **UI style** select: Comic / Flat (`localStorage`)
  7. Logout link
- Existing APIs (admin middleware):
  - `POST /api/admin/scan` — currently starts `Library.Scan()` in a goroutine and
    returns `{ success: true }` immediately (**does not** return
    `milliseconds` / `titles` that legacy JS still reads)
  - `POST /api/admin/generate_thumbnails` — starts generation async,
    `{ success: true }`
  - `GET /api/admin/thumbnail_progress` — `{ success, progress, running }`
    (covered by `thumbnail_progress_test.go`)
- React already has AppShell, i18n (zh-CN / zh-TW / en), alerts, and migrated
  admin child pages (users, missing). No `pageId: admin` yet.
- Parent task recommends this as the next child after reader; subscriptions /
  plugin download stay out of this child.
- Legacy `admin.tmpl` / `admin.js` should remain for rollback until smoke passes.

## Requirements

- Render `/admin` through React shell (`pageId: admin`) under admin auth only.
- Feature-equivalent admin home: navigation to users + missing, scan action,
  thumbnail generate + progress, theme + UI style controls, logout.
- BaseURL-aware links and API calls; no Node at runtime.
- Shared AppShell (not immersive reader chrome); admin-only page.
- zh-CN / zh-TW / en chrome strings for this page.
- Scan UX: **async start + progress polling** (thumbnail-style), not the current
  fire-and-forget silent success and not a single long-blocking POST.
- Missing-items badge: optional if cheap via existing missing-list APIs;
  not a blocker if count is omitted.
- Keep `admin.tmpl` / `admin.js` until smoke accepted.
- Focused Go tests for any scan/progress contract change; frontend typecheck/build.

## Acceptance Criteria

- [ ] `/admin` loads React admin home for admin users; non-admin stays blocked.
- [ ] Cards/links open users and missing-items (already React) under BaseURL.
- [ ] Scan starts via admin API; UI polls scan progress until not running and
      shows a clear result (e.g. title count / duration) or failure.
- [ ] Generate thumbnails starts job; progress polling updates UI; completes
      when `running` is false.
- [ ] Theme and UI style changes apply immediately and persist across reloads
      (same storage keys as React shell / legacy).
- [ ] Logout works.
- [ ] Admin home only links to already-React destinations (users, missing);
      subscriptions / plugin download / downloads stay off the home cards
      (direct URLs still work).
- [ ] Legacy assets retained; no jQuery/Alpine/UIkit on the React admin shell.
- [ ] Tests/build gates for changed contracts pass.

## Out of Scope

- Subscriptions manager, plugin download UI, OPDS.
- Download manager product revival.
- Full server config editing UI (YAML keys).
- Parent-wide legacy asset deletion.

## Decisions

- Theme / UI style controls live in **global AppShell** (all React pages), not
  only on the admin home cards. Keys match shell FOUC: `localStorage` `theme`
  (`dark`|`light`|`system`) and `ui-style` (`comic`|`flat`); apply `<html>`
  class markers the same way as `react-shell.tmpl` head script.
- Admin home still has ops cards (users, missing, scan, thumbnails) + logout
  via shell; no duplicate theme cards required on `/admin` once shell has them.
- **Scan** uses async start + progress polling (mirror thumbnails):
  - `POST /api/admin/scan` starts (or no-ops if already running) and returns
    quickly `{ success: true }`.
  - New `GET /api/admin/scan_progress` returns at least
    `{ success, running, ... }` and on completion enough metrics for UI
    (e.g. titles / milliseconds from last scan).
  - React polls while running; shows busy + final result/error.
  - Requires library-level scan job status (similar to `ThumbnailStatus`);
    `scanMu` already serializes `Scan()` — design must define start/reject
    duplicate and last-result storage.

## Open Questions

- (none blocking)
