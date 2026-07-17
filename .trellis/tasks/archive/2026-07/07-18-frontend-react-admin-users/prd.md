# React migrate admin users

## Goal

Migrate the full admin user management UI from Go templates + jQuery to the
React + Vite shell, while keeping Go as the only backend, authenticator, and
embed host.

## Confirmed Facts

- Parent migration strategy remains route-level coexistence: only migrated admin
  routes render React; other pages keep Go templates.
- Foundation already provides React shell, BaseURL boot, dual-theme tokens,
  alerts, confirm dialog, and `apiFetch`.
- Current routes:
  - `GET /admin/user` list page (`user.tmpl` + `user.js`)
  - `GET /admin/user/edit` create/edit page (`user-edit.tmpl` + `user-edit.js`)
  - `POST /admin/user/edit` create user
  - `POST /admin/user/edit/{original_username}` update user
  - `DELETE /api/admin/user/delete/{username}` delete user JSON API
- List page is mostly server-rendered; only delete uses JSON + jQuery.
- Create/edit is classic HTML form POST with redirect and query-string errors.
- Storage already supports `ListUsers`, `NewUser`, `UpdateUser`, `DeleteUser`,
  username/password validation, and last-admin protection.
- Delete already refuses deleting the current logged-in user in the API.
- Work continues on branch `feat/frontend-react-admin-users` forked from
  `origin/feat/frontend-react-vite`.

## Requirements

- Keep Go as the only long-running backend and deployable binary.
- Use the existing React shell mount pattern (`react-shell.tmpl` + pageId).
- Preserve admin auth middleware and non-root BaseURL support.
- Preserve dual-theme compatible admin subpage styling via existing React tokens.
- Scope is the full admin user management flow: list, create, edit, and delete.
- Keep independent routes:
  - list: `/admin/user`
  - create: `/admin/user/edit`
  - edit: `/admin/user/edit?username=...` or equivalent React-mounted edit route
- Add JSON contracts for list/create/update so React is not dependent on form
  posts and redirect query errors.
- Reuse existing storage rules for username/password validation, last-admin
  protection, optional password updates, and self-delete protection.
- Leave unmigrated routes on Go templates.
- Do not rewrite unrelated admin pages in this task.

## Acceptance Criteria

- [x] `/admin/user` is served by the React shell and lists users from a real
      JSON API.
- [x] Create user works through a React page mounted on the edit route.
- [x] Edit user supports rename, admin flag, and optional password change.
- [x] Delete user works from the list with confirmation and refreshes the list.
- [x] Current-user self-delete remains blocked.
- [x] Last-admin demotion/deletion remains blocked by backend rules.
- [x] Root and non-root BaseURL work.
- [x] Unrelated admin pages remain on Go templates.
- [x] Route-level rollback remains possible without removing the React
      foundation.

## Out of Scope

- Migrating login, library, reader, tags, subscriptions, or plugin download.
- Full SPA client-side ownership of all admin routes.
- Password-policy redesign beyond existing storage rules.
- Modal/drawer-only edit UX replacing the edit route.
- Cleaning remaining legacy jQuery/LESS assets for unmigrated pages.
