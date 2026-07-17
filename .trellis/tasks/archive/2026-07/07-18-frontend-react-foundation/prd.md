# React Vite foundation shell

## Goal

Create the reusable React + Vite + TypeScript foundation that Go can embed and
serve, including dual-theme shell support, BaseURL-aware assets, and production
build ordering.

## Confirmed Facts

- Go embeds `go/web/public` and currently serves template-era JS/CSS.
- Make and Docker already have or can host a Node generation stage before Go
  build.
- Comic/flat theme markers are applied on `html` before paint today.
- This child is the first milestone under `07-17-frontend-react-vite`.
- The previous asset-pipeline end state is redirected here: keep npm lockfile,
  deterministic generation, local fonts, offline delivery, and Docker/Make
  order; do not treat jQuery/LESS inventory modernization as the final design.

## Requirements

- Add a React + Vite + TypeScript application structure.
- Produce browser assets into a path Go can serve and embed.
- Provide a Go-served lightweight HTML shell for migrated routes.
- Boot comic/flat and light/dark markers without FOUC and without CDN fonts.
- Expose shared shell primitives needed by pilot pages: layout frame, page title
  area, alert/toast surface, confirm dialog primitive, loading/empty/error
  states, and BaseURL-aware fetch helper.
- Wire `npm ci`, Vite build, Make, and Docker so assets generate before
  `go build`.
- Document developer commands and ownership of generated React outputs.
- Leave unmigrated Go template routes working.

## Acceptance Criteria

- [ ] Clean checkout can install and build the React app from lockfile.
- [ ] Go binary embeds the React build outputs.
- [ ] A sample or placeholder migrated shell route proves HTML shell mounting.
- [ ] BaseURL-prefixed asset URLs work for root and non-root mounts.
- [ ] Comic/flat light/dark boot works with local assets only.
- [ ] Make and Docker generate React assets before Go compilation.
- [ ] Unmigrated template pages still render.

## Dependencies

- Parent: `07-17-frontend-react-vite`
- Redirects and supersedes the final intent of `07-17-frontend-asset-pipeline`
- No dependency on the missing-items child

## Out of Scope

- Migrating a real business page.
- Implementing missing-items storage/API behavior.
- Full multi-page client-side SPA ownership.
- Adopting a full third-party component library.
