# Legacy frontend asset retirement

## Goal

Remove dead Go-template / jQuery / Alpine / UIkit assets and unused legacy page
handlers that no longer serve the migrated React surfaces, shrinking the
embedded `go/web` footprint without breaking remaining product routes.

## Confirmed Facts

- Migrated to React shell (`renderReactShell`): login, home, library, title,
  tags, tag-detail, reader, admin, user-list, user-edit, missing-items,
  react-preview.
- Still template-rendered via `renderLayout` today:
  - `plugin-download` (`GET /download/plugins`, only if `PluginPath` set)
  - `download-manager` (`GET /admin/downloads`)
  - `subscription-manager` (`GET /admin/subscriptions`)
- Dead code still in tree but not on live handlers for migrated pages:
  - `handleHomeLegacy` → `home.tmpl`
  - `handleTitleLegacy` → `title.tmpl`
  - Page scripts: `reader.js`, `admin.js`, `missing-items.js`, `user.js`,
    `user-edit.js`, `title.js`, and related page-only CSS where unused.
- Shared legacy chrome (`head`/`top`/`bottom`, jquery/alpine/uikit, mango.css)
  is only needed while **any** template page remains.
- Product decision: subscriptions + plugin download + download manager are
  **not needed**. This child will **disable those routes and delete** their UI
  assets. Backend plugin/queue APIs may remain for CLI/future use unless proven
  dead (API deletion is optional, not required).
- Parent: `07-17-frontend-react-vite`.

## Requirements

- Do not break React-migrated routes, static `/react/` assets, or OPDS XML.
- Disable browser routes for subscriptions, download manager, and plugin
  download; delete their templates and page scripts.
- Delete pure-migrated page assets (reader/admin/home/library/login/tags/users/
  missing) and unused legacy handlers (`*Legacy`, dead data builders).
- After no template pages remain (except possibly error helpers), remove shared
  legacy chrome and vendor JS/CSS that only those pages used — **if** nothing
  else references them (verify OPDS and static file server).
- Keep `//go:embed` valid.
- `go test ./...`, frontend typecheck/build stay green.
- Document remaining legacy surface (ideally only `react-shell.tmpl` +
  `public/react/` + OPDS + favicon/manifest).

## Acceptance Criteria

- [ ] Inventory of deleted vs kept paths is recorded in `design.md`.
- [ ] `/admin/subscriptions`, `/admin/downloads`, `/download/plugins` no longer
      serve the old UIs (404 or gone from router).
- [ ] Migrated React routes still load shell + bundle under BaseURL.
- [ ] OPDS endpoints still return XML.
- [ ] Dead page tmpl/js for migrated routes removed.
- [ ] Dead legacy handlers removed when unreferenced.
- [ ] Shared jquery/alpine/uikit/mango stacks removed **only if** no remaining
      template consumer; otherwise keep minimal set and document why.
- [ ] Build/test gates pass; binary embeds only remaining assets.

## Out of Scope

- Migrating deferred features to React (explicitly cancelled).
- Deleting all plugin/queue **API** endpoints (may stay for admin automation).
- Full CSS redesign / React token rewrite.
- Database migrations.

## Decisions

- Deferred product UIs: **disable routes + delete templates/scripts** (user
  confirmed). Not “keep rare template entry points.”
- Cleanup strategy: remove page-local dead assets first, then shared chrome if
  unreferenced, then dead Go handlers/structs.

## Open Questions

- (none blocking)
