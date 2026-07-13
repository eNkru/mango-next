# Implement: remaining theme consistency

## Checklist

1. [x] F11 harden `setUIStyle`/`setTheme` for documentElement + body when present; layout/login call both
2. [x] F12 admin `uiStyleChanged` → setUIStyle + setTheme
3. [x] F03 login.html.ecr comic structural classes
4. [x] F04 comic-theme.less comic dark login bg
5. [x] F06 comic-theme-dark form labels/inputs
6. [x] F14 comic card progress CSS
7. [x] F15 comic select-bar CSS
8. [x] F09/F10 comic mobile navbar + offcanvas
9. [x] F08 comic select2 CSS
10. [x] `npm run uglify`
11. [x] trellis-check against PRD ACs

## Validation

```bash
npm run uglify
# Manual: 4 states on login, library cards, title (tags+select-bar), mobile chrome, admin style toggle
```

## Risky files

- `public/js/common.js` — FOUC/theme core
- `public/css/comic-theme.less` — large cascade
- `src/views/login.html.ecr`

## Rollback points

After each logical group (JS FOUC, login, forms, cards, chrome, select2) if build/visual break.
