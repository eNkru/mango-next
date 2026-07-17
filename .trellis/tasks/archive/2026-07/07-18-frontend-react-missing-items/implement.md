# Missing items React pilot implementation plan

## 1. Storage and API

- [x] Add list/delete missing title/entry methods on Storage.
- [x] Implement real admin missing handlers (no empty stubs).
- [x] Add focused storage tests.

## 2. React page

- [x] Add MissingItemsPage with load/delete/bulk-delete/confirm/empty/error.
- [x] Register `pageId: missing-items` in App.
- [x] Switch Go `/admin/missing` to React shell.

## 3. Verify

```bash
npm run typecheck
npm run build
make check
make test
make build
```
