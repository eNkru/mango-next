# Test coverage and CI baseline

## Goal

Create a reliable automated quality gate for the Go migration and raise coverage
at the HTTP and migration boundaries where regressions currently have the
largest impact.

## Requirements

- Add CI for supported Go versions that runs formatting checks, vet, tests, race
  tests where practical, and builds the application.
- Add direct tests for route registration, representative page/API handlers,
  authorization boundaries, error responses, and embedded static assets.
- Add migration tests that exercise upgrades from representative historical
  schema versions as well as fresh database creation.
- Add a reproducible dependency-vulnerability check with a documented update
  policy.
- Report coverage as an observable signal without gaming a repository-wide
  percentage.
- Keep tests deterministic, isolated from production paths, and free of external
  network dependencies.

## Acceptance Criteria

- [ ] Pull requests automatically run Go format, vet, test, race, and build gates,
      with any intentionally omitted platform/gate documented.
- [ ] Critical authenticated and admin routes have success, unauthorized,
      forbidden, invalid-input, and not-found coverage where applicable.
- [ ] Fresh schema creation and at least one historical migration path are tested.
- [ ] Embedded templates/static assets receive a startup or render smoke test.
- [ ] A failing check blocks CI and all checks pass on the baseline branch.
- [ ] The contributor documentation names the same commands that CI runs.

## Notes

- Dependency: none. It can start independently of security hardening.
- Coordination: security and frontend tasks own feature-specific tests; this task
  owns shared test infrastructure and CI integration.
- Evidence: the current Go checks pass, but `internal/server` has about 11%
  statement coverage and no repository CI workflow exists.
