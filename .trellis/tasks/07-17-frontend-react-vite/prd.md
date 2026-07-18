# React Vite frontend migration

## Goal

Replace Mango's current Go-template, jQuery, Alpine, and UIkit browser UI with a
React + Vite frontend while keeping the Go server as the only backend, API host,
authenticator, and single-binary deployment target.

## Confirmed Facts

- The current UI is server-rendered Go `html/template` under `go/web/views/`,
  with page scripts under `go/web/public/js/` and styles under
  `go/web/public/css/`.
- Runtime dependencies include jQuery, Alpine, UIkit, Select2, Moment, Font
  Awesome, and dual comic/flat theme CSS.
- Go embeds `go/web/views` and `go/web/public` at compile time; the production
  binary does not depend on a Node runtime.
- Browser routes already include home, library, title, tags, reader, login,
  admin, users, subscriptions, missing items, plugin download, OPDS, and the
  disabled download manager.
- Existing JSON API coverage is partial; many pages still rely on server-rendered
  template data rather than complete client-ready contracts.
- Authentication, authorization, `BaseURL` mounting, SQLite, library scanning,
  queue storage, and background tasks already live in Go and do not need to be
  rewritten for a React frontend.
- The in-progress `07-17-frontend-asset-pipeline` work introduced npm/lockfile,
  deterministic generation, Docker Node stages, and local font hosting. That
  work is redirected into the React build foundation rather than treated as the
  final jQuery/LESS inventory end state.
- The missing-items browser page already exists at `/admin/missing`, but its
  JSON handlers currently return empty success stubs rather than real missing
  inventory data.

## Task Map

- `07-18-frontend-react-foundation` (P1): React + Vite + TypeScript app shell,
  Go HTML shell mounting, BaseURL-aware assets, dual-theme tokens, Make/Docker
  generation order, and replacement of the old asset-pipeline end state.
- `07-18-frontend-react-missing-items` (P1): migrate `/admin/missing` to React,
  implement real missing-items JSON contracts, and prove the foundation with one
  complete page.
- `07-18-frontend-react-browse` (P1): migrate the authenticated home, library,
  and title-detail browsing flow with complete title administrator parity and
  shared React localization.

The recommended next child is `frontend-react-reader`. It closes the core
browse-to-read workflow before lower-frequency administration pages and keeps
the migration boundary to the two existing `/reader/{title}/{entry}[/{page}]`
routes plus their reader-specific APIs.

Later page migrations remain follow-up children and are out of the first
delivery.

## Requirements

- Use React + Vite + TypeScript as the browser UI stack.
- Keep Go as the only long-running backend and final deployable binary.
- Build React static assets at development and release time, then serve and embed
  them through Go with no public CDN runtime dependency.
- Support non-root `BaseURL` mounting for routes and static assets.
- Structure the work as a parent migration with independently verifiable child
  milestones.
- First delivery is foundation shell plus one pilot page, not a multi-page
  cutover.
- Pilot page is the admin missing-items flow at `/admin/missing`.
- Preserve comic/flat dual themes and light/dark switching in the React shell
  and pilot; reuse existing visual tokens/intent rather than inventing a new
  design system.
- During migration, only migrated routes render React. Unmigrated routes keep Go
  templates and the existing chrome. Navigation may link across both sides.
- Migrated routes use a Go-served lightweight HTML shell that loads one React
  bundle; React mounts the page for that path.
- Style React with local CSS and tokens rather than a full component library.
- Provide stable JSON API contracts for each migrated page.
- Leave disabled download-manager product cleanup as separate work.
- Document React development, build, embed, and Docker order.

## Cross-Task Acceptance Criteria

- [ ] Foundation child delivers a working React shell embedded by Go under root
      and non-root BaseURL.
- [ ] Missing-items child migrates `/admin/missing` to React against real JSON
      contracts with loading, empty, error, delete, and bulk-delete behavior.
- [ ] Unmigrated routes continue to work through Go templates.
- [ ] No Node runtime is required in the final Mango binary or Docker runtime
      image.
- [ ] Comic/flat light/dark switching works in the React shell without CDN font
      or component-library runtime dependencies.
- [ ] Documentation describes the React build order and coexistence rules.

## Out of Scope

- Rewriting the Go backend into Node, Next.js, or another server framework.
- Migrating home, library, title, reader, login, users, subscriptions, plugin
  download, or OPDS in the first delivery.
- Full design-system rewrite or public CDN assets.
- Completing every page in a single change set.
- Product removal of the disabled download manager.
