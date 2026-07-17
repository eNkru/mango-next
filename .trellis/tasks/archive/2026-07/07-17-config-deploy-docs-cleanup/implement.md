# Configuration deployment and documentation cleanup — implementation plan

## Preconditions

- User review of `prd.md`, `design.md`, this file.
- `task.py start` only after approval (status remains `planning` until then).

## Ordered checklist

1. [ ] Config wiring
   - [ ] `log_level` applied at process start (+ tests)
   - [ ] `download_timeout_seconds` → plugin HTTP/downloader (+ tests)
   - [ ] `cache_enabled` honor or document deprecate
   - [ ] Document deprecated: `session_secret`, `cache_size_mbs`, `cache_log_enabled`
2. [ ] BaseURL mount
   - [ ] Mount/register routes under base path
   - [ ] Base-aware login/auth redirects
   - [ ] Integration/unit tests for `/mango/` vs `/`
3. [ ] API docs honesty
   - [ ] Remove broken ReDoc page/route/nav; no missing asset requirement
4. [ ] Compose / env / Makefile
   - [ ] `env.example` placeholders `./data` `./config`
   - [ ] Drop `version` from compose files; fix QNAP Crystal comments
   - [ ] Align `make all` with docs / `go-all`
5. [ ] Remove top-level Crystal `spec/`
   - [ ] Update `.trellis/spec/backend/index.md` (no Crystal gate; migration version)
6. [ ] README + DEPLOY_QNAP truthful paths, ports, backup/rollback/admin notes
7. [ ] Quality gate:
   ```bash
   cd go
   GOCACHE=/tmp/mango-next-go-cache go test ./...
   GOCACHE=/tmp/mango-next-go-cache go test -race ./...
   GOCACHE=/tmp/mango-next-go-cache go vet ./...
   GOCACHE=/tmp/mango-next-go-cache go build ./...
   docker compose config
   docker compose -f docker-compose.qnap.yml config
   docker compose -f docker-compose.qnap-prebuilt.yml config
   ```

## Risky files

| File | Risk |
|---|---|
| `go/internal/server/server.go` | BaseURL mount breaks root-only assumptions |
| `go/internal/server/auth.go` | Hard-coded `/login` |
| `go/internal/plugin/*` | Timeout wiring |
| `go/cmd/mango/*` | Log level timing |
| `docker-compose*.yml`, `env.example` | NAS path regressions if over-edited |
| `spec/` deletion | Large git delete; intentional |
| `.trellis/spec/backend/index.md` | Drift if partial update |

## Rollback points

- After docs/compose-only changes: low risk.
- After BaseURL: highest; revert server package if needed.
- Config deprecations: keep parsing forever if external configs exist.

## Review gates

- Planning approval required before `task.py start`.
- Coordinate Makefile quality command names with `07-17-test-ci-baseline` if
  both land close together.
