# Float Utility Cluster for Chrome Actions

## Goal
Replace the lone top-right GitHub float and the scattered chrome actions (sidebar footer + mobile offcanvas) with one **speed-dial FAB**: language, theme (light/dark), UI style (comic/flat), GitHub, and logout.

## User value
One discoverable place for personal preferences and session exit; sidebar/offcanvas stay navigation-focused.

## Confirmed facts

1. Desktop `.sidebar-footer`: lang (`cycleLanguage`), theme (`toggleTheme`), UI style (`toggleUIStyle`), logout link.
2. Mobile offcanvas: same four after nav + divider.
3. GitHub: fixed `.github-float` (comic Kirby + flat tokens from `07-11-github-float-theme-align`).
4. No existing speed-dial; toggles live in `public/js/common.js` / `i18n.js`.
5. Reader uses its own layout — **out of MVP**.

## Decisions

| Decision | Choice |
|----------|--------|
| Pattern | Speed-dial FAB |
| Primary FAB | Dual-state: rest = `fa-sliders-h`; open = `fa-times` |
| Expand layout | Vertical stack **down** from top-right |
| Desktop footer | **Remove** four chrome actions (keep nav + collapse) |
| Mobile offcanvas | **Remove** four chrome actions |
| Close | Primary toggle + outside click + Esc + **after any satellite action** |
| Logout | Direct `logout` URL, no confirm |
| Satellite order (top→bottom under primary) | Language → Theme → UI style → GitHub → Logout |
| Labels | Icon-only + `title` / `aria-label`; language keeps `.lang-toggle-label` (简/EN/繁) |
| Empty footer | **Delete** `.sidebar-footer` container if empty (no dead markup) |

## Requirements

- R1: Single float cluster hosts all five actions; removes lone `.github-float` as standalone-only control (evolve markup into cluster).
- R2: Rest primary = sliders; open primary = close; satellites in agreed order, stack downward.
- R3: Styling matches theme axes: flat soft circle + accent; comic Kirby (border + hard shadow); both light/dark.
- R4: No chrome duplicates in desktop sidebar footer or mobile offcanvas.
- R5: Close via toggle, outside, Esc, and after action.
- R6: Keyboard/SR: `aria-expanded` on primary, labeled satellites, focus returns to primary on close when appropriate.
- R7: Reuse `cycleLanguage` / `toggleTheme` / `toggleUIStyle`; GitHub external link; logout same `base_url` logout path.
- R8: Position baseline matches current float: desktop `top:16px; right:16px`; mobile under navbar `top:64px`.
- R9: i18n titles for all actions (existing keys + any missing for primary open/close if needed).

## Acceptance criteria

1. At rest, only primary utility FAB visible top-right (no always-visible 5-icon rail).
2. Open shows five satellites stacked below primary in order: lang, theme, style, GitHub, logout.
3. Primary icon swaps to close when open; back to sliders when closed.
4. Clicking a satellite runs the action and **closes** the dial (logout/GitHub navigate away).
5. Outside click and Esc close the dial without side effects.
6. Desktop sidebar has no lang/theme/style/logout footer; mobile offcanvas has no those items.
7. Comic + light/dark and flat + light/dark all look intentional (reuse float theme tokens).
8. Language satellite shows current lang label (简/EN/繁) and still cycles languages.
9. No layout regression that permanently covers primary nav or content critically (z-index ≤ existing float ~1100 family).
10. Reader pages remain unchanged (no cluster unless layout shared — they do not).

## Out of scope

- Reader page chrome.
- Admin theme/UI selects redesign (storage keys unchanged).
- Full language picker UI (keep cycle).
- Logout confirmation.
- Arc/radial speed-dial geometry.
- Always-visible multi-button rail.

## Complexity

**Complex** — markup structure change, CSS comic/flat, small interaction JS, layout chrome cleanup. Needs `prd.md` + `design.md` + `implement.md`.

## Open questions

None — ready for design/implement artifacts and user review before `task.py start`.
