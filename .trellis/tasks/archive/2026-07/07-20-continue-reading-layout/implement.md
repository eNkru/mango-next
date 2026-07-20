# Implement: 主页继续阅读区布局优化

## Checklist

1. [ ] i18n: add `showMore` / `showLess` to zh-cn, zh-tw, en in `frontend/src/lib/i18n.tsx`
2. [ ] CSS: add hero + list + row + more-button styles in `frontend/src/styles/shell.css`; replace/remove old `.mango-continue-grid` / `.mango-continue-card` usage; update comic-theme selectors if class names change; optional thicker progress on hero
3. [ ] UI: rewrite continue section in `frontend/src/pages/HomePage.tsx`
   - Split `continue_reading` into `primary` + `rest`
   - Hero for primary; list for rest with expand (preview 3)
   - Reuse `ProgressBar`; reader links only
4. [ ] Smoke: 0 / 1 / 4 / 5+ items mental paths; comic + flat markers still look ok
5. [ ] Build check: frontend typecheck/lint if available (`npm run build` or project script)

## Validation

```bash
# from frontend/ if package scripts exist
npm run build
```

Manual:

- Home with continue data: hero + ≤3 rows, expand works
- Single continue item: hero only
- No continue data: section absent; rails still work
- Click hero 继续 + list row → reader URL

## Risky files

- `frontend/src/pages/HomePage.tsx` — main structure
- `frontend/src/styles/shell.css` — continue + comic overrides ~L639–670, ~L846+, ~L904+
- `frontend/src/lib/i18n.tsx` — message keys must exist in all 3 locales

## Rollback point

Git revert the three files above; no backend/DB changes.

## Ready for start when

- User reviewed prd + design + implement
- `task.py start` after approval
