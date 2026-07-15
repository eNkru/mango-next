# Implement: Flat Netflix restyle (3 batches)

## Batch 1 — Shell + tokens

- [x] Add Flat Netflix tokens in `_variables.less` (dark + light, red accent)
- [x] Create `flat-theme.less` + hand-synced `flat-theme.css`; link in `head.tmpl`
- [x] `common.js`: apply `flat-theme` / `flat-theme-dark`; clean comic on switch
- [x] FOUC: head script sets flat/comic classes before paint
- [x] `top.tmpl`: `.flat-topbar` markup; CSS hides sidebar under flat / topbar under comic
- [x] Restyle utility FAB + primary buttons for flat
- [ ] Smoke: toggle comic↔flat, dark↔light; comic sidebar unchanged (manual)

**Validate:** visual check + no JS errors; `go test` if any go change (likely none)
**Note:** full `mango.css` not recompiled from less (no lessc in repo); batch-1 overrides live in `flat-theme.css`.

## Batch 2 — Browse surfaces

- [x] Home: hero billboard + rails spacing/hover (flat-theme.css)
- [x] Library + title grids: 2:3 poster cards, denser grid
- [x] Title page: heading/breadcrumb/select-bar flat polish
- [x] Tags index pills
- [x] Progress bars/badges use Netflix red

**Validate:** continue reading, open book, tags navigation (manual)
**Note:** CSS-only under `body.flat-theme`; templates unchanged for batch 2

## Batch 3 — Rest of app

- [x] Login (dark/light Netflix panel, red CTA)
- [x] Admin cards + tables + action bars
- [x] Reader shell + modal (dark canvas, red buttons)
- [x] Plugin/download/admin-sub pages (forms, empty states, cards)
- [x] Select2 under flat light+dark (Netflix red choices)
- [ ] Full manual checklist from PRD acceptance (user smoke)

**Validate:** admin scan/thumb; reader open/flip/progress (manual)

## Cross-cutting

- [ ] Do not regress comic-theme.less behavior
- [ ] Commit compiled CSS with less sources
- [ ] Update frontend notes in `.trellis/spec/frontend` if stable contracts emerge

## Suggested order of files (Batch 1)

1. `_variables.less`
2. `flat-theme.less` + build
3. `head.tmpl`, `common.js`
4. `top.tmpl` / `bottom.tmpl` as needed

## Done when

All three batches merged or one branch with three logical commits matching batches; PRD acceptance checked.
