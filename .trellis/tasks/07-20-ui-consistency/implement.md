# Implement: UI Consistency Audit and Fix

## Prerequisites

- PRD decisions locked (R1–R8)
- Design contracts in `design.md`
- Do **not** start until user reviews artifacts and `task.py start` runs

## Ordered Checklist

### Phase 0 — Manifests & baseline

- [ ] Fill `implement.jsonl` / `check.jsonl` with frontend specs + this task research refs
- [ ] Baseline: `npm run typecheck` (or project-root equivalent)
- [ ] Grep snapshot (for acceptance):
  - `style={{` under `frontend/src/pages`
  - hard-coded Chinese in R2 pages
  - `react-preview` / `PlaceholderPage`

### Phase 1 — Tokens & CSS (R4/R5/R6/R7 CSS)

- [ ] `tokens.css`: add danger, success, on-accent, reader-*, optional ink
- [ ] `shell.css`:
  - primary text → `var(--mango-on-accent)`
  - danger → danger tokens
  - success alert → success token
  - ghost border → reader/ghost token
  - reader chrome hard-codes → reader tokens
  - comic `.mango-btn` thick border + offset shadow
  - form: style `label.mango-field > span` as label text; support password-row nested control
  - utility classes from design (actions flush, max-w-search, scroll-x, etc.)

### Phase 2 — Shared components

- [ ] `PosterCard`: `showProgress?: boolean` (default true)
- [ ] `BrowseToolbar`: `modes?: SortMode[]` + filter options; safe when mode not in list
- [ ] `ErrorState`: `onRetry?`, `retryLabel?`; button inside error block
- [ ] Remove Chinese defaults from StatePanels / ConfirmDialog **or** force callers to pass labels (prefer callers + i18n)
- [ ] Extract `LanguageSelect` (or shared helper) for AppShell + Login
- [ ] ConfirmDialog: no hard-coded 确认/取消 defaults without i18n path (callers already pass labels where used)

### Phase 3 — i18n dictionary (R2 prep)

- [ ] Extend `messages` in `i18n.tsx` for Login, Tags, Users, Missing, TagDetail, Confirm, Admin action errors, etc.
- [ ] Keep zh-cn / zh-tw / en keys in lockstep (`MessageKey` type enforces)

### Phase 4 — Page migrations

Order:

1. [ ] **TagDetail** (R1 + R2 + R3 + R4): adapter, PosterCard, BrowseToolbar modes, i18n, ErrorState onRetry, utilities
2. [ ] **Home**: ErrorState onRetry (drop sibling button)
3. [ ] **Library / TitleDetail**: ErrorState onRetry if missing; ensure modes default OK
4. [ ] **TagsIndex**: i18n + utilities + ErrorState onRetry
5. [ ] **UserList / UserEdit**: i18n + form field markup (R7) + utilities + ErrorState onRetry
6. [ ] **MissingItems**: i18n + utilities + ConfirmDialog labels via t()
7. [ ] **Login**: i18n + LanguageSelect + field markup
8. [ ] **Admin**: actionError ErrorState + onRetry (R3-B)

### Phase 5 — Remove scaffold (R8)

- [ ] Delete `PlaceholderPage.tsx`
- [ ] Strip App.tsx branch
- [ ] Remove Go `handleReactPreview` + route
- [ ] `DEFAULT_BOOT` → home
- [ ] Update `FRONTEND_DEV_GUIDE.md`
- [ ] Grep clean: no `react-preview` product refs (except changelog/history if any)

### Phase 6 — Verify

- [ ] `npm run typecheck`
- [ ] `npm run build`
- [ ] Go tests if route registration tested: `go test` relevant packages
- [ ] Manual theme matrix: comic/flat × light/dark — Library, TagDetail, Login, Admin, open Reader chrome
- [ ] Acceptance grep:
  - no hand-written TagDetail card markup
  - R2 pages use `t(` for user-visible strings
  - residual `style={{` only dynamic cases
  - no PlaceholderPage

## Validation Commands

```bash
# Frontend
npm run typecheck
npm run build

# Optional Go (after route delete)
cd go && go test ./internal/server/ -count=1

# Grep gates (examples)
rg -n "PlaceholderPage|react-preview" frontend go FRONTEND_DEV_GUIDE.md
rg -n "style=\{\{" frontend/src/pages --glob '*.tsx'
rg -n "正在加载|加载中|欢迎回来|请输入用户名" frontend/src --glob '*.tsx'
```

## Risky Files / Rollback Points

| Risk | File | Mitigation |
|------|------|------------|
| Toolbar regression | `BrowseComponents.tsx` | default modes = full list; Library smoke |
| Form a11y break | Login / UserEdit / shell.css field | keep clickable label wrapping; password-row nested |
| Comic button too loud | shell.css R6 block | isolated CSS revert |
| Reader visual drift | reader tokens | side-by-side before/after chrome |
| Boot default change | `boot.ts` | only affects missing `#mango-boot` (vite dev) |
| Route delete | `server.go` | simple revert |

Rollback: single git revert of the task branch/PR; no data migration.

## Review Gates Before `task.py start`

- [ ] User reviewed `prd.md` + `design.md` + this file
- [ ] No open product questions in `prd.md` Open Questions (all resolved)
- [ ] implement/check jsonl curated for sub-agent mode if used

## Notes for Implementer

- Do not expand `apiTagTitles` fields
- Prefer props-passed translations over hard-coded Chinese in shared components
- Keep BEM-ish class names; no Tailwind / CSS modules
- Match surrounding file style (some pages are dense one-liners — prefer readable multi-line when editing heavily)
