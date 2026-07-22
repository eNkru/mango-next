# Design: I18n + Boot → Zustand

## Architecture

```
main.tsx
  BrowserRouter basename={routerBasename()}
    App

// no I18nProvider / BootProvider

i18nStore     # language + setLanguage + lang side effects
bootStore     # baseUrl, isAdmin, version, pageName (init once)
themeStore / readerPrefsStore  # unchanged
prefsSync     # also rehydrates language on storage event
```

## i18nStore

| Piece | Behavior |
|-------|----------|
| `language: Language` | from `mango-language` |
| `setLanguage(lang)` | persist + set state + apply `document.documentElement.lang` |
| `rehydrateFromStorage()` | re-read key + apply lang attribute |
| `t` | **not stored** — pure function using `messages[language]`; `useI18n` composes `t` from current language |

```ts
// useI18n adapter
export function useI18n() {
  const language = useI18nStore((s) => s.language);
  const setLanguage = useI18nStore((s) => s.setLanguage);
  const t = useCallback(
    (key: MessageKey, vars?) => formatMessage(messages[language][key] ?? messages['zh-cn'][key], vars),
    [language],
  );
  return { language, setLanguage, t };
}
```

Optional: export standalone `t(key, vars)` reading `useI18nStore.getState().language` for non-React code if needed later (not required).

Init side effect on first import or `main.tsx`: apply `document.documentElement.lang` from stored language.

## bootStore

| Field | Source |
|-------|--------|
| `baseUrl` | `normalizeBaseUrl(readBoot().baseUrl)` once |
| `isAdmin` | once |
| `version` | once |
| `pageName` | once |

No setters required for MVP (session fixed until full reload / re-login which reloads shell).

```ts
export function useBoot() {
  return useBootStore(); // or selectors for fields
}
```

`routerBasename()` / `appPath()`: keep as pure helpers; basename can use `useBootStore.getState().baseUrl` or `readBoot()` — prefer store after init so single source.

**Init order in main.tsx:**

1. Create/import boot store (reads `#mango-boot` once).
2. `routerBasename()` for BrowserRouter.
3. Start prefsSync (theme + reader + language).

## Multi-tab

Extend `prefsSync.ts`:

```ts
if (event.key === 'mango-language') {
  useI18nStore.getState().rehydrateFromStorage();
}
```

## Compatibility

- Message keys and catalog text unchanged.
- `LanguageSelect` keeps calling `useI18n().setLanguage`.
- SPA: language change re-renders subscribers; boot fields stable across client routes.

## Trade-offs

| Choice | Why |
|--------|-----|
| `t` outside store state | Avoid re-creating huge message trees in state; language selector is enough |
| Boot store no multi-tab | Boot is not localStorage; full reload gets new shell |
| Keep hook names | Zero mass rename across pages |

## Rollback

Restore providers and context implementations; drop store modules.
