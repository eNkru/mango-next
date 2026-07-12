# Implement: Remove Crystal, Go-only tree

## Ordered checklist

1. **Makefile Go-first**
   - Rewrite targets: `all`/`build`/`static`/`run`/`test`/`check`/`clean`/`install` use `go/`.
   - Remove crystal/shards/yarn/gulp/arm cross crystal targets.
   - Output binary name: `mango` (Docker-aligned); keep optional `go-static` alias.

2. **Docker / compose / CI**
   - Replace root `Dockerfile` with Go multi-stage (mirror `go/Dockerfile`).
   - Point `docker-compose.yml` at Go build; update QNAP compose dockerfile path.
   - Fix `.github/workflows/docker-mango.yml` to `file: go/Dockerfile` (or root Go Dockerfile) + correct context.

3. **Delete Crystal + old frontend tree**
   - `rm -rf src lib migration public dist`
   - Delete `shard.yml` `shard.lock` `.ameba.yml` `gulpfile.js` `package.json` `yarn.lock`
   - Delete `SESSION_NOTES.md`

4. **Docs**
   - README: single Go path (dev, build, docker, deploy).
   - FRONTEND_DEV_GUIDE: `go/web/*` only.
   - DEPLOY_QNAP (+ related): Go image/build, no Crystal compile.

5. **.gitignore**
   - Remove obsolete crystal-only if safe; ignore build artifact `mango`/`mango-go` as needed.

6. **Validate**
   ```bash
   cd go && go build ./... && go vet ./... && go test ./...
   make build && make test
   # optional: docker build -t mango:go -f Dockerfile .
   ```

## Rollback points

- After step 2 only: still can build Crystal if not deleted.
- After step 3: only git restore.

## Review gate before `task.py start`

- User confirms planning artifacts OK.
- No residual “keep Crystal for compare” requirement (decision A).
