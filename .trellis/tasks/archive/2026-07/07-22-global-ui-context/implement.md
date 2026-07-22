# Implement: Global UI prefs Zustand stores

## Checklist

1. `npm install zustand` (repo root package.json).
2. Add `themeStore.ts` — init from localStorage, setTheme/setUIStyle, rehydrate, storage listener (or shared listener module).
3. Add `readerPrefsStore.ts` — prefs + setPrefs + rehydrate; storage keys identical to current `useReaderPrefs`.
4. Wire `storage` listener once (shared `prefsSync.ts` or each store module).
5. Refactor `AppShell` to use theme store selectors.
6. Refactor `useReaderPrefs` to thin Zustand wrapper (keep export path).
7. Ensure system theme watcher still works with store state.
8. `npm run typecheck` && `npm run build`.

## Validation

```bash
npm run typecheck
npm run build
```

Manual: toggle theme/style in AppShell; open second tab → change theme → first tab updates; reader prefs persist and sync across tabs.

## Risk

- Forgot selector → extra re-renders (prefer narrow selectors).
- Double-apply theme on init (idempotent `applyHtmlTheme` OK).

## Files

- `package.json`, lockfile
- `frontend/src/lib/themeStore.ts` (new)
- `frontend/src/lib/readerPrefsStore.ts` (new)
- optional `frontend/src/lib/prefsSync.ts`
- `frontend/src/shell/AppShell.tsx`
- `frontend/src/pages/reader/useReaderPrefs.ts`
- maybe thin edits to `theme.ts`
