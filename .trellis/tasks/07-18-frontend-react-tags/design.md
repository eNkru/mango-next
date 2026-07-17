# React tags migration design

## Routes

| Path | pageId |
|---|---|
| `/tags` | `tags-index` |
| `/tags/{tag}` | `tag-detail` |

## APIs

- `GET /api/tags` — unchanged string list for legacy consumers
- `GET /api/tags/index` — `{ tags: [{tag, count}] }`
- `GET /api/tags/{tag}/titles?show_hidden=0|1` — title cards + flags
- existing `PUT /api/admin/hidden/{tid}/{value}`

Register multi-segment tag API paths before `/tags/{tid}`.

## Frontend

- `TagsIndexPage`: load index, filter pills, link to detail
- `TagDetailPage`: load titles, search, admin show-hidden + hide/unhide

## Validation

```bash
npm run typecheck
npm run build
make check
make test
make build
```
