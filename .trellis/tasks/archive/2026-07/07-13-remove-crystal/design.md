# Design: Remove Crystal, Go-only tree

## Boundaries

- **Delete**: Crystal runtime surface + obsolete frontend build + dual-stack docs.
- **Keep**: `go/**`（含 `go/web`）、Trellis、业务无关运维文档中可改写的部署指引。
- **No behavior change** in Go handlers/library/plugin.

## Delete inventory

| Path / file | Why safe |
|-------------|----------|
| `src/` | Crystal app |
| `lib/` | shards vendor (gitignore already) |
| `migration/` | Crystal SQL migrations; Go has `internal/storage/migration` |
| `shard.yml`, `shard.lock`, `.ameba.yml` | Crystal deps/lint |
| `gulpfile.js`, `package.json`, `yarn.lock`, `node_modules/` if present | Crystal frontend build |
| `public/`, `dist/` | superseded by `go/web/public` |
| Root Crystal `Dockerfile` content | replace with Go |
| `SESSION_NOTES.md` | Crystal build archaeology (decision: delete) |

## Rewrite inventory

| Path | Change |
|------|--------|
| `Makefile` | Default `all`/`build`/`run`/`test`/`check`/`clean`/`install` → Go; drop crystal/yarn targets |
| Root `Dockerfile` | Copy/adapt `go/Dockerfile` (or thin wrapper `FROM` / multi-stage same as go) |
| `docker-compose.yml` | Same as `docker-compose.go.yml` (Go build context) |
| `docker-compose.go.yml` | Optional: keep as alias or remove after merge into default |
| `docker-compose.qnap.yml` | `dockerfile: go/Dockerfile` or context `go` |
| `.github/workflows/docker-mango.yml` | explicit `file: go/Dockerfile`, context `go` or `.` matching Dockerfile |
| `README.md` | Go-only develop/build/deploy |
| `FRONTEND_DEV_GUIDE.md` | paths → `go/web/views`, `go/web/public`; no ECR/gulp |
| `DEPLOY_QNAP.md` | Go build/image; remove Crystal compile notes |
| `.gitignore` | drop crystal-only noise if unused; track `mango-go` ignore or rename binary to `mango` |
| `DOCKER_HUB.md` / `build-export.sh` | ensure Go image path |

## Binary naming

- Prefer Makefile produce `mango` (or keep `mango-go` and document) for install/Docker consistency.
- Recommendation: `make build` → `./mango` from `go/cmd/mango` (align Docker ENTRYPOINT `/mango`).

## Compatibility / rollback

- Rollback: git revert delete commit; Crystal still in history.
- Data: unchanged SQLite + config paths; Go already opens same DB.

## Risks

1. Docs/scripts still reference Crystal paths after delete → broken links.
2. CI still builds root Crystal Dockerfile if not updated → build fails after delete.
3. QNAP compose `context` still expects full monorepo with root Crystal Dockerfile.
4. Someone relies on root `public/` for manual static serve (Go does not).
