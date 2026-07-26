# Implement - GitHub links and compact AppShell tools

## Checklist

1. Add semantic Lucide icon entries for GitHub, issue reporting, theme,
   UI-style, and language where missing.
2. Add localized labels for GitHub, issue-list/reporting, preference controls,
   and preference choices.
3. Build a small local AppShell dropdown menu interaction that handles trigger
   focus, menu keyboard navigation, selection, Escape/outside dismissal, and
   focus restoration without new dependencies.
4. Replace the AppShell theme, UI-style, and language selects with their
   corresponding icon triggers and menus, calling the existing setters.
5. Add icon-only GitHub and issue-list external anchors with safe new-tab
   attributes and localized tooltip/accessibility text.
6. Add responsive, token-based shell CSS for compact tools and menus, including
   comic/flat and light/dark verification.
7. Run typecheck and build; manually verify mouse, keyboard, touch-width, and
   all theme/style combinations.

## Validation

```bash
npm run typecheck
npm run build
```

Manual checks:

- Every menu opens from its icon, displays the current option, changes the
  setting, closes, and returns focus to its trigger.
- Tab, arrow navigation where implemented, Enter/Space, Escape, and outside
  click behave predictably without focusing content behind the menu.
- GitHub and issue icons show localized hover/focus titles and use the expected
  destinations in a new tab.
- Desktop and narrow mobile widths remain usable in all flat/comic and
  light/dark combinations.

## Risks

- Custom menu keyboard and dismissal behavior can regress accessibility; keep
  it small and test it directly.
- Topbar overflow must be checked at narrow widths after adding two icons.
