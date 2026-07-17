# React admin users implementation plan

## 1. API and handlers

- [x] Add `GET /api/admin/users`
- [x] Add `POST /api/admin/users`
- [x] Add `PUT /api/admin/users/{original_username}`
- [x] Keep existing delete endpoint
- [x] Add focused Go tests for list/create/update/delete guards

## 2. React pages

- [x] Add `UserListPage`
- [x] Add `UserEditPage`
- [x] Register `user-list` and `user-edit` pageIds
- [x] Reuse shell primitives for loading/empty/error/confirm/alerts

## 3. Route switch

- [x] Mount React shell for `/admin/user`
- [x] Mount React shell for `/admin/user/edit`
- [x] Pass create/edit username through boot JSON and/or query parsing

## 4. Verify

```bash
npm run typecheck
npm run build
make check
make test
make build
```

Manual:

- list users
- create user
- edit username/admin/password
- delete other user
- confirm self-delete blocked
- BaseURL still works if configured
