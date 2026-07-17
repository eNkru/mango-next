# Frontend browser smoke coverage

## Closure

Closed during planning at the user's request. No implementation was started.
The project is not ready for this level of browser automation; the requirements
below are retained only as deferred planning context.

## Goal

Provide stable browser-level evidence that critical server-rendered flows,
responsive layouts, and supported themes remain usable.

## Confirmed Facts

- No browser runner, frontend test manifest, or repository CI workflow exists.
- The application uses authenticated and admin-only routes, a SQLite-backed
  library, Go-embedded assets, and BaseURL-aware URLs.
- Two UI style families and light/dark variants are represented in the current
  templates and styles.

## Requirements

- Add a proven browser automation runner using the locked frontend toolchain.
- Treat current and previous major releases of Chromium, Firefox, and Safari as
  the supported browser baseline; Internet Explorer 11 is unsupported.
- Provide deterministic fixtures and startup/teardown suitable for local and CI
  execution without external services.
- Cover login, library browsing, title actions, reader navigation, admin actions,
  and the repaired download-manager behavior.
- Exercise representative desktop and mobile viewports across supported theme
  variants without multiplying the full suite unnecessarily.
- Use stable semantic locators or explicit test IDs only where semantic locators
  cannot express intent.
- Check for uncaught page errors, failed local assets, basic keyboard operation,
  and obvious responsive overlap in covered flows.
- Expose frontend-owned commands that the shared CI baseline can invoke.

## Acceptance Criteria

- [ ] A clean checkout can install the locked runner and execute smoke tests with
      documented local commands.
- [ ] Critical user and admin journeys pass at representative desktop and mobile
      viewports.
- [ ] Both supported UI style families and light/dark behavior receive explicit,
      non-duplicative coverage.
- [ ] Tests fail on uncaught browser errors, missing runtime assets, or broken
      BaseURL navigation in covered journeys.
- [ ] Fixtures are isolated and repeatable; a failed run does not depend on or
      corrupt developer data.
- [ ] The shared CI task can invoke one stable frontend smoke command and collect
      useful failure artifacts.

## Dependencies

- Depends on `07-17-frontend-queue-contract` for queue journey expectations.
- Depends on `07-17-frontend-asset-pipeline` for the locked runner and install
  command.
- Owns browser configuration, fixtures, selectors, and frontend smoke commands;
  `07-17-test-ci-baseline` owns shared workflow wiring.

## Out of Scope

- Exhaustive visual regression, cross-browser certification, or full end-to-end
  coverage of every page and permission combination.
- Internet Explorer 11 behavior.
- Repository-wide CI policy.
