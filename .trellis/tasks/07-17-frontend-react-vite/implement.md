# React Vite frontend migration execution plan

## Parent responsibilities

- [x] Keep the parent PRD as the source requirement set and task map.
- [x] Ensure child tasks remain independently implementable and reviewable.
- [x] Redirect `07-17-frontend-asset-pipeline` into the foundation child.
- [x] Required children completed (foundation → admin home).
- [x] Subscriptions + plugin download explicitly deferred (product: not needed).
- [ ] Parent integration review / archive.

## Child order (completed)

1. [x] `07-18-frontend-react-foundation`
2. [x] `07-18-frontend-react-missing-items`
3. [x] `07-18-frontend-react-admin-users`
4. [x] `07-18-frontend-react-tags`
5. [x] `07-18-frontend-react-login`
6. [x] `07-18-frontend-react-browse`
7. [x] `07-19-frontend-react-reader`
8. [x] `07-19-frontend-react-admin`

## Cancelled / deferred children (do not create)

- ~~`frontend-react-subscriptions`~~ — product skip
- ~~`frontend-react-plugin-download`~~ — product skip
- OPDS — keep Go XML
- `07-19-frontend-legacy-retirement` — planned: disable deferred UIs + delete dead assets

## Close-out checklist

1. Confirm PRD task map matches archived children.
2. Note unmigrated routes remain template-backed if still registered.
3. Archive parent task when user approves close-out.
4. Optional later: retire unused legacy assets in a dedicated cleanup task.

## Validation (parent integration — when archiving)

```bash
npm run typecheck
npm run build
cd go && go test ./...
```
