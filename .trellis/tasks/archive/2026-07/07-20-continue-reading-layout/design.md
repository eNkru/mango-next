# Design: 主页继续阅读区布局优化

## Summary

Replace the multi-column wide-card grid with a **hero + compact list** pattern on the home page continue-reading section. Frontend-only; reuse `api/home` data and existing `ProgressBar`.

## Boundaries

| In | Out |
|----|-----|
| `HomePage.tsx` continue section structure | Backend `apiHome` / `GetContinueReading` |
| New CSS for hero + list (+ remove/retire old continue-grid usage) | Independent history page |
| Expand/collapse local UI state | Deep link to book detail from continue card |
| i18n keys for show more / show less | Comic/Flat dual-theme LESS legacy pages |

## Layout

```
section.continue-reading
├── h2 继续阅读
├── article.hero          ← items[0]
│   ├── img cover
│   └── body: title, page text, ProgressBar (thick), [继续]
└── ul.list               ← items[1..] (omit if empty)
    ├── li.row × min(3, n) or all when expanded
    └── button.expand     ← only if items[1..].length > 3
```

### Hero (primary)

- Horizontal: cover (~100–120px) + content column
- Progress bar visually thicker than list rows (CSS modifier, e.g. `.mango-progress--lg` or hero-scoped height)
- Single primary CTA → `reader/{title_id}/{id}`
- No secondary “打开”

### List row (secondary)

- Compact row: thumb (~48–56px) | title (ellipsis) | thin ProgressBar
- Entire row is a link/button to the same reader URL
- No page text, no action buttons

### Expand

- `const LIST_PREVIEW = 3`
- `const rest = items.slice(1)`
- Default: `rest.slice(0, 3)`; expanded: full `rest`
- Toggle button only when `rest.length > 3`
- Local `useState(false)` for expanded; no URL state

## Data / contracts

- Input: `BrowseEntry[]` from existing `HomeResponse.continue_reading` (max 8, ordered by progress updated_at desc — unchanged)
- Reader URL helper: same as today via `baseUrl(\`reader/${encodeURIComponent(title_id)}/${encodeURIComponent(id)}\`)`
- Page label: keep existing `page > 0 ? \`${page} / ${pages} ${t('page')}\` : \`${pages} ${t('page')}\`` on hero only

## Component shape

Prefer keeping logic in `HomePage.tsx` (or small local helpers in same file) unless extraction is trivial:

- Optional local components: `ContinueHero`, `ContinueRow`, `ContinueSection`
- Reuse `ProgressBar` from `BrowseComponents.tsx`
- Do **not** reuse `PosterCard` / `mango-poster-rail` (explicit product requirement)

## CSS

| Class (proposed) | Role |
|------------------|------|
| `.mango-continue` | section stack (gap) |
| `.mango-continue-hero` | hero card grid (cover + body) |
| `.mango-continue-list` | vertical list, no multi-column grid |
| `.mango-continue-row` | compact clickable row |
| `.mango-continue-more` | text button for expand/collapse |

- Remove usage of `.mango-continue-grid` multi-column layout for this section
- Retire or leave unused old `.mango-continue-card` rules if fully replaced (prefer delete dead CSS in same change)
- Comic theme: keep sharp corners via existing comic selectors; update selectors if class names change
- Mobile: hero may stack or shrink cover; list rows stay single-line

## i18n

Add keys in `zh-cn` / `zh-tw` / `en`:

| Key | zh-cn | en |
|-----|-------|-----|
| `showMore` | 展开更多 | Show more |
| `showLess` | 收起 | Show less |

Optional: `showMoreCount` with `{count}` if we want “还有 N 本” — not required for MVP; plain show more/less is enough.

## Compatibility

- Empty library / new user / empty continue: unchanged (section hidden when length 0)
- start_reading / recently_added rails: untouched
- API clients: no change

## Trade-offs

| Choice | Why | Alternative rejected |
|--------|-----|----------------------|
| Hero + list vs pure rail | Differentiates from start/recent; better progress scan | Pure poster rail |
| Frontend-only expand | No backend change; keeps depth 8 | Lower API limit |
| Whole-row link on list | Low density | Per-row buttons |
| Local expand state | Simple | Persist expand preference |

## Rollback

Revert `HomePage.tsx`, `shell.css` continue rules, and i18n keys; no data migration.
