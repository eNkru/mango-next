# Delete dead server plumbing + CLI nits

## Goal

Remove small verified-dead helpers across the server package and CLI
(one-line-to-ten-line cuts, all zero-caller).

## Confirmed facts (verified 2026-09-12)

- `sendText` + `redirect` — go/internal/server/response.go:22. Zero
  references repo-wide (all handlers use sendJSON / sendImage /
  http.Redirect directly).
- `countEntries` — 5-line `DeepEntries` wrapper,
  go/internal/server/handlers_api.go:116. Zero callers. (It is a plain
  func, not a method — keep `firstEntryID`, which IS used at
  handlers_pages.go:127,200 and in first_entry_test.go.)
- `isStaticFile` — go/internal/server/middleware.go:111-119 (9 lines).
  Zero callers.
- `PersistentPreRunE` no-op (`return nil`) on the admin user command,
  go/cmd/mango/admin.go:37, with a stale comment ("avoids repeated
  open/close for subcommands" — it does nothing). Remove the field.
- `loadTemplates` — go/cmd/mango/main.go:19, a 3-line pass-through
  (`return server.NewTemplateManager(web.Views())`) called once at
  main.go:89. Inline at the call site.

## Requirements

1. Delete `sendText`, `redirect`, `countEntries`, `isStaticFile`.
2. Remove the `PersistentPreRunE` field (and its stale comment) from the
   admin user command.
3. Inline `loadTemplates()` into main.go:89 and delete the wrapper.
4. Fix imports in touched files.

## Acceptance criteria

- [ ] `cd go && go build ./... && go test ./...` pass.
- [ ] `grep -rn "sendText\|\bredirect\b\|countEntries\|isStaticFile\|loadTemplates" go --include="*.go"` returns
  only the kept `firstEntryID` context (i.e. nothing).
- [ ] `mango admin user --help` still works; server boots
  (`make run` smoke).

## Out of scope — DO NOT DELETE

- `firstEntryID` — live (2 page handlers + tests). Keep.
- `userExists` (auth.go:185) — encapsulates "treat DB error as
  not-exists"; the name earns its keep. Keep.
