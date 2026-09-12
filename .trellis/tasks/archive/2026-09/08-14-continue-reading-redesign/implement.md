# Implement — Continue Reading redesign (B + D + E)

Ordered checklist. Run validation after each milestone. Do not commit until the
final check passes.

## 1. Backend filter (D)

- [x] 1.1 In `go/internal/server/browse_api.go` `apiHome`, after computing each
      `browseEntry` via `s.browseEntry(...)`, skip the item when
      `entry.Page == -1 || entry.Page >= entry.Pages`. Keep the 8-item cap and
      order. Read `browseEntry` fields directly (they're populated by
      `makeBrowseEntry`, browse_api.go:222).
- [x] 1.2 Add a Go test in `go/internal/server/` (e.g. `browse_api_test.go`) that
      seeds progress rows: one in-progress entry, one `page = pageCount` entry,
      one `page = -1` (bulk-marked) entry, and asserts only the in-progress entry
      appears in `apiHome` `continue_reading`.
- [x] 1.3 Add/keep a regression assertion that `GET /api/library/continue_reading`
      still returns all (including finished) entries — confirming the filter is
      home-only.
- [x] 1.4 `cd go && go build ./... && go vet ./... && go test ./...`

## 2. Frontend deep-link helper (E)

- [x] 2.1 In `frontend/src/browse/`, add a `continueReaderPath(item)` helper:
      returns `reader/{title_id}/{id}` when `item.page <= 0`, else
      `reader/{title_id}/{id}/{item.page + 1}`. `encodeURIComponent` both ids.
      (Replaces the old `readerPath` in `ContinueCarousel.tsx`.)
- [x] 2.2 Typecheck: `npm run typecheck`.

## 3. Frontend rail (B)

- [x] 3.1 Create `frontend/src/browse/ContinueRail.tsx`:
      - Props: `{ items: BrowseEntry[] }`.
      - Renders `mango-browse-section` wrapper + `<h2>{t('continueReading')}</h2>`.
      - Rail shell: track ref + `canPrev`/`canNext` edge state + prev/next arrow
        buttons (mirror `PosterRail.tsx` logic, lines ~30-60).
      - Maps `items` → `ContinueCard` (cover + title + page X/Y + ProgressBar +
        Continue `AppLink` to `continueReaderPath(item)`).
      - Whole card is an `AppLink` to the same deep-link.
- [x] 3.2 Add `.mango-continue-rail*` CSS to `frontend/src/styles/shell.css`
      where the old stack block was. Reuse tokens + `mango-btn`, `mango-progress`,
      `mango-card__placeholder`. Add responsive (`max-width: 560px`) +
      `prefers-reduced-motion` rules.
- [x] 3.3 In `frontend/src/pages/HomePage.tsx`: replace
      `import { ContinueCarousel }` with `import { ContinueRail }`; swap the
      `<ContinueCarousel items={...} />` usage (line ~65) with `<ContinueRail />`.
- [x] 3.4 `npm run typecheck && npm run check` (check asserts
      `go/web/public/react/assets/main.{js,css}` exist — run `npm run build` if
      missing).

## 4. Cleanup

- [x] 4.1 Delete `frontend/src/browse/ContinueCarousel.tsx`.
- [x] 4.2 Delete the `.mango-continue-stack*` CSS block in `shell.css`
      (~lines 930-1130), including responsive + reduced-motion + comic-theme
      overrides. Grep for `mango-continue-stack` to confirm zero references remain.
- [x] 4.3 Grep repo for `ContinueCarousel` and `mango-continue-stack` → expect
      zero hits.

## 5. Final validation

- [x] 5.1 `make check` (frontend-check + go vet)
- [x] 5.2 `make test` (go test ./...)
- [x] 5.3 `npm run build` (tsc --noEmit + vite build + output check) — embeds for Go.
- [x] 5.4 Manual smoke: `make run`, read part of an entry (progress saved), reload
      home → card shows with "page X / Y"; click Continue → reader opens on saved
      page. Mark an entry fully read → it disappears from Continue on reload.
- [x] 5.5 Toggle light/dark + comic theme; confirm new card reads correctly.

## Rollback points
- After step 1: backend-only; frontend unaffected. Safe checkpoint.
- After step 3: new rail live alongside old component (old not yet deleted).
- Step 4 (deletion) is the one-way door; keep the commit isolated so `git revert`
  restores everything in one step.

## Risky files
- `go/internal/server/browse_api.go` — shared by `apiHome`, `apiBrowseBook`;
  touch only the `apiHome` continue loop.
- `frontend/src/styles/shell.css` — large file; delete only the contiguous stack
  block, don't disturb neighbors.
- `frontend/src/pages/HomePage.tsx` — only the import + one JSX line.
