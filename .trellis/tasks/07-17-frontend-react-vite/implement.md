# React Vite frontend migration execution plan

## Parent responsibilities

- [x] Keep the parent PRD as the source requirement set and task map.
- [x] Ensure child tasks remain independently implementable and reviewable.
- [x] Redirect `07-17-frontend-asset-pipeline` into the foundation child.
- [ ] Review child completion against parent cross-task acceptance criteria.
- [ ] Create follow-up children only after consent; do not implement on parent.
- [ ] Archive parent when remaining scope is deferred or completed.

## Child order (completed)

1. [x] `07-18-frontend-react-foundation`
2. [x] `07-18-frontend-react-missing-items`
3. [x] `07-18-frontend-react-admin-users`
4. [x] `07-18-frontend-react-tags`
5. [x] `07-18-frontend-react-login`
6. [x] `07-18-frontend-react-browse`
7. [x] `07-19-frontend-react-reader`

## Next child order (proposed)

1. `frontend-react-admin` (P1) — `/admin` ops home
2. `frontend-react-subscriptions` (P2) — subscription manager
3. `frontend-react-plugin-download` (P2) — plugin browser download UI
4. `frontend-legacy-retirement` (P3) — remove unused legacy assets after smoke

OPDS stays Go XML unless a separate product decision says otherwise.

## Validation commands (parent integration)

```bash
npm ci
npm run build
npm run typecheck
make check
make test
make build
```

## Rollback points

- Per-child route-local: restore that route’s template handler.
- Do not delete legacy templates until `frontend-legacy-retirement`.

## Recommended next child: admin home

Plan `frontend-react-admin` as the next independently reviewed child:

1. Inventory `/admin` template + related admin APIs (scan, thumbnails progress,
   generate thumbnails, any settings surface still on the page).
2. Define bootstrap or thin JSON contracts needed for first paint; reuse existing
   admin APIs where present.
3. React page under shared AppShell + admin nav; loading/error/empty for jobs.
4. Focused Go tests for any new/changed contracts; frontend typecheck/build.
5. Keep `admin.tmpl` (and related scripts) for rollback until smoke passes.

Do not fold subscriptions or plugin download into this child — separate auth
surfaces and queue/plugin boundaries.
