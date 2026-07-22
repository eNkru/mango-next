# Component Guidelines (React shell)

> Shared React browse/shell contracts after the UI consistency pass (2026-07-21).

## Scope / Trigger

Apply when adding or changing shared UI under `frontend/src/shell/`,
`frontend/src/browse/`, form field markup, or page-level loading/error states.

## Signatures

```ts
// frontend/src/browse/BrowseComponents.tsx
PosterCard({
  item: BrowseTitle,
  actions?: ReactNode,
  showProgress?: boolean, // default true; TagDetail uses false
})

PosterCardSkeleton() // shimmer placeholder; same .mango-card geometry as PosterCard

BrowseToolbar({
  query, onQuery, mode, onMode, ascending, onAscending,
  extra?: ReactNode,
  modes?: SortMode[], // default all: natural|title|modified|progress
})

// frontend/src/browse/PosterRail.tsx
PosterRail({
  title: string,
  items: BrowseTitle[],
  loading?: boolean,      // default false → render PosterRailSkeleton
  skeletonCount?: number, // default 6
})

PosterRailSkeleton({
  title?: string,
  count?: number, // default 6
})

// frontend/src/shell/StatePanels.tsx
LoadingState({ message?: string })
EmptyState({ message?: string })
ErrorState({
  message?: string,
  onRetry?: () => void,
  retryLabel?: string, // caller should pass t('retry')
})

// frontend/src/shell/LanguageSelect.tsx
LanguageSelect({ className?: string }) // persists mango-language via useI18n

// frontend/src/shell/Icon.tsx
Icon({
  icon: LucideIcon,   // prefer icons.* from shell/icons.ts
  size?: number,      // default 18; nav/compact often 16
  className?: string,
  decorative?: boolean, // default true → aria-hidden
  label?: string,       // if set → role=img + aria-label
})

// frontend/src/shell/icons.ts
icons.home | library | tags | admin | logout | search | sortAsc | sortDesc
  | hide | show | edit | delete | add | download | play | continue
  | markRead | markUnread | selectAll | refresh | back | forward | close | users | scan
  | missing | readerControls | exit | save

// frontend/src/lib/i18n.tsx
t(key: MessageKey, vars?: Record<string, string | number>)
// templates use {name} placeholders, e.g. t('tagTitle', { tag })
```

## Contracts

### Icon system

- Dependency: `lucide-react` with **named imports only** (tree-shake). Never
  `import * as Lucide from 'lucide-react'`.
- Pages/shell use `Icon` + `icons` semantic map; avoid raw lucide components in
  product TSX unless adding a new semantic entry.
- **Density mix**:
  | Surface | Mode |
  |---------| | ---- |
  | Topbar nav, primary/secondary actions | icon + visible label |
  | Compact tools (password toggle, dialog close, tag remove, sort direction) | icon-only + `aria-label` on the **control** |
  | Brand | `mango-mark.svg` via `baseUrl('img/icons/mango-mark.svg')` + “Mango” text |
- Decorative icons next to text: leave default (`aria-hidden`).
- Icon-only: put accessible name on button/link; keep `Icon` decorative.
- Color via `currentColor` (inherits button/link theme).
- Button CSS: `.mango-btn` is `inline-flex` + `gap: 0.4rem`; `.mango-btn--icon`
  for square icon-only (min 2.25rem hit target).
- Do **not** reintroduce Font Awesome webfonts / UIkit.

### PosterCard / TagDetail adapter

- Prefer `PosterCard` over hand-written `.mango-card` markup on browse grids.
- Tag titles API returns a **narrow** card (`id, name, cover_url, entry_count, hidden`).
- Map to `BrowseTitle` on the client with safe defaults (`progress: 0`,
  `sort_name/file_name/display_name ← name`, `modified_at: 0`). Do **not** expand
  the backend tag titles response for card reuse alone.
- Pass `showProgress={false}` when progress is synthetic/unknown (TagDetail).

### BrowseToolbar modes

