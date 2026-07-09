# Codebase Review Findings

## Findings

1. Frontend plugin and subscription data is rendered as HTML.
   - Files: `src/views/plugin-download.html.ecr`, `src/views/subscription-manager.html.ecr`, `public/js/plugin-download.js`, `public/js/subscription-manager.js`, `public/js/alert.js`.
   - Risk: plugin/search/subscription values can originate from external sites or plugin metadata and are interpolated into `x-html` or alert HTML without escaping. A malicious value can execute script in an admin session.
   - Suggested change: render plain values with `x-text` or DOM `textContent`, add one shared `escapeHtml` helper for the few intentional HTML snippets, and make `alert()` accept text plus optional trusted markup separately.

2. User management relies on frontend-only safety checks.
   - Files: `src/views/user.html.ecr`, `src/routes/api.cr`, `src/routes/admin.cr`, `src/storage.cr`.
   - Risk: the UI hides self-delete, but the API allows deleting any user and updating admin status without backend checks. An admin can delete or demote the last admin through direct API/form calls.
   - Suggested change: enforce invariants in storage or route code: cannot delete the current user through the web UI API, cannot remove the last admin, and cannot demote the last admin.

3. Runtime Docker image does not install CA certificates.
   - File: `Dockerfile`.
   - Risk: the builder installs `ca-certificates`, but the final `alpine` stage only copies the binary. HTTPS plugin requests use `HTTP::Client`, so TLS verification may fail in the runtime image.
   - Suggested change: install `ca-certificates` in the final stage and run `update-ca-certificates` if needed.

4. The declared quality gate is not reproducible.
   - Files: `Makefile`, `shard.yml`, `README.md`, `.ameba.yml`.
   - Evidence: `crystal tool format --check` reports formatting changes in multiple files. `Makefile check` calls `./bin/ameba`, but no `bin/ameba` exists and `shard.yml` has no ameba development dependency despite README claiming one.
   - Suggested change: format the listed files, either add ameba as a development dependency with a reproducible install path or remove it from `make check`, then wire the check into CI.

5. Compose defaults fail before users fill environment variables.
   - Files: `docker-compose.yml`, `env.example`.
   - Evidence: `docker compose config` fails with empty `MAIN_DIRECTORY_PATH` and `CONFIG_DIRECTORY_PATH` from the example env shape.
   - Suggested change: provide safe sample paths, document copying `env.example` to `.env`, or use required-variable syntax with clear messages.

6. Docker build context can include local metadata.
   - Files: `.dockerignore`, `Dockerfile`.
   - Risk: `COPY . .` copies everything not ignored into the builder context; `.dockerignore` omits `.git`, `.trellis`, `.agents`, `.codex`, `.opencode`, and common local artifacts.
   - Suggested change: add ignores for VCS metadata, AI/task metadata, local config, logs, cache directories, and temporary outputs.

7. Dependency audit has one low severity issue.
   - Files: `package-lock.json`, `package.json`.
   - Evidence: `npm audit --json` reports `@babel/core <= 7.29.0` arbitrary file read via sourceMappingURL comments.
   - Suggested change: update the Babel toolchain and regenerate `package-lock.json`.

8. Coverage is thin around high-risk boundaries.
   - Files: `spec/*`, `src/handlers/auth_handler.cr`, `src/routes/api.cr`, `src/upload.cr`, `src/plugin/downloader.cr`.
   - Risk: existing specs cover config, storage, utilities, rename rules, and limited plugin helpers, but not auth parsing, admin APIs, uploads, plugin download queue behavior, or XSS-prone frontend rendering.
   - Suggested change: add focused route/handler specs and frontend escaping tests before larger refactors.

## Validation

- `crystal spec`: passed, 42 examples.
- `crystal tool format --check`: failed, formatting changes required.
- `npm run uglify`: passed.
- `npm test`: failed because no `test` script exists.
- `npm audit --json`: failed with 1 low severity vulnerability.
- `docker compose config`: failed with unset/empty example variables; passed when temporary values were supplied, with an obsolete `version` warning.

## Notes

- The Trellis frontend spec files are still placeholders, so this review used source inspection and executable checks rather than project-specific frontend rules.
- The task was not archived because Trellis archive creates an automatic Git commit, and no commit was requested.

## Implementation Scope Update

- 2026-06-21: The user requested implementation work for findings 1 and 2.
- Scope is limited to safe frontend rendering for plugin/subscription/admin UI data and backend user-management invariants.
- Findings 3-8 remain recommendations only unless separately requested.

## Implementation Notes For Findings 1-2

- Frontend plugin and subscription tables now render external/plugin values with `x-text` and safe `title` attributes instead of `x-html`.
- `alert(level, text)` now treats content as plain text by default; trusted markup requires an explicit `{ allowHtml: true }` option.
- Storage now rejects deleting or demoting the last admin user.
- The admin delete API now rejects deleting the currently authenticated user before calling storage.
- Regression specs were added for deleting non-admin users, rejecting last-admin deletion/demotion, and allowing admin deletion/demotion when another admin remains.

## Implementation Validation

- `crystal spec`: passed twice after the implementation, 46 examples.
- `crystal tool format --check src/storage.cr src/routes/api.cr spec/storage_spec.cr`: passed.
- `npm run uglify`: passed.
- `rg` check for risky `x-html`/table-render helpers: no app-owned risky usages remain; only the explicit trusted `allowHtml` alert path and vendored `dotdotdot.js` internal HTML operations remain.
- `crystal tool format --check src spec`: still fails on pre-existing files not modified by this implementation (`src/util/proxy.cr`, `src/util/web.cr`, `src/server.cr`, `src/library/library.cr`, `src/library/archive_entry.cr`, `src/library/entry.cr`, `src/routes/admin.cr`, `spec/spec_helper.cr`).
- `npm test`: still unavailable because `package.json` has no `test` script.
