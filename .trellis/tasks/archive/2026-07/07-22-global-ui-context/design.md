# Design: Global UI prefs with Zustand

## Architecture

```
main.tsx
  BootProvider          # unchanged
    BrowserRouter
      I18nProvider      # unchanged
        App

themeStore (zustand)    # theme + uiStyle + localStorage + applyHtmlTheme
readerPrefsStore        # ReaderPrefs + mango.reader.* keys
```

No ThemeProvider required if components use `useThemeStore(selector)`.

## themeStore

| Field / action | Behavior |
|----------------|----------|
| `theme: ThemeSetting` | load from `theme` key |
| `uiStyle: UIStyle` | load from `ui-style` key |
| `setTheme(t)` | state + `localStorage` + `applyHtmlTheme` |
| `setUIStyle(s)` | same |
| `rehydrateFromStorage()` | re-read keys (for storage event) |
| init | `watchSystemTheme` once (module or AppShell effect) when theme===system |

Keys unchanged: `theme`, `ui-style`.

## readerPrefsStore

| Field / action | Behavior |
|----------------|----------|
| `prefs: ReaderPrefs` | full object matching today |
| `setPrefs(patch)` | merge + write all `mango.reader.*` keys |
| `rehydrateFromStorage()` | reload prefs from keys |

Keep `useReaderPrefs()` as:

```ts
export function useReaderPrefs() {
  const prefs = useReaderPrefsStore((s) => s.prefs);
  const setPrefs = useReaderPrefsStore((s) => s.setPrefs);
  return { prefs, setPrefs };
}
```

## Multi-tab

```ts
window.addEventListener('storage', (e) => {
  if (!e.key) return;
  if (e.key === 'theme' || e.key === 'ui-style') themeStore.getState().rehydrateFromStorage();
  if (e.key?.startsWith('mango.reader.')) readerPrefsStore.getState().rehydrateFromStorage();
});
```

Register once at module load (or from `main.tsx`). Own-tab writes do not fire `storage` — no loop.

## AppShell

Replace local `useState(loadThemeSetting)` with:

```ts
const theme = useThemeStore((s) => s.theme);
const uiStyle = useThemeStore((s) => s.uiStyle);
const setTheme = useThemeStore((s) => s.setTheme);
const setUIStyle = useThemeStore((s) => s.setUIStyle);
```

## Compatibility

- Same keys → no data migration.
- `theme.ts` helpers remain for FOUC / pure apply if useful.
- I18n/Boot untouched.

## Trade-offs

| Choice | Why |
|--------|-----|
| Split stores | Clear domain boundary (shell theme vs reader) |
| Zustand | User-approved; lightweight first state lib |
| Keep useReaderPrefs API | Minimal reader call-site churn |

## Rollback

Remove zustand dep; restore AppShell useState + original useReaderPrefs.
