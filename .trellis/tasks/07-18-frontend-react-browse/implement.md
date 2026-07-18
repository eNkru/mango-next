# React home, library, and title migration execution plan

## 1. Backend metadata and contracts

- [x] Add library metadata helpers for title/entry display names that preserve
      unrelated `info.json` fields and update loaded objects safely.
- [x] Apply persisted display names during fresh scans and cache hydration.
- [x] Implement JSON-body display-name mutations and title/entry sort-name
      mutations with validation, admin checks, and compatibility routes.
- [x] Add one enriched home snapshot API.
- [x] Normalize library data with hidden filtering, deep counts, progress, and
      sort fields.
- [x] Expand book detail data with breadcrumbs, nested titles, entries, tags,
      progress, timestamps, sort names, and admin status.
- [x] Add focused storage/library/server tests for persistence and contracts.

## 2. Shared React foundations

- [x] Add typed browse DTOs and shared sort/filter helpers.
- [x] Add the three-language provider using `mango-language`, interpolation,
      document language/title updates, and Simplified Chinese fallback.
- [x] Add the language selector to `AppShell` and translate shared navigation.
- [x] Add reusable poster card, progress, rail, search/sort, dialog, entry
      action, and upload controls using existing CSS tokens.
- [x] Extend responsive comic/flat light/dark styles without importing legacy
      UIkit/jQuery CSS behavior.

## 3. Route migrations

- [x] Switch `/` to page ID `home` and implement all onboarding/empty states,
      continue-reading hero, start-reading rail, and recently-added rail.
- [x] Switch `/library` to page ID `library` and implement search, all sort
      modes, progress cards, hidden visibility, and hide/show mutations.
- [x] Switch `/book/{title}` to page ID `title-detail` and implement
      breadcrumbs, tags, nested titles, entries, search/sort, selection, reader
      and download actions, and progress mutations.
- [x] Implement complete administrator edit flows for title/entry display name,
      sort name, cover upload, hidden state, tags, and bulk progress.
- [x] Register the page IDs in `App.tsx` and keep all links/API calls
      `BaseURL`-aware.

## 4. Validation and review

- [x] Run `npm run typecheck` and `npm run build`.
- [x] Run focused Go tests during implementation, then `make check`,
      `make test`, and `make build`.
- [x] Verify root and `/mango/` mounts for all three routes and API calls.
- [ ] Smoke-test desktop/mobile in comic/flat and light/dark modes.
- [ ] Smoke-test Simplified Chinese, Traditional Chinese, and English.
- [ ] Exercise loading, empty, error, non-admin, and admin mutation states.
- [x] Confirm the three React shells load no jQuery, Alpine, UIkit, or legacy
      page scripts and that unmigrated pages still render.

## Risk and rollback points

- Metadata writes are the highest-risk backend change: preserve unknown
  `info.json` keys, validate paths/IDs through loaded library objects, and test
  restart/cache behavior before switching routes.
- Keep legacy display-name and sort-name routes until all legacy pages using
  them are retired.
- Commit-ready rollback is route-local: restore `renderLayout` calls for home,
  library, and title while leaving additive APIs and React components unused.
- Do not delete templates, scripts, or dependencies still required by reader,
  admin, subscriptions, plugins, or OPDS.
