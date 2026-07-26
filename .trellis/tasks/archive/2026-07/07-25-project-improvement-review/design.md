# Design — Reader UI/UX polish

## Architecture boundary

- **Touch only:** `frontend/src/pages/reader/*`, `frontend/src/styles/shell.css` (reader + reduced-motion rules), `frontend/src/lib/i18n.tsx` (labels), `.trellis/spec/frontend/react-reader.md`.
- **Do not change:** Go handlers, boot JSON, API paths, `App.tsx` routes, global state architecture, CSS methodology.

## Data flow (unchanged)

```text
URL / boot → useReaderBootstrap → data
prefs (localStorage mango.reader.*) → useReaderPrefs
page state → useReaderNavigation (navigate replace + preload + progress callback)
viewport zones / keys → flipWithRtl / setPage
controls modal → onJumpPage / onPrefs / entry nav / exit
```

Polish sits in presentation + small interaction handlers; page math stays in `readerMath.clampPage`.

## R1 — Page jump control

**Replace** the `Array.from({ length: pages }, …)` `<select>` with:

```text
<label>
  Jump to page
  <input type="number" min={1} max={pages} inputMode="numeric" value={draft} />
  <button type="submit">Go</button>   // or form onSubmit
</label>
```

- Keep a controlled **draft** string/number separate from live `page` until submit (avoids thrashing URL on every keystroke).
- On open, seed draft from current `page`.
- On submit / blur-commit: `onJumpPage(clampPage(parsed, pages))`.
- Optional: small text hint `{page}/{pages}` already shown above — keep progress line.
- **Not** required: virtualized select libraries.

## R2 — Touch chrome discovery

Current behaviors to preserve:
- Top edge pointer enter shows bar (`EDGE_PX`).
- Escape opens/closes controls.
- Image click / `onImageClick` opens controls (paged + continuous).

Additive polish (minimal):
1. When `openControls` runs, always `showBar(true)` (already true) — ensure top bar stays visible while modal open (already `visible={barVisible || controlsOpen}`).
2. On first pointer **tap** (click) on the page image that opens controls, also force bar visible (already via openControls).
3. For coarse pointers: treat `pointerdown` near top edge (same `EDGE_PX`) as show-bar (in addition to `pointermove`) so touch-from-top works without hover.
4. Spec: document “tap page → settings; top edge / bar → chrome; Escape → toggle settings”.

No permanent bottom chrome in this slice.

## R3 — Focus management

Implement a tiny local helper in reader (or `shell/useFocusTrap.ts` if preferred without expanding scope):

On `open === true`:
- `previousActive = document.activeElement`
- focus first focusable in dialog (close button or page input)
- `keydown` Tab: cycle focusables inside dialog
- Escape: already handled at page level — ensure trap does not fight; page Escape should close controls first (current logic OK)

On close:
- restore `previousActive` if still in document (prefer topbar “Reading settings” button — pass `openerRef` from `ReaderTopBar` / `ReaderPage`)

Backdrop click already calls `onClose`.

## R4 — Zone i18n

```tsx
aria-label={t('previousEntry')} // or dedicated t('pagePrevious') / t('pageNext')
aria-label={t('nextEntry')}
```

Prefer dedicated `pagePrevious` / `pageNext` if “previous entry” is wrong semantics for page flip; add zh/zh-TW/en in `i18n.tsx`.

## R5 — Reduced motion

In `shell.css`:

```css
@media (prefers-reduced-motion: reduce) {
  .mango-reader-page--flip-left,
  .mango-reader-page--flip-right {
    animation: none;
  }
}
```

Optional belt-and-suspenders: in `ReaderPage`, skip setting `flipSide` when `matchMedia('(prefers-reduced-motion: reduce)')`.

## R6 — Actions row

Logic:

```text
if previousEntryUrl && onPreviousEntry → Previous button
if nextEntryUrl → primary Next
else → primary Exit
if nextEntryUrl → secondary Exit once
// never render Exit twice
```

Remove the branch that always adds Exit when `exitUrl` after already rendering Exit as primary.

## R7 — Form submit

`onSubmit` of jump form: parse draft, `onJumpPage(clamp)`, preventDefault. Remove empty `submitJump`.

## Compatibility

- Prefs keys unchanged.
- URL scheme `/reader/:tid/:eid/:page` unchanged.
- Keyboard page navigation in `useReaderNavigation` unchanged.

## Trade-offs

| Choice | Why | Cost |
|--------|-----|------|
| Number input vs sparse select | Scales to any page count | Slightly less “browse all pages” UX |
| Local focus trap vs library | No new dependency | Must test Tab cycle carefully |
| Edge pointerdown only | Minimal code | Not a full gesture system |

## Rollback

Revert the few frontend files + CSS + i18n + spec; no migrations.
