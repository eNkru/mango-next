# Task PRD: Migrate I18n and Boot providers to Zustand

## Goal

Replace React `I18nProvider` and `BootProvider` with Zustand stores (or slices) so all client global state uses one paradigm after the UI prefs store lands.

## Target Area

- `frontend/src/lib/i18n.tsx` (`I18nProvider`, `useI18n`)
- `frontend/src/lib/bootContext.tsx` (`BootProvider`, `useBoot`)
- `frontend/src/main.tsx` provider tree
- All `useI18n` / `useBoot` call sites (API shape may stay compatible via wrappers)

## Requirements

1. Language + `t()` available via Zustand without breaking message catalogs.
2. Boot session (`baseUrl`, `isAdmin`, `version`) initialized once from `readBoot()` into store.
3. Prefer keeping hook names (`useI18n`, `useBoot`) as thin adapters to reduce call-site churn.
4. Depends on / follows `07-22-global-ui-context` (Zustand already in project).

## Acceptance Criteria

- [ ] No `I18nProvider` / `BootProvider` required in `main.tsx` (or only if needed for tests).
- [ ] Language switch and boot fields work as today.
- [ ] Typecheck/build pass; no user-visible i18n regressions.

## Out of scope

- Changing translation strings.
- Server-side prefs.

## Notes

- Parent: `07-22-project-review`.
- Status: planning; do not start until global-ui-context (prefs Zustand) is done.
