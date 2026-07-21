# Design: Continue-read focus carousel

## Architecture

Frontend-only redesign of Home `ContinueSection`.

```
GET /api/home → continue_reading: BrowseEntry[]
                ↓
        ContinueCarousel (new or rewritten ContinueSection)
          - activeIndex state (default 0)
          - horizontal track + scroll-snap
          - optional prev/next controls (md+)
          - active card: full meta + Continue <a>
          - inactive: cover button/div → setActive(i)
```

No API or ranking changes. No new npm deps.

## Component boundary

| Piece | Responsibility |
|-------|----------------|
| `HomePage` | Load data; pass `continue_reading` into carousel when non-empty |
| `ContinueCarousel` (in `HomePage.tsx` or `browse/ContinueCarousel.tsx`) | Track, active index, snap sync, arrows, cards |
| `shell.css` | Layout, peek, scale, theme overrides; remove/replace old continue list styles |
| `ui-theme-layout.md` | Spec contract update |

Prefer a dedicated `frontend/src/browse/ContinueCarousel.tsx` if markup exceeds ~80 lines; otherwise keep colocated in `HomePage.tsx` to match current style.

## Interaction model

1. **Initial**: `activeIndex = 0`.
2. **Swipe/scroll**: On `scrollend` (or rAF-throttled scroll), compute nearest snap child → set `activeIndex`.
3. **Inactive press**: `scrollTo` that slide + set `activeIndex` (no navigation).
4. **Continue**: Only on active card; `href` = existing `readerUrl(entry)`.
5. **Arrows**: `activeIndex ± 1` clamped; programmatically scroll into view. Hidden when `length <= 1` or viewport < ~768px.
6. **Keyboard** (track or region `tabIndex={0}`): ArrowLeft/Right adjust index.

## Layout (CSS)

### Mobile

- Track: `display: flex; overflow-x: auto; scroll-snap-type: x mandatory;`
- Slide width: ~`88%` of track (leave ~6% each side for peek) or fixed padding + `scroll-padding`
- Active: full opacity, scale 1, full meta block under/beside cover
- Inactive: scale ~0.9, reduced opacity optional; **cover only** in the meta area (hide text/CTA via CSS or conditional render)
- Hide arrows

### Desktop

- Same track; slightly smaller slide fraction so 1 full + partial neighbors show
- Active scale 1, inactive ~0.85–0.9
- Prev/next absolute or flanking the track

### Comic / flat

- Reuse surface tokens: border, radius, shadow patterns from old `.mango-continue-hero`
- Comic: `border-radius: 0`, thicker border on cards

### Reduced motion

```css
@media (prefers-reduced-motion: reduce) {
  /* disable transform transitions; snap still ok */
}
```

## Data / edge cases

| Case | Behavior |
|------|----------|
| 0 items | Section not rendered |
| 1 item | Single active card, no arrows, no peek needed |
| Many items | All in track; no show more |
| Missing cover_url | Placeholder (reuse card placeholder pattern if available) |

## A11y

- Section `aria-roledescription="carousel"` (or region with label = continue reading)
- Slides: inactive are `button` or focusable controls that “select”; active Continue is a real link
- Live region optional for title change (nice-to-have, not required for MVP)
- Arrow buttons: `aria-label` prev/next; `disabled` at ends

## Compatibility

- Replaces classes: `.mango-continue-list`, `.mango-continue-row`, `.mango-continue-more` usage in continue section
- May keep/rename `.mango-continue-hero` for active card chrome or introduce `.mango-continue-carousel*` namespace for clarity
- Spec file must drop “hero + list + LIST_PREVIEW” contract

## Trade-offs

| Choice | Pro | Con |
|--------|-----|-----|
| scroll-snap native | No lib, touch-friendly | Sync activeIndex from scroll is slightly fiddly cross-browser |
| Cover-only inactive | Clear focus hierarchy | Less progress glance on neighbors |
| Arrows desktop only | Clean mobile | Desktop-only discoverability of arrows |

## Rollback

Revert Home continue markup + CSS + spec; no data migration.
