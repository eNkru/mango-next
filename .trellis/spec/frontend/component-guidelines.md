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

BrowseToolbar({
  query, onQuery, mode, onMode, ascending, onAscending,
  extra?: ReactNode,
  modes?: SortMode[], // default all: natural|title|modified|progress
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

// frontend/src/lib/i18n.tsx
t(key: MessageKey, vars?: Record<string, string | number>)
// templates use {name} placeholders, e.g. t('tagTitle', { tag })
```

## Contracts

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

### Error / loading

- Pages with a load function: pass `onRetry` + `retryLabel={t('retry')}` to `ErrorState`.
- Do not place a sibling retry button next to `ErrorState` (Home used to; removed).
- Prefer callers pass translated `message` / labels; shared defaults stay neutral (`…` / `Retry`).
- Admin is an action panel: **no** full-page LoadingState. Scan/thumb **start** failures
  use `ErrorState` + `onRetry` under the card grid (see `react-admin.md`).

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

#### Correct
- Adapter + `PosterCard` + `modes={['natural','title']}` + `showProgress={false}`
- All user-visible page copy via `useI18n().t`
- Delete dead foundation preview page and route

## Tests / checks

- `npm run typecheck` / `npm run build`
- Grep: no `PlaceholderPage` / `react-preview` in product sources
- Grep: `style={{` only on dynamic width/margin cases
- Manual: four theme combos + Login language switch + TagDetail grid
