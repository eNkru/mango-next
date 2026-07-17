# React admin users design

## Scope

Migrate full admin user management:

- list users
- create user
- edit user
- delete user

Keep independent routes and route-level React mounting. Do not migrate other
admin pages.

## Routes and pageIds

| Browser route | Go handler result | React pageId |
|---|---|---|
| `GET /admin/user` | React shell | `user-list` |
| `GET /admin/user/edit` | React shell | `user-edit` |
| `GET /admin/user/edit?username=foo` | React shell with boot username | `user-edit` |

Deep links and BaseURL remain under Go control. List ↔ edit navigation uses
normal same-origin links (`baseUrl('admin/user')`, `baseUrl('admin/user/edit?...')`).

## JSON contracts

All endpoints stay admin-authenticated under BaseURL.

### List

`GET /api/admin/users`

```json
{
  "success": true,
  "users": [
    { "username": "admin", "admin": true }
  ],
  "current_username": "admin"
}
```

### Create

`POST /api/admin/users`

```json
{ "username": "alice", "password": "secret", "admin": false }
```

Response: `{ "success": true }` or `{ "success": false, "error": "..." }`.

### Update

`PUT /api/admin/users/{original_username}`

```json
{ "username": "alice2", "password": "", "admin": true }
```

Empty password means keep existing password. Backend still enforces last-admin
and validation rules.

### Delete

Existing:

`DELETE /api/admin/user/delete/{username}`

Keep as-is for compatibility; React list uses this endpoint.

## Frontend

- `UserListPage`
  - load users
  - show admin badges
  - hide delete for current user
  - confirm delete
  - link to create/edit
- `UserEditPage`
  - create mode when no username boot/query
  - edit mode when username present
  - optional password change
  - submit via JSON
  - surface validation/backend errors without false success

Register both pageIds in `App.tsx`.

## Go changes

- Switch `handleUserList` and `handleUserEdit` to React shell render helpers.
- Leave old form POST handlers unused by React, but keep temporarily for
  rollback or remove only if no remaining callers.
- Add list/create/update JSON handlers and routes under `/api/admin`.
- Prefer focused Go tests for API success/error paths.

## Rollback

- Restore template handlers for `/admin/user` and `/admin/user/edit`.
- Keep React foundation and missing-items migration intact.

## Validation

```bash
npm run typecheck
npm run build
make check
make test
make build
```
