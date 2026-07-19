# React reader migration design

## Boundaries

Go remains responsible for authentication, library lookup, page image serving,
progress persistence, and embedded static delivery. React owns reader rendering,
navigation math, preference storage, and immersive chrome for:

| Route | pageId | Notes |
|---|---|---|
| `/reader/{title}/{entry}` | n/a (redirect) | Keep existing 302 to page 1 |
| `/reader/{title}/{entry}/{page}` | `reader` | React immersive reader |

Unmigrated routes stay on Go templates. Image bytes stay on
`GET /api/page/{tid}/{eid}/{page}` and are never embedded in HTML/JSON.

Reader does **not** use shared `AppShell` (nav + page header). It mounts under
the same React providers (i18n / theme / alerts) with a dedicated immersive
layout and auto-hide top bar.

## Data flow

```
GET /reader/{tid}/{eid}            → 302 → /reader/{tid}/{eid}/1
GET /reader/{tid}/{eid}/{page}     → react-shell boot { pageId: "reader", tid, eid, page, baseURL, isAdmin }
                                     React ReaderPage
                                       → GET /api/reader/{tid}/{eid}
                                       → render continuous | paged
                                       → GET /api/page/... for visible + preload images
                                       → PUT /api/progress/... on throttle / complete
```

Boot JSON stays small (route identity only). Content metadata comes from the
bootstrap API so error states (missing title/entry, zero pages) are explicit
JSON rather than blank shells.

## Bootstrap contract

`GET /api/reader/{tid}/{eid}` (auth required)

Success shape:

```json
{
  "success": true,
  "data": {
    "title": { "id": "...", "name": "..." },
    "entry": {
      "id": "...",
      "name": "...",
      "pages": 120,
      "progress": 12
    },
    "dimensions": [{ "width": 1000, "height": 1500 }],
    "entries": [
      { "id": "...", "name": "...", "pages": 100, "progress": 0 }
    ],
    "exit_url": "/book/{tid}",
    "next_entry_url": "/reader/{tid}/{nextEid}/1",
    "previous_entry_url": "/reader/{tid}/{prevEid}/1"
  }
}
```

Rules:

- One-based page numbers in public URLs; zero-based indices only inside pure
  frontend helpers and when calling `GET /api/page` (legacy page endpoint is
  zero-based — preserve that).
- `dimensions.length === entry.pages`; empty/corrupt entry returns structured
  error, not empty success.
- Adjacent URLs and `exit_url` are BaseURL-prefixed absolute path strings (same
  pattern as other React APIs), or empty string when absent.
- Sibling `entries` list is the same parent title order used by legacy reader
  for entry jump.
- Dimensions may be computed with the existing dimensions helper; do not invent
  a second page-count source of truth.

Error shape: existing `sendJSONError` contract (`success: false`, message,
status 404/400).

## Frontend architecture

```
frontend/src/
  pages/reader/
    ReaderPage.tsx          # route owner: load bootstrap, gate states
    ReaderViewport.tsx      # continuous vs paged host
    ReaderTopBar.tsx        # auto-hide immersive bar
    ReaderControls.tsx      # modal: mode, prefs, jump, entry, exit
    useReaderBootstrap.ts
    useReaderNavigation.ts  # page index, URL sync, keyboard/click zones
    useReaderProgress.ts    # throttle + complete
    useReaderPrefs.ts       # mango.reader.* localStorage
    readerMath.ts           # pure helpers + unit tests
    types.ts
```

### Chrome / interaction

- Default: full-viewport black reader, top bar hidden.
- Show top bar when pointer enters a top-edge hit zone (~24–40px), or when the
  control panel is opened / Escape toggles chrome intentionally.
- Animate opacity/transform (~150–200ms); auto-hide after short idle when pointer
  leaves the bar and controls are closed.
- Language selector lives in the top bar (reuse shared language control).
- Control panel retains legacy capabilities: mode, fit, margin, RTL, flip
  animation, preload, page jump, entry jump, prev/next entry, exit.

### Navigation / progress invariants

- Single source of truth for current page index in `useReaderNavigation`.
- URL page updates via history.replaceState (no full reload); invalid page
  clamps to `[1, pages]`.
- Keyboard: left/right and j/k; direction inverted when RTL is on in paged mode.
- Progress save rules match legacy intent:
  - save when first/last page
  - save when distance from last saved page >= 5
  - always allow save for long-page titles (avg height/width ratio > 2)
  - next-entry / exit mark complete (`pages`) before navigate
- Prefs keys:
  - `mango.reader.mode`
  - `mango.reader.margin`
  - `mango.reader.fitType`
  - `mango.reader.preloadLookahead`
  - `mango.reader.enableFlipAnimation`
  - `mango.reader.enableRightToLeft`
- Do not read legacy unprefixed keys.

### Image loading

- Page images: `baseURL + "api/page/" + tid + "/" + eid + "/" + zeroBasedIndex`
- Preload next N pages in paged mode using the same URL builder.
- Continuous mode renders the strip with dimension-driven layout; do not block
  first paint on every image.

## Go route wiring

- Keep `handleReaderNoPage` redirect as-is.
- Change `handleReader` success path to `renderReactShell` with
  `pageId: "reader"` and boot fields `{ tid, eid, page }`.
- Prefer React error state after shell loads for missing/corrupt entries.
- Register `GET /api/reader/{tid}/{eid}` next to existing reader APIs.
- Leave `reader.tmpl` / `reader.js` on disk for rollback until smoke passes.

## Compatibility and rollout

- Public URLs unchanged.
- Additive API only; no breaking change to `/api/page`, `/api/dimensions`,
  `/api/progress`.
- Title-detail "Read" links already point at `/reader/...` and need no change.
- BaseURL must work for shell assets, bootstrap, page images, and adjacent URLs
  under root and non-root mounts.

## Tradeoffs

| Choice | Why | Cost |
|---|---|---|
| Dedicated bootstrap API | One coherent first paint + testable errors | Small new endpoint |
| Immersive chrome + auto-hide bar | Matches reading UX; keeps language/exit reachable | More UI state than plain AppShell |
| Namespaced prefs | Avoid global key collisions | One-time loss of legacy prefs |
| Keep zero-based page API | No backend image-route rewrite | Frontend must convert 1-based URL vs 0-based API carefully |

## Rollback

Route-local: restore `handleReader` to template render and stop registering the
React `reader` pageId. Leave bootstrap API in place (unused) or delete later.
Do not delete templates/scripts until smoke is accepted.
