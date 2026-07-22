# Task PRD: Migrate I18n and Boot providers to Zustand

## Goal

Replace React `I18nProvider` and `BootProvider` with Zustand stores so all client global state uses one paradigm (alongside existing theme/reader stores).

## Target Area

- `frontend/src/lib/i18n.tsx` (`I18nProvider`, `useI18n`, language storage, messages)
- `frontend/src/lib/bootContext.tsx` (`BootProvider`, `useBoot`, `routerBasename`, `appPath`)
- `frontend/src/main.tsx` provider tree
- `frontend/src/lib/prefsSync.ts` (extend for language multi-tab)
- Call sites keep using `useI18n` / `useBoot` via thin adapters

## Confirmed facts (from repo)

- **I18n:** Context + `localStorage['mango-language']`; `setLanguage` + `document.documentElement.lang`; `t(key, vars)` formats `{name}` placeholders; large static `messages` map.
- **Boot:** `BootProvider` memoizes once from `readBoot()` → `{ baseUrl, isAdmin, version, pageName }`; `useBoot` falls back to `readBoot()` outside provider; `routerBasename` / `appPath` live in same file.
- **Already Zustand:** `themeStore`, `readerPrefsStore`, multi-tab `prefsSync`.
- **global-ui-context** merged; this is the planned follow-up.

## Requirements

1. Language + `t()` via Zustand; same key `mango-language`; message catalogs unchanged.
2. Boot session fields initialized once from `readBoot()` into a Zustand store.
3. Keep hook names `useI18n` / `useBoot` as thin store wrappers.
4. Remove `I18nProvider` / `BootProvider` from `main.tsx` (BrowserRouter remains).
5. Preserve `routerBasename` / `appPath` exports.
6. Multi-tab sync for `mango-language` via `storage` (same pattern as prefs).

## Acceptance Criteria

- [x] `main.tsx` no longer wraps with `I18nProvider` / `BootProvider`.
- [x] Language switch + `t()` work as today; `document.documentElement.lang` still updates.
- [x] Changing language in one tab updates other open tabs of the same origin.
- [x] `useBoot()` returns stable session fields matching first-paint `#mango-boot` / pathname fallback.
- [x] Typecheck and build pass.

## Out of scope

- Changing translation strings / adding languages.
- Server-side prefs.
- Merging theme/reader/i18n/boot into one mega store.
- Multi-tab sync for boot (not stored in localStorage).

## Decisions

- **Shape:** split `i18nStore` + `bootStore`.
- **Multi-tab:** yes for `mango-language`.
- **Hooks:** keep `useI18n` / `useBoot` as thin adapters over stores.
- **Messages:** stay in module (not in store state) to avoid bloating subscriptions.

## Open questions

- None blocking planning.

## Notes

- Parent: `07-22-project-review`.
- Complex: needs `design.md` + `implement.md` before `task.py start`.
