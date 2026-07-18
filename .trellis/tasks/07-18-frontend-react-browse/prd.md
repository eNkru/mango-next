# React home, library, and title migration

## Goal

Migrate Mango's core authenticated browsing flow from Go templates, jQuery,
Alpine, and UIkit to the existing React + Vite shell while preserving the Go
backend, route URLs, user progress, admin controls, themes, and non-root
`BaseURL` behavior.

## Confirmed Facts

- The routes are `/`, `/library`, and `/book/{title}` and currently render
  `home.tmpl`, `library.tmpl`, and `title.tmpl` after authentication.
- The existing React shell already serves login, missing items, admin users,
  tags, and tag detail pages through route-specific `pageId` values.
- Existing authenticated JSON endpoints include `/api/library`,
  `/api/library/continue_reading`, `/api/library/start_reading`,
  `/api/library/recently_added`, and `/api/book/{tid}`.
- Existing mutation endpoints cover progress, bulk progress, hidden titles,
  display and sort names, tags, cover upload, and downloads, but their response
  contracts were not designed as one complete browse-page API.
- The current display-name mutation handler returns success without persisting
  anything; complete parity therefore requires a real backend implementation
  rather than only a React client.
- The legacy shell supports Simplified Chinese, Traditional Chinese, and
  English through a runtime language selector. The current React shell uses
  hard-coded Simplified Chinese and has no language selector.
- The legacy home page has new-user/empty-library states plus continue-reading,
  start-reading, and recently-added rails.
- The legacy library page has search, sort, per-title progress, admin-only
  hidden-title visibility, and hide/show actions.
- The legacy title page supports nested titles, entries, progress, search,
  sorting, entry actions, bulk progress, tags, and admin editing controls.
- The title detail migration must preserve complete administrator feature
  parity in this delivery; a read-only or browse-only first cut is not
  acceptable.
- Migrated routes must remain same-origin and `BaseURL`-aware; the final Go
  binary must not require Node at runtime.

## Requirements

- Render all three routes through the existing React shell without changing
  their public URLs or authentication boundary.
- Provide stable JSON response contracts for the data each page needs rather
  than injecting full page models into boot JSON.
- Preserve loading, empty, error, and successful data states on every page.
- Preserve links into the still-legacy reader route and all `BaseURL` prefixes.
- Preserve complete title-detail administrator behavior: title and entry
  display-name editing, sort-name editing, cover upload, tag maintenance,
  hidden/show controls, and bulk progress actions.
- Preserve comic/flat and light/dark presentation through existing React theme
  tokens and shell conventions.
- Keep unmigrated routes and OPDS behavior unchanged.
- Add focused Go contract/route tests and React component or behavior tests in
  proportion to the migrated interactions.
- Make display-name changes persist and be reflected by subsequent browse API
  responses.
- Add a shared React language selector with Simplified Chinese, Traditional
  Chinese, and English, preserving the selected language across navigation.
- Translate the shared shell and the new home, library, and title-detail pages
  in this task. Previously migrated React page bodies may remain Simplified
  Chinese and are not part of this task's translation acceptance scope.
- Use feature-equivalent layouts in the established React design language.
  Preserve the legacy information hierarchy and interactions, but do not
  require pixel-level reproduction of UIkit templates.

## Acceptance Criteria

- [ ] `/` renders React and correctly handles empty library, new reader,
      continue-reading, start-reading, and recently-added states.
- [ ] `/library` renders React with title covers/counts/progress, search, sort,
      and admin-only hidden-title controls.
- [ ] `/book/{title}` renders React with nested titles, entries, progress, and
      working links to reader and download routes.
- [ ] Administrators can edit title and entry display/sort names, upload title
      and entry covers, maintain tags, hide/show a title, and apply bulk
      read/unread progress without falling back to the legacy template.
- [ ] All retained mutations refresh or update the affected React state and
      surface API errors through the shared alert/error UI.
- [ ] All routes and API calls work under `/` and a non-root BaseURL such as
      `/mango/`.
- [ ] Existing migrated and legacy routes continue to work.
- [ ] No jQuery, Alpine, UIkit, or legacy page script is loaded by these three
      migrated routes.
- [ ] The shared React shell and the three browse routes can switch between
      Simplified Chinese, Traditional Chinese, and English without reloading
      legacy i18n code.
- [ ] Frontend build/typecheck and relevant Go tests pass.

## Out of Scope

- Migrating the reader itself.
- Migrating admin settings, subscriptions, plugin download, or OPDS.
- Replacing the established React shell or introducing a component library.
- Deleting legacy assets still used by unmigrated routes.

## Open Questions

- None.
