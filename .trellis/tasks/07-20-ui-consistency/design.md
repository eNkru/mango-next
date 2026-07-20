# Design: UI Consistency Audit and Fix

## Scope Boundaries

| In | Out |
|----|-----|
| React frontend under `frontend/src/**` | Business data API shape changes |
| Shared shell / browse components | Reader rendering math / navigation logic |
| `tokens.css` + `shell.css` | Mobile layout redesign |
| Delete admin `react-preview` scaffold (page + route + docs) | New features / new pages |
| i18n dictionary expansion | Crystal / legacy LESS theme pages |

## Architecture

```
tokens.css          → design tokens (theme + reader + semantic colors)
shell.css           → layout, components, utility classes
BrowseComponents    → PosterCard (showProgress), BrowseToolbar (modes)
StatePanels         → Loading / Empty / Error (+ onRetry)
LanguageSelect?     → optional shared language control (AppShell + Login)
pages/*             → consume shared pieces; no local card/toolbar clones
App.tsx + Go route  → drop react-preview
```

Implementation order (dependency-safe):

1. **Tokens & CSS foundation** (R4 partial, R5, R6)
2. **Shared component API** (PosterCard, BrowseToolbar, ErrorState, LanguageSelect, form field CSS)
3. **i18n dictionary** keys for all target pages + ConfirmDialog defaults strategy
4. **Page migrations** (R1 TagDetail first among pages, then R2 pages, Admin R3)
5. **Remove preview scaffold** (R8)
6. **Theme smoke + typecheck/build**

## Contracts

### PosterCard

```ts
// existing + showProgress
function PosterCard({
  item,
  actions,
  showProgress = true,
}: {
  item: BrowseTitle;
  actions?: ReactNode;
  showProgress?: boolean;
})
```

- TagDetail: `showProgress={false}`
- Home / Library: default (progress bar shown)

### BrowseToolbar

```ts
modes?: SortMode[]  // default: ['natural','title','modified','progress']
```

- Render only options present in `modes`
- If `mode` ∉ `modes`, parent should clamp; toolbar may also clamp on change options for safety
- TagDetail: `modes={['natural','title']}`

### TagDetail → BrowseTitle adapter

```ts
function toBrowseTitle(card: TagApiTitle): BrowseTitle {
  return {
    id: card.id,
    name: card.name,
    display_name: card.name,
    file_name: card.name,
    sort_name: card.name,
    cover_url: card.cover_url,
    entry_count: card.entry_count,
    progress: 0,
    modified_at: 0,
    hidden: card.hidden,
  };
}
```

No backend change. Sorting by title/natural only is honest.

### ErrorState

```ts
function ErrorState({
  message,
  onRetry,
  retryLabel,
}: {
  message?: string;
  onRetry?: () => void;
  retryLabel?: string;
})
```

- Prefer **callers pass translated** `message` / `retryLabel` (StatePanels stays free of hard-coded Chinese defaults long-term)
- Defaults: keep neutral short fallbacks or empty; pages always pass `t(...)` where possible
- When `onRetry` set: render `mango-btn` inside error block
- Home: remove sibling retry button

### LanguageSelect

Extract from AppShell the language `<select>` into a small component (or shared render helper) used by:

- AppShell topbar
- Login card footer/header

Same storage key `mango-language` via existing `setLanguage`.

### Form field markup

Canonical:

```html
<label class="mango-field">
  <span>Label text</span>
  <input class="mango-input" />
</label>
```

Checkbox:

```html
<label class="mango-field mango-field--inline">
  <input type="checkbox" />
  <span>Label</span>
</label>
```

Login password row: outer `label.mango-field` + span; control is the existing `.mango-login__password-row` wrapper (not a raw input). CSS must allow nested control containers.

Update `.mango-field` CSS so `label.mango-field > span` is the label text style (today rules target `.mango-field label` which assumes div-wrapper pattern).

### Tokens (R5 / R6)

Add to `tokens.css` (names illustrative; keep `--mango-*` prefix):

| Token | Purpose |
|-------|---------|
| `--mango-danger` / `--mango-danger-hover` | destructive buttons (flat vs comic scopes) |
| `--mango-success` | success alert border |
| `--mango-on-accent` | primary button text (was `#fff`) |
| `--mango-reader-bg` | reader page background |
| `--mango-reader-fg` | reader text |
| `--mango-reader-chrome` | topbar/chrome surface |
| `--mango-reader-border` | chrome border |
| `--mango-reader-ghost-border` | ghost button border |
| `--mango-reader-img-bg` | strip image placeholder bg |
| comic ink black utilities | prefer tokens where repeated `#000` for shadows is intentional comic ink — optional `--mango-ink: #000` |

