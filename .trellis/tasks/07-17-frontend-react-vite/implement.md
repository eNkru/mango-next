# React Vite frontend migration execution plan

## Parent responsibilities

- [ ] Keep the parent PRD as the source requirement set and task map.
- [ ] Ensure child tasks remain independently implementable and reviewable.
- [ ] Redirect `07-17-frontend-asset-pipeline` into the foundation child instead
      of finishing the old jQuery/LESS inventory end state.
- [ ] Review child completion against the parent cross-task acceptance criteria.

## Child order

1. `07-18-frontend-react-foundation`
   - Establish React/Vite/TypeScript, Go shell mounting, theme tokens, BaseURL
     assets, Make/Docker generation, and docs.
2. `07-18-frontend-react-missing-items`
   - Depends on the foundation child.
   - Implement real missing-items APIs and migrate `/admin/missing`.

## Validation commands

```bash
npm ci
npm run build
make check
make test
make build
docker build -t mango-react-check .
```

## Rollback points

- After foundation only: no user-facing route switch required.
- After pilot: restore Go `missing-items` template handler if React page fails.
- Do not delete legacy templates until later migrations explicitly retire them.

## Follow-up children

- login shell
- `07-18-frontend-react-browse`: home / library / title browse (created)
- reader
- users / admin settings
- subscriptions / plugin download
- retirement of remaining jQuery/Alpine/UIkit pages

## Recommended next child: reader

Plan `frontend-react-reader` as the next independently reviewed child after the
browse migration:

1. Define one reader bootstrap contract for entry identity, page count,
   dimensions, current progress, adjacent entries, exit URL, and reading
   preferences. Keep image bytes on the existing page endpoint.
2. Migrate both reader routes to `pageId=reader`, preserving direct page URLs,
   keyboard/touch navigation, right-to-left and long-strip modes, preload,
   progress saving, adjacent-entry navigation, and BaseURL behavior.
3. Reuse the shared React language/theme providers and error UI; isolate reader
   state in a reducer so URL page, loaded page, and persisted progress cannot
   drift independently.
4. Add Go contract tests for missing/corrupt entries and adjacent navigation,
   pure frontend tests for the reducer/navigation math, then browser smoke on
   desktop and mobile with one archive and one loose-image directory.
5. Keep `reader.tmpl`, `reader-error.tmpl`, and `reader.js` as route-local
   rollback assets until the child passes production build and browser smoke.

Do not combine admin settings, subscriptions, plugin download, or OPDS into
this child; they have separate authorization and data-contract boundaries.
