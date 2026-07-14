# Fix hide/show title toggle not working

## Goal

Restore the admin hide/show title feature so library and tag pages correctly persist and display hidden titles, and the show/hide toggle button works.

## Background

Frontend already calls:

- `PUT /api/admin/hidden/{tid}/{value}` via `toggleHidden`
- `?show_hidden=1` via `toggleShowHidden`

But server-side page rendering is incomplete:

1. `handleLibrary` hardcodes `Hidden: false` and `ShowHidden: false`
2. Tag page toggle button always calls `toggleShowHidden(0)`, so it cannot switch on

## Requirements

1. Library page must load each title's `hidden` flag from storage.
2. Library page must honor `?show_hidden=1`:
   - default: hide titles with `hidden=1`
   - with flag: include hidden titles and mark them for UI (`is_hidden` / badges / unhide button)
3. Hide/unhide action on library cards and title detail must persist via storage and be visible after reload.
4. Library show/hide toggle button must switch between normal and show-hidden modes.
5. Tag page show/hide toggle button must correctly switch based on current `ShowHidden` state (same pattern as library).
6. Tag page should surface hidden status consistently when show-hidden mode is on (at least title list inclusion; prefer matching library UI markers if cheap).
7. Non-admin users must continue not to see hide controls / show-hidden toggle.

## Out of Scope

- Redesigning hide UX/animation
- Changing storage schema
- API redesign for mobile clients beyond existing endpoints

## Acceptance Criteria

- [x] Admin can hide a title from library card hover action; after reload, title is no longer listed in default library view
- [x] Admin can click "显示隐藏" and see previously hidden titles with hidden UI markers and unhide action
- [x] Admin can unhide a title; after reload in default mode, title returns to normal list
- [x] Title detail page hide/show button still works and reflects stored hidden state
- [x] Tag page toggle switches `show_hidden` on/off correctly
- [x] Existing storage tests for Set/Get hidden continue to pass; add/adjust page-level coverage if practical

## Constraints

- Keep frontend `title.js` API contract unchanged unless a bug is found there
- Prefer minimal changes in `handlers_pages.go` + templates
- Follow existing Trellis frontend/backend conventions

## Notes

Lightweight bugfix: PRD-only is sufficient; no separate design.md required.
