# Implement: Comic Netflix-layout (3 batches)

## Batch 1 — Shell + top bar for comic

- [x] flat-theme visibility shared with comic
- [x] comic hide sidebar + topbar skin + full-width content
- [x] medium card motion; lessc comic-theme.css
- [x] top.tmpl comment update

## Batch 2 — Browse surfaces

- [x] continue-reading full-height side posters (Flat parity)
- [x] library equal-height cards (2-line title slot)
- [x] poster 2:3 + sharp corners

## Batch 3 — Rest of app

- [x] existing comic login/admin/reader/select2 under new shell
- [x] global `border-radius: 0` under comic

## Cross-cutting

- [x] Flat zero visual regression (visibility rule only)
- [x] comic palette retained (not Netflix red)
- [x] spec: `.trellis/spec/frontend/ui-theme-layout.md`

## Done

Code on branch `ui/comic-netflix-layout` (`259c5b5` + follow-up commits).
