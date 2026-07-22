# Implement: I18n + Boot Zustand migration

## Checklist

1. Add `bootStore.ts` — init from `readBoot()`, export `useBootStore` + thin `useBoot` (move from bootContext or re-export).
2. Keep `routerBasename` / `appPath` working (from boot store or boot.ts helpers).
3. Refactor `i18n.tsx` — replace provider with `i18nStore`; keep `messages`, `Language`, `MessageKey`, `formatMessage`; `useI18n` as adapter.
4. Apply `document.documentElement.lang` in `setLanguage` / rehydrate / bootstrap.
5. Extend `prefsSync` for `mango-language`.
6. Simplify `main.tsx` — drop providers; ensure boot store init before basename.
7. Delete or gut unused `BootProvider` / `I18nProvider` / context objects.
8. `npm run typecheck` && `npm run build`.

## Validation

```bash
npm run typecheck
npm run build
```

Manual: language switch updates UI + html lang; second tab syncs language; boot isAdmin still gates admin nav; deep link still works.

## Risk

- Init order: basename before boot store hydrated → use sync module-level init.
- Components calling `useI18n` outside React tree — only via hooks (OK today).

## Files

- `frontend/src/lib/bootStore.ts` (new) and/or rewrite `bootContext.tsx`
- `frontend/src/lib/i18n.tsx`
- `frontend/src/lib/prefsSync.ts`
- `frontend/src/main.tsx`
