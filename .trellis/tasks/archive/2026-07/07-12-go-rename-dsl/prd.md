# Move rename.cr DSL to Go

## Goal

Port Crystal `src/rename.cr` (Rename rule DSL for templated filenames) to Go
with behavior matching `spec/rename_spec.cr`. Provides parse + render for
patterns used when building chapter/file names from variable maps.

## Confirmed Facts

- Source: `src/rename.cr` (~151 lines) — `Rename::Rule` with `{var}`, `{a|b}`,
  `[optional group]`, post_process illegal chars.
- **Production call sites**: none in current `src/` runtime (not required by
  mango/server). Only consumer: **`spec/rename_spec.cr`**.
- Still in scope for migration parity: pure library + tests; optional future
  use by downloader naming.
- Go today: `sanitizeFilename` in plugin/downloader only — different helper.

## Requirements

### R1 — Port Rule DSL

- Package `go/internal/rename` (or `library` — prefer **`internal/rename`** to
  mirror standalone Crystal file).
- `Parse(rule string) (*Rule, error)` / `NewRule` — parse failures match Crystal
  error cases (nested brackets, unclosed, illegal `/` in rule text, etc.).
- `Render(vars map[string]string) string` — variable / pattern / group semantics
  + post_process (`..` → `_`, strip trailing space/dot, replace illegal chars).

### R2 — Tests from rename_spec.cr

Port every example in `spec/rename_spec.cr` as Go table tests.

### R3 — No forced production wire

- Do **not** require changing downloader unless a call site appears.
- Export package for gap-report “ported + tested”.

## Acceptance Criteria

- [ ] `go/internal/rename` with Parse + Render
- [ ] Tests cover nested/unclosed brackets, `|` patterns, spaces, post_process
- [ ] `go build/vet/test ./...` green

## Out of Scope

- UI “rename display name” APIs (those are sort_title/display_name, not this DSL)
- Binding into plugin download path (optional follow-up)

## Dependencies

- parent `07-12-crystal-go-migration`
- no dependency on signature/upload

## Decisions

- Standalone package + full unit tests; production wire optional later.
