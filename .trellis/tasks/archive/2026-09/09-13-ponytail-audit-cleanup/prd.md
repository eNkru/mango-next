# Ponytail audit cleanup — tracker

## Goal

Work through the 2026-09-12 whole-repo ponytail (over-engineering) audit one
batch at a time. Each subtask is one reviewable commit: delete verified-dead
code or shrink a confirmed duplication, then re-verify with build + tests.

**Source:** ponytail-audit (inline pattern-scan + single-agent whole-file read).
Every `delete:` below was verified against callers / route table / render
sites / template actions. Total: **~-465 lines confirmed actionable**,
-0 deps (dep list already lean).

## Subtasks (work in order 1→7)

| # | Task dir | Cut | Size |
|---|----------|-----|------|
| 1 | `09-13-ponytail-dead-handlers` | apiBook + apiLibrary (unrouted legacy) | ~145 |
| 2 | `09-13-ponytail-dead-storage` | BulkMark×2, MigrateProgressTable, CountEntries, library_cache.go | ~82 |
| 3 | `09-13-ponytail-dead-templates` | OPDS tmpls, funcMap+Lookup, io.Writer | ~76 |
| 4 | `09-13-ponytail-dead-plumbing` | sendText/redirect, countEntries, isStaticFile, PersistentPreRunE, loadTemplates | small |
| 5 | `09-13-ponytail-dead-frontend` | theme.ts funcs, AppLink exports, .mango-meta, LoadingState alias, useReaderPrefs | ~60 |
| 6 | `09-13-ponytail-shrinks` | Image structs, entryFileName, chapterSorter, slices.Reverse (live code — care) | ~30 |
| 7 | `09-13-ponytail-build-config` | docker-compose.go.yml, Makefile aliases | ~27 |

## Global acceptance

- After each subtask: `cd go && go build ./... && go test ./...`
  (frontend subtask: `npm run typecheck && npm run check && npm run build`).
- At the end: `make check && make test`, plus smoke (home, reader,
  /opds endpoints, `mango admin user --help`).
- One commit per subtask on a `chore/ponytail-cleanup` branch (or per-task
  branches if preferred); PRs reference this tracker.

## DO NOT TOUCH (refuted audit claims — verified live)

These were claimed dead by the audit agent and proven wrong. Never delete:

- `GetAllTitles` / `GetAllEntries` / `TitleRecord` / `EntryRecord` —
  called from `library_test.go`. Test-covered, not dead.
- `GetEntriesSortTitle`, `CountTitles`, `DeleteThumbnail` — called from
  `storage_test.go`. Test-covered, not dead.
- `buildLibraryPageData` + `LibraryPageData` — called from
  `hidden_pages_test.go` (3 sites); struct constructed at
  handlers_pages.go:133. Not dead.
- **`ContinueReadingItem.Percentage` (and sibling progress fields)** —
  `apiContinueReading` (handlers_api.go:498) reads `item.Percentage` at
  line 518. Deleting the field BREAKS compilation. (The getter never
  populates it, so the endpoint returns 0 — that is a correctness bug,
  out of scope here, not a deletion.)
- `userExists` (auth.go:185) — encapsulates "treat DB error as
  not-exists" for auth-proxy provisioning; the name earns its keep.
- `GithubOctocat.tsx` — named SVG component, idiomatic, keep.
- `continueReaderPath.ts` — deliberate shared URL-builder seam from the
  Continue Reading redesign; single-caller by design.
- `randomStr` dup (upload vs storage) — 3-line func; a cross-package
  import to save 3 lines is worse. Acceptable Go duplication.
- mount-prefix strip (auth/middleware/server) — the 3 sites differ
  (HasPrefix guard present/absent); not a mechanical merge.

## Deliberate duplication (tracked, revisit later — not a subtask)

- `ContinueRail` / `PosterRail` rail-shell dup (~35 lines). The redesign's
  design.md explicitly chose duplication over generalizing PosterRail
  ("Acceptable for MVP; revisit if a third wide-card rail appears").

## Unverified leads (spot-check before promoting to tasks)

- Table-param merges in storage: GetOrCreateTitleID/GetOrCreateEntryID
  (identity.go:55), Get/SetTitleSortTitle vs Get/SetEntrySortTitle
  (title.go:60), ListMissingTitles/ListMissingEntries (missing.go:119);
  mimeFromFilename dup (library vs upload). Touch live queries — verify
  before cutting.

## Out of scope

Correctness/security/perf (e.g. the Percentage-always-zero bug above —
file separately). New dependencies (we want less, not more).
