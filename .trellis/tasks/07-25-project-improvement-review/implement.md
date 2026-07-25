# Implement — Reader UI/UX polish

## Checklist (order)

1. **i18n** — Add `pagePrevious` / `pageNext` (or reuse carefully) in zh / zh-TW / en in `frontend/src/lib/i18n.tsx`.
2. **ReaderViewport** — Localized zone `aria-label`s via `useI18n`.
3. **ReaderControls**
   - Number jump form + clamp commit (use `clampPage` from `readerMath`).
   - Fix actions row (single Exit; Next primary when available).
   - Focus trap + restore on open/close (`useEffect` when `open`).
4. **ReaderPage / ReaderTopBar**
   - Pass opener ref or callback for focus restore.
   - Coarse/touch: `pointerdown` on top edge shows bar (keep existing move/enter).
5. **CSS** — `prefers-reduced-motion` disables flip animations.
6. **Spec** — Update `.trellis/spec/frontend/react-reader.md` (chrome, jump, a11y, navigation note: react-router navigate).
7. **Validate** — typecheck + build + readerMath test.

## Validation commands

```bash
npm run typecheck
npm run build
# optional pure test runner for readerMath if project invokes it:
node --experimental-strip-types frontend/src/pages/reader/readerMath.test.ts
# or existing package script if present
```

Manual browser (after `make run` or vite+server):
- Open a long entry; open settings; jump to middle page; confirm URL + image.
- Tab through dialog; Escape closes; focus returns.
- Enable OS reduced motion; flip pages — no slide animation.
- Touch or DevTools device mode: top-edge interaction shows bar; tap page opens settings.
- Entry with next chapter: controls show Previous? / Next / one Exit.

## Risky files

| File | Risk |
|------|------|
| `ReaderControls.tsx` | Focus trap regressions, jump commit loops |
| `ReaderPage.tsx` | Chrome timers vs controlsOpen interaction |
| `shell.css` | Accidental global reduced-motion scope — keep selectors reader-specific |

## Rollback points

- After step 3: controls usable even if touch polish incomplete.
- Full revert: git restore reader + i18n + shell.css + react-reader.md.

## Out of this implement pass

- Library skeletons, title overflow menus, shell topbar, ConfirmDialog trap (unless shared helper is 1 file and used only by reader first).
