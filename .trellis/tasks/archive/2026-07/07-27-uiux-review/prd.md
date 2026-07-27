# UI/UX polish: ease-of-use, mobile, page-load

## Goal

Improve the Mango React shell across three axes the user named: ease-of-use,
mobile design correctness, and page-load performance. This is an
**implementation** task (the review findings below drive the work), not a
doc-only audit.

## Scope

Frontend only: `frontend/src/**`, `frontend/index.html`, and the production
HTML shell `go/web/views/react-shell.tmpl` (it — not the Vite `index.html` — is
what production serves, and it owns the FOUC script + asset `<link>`s). Vite
config (`vite.config.ts`) is in scope for bundle splitting / cache-busting.

Out of scope: Go API behavior, LESS legacy pages, content/i18n copy rewrites.

## Findings (evidence-based, file:line)

### Ease-of-use
- **Initial `<html lang>` is always `zh-CN`** before JS runs.
  `go/web/views/react-shell.tmpl:4` and `frontend/index.html:3` hardcode
  `lang="zh-CN"`. `i18n.tsx:280-282` corrects it on module load and on
  `setLanguage`, but the pre-JS paint (and any no-JS path) is wrong for
  `en` / `zh-tw` users — screen readers get the wrong language on first paint.
- **Topbar tool cluster is crowded** (`AppShell.tsx:154-231`): 6 icon controls
  (GitHub, Issues, Theme, UI-style, Language, Logout) sit right of the brand.
  On narrow viewports this wraps and stacks the topbar tall (see mobile).
- **Title-detail page is dense** (`TitleDetailPage.tsx`): each entry card carries
  up to 4 action buttons (Continue/Begin, Download, Mark read/unread, Edit) +
  an admin checkbox; the page packs many `useState` and a long JSX tree. Hard
  to scan, busy on mobile.
- **No skip-to-content link** and **focus is not moved to the new page heading
  on client-side route change** (`AppLink`/`App.tsx`). Keyboard/screen-reader
  users lose orientation between pages.
- **Quick-jump select always opens page 1** (`TitleDetailPage.tsx:94`):
  `navigate(.../${id}/1)` ignores any saved progress on the target entry.

### Mobile design
- **Reader top bar has no mobile breakpoint** (`ReaderTopBar.tsx`,
  `shell.css:1404-1474`): three groups (brand+controls / title+progress /
  language+exit) in a `space-between` flex with `gap:1rem`, no `flex-wrap`,
  no `@media` for it. On a ~360px phone this overflows / squishes the center
  title to nothing.
- **App topbar stacks tall on mobile** (`shell.css:1334-1355`): at ≤820px the
  nav wraps to a full-width scroll row (good), but the 6-button tool cluster
  can wrap to its own line, yielding a 3-row topbar that eats vertical space.
- **Touch targets below the 44px minimum**: `.mango-btn--icon` is
  `min-width/height: 2.25rem` = 36px (`shell.css:287-292`); the tag-remove
  button is `1.5rem` = 24px (`shell.css:1248-1256`). Below Apple/Google 44px
  guidance — hard to hit on phones.
- **`backdrop-filter: blur(10px)` on the sticky topbar** (`shell.css:36`):
  expensive on low-end mobile, can cause scroll jank.
- **Missing-items table on mobile** (`MissingItemsPage.tsx:120-175`): 4-column
  table in `.mango-scroll-x` — horizontal scroll works but is not a native
  mobile pattern. Acceptable fallback, not great.

### Page-load / performance
- **No code-splitting anywhere** — no `React.lazy`/`Suspense`/dynamic `import`
  in the repo (verified by grep). `App.tsx` eagerly imports every page
  including the heavy **Reader** (`ReaderPage` + `ReaderControls` +
  `ReaderViewport` + `useReaderNavigation` + `useFocusTrap` + readerMath).
  The entire app ships in one `main.js`, so a user on the login page downloads
  the reader. **Biggest available page-load win.**
- **No content-hash on asset filenames** (`vite.config.ts:31-39`):
  `entryFileNames: 'assets/main.js'`, CSS forced to `assets/main.css`, fixed
  names — and `react-shell.tmpl:11,14` hardcodes those exact URLs. After an
  update the URLs are identical, so browser/CDN caches can serve **stale**
  `main.js`/`main.css`. Real correctness risk for a self-hosted app that
  upgrades. Needs a design decision (see design.md §Cache-busting).
- **Fredoka font not preloaded** (`fonts.css`, `react-shell.tmpl`): `@font-face`
  with `font-display: swap` (good — no FOIT) but the woff2 is only discovered
  after CSS parses. A `<link rel="preload" as="font" ...>` in the shell would
  start the fetch earlier for comic-theme users.
- **CSS is one 1825-line `shell.css`** covering all themes + reader + admin.
  Not easily splittable under the single-`main.css` Go-embed model; low
  priority given size.

### Acknowledged strengths (do not regress)
i18n store + `applyDocumentLanguage`; `:focus-visible` on inputs/menus
(`shell.css:156,190,545`); `prefers-reduced-motion` respected
(`shell.css:668,1109,1523`); rail skeletons to avoid CLS (`PosterRail.tsx`);
`loading="lazy"` on poster images (`BrowseComponents.tsx:48`); semantic
`<header role="banner">` + `<main>`; login password toggle + language select
before session; ARIA menu keyboard model in `AppShell.tsx`; decorative `alt=""`.

## Requirements

1. Split route bundles so the Reader (and Admin) code is not in the initial
   `main.js` loaded by every page (login/home/library).
2. Fix mobile overflow on the reader top bar and reduce app-topbar height on
   small screens.
3. Bring icon-button touch targets up to ≥44px on mobile without breaking
   desktop density.
4. Set `<html lang>` correctly before first paint for all three languages.
5. Reduce mobile scroll-jank from the blurred sticky topbar.
6. Do not regress any acknowledged strength above; keep all four theme combos
   (flat/comic × light/dark) and i18n working.

## Acceptance Criteria

- [ ] `main.js` no longer contains the Reader; a separate chunk loads on
      first entry to `/reader/...`. Login page network payload excludes reader
      code (verify via build output / network).
- [ ] Reader top bar fits a 360px viewport without horizontal overflow and the
      center title remains legible (or is intentionally hidden with an
      accessible alternative).
- [ ] App topbar on a 360px viewport does not exceed ~2 visual rows and all
      controls remain reachable.
- [ ] Every icon-only control has a ≥44px hit area on touch devices
      (including tag-remove).
- [ ] With `localStorage['mango-language']='en'`, the very first paint has
      `<html lang="en">` (not `zh-CN`).
- [ ] No stale-asset risk after upgrade: either filenames are content-hashed
      with the shell template resolving them, **or** caching headers are set
      to force revalidation (decision recorded in design.md).
- [ ] `npm run typecheck` and `npm run build` pass; `scripts/check-react-outputs.mjs`
      passes; no new inline styles except dynamic width/margin.
- [ ] Manual smoke: four theme combos, login language switch, library grid,
      title-detail, reader (continuous + paged), mobile widths 360/414/768.

## Notes

- Keep `prd.md` focused on requirements/constraints/acceptance. Technical
  design in `design.md`; execution checklist in `implement.md`.
- Some findings (title-detail density, missing-items cards, skip-link,
  route-change focus, font preload) are lower-severity and may be deferred —
  see `implement.md` tiers. Scope confirmed at the review gate.