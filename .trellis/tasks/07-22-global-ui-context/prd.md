# Task PRD: Global UI Context and State Store

## Goal

Consolidate client UI preferences (theme, UI style, reader prefs) into **Zustand** stores synced with existing `localStorage` keys (including multi-tab via `storage` events), so AppShell and reader share one source of truth under SPA navigation.

## Target Area

- `frontend/src/lib/theme.ts` (keep pure apply helpers; store owns state)
- `frontend/src/pages/reader/useReaderPrefs.ts` → thin wrapper over reader store
- `frontend/src/shell/AppShell.tsx` theme/style controls
- New: `frontend/src/lib/themeStore.ts`, `frontend/src/lib/readerPrefsStore.ts` (names flexible)
- `package.json` — add `zustand`

## Confirmed facts (from repo)

- Theme keys: `theme`, `ui-style` via `theme.ts`; AppShell local state today.
- Reader keys: `mango.reader.*` via `useReaderPrefs`.
- Language / Boot: stay Context; follow-up `07-22-zustand-i18n-boot`.
- No existing third-party state lib.

## Requirements

1. Add Zustand **theme store** (theme + uiStyle) and **reader prefs store**.
2. Persist with the **same** localStorage keys (no user migration break).
3. AppShell reads/writes theme + ui-style via theme store.
4. `useReaderPrefs` (or equivalent) reads/writes via reader store.
5. Theme/style changes call `applyHtmlTheme`.
6. Multi-tab: listen to `window` `storage` and rehydrate the affected store.
7. Do **not** migrate I18n or Boot.

## Acceptance Criteria

- [x] Theme / UI style controls use the theme Zustand store.
- [x] Reader prefs load/save same `mango.reader.*` keys through the reader store.
- [x] Changing prefs in one tab updates other open tabs of the same origin.
- [x] Existing localStorage keys preserved.
- [x] I18nProvider and BootProvider remain as-is.
- [x] `npm run typecheck` and `npm run build` pass.

## Out of scope

- Migrating I18n/Boot → `07-22-zustand-i18n-boot`.
- Server-side preference APIs.
- Changing i18n message catalogs.

## Decisions

- **Library:** Zustand.
- **Scope:** theme + ui-style + reader prefs only.
- **I18n / Boot:** Context; follow-up task exists.
- **Multi-tab:** yes — `storage` event rehydrate.
- **Shape:** **split stores** — `themeStore` + `readerPrefsStore` (not one mega prefs store).

## Open questions

- None blocking planning.

## Notes

- Parent: `07-22-project-review`.
- Complex: `design.md` + `implement.md` required before `task.py start`.
