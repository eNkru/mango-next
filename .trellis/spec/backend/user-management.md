# User Management

## Scenario: Admin Removal Invariants

### 1. Scope / Trigger

- Trigger: Any change to `Storage#delete_user`, `Storage#update_user`,
  `/api/admin/user/delete/:username`, or admin user edit/delete UI behavior.
- Goal: The application must always retain at least one admin user, and the web
  API must not let the authenticated admin delete their own account.

### 2. Signatures

- `Storage#delete_user(username : String)`
- `Storage#update_user(original_username : String, username : String, password : String, admin : Bool)`
- `DELETE /api/admin/user/delete/:username`

### 3. Contracts

- `Storage#delete_user` deletes non-admin users.
- `Storage#delete_user` deletes an admin only when another admin remains.
- `Storage#update_user` may change username/password/admin status.
- `Storage#update_user` must reject demoting the only admin.
- `DELETE /api/admin/user/delete/:username` compares `:username` with
  `get_username env` and rejects deleting the current user before calling
  storage.
- Existing JSON error shape is preserved for the delete API:
  `{"success": false, "error": "<message>"}`.

### 4. Validation & Error Matrix

- Delete current web user -> `"Cannot delete the current user"`.
- Delete the only admin through storage/API -> `"Cannot remove the last admin user"`.
- Demote the only admin through storage/form -> `"Cannot remove the last admin user"`.
- Delete/demote one admin while another admin exists -> allowed.
- Delete/demote a non-admin -> allowed if other existing validation passes.

### 5. Good/Base/Bad Cases

- Good: Create `admin2`, delete `admin2`, keep original `admin`.
- Good: Create `admin2`, update `admin2` with `admin = false`, keep original
  `admin`.
- Base: Delete a non-admin user.
- Bad: Delete `admin` when it is the only admin.
- Bad: Update `admin` with `admin = false` when it is the only admin.

### 6. Tests Required

- Storage spec for deleting a non-admin user.
- Storage spec rejecting deletion of the last admin and asserting the admin row
  still exists.
- Storage spec rejecting demotion of the last admin and asserting the admin flag
  remains true.
- Storage spec allowing admin deletion/demotion when another admin remains.
- Route-level specs should cover the current-user delete guard if route test
  infrastructure is added.

### 7. Wrong vs Correct

#### Wrong

```crystal
def delete_user(username)
  db.exec "delete from users where username = (?)", username
end
```

#### Correct

```crystal
def delete_user(username)
  db.transaction do |tran|
    conn = tran.connection
    ensure_admin_remains conn, username
    conn.exec "delete from users where username = (?)", username
  end
end
```
