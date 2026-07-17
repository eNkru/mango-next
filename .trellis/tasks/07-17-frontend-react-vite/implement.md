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

## Follow-up children (not created yet)

- login shell
- home / library / title browse
- reader
- users / admin settings
- subscriptions / plugin download
- retirement of remaining jQuery/Alpine/UIkit pages
