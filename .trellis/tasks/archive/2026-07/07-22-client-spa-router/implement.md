# Implement: Client-Side SPA Router

## Checklist

1. Add `react-router-dom` (match React 19 compatible version) in root/frontend package as used by repo.
2. `BootProvider` + hook (`useBoot`) from initial `readBoot()`.
3. Wire `BrowserRouter` with basename from boot `baseUrl`.
4. Replace `App.tsx` page switch with `<Routes>` / `<Route>`.
5. Add `AppLink` (or thin wrapper) for SPA links; convert priority surfaces.
6. Migrate pages off `readBoot()` for route params (useParams / useSearchParams); keep boot for isAdmin/baseUrl.
7. Reader: route params + navigate/replace for page and entry changes; SPA exit links.
8. Typecheck + build.

## Validation

```bash
# from repo package root used for frontend
npm run typecheck
npm run build
```

Manual: shell nav no flash; deep link refresh; back/forward; open reader from title; exit reader; logout still full reload; download still full.

## Risk

- Basename + trailing slash mismatch with `baseUrl()`.
- Reader history fight between replaceState and router.
- Missed hard link → accidental full reload (OK) or wrong SPA path.

## Files likely touched

- `package.json` / lockfile
- `frontend/src/main.tsx`, `App.tsx`
- `frontend/src/lib/boot.ts` (+ new boot context / router helpers)
- `frontend/src/shell/AppShell.tsx`
- `frontend/src/browse/*`, pages listed above, reader pages
