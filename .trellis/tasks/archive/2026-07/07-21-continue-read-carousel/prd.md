# Continue-read carousel: large first + minor rest

## Goal

Redesign the Home **Continue reading** section as a **focus carousel**: the **active** title is large; adjacent titles are smaller. Swiping/selecting changes which book is featured (not a fixed `items[0]`-only hero).

## User value

Users should immediately see “what to open next” on the focused card (cover + progress + continue), while still scanning other in-progress titles and promoting any of them to the large slot without a dense text list.

## Confirmed facts (from codebase)

- Source: `frontend/src/pages/HomePage.tsx` → `ContinueSection`
- Current layout (also in `.trellis/spec/frontend/ui-theme-layout.md`):
  - `items[0]` → `.mango-continue-hero` (cover + title + page text + thicker progress + Continue CTA)
  - `items[1..]` → `.mango-continue-list` / `.mango-continue-row` (thumb + title + thin progress); whole row links to reader
  - When rest length > 3, `LIST_PREVIEW = 3` + show more / less
- Data: `GET /api/home` → `continue_reading: BrowseEntry[]` (no new API needed for a pure UI redesign)
- No UI component library; custom `mango-*` CSS in `frontend/src/styles/shell.css`
- Start reading / recently added use horizontal `.mango-poster-rail` + `PosterCard` (continue intentionally different)
- Comic theme: thick border / sharp corners on continue surfaces

## Decisions

| Topic | Decision |
|-------|----------|
| Large slot model | **B — true focus carousel**: active index is large; changing focus promotes another item |
| Initial active index | **0** (API order / most relevant continue item) |
| Mobile | Full-width active card + left/right peek (~8–12%) + native horizontal scroll-snap; all items in track; no show more |
| Inactive click | **Promote only** (focus/snap); reader only via **Continue** on the active card |
| Nav chrome | **Prev/next arrows only** (no dots); hide when `length <= 1`; show from ~md/desktop, hide on narrow mobile (swipe + peek) |
| Card density | **Active**: cover + title + page text + progress + Continue. **Inactive**: cover only |
| Library | No third-party carousel library |
| Single item | One large active card; no arrows; no empty track chrome |
| Keyboard | When carousel focused: Left/Right (or Home/End optional) moves active index; Tab reaches Continue |
| Reduced motion | Prefer instant snap / minimal scale transition under `prefers-reduced-motion` |
| Theme | Flat + comic tokens; comic sharp corners / thick borders on carousel surfaces |

## Requirements

1. Replace Home continue **hero + list + show more** with a focus carousel driven by `continue_reading`.
2. Active slide is visually dominant; neighbors are smaller / peek; changing active does not navigate away.
3. **Continue** on active slide opens reader for that entry (`reader/{title_id}/{id}`).
4. Click/tap on inactive slide only promotes focus (scroll-snap to index).
5. Mobile: full-width focus + side peek + scroll-snap; desktop: multi-visible focus scale + arrows when >1 item.
6. All continue items participate in the track (no `LIST_PREVIEW` / show more).
7. Empty `continue_reading` → section not rendered (same as today).
8. No new backend fields required.

## Acceptance Criteria

- [ ] Home continue section uses focus carousel; old vertical list + show more are gone
- [ ] Default active index is `0` after load
- [ ] Active card shows cover, title, page info, progress bar, and Continue CTA to correct reader URL
- [ ] Inactive cards show cover only; activating them does not navigate to reader
- [ ] Swipe / scroll-snap on mobile changes active item; side peeks visible when neighbors exist
- [ ] Desktop (or ≥~768px) shows prev/next arrows when `items.length > 1`; first/last disable or no-op at ends
- [ ] Single-item list: large card only, no arrows
- [ ] Flat and comic themes still look correct (radius/border/shadow patterns)
- [ ] `prefers-reduced-motion: reduce` avoids distracting scale/scroll animation
- [ ] Spec `ui-theme-layout.md` continue-reading section updated to match new contract
- [ ] `npm run typecheck` / build outputs still pass

## Out of scope

- Backend / `GetContinueReading` ranking or API shape changes
- Start-reading / recently-added rails
- Third-party carousel library
- Autoplay

## Open questions

- None blocking planning (defaults above for keyboard / reduced-motion / single-item). Confirm in design if arrow always-visible on mobile is preferred later.
