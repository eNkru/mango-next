# React reader migration execution plan

## 0. Preconditions

- Parent: 07-17-frontend-react-vite
- Depends on completed browse child (title-detail already links into /reader/...)
- Working tree should start clean before implementation commits

## 1. Backend bootstrap and route switch

- [x] Add GET /api/reader/{tid}/{eid} returning title/entry, pages, progress, dimensions, sibling entries, exit/next/previous URLs (BaseURL-aware).
- [x] Reuse existing library/entry lookup and dimensions helpers; structured 404/400 on missing/corrupt/empty entries.
- [x] Switch handleReader success path to renderReactShell with pageId reader and boot { tid, eid, page }.
- [x] Keep redirect /reader/{tid}/{eid} to .../1 unchanged.
- [x] Add focused Go tests for bootstrap success/error and React shell pageId registration for reader routes.

## 2. Shared frontend foundations for reader

- [x] Add reader DTOs/types and pure helpers (readerMath: clamp page, URL/index conversion, RTL direction, progress-throttle decision).
- [x] Add unit tests for pure helpers.
- [x] Add useReaderPrefs with mango.reader.* keys and defaults matching legacy behavior.
- [x] Extend i18n catalogs (zh-CN / zh-TW / en) for reader chrome strings.

## 3. Reader UI and interactions

- [x] Implement ReaderPage load/error/empty gates from bootstrap.
- [x] Implement immersive layout: ReaderTopBar auto-hide on top-edge hover / intentional open, animated show/hide, idle hide.
- [x] Implement continuous and paged viewports with fit/margin/RTL/flip-animation preferences.
- [x] Wire keyboard, click-zone, control-panel navigation; sync URL via replace.
- [x] Preload lookahead images via existing page endpoint.
- [x] Wire progress save + complete-on-exit/next-entry.
- [x] Register reader pageId in App.tsx without wrapping in AppShell.

## 4. Validation and review

- [x] npm run typecheck and npm run build in frontend/.
- [x] Focused Go tests for reader API/route; then make check, make test, make build as available.
- [ ] Smoke root and non-root BaseURL: open reader from title-detail, deep link to mid-page, flip pages, switch mode, next entry, exit.
- [ ] Smoke auto-hide bar, language switch.
- [ ] Confirm unmigrated routes still template-render; no jQuery/Alpine/UIkit on reader shell.
- [x] Leave legacy reader.tmpl / reader.js in place until smoke accepted.

## Risk and rollback points

- Highest risk: 1-based URL page vs 0-based /api/page index mismatch.
  - Runtime note: live `/api/page` + `ReadPage` are **1-based** (legacy reader.js used `i+1`); React matches that.
- Progress throttle regressions: port decision table from legacy and unit-test.
- Auto-hide bar must not steal clicks from page turn zones.
- Rollback: restore handleReader template render; do not delete legacy assets in this child.
