# Task PRD: Client-Side SPA Router Integration

## Goal

Eliminate full-document reloads for in-app navigation so the React shell stays mounted (no white flash) when moving between Home, Library, Title, Tags, Admin, Reader, and related pages—while keeping deep links and browser back/forward.

## Target Area

- `frontend/src/App.tsx`, `frontend/src/main.tsx`
- `frontend/src/lib/boot.ts`, `frontend/src/lib/baseUrl.ts`
- New router helpers / `Link` wrappers under `frontend/src/lib/` or `frontend/src/shell/`
- Nav links: `shell/AppShell.tsx`, browse cards, admin cards, breadcrumbs, reader entry/exit
- `package.json` — add `react-router` (or `react-router-dom`)

## Confirmed facts (from repo)

- App selects page via `readBoot().pageId` switch; production boot from `#mango-boot`; Vite uses `bootFromPathname`.
- Most in-app links are plain `<a href={baseUrl(...)}>` → full document navigation.
- Some pages already use `history.replaceState` for query/reader page.
- No router library today; Go still serves React shell + APIs for all paths.
- Logout / download must stay full navigations.

## Requirements

1. Integrate **react-router** for client-side path transitions under configured `baseUrl`.
2. Convert shell + main browse/admin links (and reader entry/exit) to SPA navigation.
3. Reader is an SPA route with immersive chrome (no AppShell).
4. Session fields (`baseUrl`, `isAdmin`, `version`) come from a sticky boot context initialized once from `#mango-boot` / first paint; route params supply page-specific IDs.
5. Keep full navigation for logout, `api/download/*`, and login success assign (MVP).
6. Deep links and back/forward work; Go continues to serve shell HTML for first load of each path.

## Acceptance Criteria

- [x] In-app navigation among Home / Library / Title / Tags / Admin / Reader does not full-reload the document.
- [x] Direct URL open and browser back/forward land on the correct page.
- [x] Reader open/close and entry switches work via router without breaking page replaceState behavior for page number (or equivalent navigate replace).
- [x] Logout and download links still full-navigate.
- [x] `baseUrl` mount prefix still works.
- [x] Frontend typecheck/build passes.

## Out of scope

- Changing chi API contracts.
- Global prefs store (`global-ui-context` task).
- Making login success a soft SPA navigate (optional follow-up).
- Server-side rendering.

## Decisions

- **Router:** react-router (client-side); Go still serves shell + API.
- **Link scope:** shell topbar + main browse/admin paths + reader entry/exit.
- **Reader:** included in SPA; immersive UI without AppShell.
- **Boot session:** sticky React context from initial `#mango-boot`; route params for page-specific identity.

## Open questions

- None blocking planning.

## Notes

- Parent: `07-22-project-review`.
- Complex: needs `design.md` + `implement.md` before `task.py start`.
