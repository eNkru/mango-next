# Implement: Float Utility Cluster

## Checklist

1. **Markup** (`src/views/layout.html.ecr`)
   - [ ] Replace `.github-float` with `.utility-fab` structure (primary + menu items).
   - [ ] Delete `.sidebar-footer` chrome actions (entire footer if empty).
   - [ ] Remove mobile offcanvas lang/theme/style/logout (+ orphan divider if any).
   - [ ] Wire titles/aria via existing i18n keys; add keys for primary if missing (`utility_menu` or reuse).

2. **i18n** (`public/js/i18n.js`)
   - [ ] Ensure keys: language, theme_toggle, ui_style_toggle, view_on_github, logout, plus primary open/close label if new.
   - [ ] en / zh-cn / zh-tw.

3. **CSS flat** (`public/css/mango.less`)
   - [ ] Position cluster (reuse github-float coordinates).
   - [ ] Style primary + items (flat light/dark).
   - [ ] Open state stack (gap, visibility).
   - [ ] Remove or repoint obsolete `.github-float` rules.

4. **CSS comic** (`public/css/comic-theme.less`)
   - [ ] Kirby styles for primary + items light/dark.
   - [ ] Remove/repurpose `.github-float` comic rules if present.

5. **JS** (`public/js/common.js`)
   - [ ] Open/close state machine; outside click; Esc; after-action close.
   - [ ] Call existing cycleLanguage / toggleTheme / toggleUIStyle from item handlers.
   - [ ] Init on DOM ready; no FOUC flash if possible (menu starts closed/hidden).

6. **Build**
   - [ ] `npx gulp less` (css gitignored but local).

7. **Manual QA**
   - [ ] Desktop: open/close, each action, outside, Esc.
   - [ ] Mobile width: under navbar, no clip.
   - [ ] Comic/flat × light/dark.
   - [ ] Lang label updates on cycle.
   - [ ] Sidebar/offcanvas free of old chrome actions.
   - [ ] Reader unchanged.

## Validation commands

```bash
npx gulp less
# grep sanity
rg -n "sidebar-footer|github-float|utility-fab" src/views/layout.html.ecr public/css public/js
```

## Risky files

- `layout.html.ecr` — single layout for all authenticated pages.
- `common.js` — global boot; keep handlers isolated.
- Comic/flat CSS — specificity wars with `.uk-light`.

## Rollback point

Before commit: restore previous layout float + footer/offcanvas from git if UX fails review.

## Branch note

Implement on a branch from current work (`feature/github-float-theme-align` or fresh from main after merge) — confirm with user at start if needed. Prefer stacking on the github-float feature branch if still unmerged so styles stay consistent.
