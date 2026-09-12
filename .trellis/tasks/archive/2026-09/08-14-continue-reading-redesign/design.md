# Design — Continue Reading redesign (B + D + E)

## Architecture / boundaries

Three independent, verifiable changes:

1. **Backend filter (D)** — `go/internal/server/browse_api.go`, `apiHome` loop.
2. **Frontend rail (B)** — new `frontend/src/browse/ContinueRail.tsx` + CSS in
   `frontend/src/styles/shell.css`; `HomePage.tsx` wiring swap.
3. **Frontend deep-link (E)** — URL builder in `ContinueRail` (or a shared
   `frontend/src/browse` helper); no reader change.

No DB migration, no storage query change, no new route.

## Data flow & contracts

```
GetContinueReading(username)   → []ContinueReadingItem (progress.go:155)
  (unchanged; still returns finished entries)
        │
        ▼
apiHome (browse_api.go:40)
  for each item: browseEntry(...) computes page + pages (makeBrowseEntry:222)
  FILTER: skip when page == -1 || page >= pages      ← (D) new
  cap 8, map to browseEntry JSON (page, pages, progress, cover_url)
        │
        ▼
GET /api/home → continue_reading: BrowseEntry[]   (frontend/src/lib/browse.ts:6)
        │
        ▼
ContinueRail (new)
  wide card: cover | title + "page X / Y" + ProgressBar + Continue btn
  Continue link: reader/{title_id}/{id}/{page+1}     ← (E) new
        │
        ▼
Route /reader/:tid/:eid/:page (App.tsx:98) → ReaderRoute → initialPage
  (existing; no change)
```

### Contract notes
- `BrowseEntry.page` is **0-based** (from storage `progress.page`). Reader route
  `:page` is **1-based**. Deep-link emits `page + 1`; clamp to omit segment when
  `page <= 0`.
- `page == -1` = bulk-marked-read entry (`BulkMarkRead`). `page >= pages` =
  fully read entry (incl. `BulkMarkTitleRead` setting `page = pageCount`). Both
  are "finished" for the filter.
- The standalone `GET /api/library/continue_reading` endpoint
  (`handlers_api.go:498`) calls `GetContinueReading` directly and returns only
  `{entry_id, title_id, percentage}` — **untouched** by this task.

## Component design — ContinueRail

Reuse `PosterRail`'s scroll-shell pattern (edge detection + arrows), but
`PosterRail` is typed to `BrowseTitle`. Two options:

- **Chosen:** Add a `ContinueRail` that owns a thin copy of the rail shell
  (track + arrows + `canPrev`/`canNext` edge state) and renders `ContinueCard`
  children. Keeps `PosterRail` stable (its `BrowseTitle` API is shared with
  start-reading / recently-added) and avoids a risky generic refactor.

### `ContinueCard` layout (wide card)
```
┌─────────────────────────────────────┐
│ ┌────┐  Title Name              (h3)│
│ │cover│  page 12 / 24                │
│ │    │  ▓▓▓▓▓▓▓░░░░░  50%  (progress)│
│ └────┘  [ Continue ▶ ]              │
└─────────────────────────────────────┘
```
- Cover: `item.cover_url` or `mango-card__placeholder`.
- Meta grid: `grid-template-columns: cover-w minmax(0,1fr)`.
- Page text: `item.page > 0 ? \`${page} / ${pages} ${t('page')}\` : \`${pages} ${t('page')}\`` (same rule as the old card).
- Progress: `<ProgressBar value={item.progress} />` from `BrowseComponents`.
- Continue: `AppLink` to the deep-link; primary button with `icons.continue`.
- Whole card clickable → same deep-link (cover + title). Continue button is the
  explicit affordance; both navigate to the same place.

### CSS
- New `.mango-continue-rail*` classes in `shell.css`, placed where the old
  `.mango-continue-stack*` block was. Reuse:
  - `mango-browse-section` wrapper + `mango-poster-rail-shell` track where
    possible.
  - tokens: `--mango-bg-surface`, `--mango-border`, `--mango-radius`,
    `--mango-accent`, `--mango-text-muted`.
  - `mango-btn mango-btn--primary`, `mango-progress`, `mango-card__placeholder`.
- Responsive: card height clamp (mirror old `--stack-card-h` approach), shrink
  meta padding on `max-width: 560px`. `prefers-reduced-motion` disables scroll
  snap transitions only.
- Comic/flat theme: the new card uses `--mango-*` tokens, so it inherits the
  dual-shell theme like other cards; add `html.comic-theme .mango-continue-rail__card`
  overrides only if tokens don't cover it.

## Compatibility & migration
- No DB schema change; no migration version bump.
- `GetContinueReading` and `/api/library/continue_reading` unchanged → any
  legacy/OPDS consumer unaffected.
- Removing `ContinueCarousel.tsx`: confirm no other import site (only
  `HomePage.tsx:5`). Removing CSS: confirm no other selector references
  `.mango-continue-stack*` (only the removed component used them).

## Trade-offs
- **Duplicate rail shell vs generalize PosterRail.** Chose duplicate to avoid
  touching the shared `BrowseTitle` rail used by two other sections. Cost: two
  rail implementations. Acceptable for MVP; revisit if a third wide-card rail appears.
- **Filter in handler vs storage.** Handler-only keeps the storage query and the
  standalone endpoint stable. Cost: the standalone endpoint still shows finished
  titles — explicitly accepted (out of scope).
- **Deep-link page+1** assumes storage page is 0-based reader page 0 = "not
  started". When `page == 0` we omit the segment (reader defaults to 1).

## Rollback
- Revert is a single PR/commit: restore `ContinueCarousel.tsx`, restore the CSS
  block, revert `HomePage.tsx` import, revert the `apiHome` filter line. No data
  migration to undo.

## Validation
- `cd go && go test ./...` (add/extend a test asserting finished entries are
  excluded from `apiHome` continue list; keep standalone endpoint test unchanged).
- `npm run typecheck && npm run check && make check`.
- Manual: read part of an entry, mark it fully read, reload home — it should
  disappear from Continue; the reader deep-link should open on the saved page.
