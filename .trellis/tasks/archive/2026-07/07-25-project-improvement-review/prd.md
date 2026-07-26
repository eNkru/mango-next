# Project improvement review — Reader UI/UX polish

## Goal

Improve **daily reading UX** in the existing React immersive reader without changing app architecture (no new frameworks, no route/package redesign, no backend restructure).

User value: easier page jump on long books, better touch chrome discovery, fewer a11y/i18n gaps, cleaner controls modal actions, and respect for reduced motion.

## Scope

### In scope
- Reader surfaces only: `frontend/src/pages/reader/*` and reader-related CSS in `frontend/src/styles/shell.css`.
- Small shared helpers only if required by reader (e.g. focus trap used by `ReaderControls`; optional reuse by existing dialogs later is fine but **not required** this slice).
- i18n strings in `frontend/src/lib/i18n.tsx` for new/fixed reader labels.
- Spec touch-up: `.trellis/spec/frontend/react-reader.md` chrome/jump contracts after implementation.

### Out of scope
- Go server, API contracts, CI, storage, package layout.
- Home/library/title/admin polish (deferred).
- Replacing CSS system, Tailwind, component library.
- New reader features (double-page spread, bookmarks UI, fullscreen API, etc.) unless listed below as explicit acceptance items.

## Background (confirmed)

| Fact | Evidence |
|------|----------|
| Reader is not in `AppShell` | `react-reader.md`, `ReaderPage.tsx` |
| Chrome: top edge ~36px, idle hide 1.8s, Escape toggles controls | `ReaderPage.tsx` `EDGE_PX` / `IDLE_MS` |
| Page jump = one `<option>` per page | `ReaderControls.tsx` ~83–99 |
| Zone labels hard-coded English | `ReaderViewport.tsx` `"prev"` / `"next"` |
| Flip keyframes not in reduced-motion | `shell.css` flip animations; reduced-motion only for skeleton/carousel |
| Controls dialog: role only, no focus trap | `ReaderControls.tsx` |
| Duplicate exit when `nextEntryUrl` set | `ReaderControls.tsx` ~200–216: Next **and** Exit both rendered |
| Continuous: image click opens controls; paged: side zones flip, image click opens controls | `ReaderViewport` + `ReaderPage` |
| Spec still mentions `history.replaceState`; code uses react-router `navigate` | `useReaderNavigation.ts` vs `react-reader.md` §3 |

## Requirements

- **R1 Page jump:** Replace full page `<select>` with a compact control (number input + optional Go, or input + range) that clamps via existing `clampPage` semantics; usable for 1–thousands of pages.
- **R2 Touch chrome:** Provide an obvious way on touch/coarse pointers to show the top bar without relying only on top-edge hover (e.g. tap center already opens controls; also show top bar on first intentional center interaction and/or short tap on non-zone area; document in spec). Prefer behavior tweak over new permanent chrome.
- **R3 Controls modal a11y:** When open, focus moves into dialog; Tab cycles inside; Escape closes (already partially handled on `window`); restore focus to opener (controls button) on close.
- **R4 i18n zones:** Zone `aria-label`s use `t(...)` (add keys if missing), not raw `"prev"`/`"next"`.
- **R5 Reduced motion:** Under `prefers-reduced-motion: reduce`, flip page animations do not run (CSS and/or skip `flipSide` class).
- **R6 Controls actions cleanup:** When next entry exists, primary = Next entry; Exit remains once (no duplicate Exit). When no next, primary = Exit once.
- **R7 Dead form:** Remove or wire `submitJump` so Enter on jump control commits the same clamp/jump as explicit action.
- **R8 Spec:** Update `react-reader.md` chrome, jump control, a11y, and navigation URL notes to match code after changes.
- **R9 Validation:** `npm run typecheck` and `npm run build` (or project-equivalent) pass; `readerMath` unit test still passes if still used.

## Acceptance criteria

- [x] AC1: Opening controls on a 500+ page entry does not render 500+ `<option>` nodes; user can jump to an arbitrary page within [1, pages] and URL/progress update as today.
- [x] AC2: On a coarse-pointer / touch path, user can reveal top bar or open settings without desktop-only hover on the top 36px strip alone (behavior documented in spec).
- [x] AC3: With controls open, keyboard Tab does not focus elements under the backdrop; Escape closes; focus returns to the control that opened the panel.
- [x] AC4: Flip animation does not play when OS reduced-motion is on (visual check or CSS rule present covering `.mango-reader-page--flip-*`).
- [x] AC5: Controls footer never shows two identical Exit buttons; Next + single Exit when next exists.
- [x] AC6: Zone buttons expose localized accessible names.
- [x] AC7: No new npm dependencies required (or justified if any); no Go changes.
- [x] AC8: `react-reader.md` matches implemented chrome/jump/a11y contracts.

Implementation note (2026-07-25): code on branch `feat/reader-ui-ux-polish`; `npm run typecheck` + `npm run build` green. Manual browser smoke still recommended before merge.

## Key decisions (resolved)

| Decision | Choice |
|----------|--------|
| Outcome type | Implementable polish slice (not report-only) after plan approval |
| Surface | Reader only (option A) |
| Structure | No architecture / package / routing redesign |
| Dependencies | Prefer zero new libs; native focus trap in ~30–50 lines OK |

## Risks / deferred

- **Risk:** Continuous-mode image click currently opens full controls; adding “show bar only” vs “open controls” may need a short intentional distinction (tap vs long-press is out of scope — keep simple: show bar when opening controls; optional light bar-only on edge remains).
- **Deferred:** Library skeleton, title action overflow, shell topbar mobile, progressbar roles, carousel live regions, ConfirmDialog focus trap (unless shared helper is trivial).
- **Deferred:** Full keyboard page flip help overlay / onboarding.

## Open questions

_None blocking._ Implementation details (number input vs input+button layout) are design choices within R1.

## Notes

- Host may lack `go` on PATH; reader work is frontend-only.
- Do not start implementation until user **explicitly approves this final planning summary** and `task.py start` is run.
