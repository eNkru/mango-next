# Implement: Theme CSS consolidation

## Ordered Checklist

1. Create `_variables.less` with all shared palette colors.
2. Update `mango.less` to `@import "_variables.less"` and remove inline
   duplicate variable definitions.
3. Update `tags.less` to `@import "_variables.less"` and scope Select2
   dark rules to `body:not(.comic-theme):not(.comic-theme-dark)`.
4. Update `comic-theme.less` to `@import "_variables.less"` and remove
   inline duplicate variable definitions.
5. Remove ~30 unused comic decorative classes and keyframes from
   `comic-theme.less`.
6. Remove the flat-fallback and dark-flat-override sections from
   `comic-theme.less`.
7. Mirror changes 3-6 in the compiled `comic-theme.css` and `tags.css`.
8. Fix dark background consistency: change `#101116` to `#121212` in
   `mango.less` and `mango.css` (mango-app-shell dark gradient).
9. Update `head.tmpl` to conditionally load Google Fonts.
10. Verify no template references break (grep for removed classes in
    templates).
11. Build and verify Go binary compiles (embed still works).
12. Run Docker container and visually verify all 4 theme variants
    (flat-light, flat-dark, comic-light, comic-dark).

## Validation Commands

```bash
cd go && go build ./...
cd go && go vet ./...
cd go && go test ./...
# Verify no removed classes are referenced
rg 'sound-boom|speech-bubble|halftone-overlay|comic-panel-tilt|comic-border|comic-flip|comic-title-3d|comic-alert|comic-breadcrumb|comic-btn-info|comic-btn-success|comic-btn-warning' go/web/views/ go/internal/
```

## Risk And Rollback Points

- Removing flat-fallback may break flat-mode rendering for structural
  `comic-*` classes. If a class breaks, add its flat base to `mango.less`
  instead of restoring the fallback.
- The JS font loader must run before body paint to avoid FOUC in comic mode.
- Manual .css edits must match .less edits exactly; diff the two after
  editing to verify consistency.