Reader **does not** switch with flat/comic; reader tokens can live under `:root` as fixed immersive dark values.

### Buttons (R6)

```css
/* danger uses --mango-danger, not accent */
.mango-btn--danger { border-color: var(--mango-danger); color: var(--mango-danger); }

html.comic-theme .mango-btn,
html.comic-theme-dark .mango-btn {
  border-width: 2px; /* or 3px to match panels */
  border-color: #000; /* or var(--mango-ink) */
  box-shadow: 3px 3px 0 #000;
}
/* primary/danger comic variants keep fill + thick border readable */
```

ConfirmDialog already uses `mango-btn--danger` for confirm — color change applies automatically.

### Utility classes (R4)

Minimal set only where inline styles repeat:

| Class | CSS |
|-------|-----|
| `.mango-actions--flush` | margin-top: 0 (toolbar action rows) |
| `.mango-actions--stack-sm` | margin-top: 0; margin-bottom: 1rem |
| `.mango-max-w-search` | max-width: 16rem (or 18rem — pick one, TagsIndex 18rem → normalize 18rem) |
| `.mango-scroll-x` | overflow-x: auto |
| `.mango-mt-1` | margin-top: 1rem |
| `.mango-ml-2` | margin-left: 0.5rem |
| `.mango-muted-copy` | color: var(--mango-text-muted); line-height: 1.6 |

Do **not** add a full spacing scale.

Keep dynamic inline: ProgressBar width %, ReaderViewport margins.

### Admin action error (R3-B)

State:

```ts
const [actionError, setActionError] = useState<null | { kind: 'scan' | 'thumb'; message: string }>(null);
```

On startScan / startThumbnails failure: set actionError; render:

```tsx
{actionError ? (
  <ErrorState
    message={actionError.message}
    onRetry={() => void (actionError.kind === 'scan' ? startScan() : startThumbnails())}
    retryLabel={t('retry')}
  />
) : null}
```

Prefer ErrorState over duplicate danger alert for the same failure (or alert only for poll mid-run errors).

### R8 Delete preview

| Remove / change | Path |
|-----------------|------|
| Delete file | `frontend/src/pages/PlaceholderPage.tsx` |
| App switch | drop `react-preview` case |
| Go handler | `handleReactPreview` |
| Go route | `GET /admin/react-preview` |
| boot default | `DEFAULT_BOOT.pageId/pageName` → `'home'` |
| Docs | `FRONTEND_DEV_GUIDE.md` |

No test hard-depends on react-preview beyond docs (verify with grep).

## Data Flow

TagDetail load:

```
GET api/tags/{tag}/titles[?show_hidden=1]
  → titles[] (narrow)
  → map toBrowseTitle
  → filterBrowseItems / sortBrowseItems (modes natural|title)
  → PosterCard grid
```

i18n:

```
I18nProvider → useI18n().t
Login LanguageSelect → setLanguage → re-render Login strings
```

## Compatibility

- Library / TitleDetail / Home: default props preserve behavior
- Tag titles API: unchanged
- Reader visuals: token swap only; pixel-close dark chrome
- Comic: sharper buttons; danger red independent of Netflix/comic accent

## Trade-offs

| Choice | Benefit | Cost |
|--------|---------|------|
| Frontend title adapter | No backend work | Tag cards lack real progress |
| modes filter | Honest toolbar | Small API surface on toolbar |
| showProgress prop | Clean tag grid | One more prop |
| ErrorState onRetry | Consistent retry UX | Slightly heavier StatePanels |
| Delete preview | Less dead code | Lose quick component playground |

## Rollback

- Pure frontend + one admin GET route removal
- Revert single PR / commit; no DB migration
- If comic button skin too heavy: CSS-only revert of R6 block

## Validation Surfaces

- `npm run typecheck` / `npm run build`
- Manual: 4 theme combos on Library + TagDetail + Login + Admin + Reader open/close chrome
- Grep gates: no PlaceholderPage; reduced `style={{` in pages; no hard-coded user strings on R2 pages
