# React Admin Home

## 1. Scope / Trigger

Apply when changing `/admin` React page, AppShell theme controls used by admin,
or admin scan/thumbnail client polling.

## 2. Signatures / Layout

```text
GET /admin → pageId admin → AdminPage (AppShell)
frontend/src/pages/AdminPage.tsx
frontend/src/lib/theme.ts
POST /api/admin/scan
GET  /api/admin/scan_progress
POST /api/admin/generate_thumbnails
GET  /api/admin/thumbnail_progress
```

## 3. Contracts

- Home cards only: users, missing, scan, thumbnails (no subscriptions/plugin/downloads).
- Scan/thumb: start then poll ~1s while `running`; show last titles/ms or %.
- Theme/UI style: AppShell globals; keys and class rules in `ui-theme-layout.md`.
- Keep `admin.tmpl` / `admin.js` until explicit retirement.
- **No full-page LoadingState** (admin is an action panel, not a data list).
- Scan / generate_thumbnails **start** failures: set `actionError` and render
  `ErrorState` with `onRetry` re-running that action (prefer ErrorState over a
  duplicate danger alert for the same failure). Mid-run poll errors may still
  use `pushAlert`.

## 4. Wrong vs Correct

Wrong: wrap admin in immersive reader chrome or link half-migrated admin routes.

Wrong: force a whole-page LoadingState on first paint just to “match other pages”.

Correct: `AppShell` + already-React destinations only; action failures use
`ErrorState` + `onRetry` under the card grid.
