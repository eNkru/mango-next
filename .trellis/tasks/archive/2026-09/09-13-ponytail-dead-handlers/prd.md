# Delete dead legacy handlers apiBook + apiLibrary

## Goal

Remove two large HTTP handlers that are defined but never routed — leftovers
from the browse-API migration. Biggest single cut in the audit (~145 lines).

## Confirmed facts (verified 2026-09-12)

- `apiLibrary` = go/internal/server/handlers_api.go:59-119 (61 lines).
- `apiBook` = go/internal/server/handlers_api.go:120-203 (84 lines).
- The route table in go/internal/server/server.go wires ~60 handlers,
  including `apiBrowseLibrary` (`r.Get("/library", …)`) and `apiBrowseBook`
  (`r.Get("/book/{tid}", …)`). `apiLibrary` / `apiBook` appear **zero**
  times in server.go and zero times anywhere else (grep over all .go files
  finds only the two definitions). Not even tests reference them.

## Requirements

1. Delete `func (s *Server) apiLibrary` (handlers_api.go:59-119).
2. Delete `func (s *Server) apiBook` (handlers_api.go:120-203).
3. Remove any now-unused imports in handlers_api.go (check with
   `goimports` / compiler).

## Acceptance criteria

- [ ] `cd go && go build ./...` passes with no unused-import errors.
- [ ] `cd go && go test ./...` passes.
- [ ] `grep -rn "apiBook\|apiLibrary" go --include="*.go"` returns nothing.
- [ ] Smoke: `GET /api/library` and `GET /api/book/{tid}` still serve
  (proving the live `apiBrowse*` handlers were untouched).

## Out of scope

- `apiBrowseLibrary` / `apiBrowseBook` and all other routed handlers —
  live, keep.
- No route-table changes needed (nothing references the deleted funcs).
