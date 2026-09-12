# Redesign Continue Reading section (rail + finish filter + page deep-link)

## Goal

Make the Home page "Continue reading" section fast to scan and precise to
resume. Replace the cover-stack carousel with a horizontal wide-card rail,
stop surfacing titles whose saved entry is already finished, and deep-link the
resume target to the saved page so the user lands exactly where they left off.

## Background / Confirmed facts (from repo evidence)

- Home data comes from `GET /api/home` (`go/internal/server/browse_api.go:40`),
  which builds `continue_reading` from `Storage.GetContinueReading(username)`
  (`go/internal/storage/progress.go:155`), capped to 8 entries, each mapped via
  `browseEntry` (cover, pages, page, progress, modified_at).
- `GetContinueReading` (`progress.go:155`) orders by `updated_at DESC`, LIMIT 20,
  and only excludes `t.unavailable = 0 AND t.hidden = 0`. It does **not** filter
  finished entries. `page = -1` marks a bulk-read entry
  (`BulkMarkRead`, `progress.go:99`); `page = pageCount` marks a title fully read
  (`BulkMarkTitleRead`, `progress.go:122-130`).
- `BrowseEntry` DTO already carries `page`, `pages`, `progress`, `cover_url`
  (`frontend/src/lib/browse.ts:6`), so the frontend has all data for the wide
  card and the deep-link already.
- The reader route already supports an optional page segment:
  `/reader/:tid/:eid/:page` (`frontend/src/App.tsx:97-98`), and `ReaderRoute`
  passes a validated `initialPage` to `ReaderPage` (`App.tsx:57-67`). Storage
  `page` is 0-based; the reader is 1-based, so the link must emit `page + 1`.
- Continue items are also exposed by the standalone endpoint
  `GET /api/library/continue_reading` (`handlers_api.go:498`), which returns
  `{entry_id, title_id, percentage}` only (no page/cover).

## Requirements

### D — Backend: filter finished entries from the home continue list
- In `apiHome` (`browse_api.go:40`), when building `continueItems`, skip any
  entry whose saved `page` indicates finished: `page == -1` OR `page >= pageCount`.
  Use the `browseEntry.Page` / `browseEntry.Pages` already computed by
  `makeBrowseEntry` (the filter belongs in the apiHome loop, not in storage).
- Apply the filter in **apiHome only**. Do not change `GetContinueReading` or the
  standalone `/api/library/continue_reading` endpoint contract.
- Keep the existing 8-entry cap and `updated_at DESC` ordering.

### B — Frontend: wide-card horizontal rail
- Replace `ContinueCarousel` with a new `ContinueRail` component rendering
  horizontal-scrolling wide cards: cover thumbnail (left) + title + "page X / Y"
  + chunky progress bar + Continue button (right).
- Reuse the `PosterRail` scroll/arrow pattern (edge detection, arrows, keyboard
  where applicable). Because `PosterRail` is typed to `BrowseTitle`, not `BrowseEntry`.

### E — Frontend: deep-link the resume target
- Generate the reader link using the same logic as `ContinueCarousel`: when
  `page == -1` or `page >= pages` omit the page segment; otherwise emit
  `reader/{title_id}/{id}/{page + 1}`.
- Use the helper `continueReaderPath(item: BrowseEntry)` (new or reused).

## Acceptance Criteria
- Only in-progress entries appear in the rail (filtered in backend).
- Clicking Continue lands exactly on the saved page (or last read page).
- Horizontal rail with arrows, responsive.
- No regression on `/api/library/continue_reading`.
- Clean deletion of old `ContinueCarousel.tsx` and CSS.

**PRD signed off.**
