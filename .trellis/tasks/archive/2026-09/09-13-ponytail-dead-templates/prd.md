# Delete dead template layer (OPDS tmpls, funcMap, Lookup)

## Goal

Remove unrendered templates and unused template plumbing (~76 lines). The Go
HTML shell now serves exactly one template; the rest is dead weight.

## Confirmed facts (verified 2026-09-12)

- go/web/views/ contains 3 templates: opds/index.tmpl (21 lines),
  opds/title.tmpl (36 lines), react-shell.tmpl (live).
- The ONLY `renderPage` call site is handlers_pages.go:312, rendering
  `"views/react-shell"`. OPDS handlers (handleOPDSIndex/handleOPDSTitle)
  build XML inline with fmt.Sprintf (see handlers_pages.go:332+). The two
  OPDS .tmpl files are never rendered (57 lines dead).
- `TemplateManager` funcMap (slice/seq/add/sub/html/url/js, web.go:60):
  precise grep for template actions (`{{…seq/add/sub/slice…}}`) across all
  .tmpl files returns nothing. Truly unused (~15 lines).
- `TemplateManager.Lookup` (web.go:109): zero callers (~3 lines).
- `TemplateManager.Render` (web.go:108) takes
  `w interface{ Write([]byte) (int, error) }` — a hand-rolled `io.Writer`.
  Its sole call site (server.go:236) passes `http.ResponseWriter`, which
  is an `io.Writer`. Use `io.Writer` (stdlib tag).

## Requirements

1. Delete go/web/views/opds/index.tmpl and go/web/views/opds/title.tmpl.
2. Delete the funcMap block and pass no Funcs (or delete the FuncMap
   construction entirely); delete the `Lookup` method.
3. Change `Render` signature to `w io.Writer` (add `"io"` import if
   missing).

## Acceptance criteria

- [ ] `cd go && go build ./... && go test ./...` pass.
- [ ] Home page renders (react-shell path untouched).
- [ ] `GET /opds` and `GET /opds/book/{title_id}` still serve valid XML
  (proving the inline-XML path was untouched).
- [ ] `ls go/web/views/opds/` is empty or the dir is removed (keep
  .gitkeep only if the embed needs the dir).

## Out of scope

- react-shell.tmpl and `ReactShellData` — live, keep.
- OPDS inline-XML handlers — live, keep.
- `TemplateManager` struct + `NewTemplateManager` walk/parse — live, keep.
