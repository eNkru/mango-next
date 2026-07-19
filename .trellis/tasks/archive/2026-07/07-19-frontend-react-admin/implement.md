# React admin home execution plan

## 0. Preconditions

- Parent: 07-17-frontend-react-vite
- Reader / browse / users / missing already React
- Branch: `feature/react-admin`

## 1. Backend scan job + route switch

- [x] Add library scan job status (start / finish / status) without deadlocking `scanMu`.
- [x] Change `POST /api/admin/scan` to claim job + background `Scan()` + quick JSON.
- [x] Add `GET /api/admin/scan_progress` with running + last titles/ms/error.
- [x] Register scan_progress next to other admin APIs.
- [x] Switch `handleAdmin` to `renderReactShell` `pageId: admin`.
- [x] Keep `admin.tmpl` / `admin.js`.
- [x] Go tests: scan start, progress idle/complete; admin pageId on `/admin`.

## 2. Frontend theme + admin page

- [x] Add `lib/theme.ts` (keys + applyHtmlTheme matching react-shell FOUC).
- [x] Extend AppShell with theme + ui-style selects (all React pages).
- [x] Implement AdminPage: users/missing links, scan + thumb actions with poll.
- [x] i18n zh-CN / zh-TW / en for admin + theme strings.
- [x] Register `admin` in App.tsx inside AppShell pattern (page component uses AppShell).

## 3. Validation

- [x] `npm run typecheck` && `npm run build`
- [x] Focused Go tests; server + library packages green
- [ ] Smoke: deferred by request (browser not run this session)
- [x] Legacy admin assets retained on disk

## Risk and rollback

- Highest risk: scan job status races / double Start.
- Theme class apply must not leave both comic and flat markers.
- Rollback: restore template `handleAdmin`; leave scan_progress API if harmless.