- Optional `modes` filters the sort `<select>`.
- TagDetail: `modes={['natural','title']}` only (no real modified/progress fields).
- Library / TitleDetail: omit `modes` (full four).
- If current `mode` ∉ allowed list, toolbar displays the first allowed mode.
- Search field shows a decorative search icon; sort direction is icon-only
  (`.mango-btn--icon` + aria-label for ascending/descending).

### Error / loading

- Pages with a load function: pass `onRetry` + `retryLabel={t('retry')}` to `ErrorState`.
- Do not place a sibling retry button next to `ErrorState` (Home used to; removed).
- Prefer callers pass translated `message` / labels; shared defaults stay neutral (`…` / `Retry`).
- Admin is an action panel: **no** full-page LoadingState. Scan/thumb **start** failures
  use `ErrorState` + `onRetry` under the card grid (see `react-admin.md`).

### PosterRail skeleton (CLS)

- Prefer rail-shaped skeleton over full-page `LoadingState` when the loaded UI is poster rails
  (Home: start-reading / recently-added).
- `PosterRail` with `loading` renders `PosterRailSkeleton` (same section +
  `.mango-poster-rail-shell` / `.mango-poster-rail` shells; no arrow buttons while loading).
- Skeleton cards: `PosterCardSkeleton` reuses `.mango-card` + `.mango-card__media`
  (`aspect-ratio: 2 / 3`) + body line heights aligned to title min-height / meta / progress.
- Shimmer uses theme tokens (`--mango-text-muted`, `--mango-bg-surface`); disable animation
  under `prefers-reduced-motion: reduce`.
- Do **not** invent a second card geometry for placeholders — mismatch causes CLS when data
  arrives.

### LanguageSelect

- Single control for AppShell topbar, Login footer, Reader top bar.
- Storage key: `localStorage['mango-language']` (`zh-cn` | `zh-tw` | `en`).
- Login does **not** use full AppShell; still must offer language switch.

### Form fields

Canonical markup:

```html
<label class="mango-field">
  <span>Label</span>
  <input class="mango-input" />
</label>

<label class="mango-field mango-field--inline">
  <input type="checkbox" />
  <span>Label</span>
</label>
```

- Password rows may nest `.mango-login__password-row` inside `label.mango-field`.
- CSS styles `label.mango-field > span` as the label text (not only `.mango-field label`).

### Inline styles

- Forbidden for static layout (use utilities: `.mango-actions--stack-sm`,
  `.mango-max-w-search`, `.mango-scroll-x`, `.mango-mt-1`, `.mango-ml-2`, `.mango-muted-copy`).
- Allowed dynamic only: ProgressBar `width: ${n}%`, Reader page margins.

## Validation & Error Matrix

| Condition | Behavior |
|-----------|----------|
| TagDetail missing tag in path | `ErrorState` with `t('missingTag')` |
| Load throws | `ErrorState` + `onRetry` re-invokes load |
| Admin scan POST fails | `actionError` → `ErrorState` retry `startScan` |
| ConfirmDialog open without labels | Neutral English defaults; prefer `t('delete')` / `t('cancel')` |

## Wrong vs Correct

#### Wrong
- Hand-roll tag cards that duplicate PosterCard DOM
- Show full sort modes on TagDetail (progress/modified are always 0)
- Hard-code Chinese strings in page TSX instead of `t()`
- Keep `/admin/react-preview` Placeholder playground in production routes
- Icon-only button without aria-label / title
- Import entire lucide-react barrel in a page

#### Correct
- Adapter + `PosterCard` + `modes={['natural','title']}` + `showProgress={false}`
- All user-visible page copy via `useI18n().t`
- Delete dead foundation preview page and route
- `<Icon icon={icons.edit} />` + label on main actions; icon-only + aria-label
  on compact tools
- AppShell brand: mark image (`alt=""`) + visible “Mango”

## Tests / checks

- `npm run typecheck` / `npm run build`
- Grep: no `PlaceholderPage` / `react-preview` in product sources
- Grep: `style={{` only on dynamic width/margin cases
- Grep: no Font Awesome / `fa-` webfont usage in product sources
- Manual: four theme combos + Login language switch + TagDetail grid + topbar
  icons + Login password toggle
