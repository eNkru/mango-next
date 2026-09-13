# Delete dead frontend code

## Goal

Remove verified-unused frontend exports, dead CSS, and two single-use
wrappers (~60 lines). All verified with repo-wide import/reference grep.

## Confirmed facts (verified 2026-09-12)

- `saveThemeSetting` / `saveUIStyle` / `watchSystemTheme` in
  frontend/src/lib/theme.ts (~25 lines): zero importers outside theme.ts.
  All components use `themeStore`. (theme.ts itself may become empty —
  delete the file if nothing remains, else keep live exports.)
- `hardHref` + `isModifiedClick` in frontend/src/lib/AppLink.tsx: zero
  references outside AppLink.tsx. (AppLink itself is used by 11 files —
  keep the component, delete only the two dead exports.)
- `.mango-meta` ruleset (grid/dt/dd, ~20 lines, shell.css:431): zero
  references in any tsx/ts. Delete the ruleset.
- `LoadingState` and `EmptyState` (StatePanels.tsx:13) have byte-identical
  bodies (`<div className="mango-state">{message}</div>`). Alias one to
  the other (keep both names, one implementation) — ~4 lines saved.
- `useReaderPrefs.ts` (8 lines): thin wrapper over `useReaderPrefsStore`
  (already a hook), single caller ReaderPage.tsx:39; its own comment
  admits "same API as before" (legacy shim). Inline the two selector
  calls in ReaderPage and delete the file.

## Requirements

1. Delete the three dead theme.ts functions (or the file if emptied).
2. Delete the two dead AppLink exports.
3. Delete the `.mango-meta` CSS ruleset.
4. Merge LoadingState/EmptyState bodies (keep both exported names).
5. Inline `useReaderPrefs()` into ReaderPage.tsx:39; delete
   pages/reader/useReaderPrefs.ts.

## Acceptance criteria

- [ ] `npm run typecheck && npm run check && npm run build` pass.
- [ ] `grep -rn "hardHref\|isModifiedClick\|saveThemeSetting\|saveUIStyle\|watchSystemTheme\|mango-meta\|useReaderPrefs" frontend/src`
  shows only live uses (themeStore, AppLink component, ReaderPage's
  inlined selectors).
- [ ] Smoke: theme toggle, home, reader prefs (margin/fit/mode) still work.

## Out of scope — DO NOT DELETE

- `themeStore`, `AppLink` component, `Icon`/`icons` (centralized icon
  vocab — deliberate), `GithubOctocat` (idiomatic), `continueReaderPath`
  (deliberate seam from the redesign), `readerPrefsStore`, `prefsSync`.
