# React migrate tags pages

## Goal

Migrate the tags index and tag detail browse pages from Go templates + jQuery to
the React shell, while keeping Go as the only backend and preserving hidden-title
admin behavior.

## Confirmed Facts

- Routes: `GET /tags`, `GET /tags/{tag}` with optional `show_hidden=1`.
- Existing `GET /api/tags` returns plain tag name strings for title-page Select2.
- React needs count-aware tag index and title cards for a tag detail page.
- Admin can toggle title hidden via `PUT /api/admin/hidden/{tid}/{value}`.
- Work branch: `feat/frontend-react-tags` from latest main.

## Requirements

- Mount React shell for `/tags` and `/tags/{tag}`.
- Provide JSON APIs for tags index (with counts) and tag titles list.
- Support client-side search/filter on both pages.
- Preserve admin show-hidden toggle and per-title hide/unhide.
- Do not break existing `GET /api/tags` consumers.
- Keep dual-theme shell styling and BaseURL support.

## Acceptance Criteria

- [x] `/tags` is React-mounted and lists tags with counts.
- [x] `/tags/{tag}` is React-mounted and lists title cards.
- [x] Admin show-hidden and hide/unhide remain available.
- [x] Non-admin cannot enable show-hidden.
- [x] Existing `GET /api/tags` remains string-list compatible.
- [x] Unrelated routes stay on Go templates.

## Out of Scope

- Migrating library/home/title/reader.
- Full sort-option persistence rewrite for tag pages.
