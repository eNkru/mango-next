# UI review: 2 themes × 2 styles

## Goal

Walk the UI across comic/flat × light/dark and produce a **findings list with priorities** — what looks wrong, what is inconsistent, what is taste vs bug — without implementing fixes in this task.

## User value

Users experience four distinct looks. A prioritized audit gives a clear backlog and stops drive-by half-theming.

## Confirmed facts (from codebase)

1. **Two axes, four effective looks**
   - **Theme** (`localStorage.theme`): `light` | `dark` | `system` → resolves to light/dark; dark applies `uk-light` (and comic dark extras).
   - **UI style** (`localStorage.ui-style`): `comic` (default) | `flat` → `comic-theme` / `comic-theme-dark` on html/body.
   - Combinations: comic+light, comic+dark, flat+light, flat+dark.

2. **Not "two hex colors"** — two design languages:
   - **Flat**: single coral accent (`@accent` `#D96A4B`, dark `@dark-accent` `#E8845F`) in `public/css/mango.less`.
   - **Comic**: multi-color Kirby palette (`@comic-red`, `@comic-blue`, `@comic-yellow`, …) in `public/css/comic-theme.less`.

3. **Application mechanism**: class + LESS scopes (no Tailwind theme layer, almost no CSS custom properties). Flat dark often via `.uk-light`; comic via `body.comic-theme*`. Flat mode remaps many `comic-*` classes under `body:not(.comic-theme)`.

4. **Switch UX**: utility FAB (`initUtilityFab` in `public/js/common.js` + `layout.html.ecr`) and Admin theme/style selects.

5. **Key surfaces for visual audit**: layout chrome (sidebar, navbar, FAB), home/library cards, admin, tags, login; reader is a separate document (own theme handling).

6. **Known gaps already documented elsewhere**
   - Dead `mango-app-shell` styles never mounted.
   - Token duplication (`@accent` in mango.less + tags.less; hardcoded coral in flat-compat blocks).
   - Dark paths differ between flat and comic (easy to miss components).
   - Related tasks: `07-11-float-utility-cluster`, `07-11-github-float-theme-align` (chrome/FAB focus).

7. **No formal theming spec** under `.trellis/spec/`; knowledge lives in LESS + task PRDs + `FRONTEND_DEV_GUIDE.md`.

## Decisions

| Decision | Choice | Notes |
|----------|--------|--------|
| Review axes | comic/flat × light/dark (4 states) | Matches runtime model |
| "2 colors" interpretation | 2 **UI styles** (design languages), not 2 hex swatches | Confirmed in code |
| **Primary outcome** | **Findings list + priority only** | No code fixes in this task; follow-up tasks later |
| **Surfaces in scope** | **Full main app + Reader** | chrome, home/library, admin, tags, login, **reader** |
| **Pass bar** | **Consistency first** | Intentional 4-state styling; no unthemed defaults; comic ≠ flat readable as different languages |
| **Priority scheme** | **P0 / P1 / P2** | See definitions below |
| Task type | Audit / planning | Lightweight PRD + findings artifact likely enough |

## Surfaces (must-cover)

| Surface | Notes |
|---------|--------|
| Layout chrome | sidebar, navbar, utility FAB, theme/style toggles |
| Home / library | cards, headings, list chrome |
| Admin | forms, selects, theme/style controls |
| Tags | tag chips / related lists |
| Login | standalone auth chrome if themed |
| **Reader** | Own document (`reader.html.ecr`); still loads `head` → `common.js` + comic CSS, so theme/style classes apply; chrome = modals/nav/buttons (`comic-reader-*`), content = page images |

## Requirements

- R1: Structured review of the four combinations against agreed surfaces and criteria.
- R2: Separate **bugs / inconsistencies** from **taste / redesign** notes.
- R3: Each finding: surface, which of 4 states, severity/priority, short note.
- R4: Prioritized backlog suitable to spawn later fix tasks (not implement here).
- R5: Do not invent a third theme system in recommendations unless product later decides.
- R6: Include **Reader** document page in the audit (not only layout-based pages).

## Quality criteria (pass bar)

**Primary: consistency** across comic/flat × light/dark.

A surface **passes** when:
1. Each of the 4 states has an **intentional** look (not leftover default UIkit / browser chrome).
2. No obvious **half-themed** pieces (e.g. light card on dark page, comic button in flat that still screams Kirby, or flat coral missing only in one state).
3. **comic vs flat** remain distinguishable design languages on the same surface.
4. Theme/style toggle updates the surface without reload (where that surface uses global classes).

**Not primary** (note only if glaring, do not drive ranking unless P0 unreadable):
- WCAG contrast scores
- Brand “Kirby purity” taste debates beyond “this state has no style at all”

## Priority scheme

| Priority | Meaning | Typical follow-up |
|----------|---------|-------------------|
| **P0** | A state is clearly **unthemed or half-themed** (broken chrome, white flash blocks, wrong light surface on dark, controls invisible/unusable) | Fix soon; likely own task |
| **P1** | Component **works** but **inconsistent** across states or axes (e.g. comic hover missing in dark only; flat remaps wrong token) | Grouped fix task |
| **P2** | Polish, dead CSS (`mango-app-shell`), docs, token duplication, taste notes | Backlog / cleanup |

Each finding records: **id**, **surface**, **states** (which of 4), **priority**, **type** (bug vs taste), **evidence** (selector/file if known), **note**.

## Acceptance criteria

- [x] Shared outcome: findings + priority (no implementation this task).
- [x] Explicit surface list: chrome + home/library + admin + tags + login + reader.
- [x] Explicit quality criteria: consistency-first (see above).
- [x] Priority scheme: P0 / P1 / P2 (see above).
- [x] Written findings document in task dir (`findings.md`) covering must-cover surfaces × 4 states.
- [x] Prioritized backlog summary suitable to spawn later fix tasks.
- [x] Out-of-scope list agreed.

## Deliverable

- `findings.md` — 20 findings (2×P0, 13×P1, 5×P2), four-state scorecard, top-5 fix backlog, false-positive list.

## Out of scope

- Implementing visual fixes in this task.
- Full redesign of comic or flat language (unless a finding is only "needs redesign").
- Full `mango-app-shell` dead-code purge (note only if it confuses review).
- Replacing LESS with CSS-variables / Tailwind theme system.
- Systematic a11y audit / contrast tooling (unless a state is clearly unreadable → file as P0 bug).
- Re-doing work already scoped solely in `07-11-float-utility-cluster` / `07-11-github-float-theme-align` except cross-check and reference if still broken.

## Open questions

None blocking planning. Ready for user review → `task.py start` → code/visual audit → `findings.md`.

## Complexity

**Lightweight**: audit + `findings.md`; no `design.md`/`implement.md` required for this task. Fix work becomes separate task(s).

## Related tasks

- `07-11-float-utility-cluster` — FAB chrome actions
- `07-11-github-float-theme-align` — GitHub float theming
- `07-09-fix-reader-blank` — reader (separate)
