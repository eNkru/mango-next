# Design: restore icons UI/UX

## Architecture

```
package.json (lucide-react)
        │
        ▼
frontend/src/shell/Icon.tsx     # thin wrapper: size, stroke, aria
frontend/src/shell/icons.ts     # optional semantic map (nav/actions)
        │
        ├── AppShell / ReaderTopBar / BrowseToolbar / pages
        └── shell.css           # .mango-btn icon slots, brand mark
```

## Boundaries

| Layer | Owns | Does not own |
|-------|------|--------------|
| `Icon` | SVG size, stroke, currentColor, decorative vs labeled | Business labels, routing |
| Pages / shell | Which semantic icon + label density | Raw lucide imports preferred via Icon |
| Brand | `mango-mark.svg` via `baseUrl('img/icons/mango-mark.svg')` | Lucide brand icons |
| Spec | Icon + button icon conventions | Per-page pixel layouts |

## Icon component contract

```ts
// frontend/src/shell/Icon.tsx
type IconProps = {
  icon: LucideIcon;       // or semantic name if using map
  size?: number;          // default 18 (nav/btn), 16 compact
  className?: string;
  decorative?: boolean;   // default true → aria-hidden
  label?: string;         // if set, role=img + aria-label; decorative=false
};
```

Rules:

- Default: decorative next to visible text (`aria-hidden`).
- Icon-only control: put `aria-label` on the **button/link**, keep icon decorative.
- Color: `currentColor` so theme/accent inherit.
- Stroke: lucide default; comic theme may bump stroke slightly via CSS if needed.

## Density rules (UX)

| Surface | Mode | Examples |
|---------|------|----------|
| Topbar nav | icon + label | Home, Library, Tags, Admin |
| Primary/secondary actions | icon + label | Begin/Continue, Download, Save, Delete, New user |
| Compact tools | icon-only + aria-label | Password visibility, dialog close, tag remove |
| Reader chrome | icon + label for Exit/Controls; icon-only for close | ReaderTopBar / ReaderControls |
| Brand | mark image + text | AppShell brand link |

## Button CSS

Extend `.mango-btn` (and topbar links as needed):

- `.mango-btn` children: `inline-flex; align-items: center; gap: 0.4rem`
- Optional `.mango-btn--icon` for square icon-only (min 2.25rem hit target)
- Topbar brand: `img.mango-topbar__mark` ~24–28px, `gap` already 0.6rem

## Brand mark

- Path: `baseUrl('img/icons/mango-mark.svg')` (served from `go/web/public/`).
- Markup: `<img src=... alt="" class="mango-topbar__mark" />` + visible “Mango” text (mark decorative when text present).
- Do not duplicate as React SVG component unless needed later.

## Semantic icon mapping (initial)

| Semantic | Lucide | Where |
|----------|--------|-------|
| home | Home | nav |
| library | Library / BookOpen | nav, library |
| tags | Tags | nav |
| admin | Settings / Shield | nav |
| logout | LogOut | topbar |
| search | Search | BrowseToolbar |
| sortAsc / sortDesc | ArrowUpDown / ArrowUp / ArrowDown | BrowseToolbar |
| hide / show | EyeOff / Eye | Library, Title, Login password |
| edit | Pencil | Title, Users |
| delete | Trash2 | Users, Missing |
| add | Plus | tags, new user |
| download | Download | Title entries |
| play / continue | Play / BookOpen | Begin/Continue |
| markRead / unread | CheckCircle / Circle | progress actions |
| refresh | RefreshCw | lists |
| back | ArrowLeft | TagDetail, forms |
| close | X | dialogs, reader controls |
| users | Users | Admin card |
| scan | ScanSearch / FolderSearch | Admin |
| missing | FileWarning | Admin / Missing |
| reader controls | SlidersHorizontal | ReaderTopBar |
| exit | X / DoorOpen | Reader exit |

Exact lucide names finalized at implement time if a name is unavailable.

## Trade-offs

| Choice | Benefit | Cost |
|--------|---------|------|
| lucide-react | tree-shake, consistent stroke, React-first | new dependency |
| Thin Icon wrapper | one place for size/a11y | slight indirection |
| Density mix | readable + compact | needs clear rules in spec |
| Keep mark as static SVG URL | no bundle bloat | depends on public path + baseUrl |

## Compatibility / rollout

- Additive: no API/backend change.
- Build: `npm install lucide-react` at repo root (mango-frontend package).
- Rollback: revert frontend + package-lock; no DB/migration.

## Out of design

- Full icon button component library
- Animated icons
- Reader page-turn chevrons as permanent chrome
