# Design - GitHub links and compact AppShell tools

## Boundary

Change only the normal React application shell. The immersive reader has separate
chrome and remains unchanged. No server or persistence contract changes are
needed: existing theme, UI-style, and language setters keep owning storage.

## AppShell tools

Replace the three visible `<select>` controls with three icon buttons:

- Theme button opens System, Light, and Dark choices.
- UI-style button opens Comic and Flat choices.
- Language button opens Simplified Chinese, Traditional Chinese, and English
  choices.

Each menu is a local AppShell interaction with semantic menu/button roles,
current choice state, keyboard focus management, Escape dismissal, outside-click
dismissal, and focus restored to its trigger. A selection calls the existing
setter and closes the menu. Only one menu may be open at a time.

Add two icon-only external anchors beside the preference controls:

- GitHub repository: `https://github.com/eNkru/mango-next`
- Issue list: `https://github.com/eNkru/mango-next/issues`

They include a localized `aria-label` and `title`; the issue title provides the
requested feedback/reporting hint on both pointer hover and keyboard focus.

## Accessibility and visual design

- Reuse `Icon` and semantic `icons` entries. Icons are decorative because the
  containing button/link owns the accessible name.
- Use `.mango-btn--icon` hit-target conventions and add shell-scoped menu CSS.
- Menus inherit existing theme tokens and comic sharp-corner behavior rather
  than adding a component library or a new global style system.
- On narrow screens, tools remain reachable without overflowing the top bar;
  compact icon controls replace the removed select widths.

## Compatibility and rollback

- `localStorage.theme`, `localStorage['ui-style']`, and
  `localStorage['mango-language']` are unchanged.
- Revert the small shell, icon, i18n, and CSS changes to restore selects.
