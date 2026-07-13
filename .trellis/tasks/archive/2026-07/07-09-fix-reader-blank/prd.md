# Fix Reader Page Blank Display

## Problem
When clicking to view a book in the Go-rewritten Mango app, the reader page displays nothing (blank page).

## Root Cause
`ReaderPageData` struct in `go/internal/server/web.go` does not include `PageName` field, but `head.tmpl` template requires `{{.PageName}}` to render the page title. Other pages embed `LayoutData` which contains `PageName`, but `ReaderPageData` is standalone.

Additionally, `reader.tmpl` uses `{{template "head" .}}` which expects both `BaseURL` and `PageName` from the data context.

## Acceptance Criteria
1. Reader page renders correctly when clicking a book entry
2. Page title shows "Mango - reader" (or appropriate name)
3. No template rendering errors in server logs
4. All existing reader functionality (page navigation, mode switching, entry selection) continues to work

## Scope
- File: `go/internal/server/web.go` — add `PageName` field to `ReaderPageData` or embed `LayoutData`
- File: `go/internal/server/handlers_pages.go` — populate `PageName` when constructing `ReaderPageData` in `handleReader()`
- Verify `ReaderErrorPageData` has the same issue and fix if needed

## Constraints
- Minimal change — only fix the missing field, no refactoring
- Must not break existing reader JavaScript functionality
