# React home, library, and title migration design

## Boundaries

Go remains responsible for authentication, authorization, library state,
storage, file mutation, and embedded asset delivery. React owns rendering and
browser interaction for `/`, `/library`, and `/book/{title}`. The reader and all
other unmigrated routes remain server-rendered.

The three page handlers keep their current auth middleware and public paths but
return `views/react-shell` with these page IDs:

| Route | pageId |
|---|---|
| `/` | `home` |
| `/library` | `library` |
| `/book/{title}` | `title-detail` |

Boot JSON stays small: page identity, `BaseURL`, admin status, and title ID when
needed. Page content is loaded through authenticated same-origin JSON APIs.

## Read contracts

### Home

Add `GET /api/home` as a single coherent snapshot. It returns:

- `new_user`, `empty_library`, and admin/config guidance for empty states
- enriched `continue_reading`, `start_reading`, and `recently_added` arrays
- stable title/entry names, cover URLs, page counts, and progress percentages
- the existing eight-item section cap

The existing three `/api/library/*` endpoints remain registered for
compatibility; React does not fan out across partial responses.

### Library

Extend or normalize `GET /api/library?show_hidden=1` so it returns `is_admin`,
effective `show_hidden`, and title cards containing ID, display name, cover,
deep entry count, hidden flag, progress, and modified time. Hidden titles are
excluded unless an administrator explicitly requests them.

Search and sorting are client-side over the returned snapshot. Sort modes are
natural/automatic, title, modified time, and progress, each ascending or
descending. Query-string state is preserved for shareable/back-forward-safe
views. The legacy no-op `/api/sort_opt` contract is not expanded unless a
confirmed consumer requires server-persisted sort preferences.

### Title detail

Expand `GET /api/book/{tid}` without removing existing keys. The response
contains:

- title identity, file/display/sort names, cover, hidden state, tags, and
  breadcrumb parents
- direct child titles and entries with cover, count/pages, progress, modified
  time, and sort name
- `is_admin` for rendering privileged controls

Entry and child-title sorting/filtering remains client-side. A 404 JSON error
is rendered as the React error state without changing reader behavior.

## Mutation contracts

Keep existing authenticated/admin boundaries and reuse current endpoints for
hidden state, tags, progress, bulk progress, upload, and download.

Replace the unsafe path-encoded display-name write for React with a JSON-body
form (`PUT /api/admin/display_name/{tid}` with `name` and optional `eid`), while
keeping the old route registered during coexistence. Extend sort-name writes to
accept optional `eid` and persist either title or entry sort names.

Display names are real persistent library metadata:

1. Library helpers update `info.json` while preserving unrelated keys.
2. Title and entry constructors/cache hydration apply saved display names.
3. Successful mutations update the in-memory object under the library lock so
   the next API response changes immediately without a rescan.
4. Validation rejects empty names, missing titles/entries, and failed writes
   with non-2xx JSON errors.

Cover upload continues to use multipart form data and the existing size/MIME
limits. React shows busy state, reports failures through `AlertHost`, and reloads
the page snapshot after each successful mutation.

## Frontend structure

- `HomePage`, `LibraryPage`, and `TitleDetailPage` own page-level loading and
  mutations.
- Shared browse components cover poster cards, progress bars, search/sort
  controls, entry actions, edit dialog, upload control, and horizontal rails.
- Shared DTO types and sort/filter utilities live outside page components.
- Numeric progress and fixed card geometry prevent state changes from shifting
  the layout.
- Native controls and the existing dialog/button/card styling are used; no UI
  component library is introduced.

The home page keeps one prominent continue-reading item and compact horizontal
rails. Library and title detail use responsive poster grids. Title selection
and bulk actions remain visible and keyboard-operable without Alpine/UIkit.

## Localization

Add a small typed React localization provider with `zh-cn`, `zh-tw`, and `en`
dictionaries. It reads and writes the legacy `mango-language` localStorage key,
sets the document language/title, exposes interpolation, and renders a language
menu in `AppShell`.

Only `AppShell` and the three new page bodies are translated in this task.
Existing React page bodies may remain Simplified Chinese, but shared navigation
changes language consistently. Missing keys fall back to Simplified Chinese.

## Compatibility and rollback

- Every URL and middleware boundary remains unchanged.
- Asset and API URLs use the existing `baseUrl` helper.
- Existing API fields/routes remain available while new fields are additive.
- Legacy templates and scripts remain embedded for unmigrated pages.
- Route rollback is limited to switching the three page handlers back to
  `renderLayout`; backend contract additions remain harmless.

## Verification

- Go tests cover auth/admin boundaries, JSON shapes, hidden filtering,
  progress, display/sort persistence, BaseURL, and page shell IDs.
- Frontend checks cover type safety and production output; pure sort/filter and
  localization behavior should be isolated for focused tests where the current
  toolchain permits.
- Browser smoke checks cover desktop/mobile, comic/flat, light/dark, all three
  languages, empty/error states, and administrator mutations.
