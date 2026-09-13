# Shrink Go duplications (Image, entryFileName, chapterSorter)

## Goal

Consolidate four confirmed duplications. Unlike the delete-tasks, these
touch LIVE code — read both sides before changing, keep behavior identical,
and verify with tests + a library rescan.

## Confirmed facts (verified 2026-09-12)

1. **Duplicate Image structs.** `thumbnail.Image`
   (go/internal/thumbnail/thumbnail.go:18) and `storage.Image`
   (go/internal/storage/thumbnail.go:6) have identical shape
   (Data/Filename/Mime/Size); go/internal/library/library.go:324 does a
   manual field-by-field copy to convert. Fix: have
   `thumbnail.Generate` return `*storage.Image` directly (preferred —
   kills the struct AND the conversion), or have storage accept the
   thumbnail type. ~10 lines.
2. **Duplicate entryFileName.** Same-named func in
   go/internal/server/browse_api.go:306 (`entry library.Entry`) and
   go/internal/library/title_info.go:165 (`entry Entry` — same type).
   Fix: export one (e.g. `library.EntryFileName`), reuse in browse_api.
   (~7 lines; agent line anchors were off — use these corrected ones.)
3. **chapterSorter scaffolding.** go/internal/library/sort.go:105 —
   struct + `titles` field + constructor, but `compare` (line 116) purely
   delegates to `compareNumerically` ("For Phase 2 it delegates" = the
   phase was never built) and never reads `titles`. Fix: call
   `compareNumerically` directly at the use site; delete the struct,
   constructor, and `entryNames` if it becomes unused.
4. **Manual reversal.** `browseParents` (browse_api.go:276) uses a 3-line
   swap loop. Fix: `slices.Reverse(reverse)` (stdlib, Go ≥1.21; repo is
   Go 1.26). ~2 lines.

## Requirements

Implement the four fixes above, one at a time, building + testing between
each.

## Acceptance criteria

- [ ] `cd go && go build ./... && go test ./...` pass after each fix.
- [ ] Thumbnail generation unchanged (covers fix 1): run a library scan
  in dev and confirm thumbnails generate.
- [ ] Chapter/entry sort order unchanged (covers fixes 3–4): compare
  before/after on a multi-chapter title.
- [ ] `grep -rn "func entryFileName" go --include="*.go"` shows exactly
  one definition (covers fix 2).

## Out of scope — DO NOT TOUCH

- `ContinueRail` / `PosterRail` rail-shell dup (~35 lines) — deliberate
  MVP trade-off documented in the redesign's design.md ("revisit if a
  third wide-card rail appears"). Revisit later, not here.
- `randomStr` dup (upload vs storage, 3 lines) — acceptable Go
  duplication; a cross-package import to save 3 lines is worse.
- mount-prefix strip (auth/middleware/server) — sites differ subtly
  (HasPrefix guard); not a mechanical merge.
- Storage table-param merges (GetOrCreate*, sort-title getters/setters,
  ListMissing*) — unverified leads on live queries; promote to a new
  task only after spot-checking.
