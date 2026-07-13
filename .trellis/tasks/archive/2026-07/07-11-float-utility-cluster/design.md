# Design: Float Utility Cluster

## Architecture

### Markup (layout.html.ecr)

Replace standalone `.github-float` with a cluster root:

```html
<div class="utility-fab" id="utility-fab">
  <button type="button" class="utility-fab__primary" aria-expanded="false" aria-controls="utility-fab-menu" …>
    <i class="fas fa-sliders-h utility-fab__icon-rest"></i>
    <i class="fas fa-times utility-fab__icon-open" hidden></i>
  </button>
  <ul id="utility-fab-menu" class="utility-fab__menu" hidden>
    <li><button type="button" class="utility-fab__item lang-toggle" …><span class="lang-toggle-label">简</span></button></li>
    <li><button type="button" class="utility-fab__item" onclick/handler theme>…</button></li>
    <li><button type="button" class="utility-fab__item" … style>…</button></li>
    <li><a class="utility-fab__item" href="github…" target="_blank" rel="noopener noreferrer">…</a></li>
    <li><a class="utility-fab__item" href="<%= base_url %>logout">…</a></li>
  </ul>
</div>
```

- Prefer **button** for toggles (not `<a onclick>`) for a11y; keep links for GitHub/logout.
- Remove desktop `.sidebar-footer` block entirely.
- Remove mobile offcanvas lang/theme/style/logout + trailing divider if it only separated those actions (keep nav).

### CSS

| Layer | File | Role |
|-------|------|------|
| Position + flat look | `mango.less` | Fixed top-right; soft circle primary + items; open stack gap; `.uk-light` dark flat |
| Comic overrides | `comic-theme.less` | Kirby border/hard shadow on primary + items under `comic-theme` / `comic-theme-dark` |
| Retire | `mango.less` / comic | Standalone `.github-float` rules → map to `.utility-fab__primary` / `__item` or remove after migrate |

**Open state**: `.utility-fab.is-open` shows menu; items animate with simple translate/opacity (respect `prefers-reduced-motion`).

**z-index**: keep ~1100 so above content, similar to current float.

### JS

Small module in `public/js/common.js` (or tiny `utility-fab.js` if preferred — **prefer common.js** to avoid new script tag unless file already growing badly):

- `toggleUtilityFab(open?: boolean)`
- Click primary → toggle
- Document click outside → close
- `keydown` Escape → close
- Satellite click (capture/bubble) → run action then close (links navigate naturally)
- Set `aria-expanded`, swap icon visibility / `hidden` on open icons
- Expose on `window` only if needed for inline handlers; prefer delegated listeners on `#utility-fab`

No change to theme/lang storage logic.

### Data / contracts

- Unchanged: `localStorage` theme / ui-style / language.
- GitHub URL unchanged: `https://github.com/eNkru/mango-next`.

### Compatibility

- Feature branch may already have github-float theming; this design **supersedes** standalone float markup while reusing visual tokens.
- Admin theme dropdowns still work (same storage).

### A11y

- Primary: `aria-expanded`, `aria-controls`, `aria-label` (i18n key e.g. `utility_menu` / open+close can share “工具菜单”).
- Menu: `role="menu"` optional; items `role="menuitem"` if menu pattern, or keep toolbar pattern with labeled buttons — **recommend simple group of labeled buttons** without full ARIA menu keyboard roving for MVP.
- Focus: on open, optional focus first item; on close via Esc, return focus to primary.

### Trade-offs

| Choice | Why | Cost |
|--------|-----|------|
| Speed-dial vs popover list | User chose A | Custom open/close CSS |
| Icon-only satellites | Less clutter | Tooltips required |
| Remove all duplicates | Single source of truth | Users relearn location once |
| Buttons for toggles | Better a11y than `<a onclick>` | Slight markup change |

### Rollback

Revert layout + CSS + JS; restore sidebar-footer and offcanvas actions; restore standalone github-float if needed.

### Risks

- Mobile: FAB under navbar must not collide with hamburger (left) — right side OK.
- Long lang label “EN” vs “简” sizing — keep min hit target 40×40 (36 mobile).
- Content under top-right cards: existing float already occupies space; stack when open is temporary.
