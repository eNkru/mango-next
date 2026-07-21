# Implement: Continue-read focus carousel

## Checklist

1. [x] Replace `ContinueSection` with focus carousel (component in `HomePage.tsx` or `browse/ContinueCarousel.tsx`)
2. [x] Wire `activeIndex`, scroll-snap track, inactive promote, active Continue link
3. [x] Desktop prev/next arrows (hide ≤1 item and on narrow viewports)
4. [x] CSS in `shell.css`: carousel track, peek, active/inactive scale, theme + reduced-motion; remove obsolete continue-list/more rules
5. [x] Update `.trellis/spec/frontend/ui-theme-layout.md` continue-reading section
6. [ ] Manual pass: 0 / 1 / many items; mobile swipe; desktop arrows; flat + comic
7. [x] `npm run typecheck` and `npm run build` (or `make frontend-check`)

## Validation

```bash
npm run typecheck
npm run build
# manual: npm run server + npm run dev — Home continue section
```

## Risky files

- `frontend/src/pages/HomePage.tsx`
- `frontend/src/styles/shell.css` (large; careful not to break poster-rail / other home sections)
- `.trellis/spec/frontend/ui-theme-layout.md`

## Rollback

Git revert the feature branch; no backend changes.

## Before `task.py start`

- User review of `prd.md` + `design.md` + this file
- Curate `implement.jsonl` / `check.jsonl` if dispatching sub-agents
