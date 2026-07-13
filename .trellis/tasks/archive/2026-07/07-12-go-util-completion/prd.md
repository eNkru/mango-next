# Complete Crystal util (proxy / validation / web leftovers)

## Goal

Close remaining gaps from `src/util/*` after signature/sort/upload: ensure HTTP
proxy env behavior for plugin downloads, document validation parity, and note
web helpers already covered by server middleware.

## Confirmed Facts

| Crystal | Go status |
|---------|-----------|
| `proxy.cr` | ❌ `http.DefaultClient` / bare Client — no ProxyFromEnvironment |
| `validation.cr` username/password | ✅ `storage.validateUsername/Password` |
| `validation.cr` validate_archive | ✅ `validateZip` in downloader + archive pkg |
| `web.cr` macros | ✅ chi middleware/auth/response helpers |
| `chapter_sort` / `numeric_sort` | ✅ `library/sort.go` |
| `util.cr` sanitize_filename | ⚠️ partial `sanitizeFilename` in downloader |

## Requirements

### R1 — Proxy (main code work)

- Use `http.ProxyFromEnvironment` (or equivalent) on plugin **Sandbox** and
  **Downloader** HTTP clients so `HTTP_PROXY` / `HTTPS_PROXY` / `NO_PROXY` work
  like Crystal `proxy.cr`.

### R2 — Validation

- No code if tests already cover; confirm parity in gap notes.

### R3 — sanitize_filename (optional polish)

- Align with Crystal `sanitize_filename` if easy; else leave and note in gap.

### R4 — web.cr

- Out of scope for new package; already middleware.

## Acceptance Criteria

- [ ] Plugin HTTP clients respect proxy env
- [ ] Tests for proxy transport or documented manual check
- [ ] `go test ./...` green

## Out of Scope

- Full Kemal web macros port
