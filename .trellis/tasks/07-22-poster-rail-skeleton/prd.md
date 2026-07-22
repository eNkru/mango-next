# Task PRD: Poster Rail Skeleton Loaders

## Goal

Enhance frontend UX during library loading by adding shimmer skeleton placeholders to `PosterRail.tsx`, eliminating Cumulative Layout Shift (CLS) when browsing books/manga.

## Target Area

- `frontend/src/browse/PosterRail.tsx`
- `frontend/src/browse/BrowseComponents.tsx`

## Requirements

1. Create a `PosterRailSkeleton` component displaying animated shimmer poster cards matching the dimensions of standard poster rails.
2. Render the skeleton while `loading` state is true in `PosterRail`.
3. Preserve existing responsive layout boundaries across desktop and mobile screens.

## Acceptance Criteria

- [ ] `PosterRail` displays skeleton placeholders while asynchronous data is fetching.
- [ ] No Cumulative Layout Shift (CLS) occurs when real poster cards replace the loading placeholders.
- [ ] Visual appearance conforms to existing dark/light UI styling.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
