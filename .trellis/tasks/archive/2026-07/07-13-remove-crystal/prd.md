# Remove Crystal code and finalize Go-only tree

## Goal & User Value

仓库以 Go 为唯一实现：删除 Crystal 运行时/构建链/配置与相关文档表述，降低维护成本，避免双栈误导。最终用户与部署只接触 Go 二进制与 `go/` 工具链。

## Confirmed Facts（代码扫描）

- 分支 `migrate-to-go`；Go 测试 192 通过；迁移 gap-report 认为核心功能已齐。
- Go 静态资源/模板嵌入：`go/web/embed.go` → `go/web/views` + `go/web/public`（**不依赖**根目录 `public/`、`src/views/`）。
- Crystal 树仍在：`src/`、`lib/`（shards）、`migration/*.cr`、`shard.yml`/`shard.lock`、`.ameba.yml`、`gulpfile.js`、`package.json`/`yarn.lock`、根 `Dockerfile`（Crystal）、`docker-compose.yml`（根 Dockerfile）。
- Go 侧：`go/Dockerfile`、`docker-compose.go.yml`、Makefile 中 `go-build`/`go-run`/`go-test` 等。
- DB 迁移已在 Go 内：`go/internal/storage/migration/migrations.go`；运行时不读 `migration/*.cr`。
- 文档双栈：`README.md`、`SESSION_NOTES.md`、`FRONTEND_DEV_GUIDE.md`、`DEPLOY_QNAP.md` 等。
- CI：`.github/workflows/docker-mango.yml` 未锁定 `go/Dockerfile`。
- `.gitignore` 仍面向 Crystal（`/lib/`、`/.shards/`、`yarn.lock`、`mango`）并忽略 `mango-go`。

## Decisions

1. **清理激进度：A 全删** — 删除 Crystal 源码/依赖/构建链；主树 Go-only；对照用 git 历史。
2. **前端真相源：仅 `go/web/*`** — 一并删除根 `public/`、`dist/`、`src/views/`、`gulpfile.js`、`package.json`、`yarn.lock`。
3. **文档：删 SESSION_NOTES + 主文档改写为 Go-only** — 含 README、FRONTEND_DEV_GUIDE、部署相关 Crystal 表述。

## Requirements

### R1 — 删除 Crystal 实现与前端旧树
- 删除应用源码、shards、Crystal migrations、Node/gulp 构建链、根 public/dist（见 design 清单）。

### R2 — 构建/部署只保留 Go
- Makefile 默认目标为 Go；去掉 crystal/shards/yarn/gulp。
- 根 `Dockerfile` 与默认 compose 为 Go；QNAP compose 指向 `go/Dockerfile`。

### R3 — 文档与 CI 一致
- README / FRONTEND_DEV_GUIDE / DEPLOY 文档 Go-only。
- 删除 `SESSION_NOTES.md`。
- GitHub workflow 构建 Go 镜像。

### R4 — 不破坏 Go 行为
- `cd go && go build ./... && go test ./...` 保持绿。
- 不改业务逻辑、不改 schema、不引入新功能。

## Acceptance Criteria

- [ ] 无 `shard.yml`、无 `src/*.cr`、无根 `public/`/`dist/` Crystal 前端树。
- [ ] `make` / `make build` / `make run` / `make test` 仅走 Go。
- [ ] 根 `Dockerfile` + 默认 `docker-compose.yml` 为 Go。
- [ ] README / FRONTEND_DEV_GUIDE 无 Crystal 主路径；`SESSION_NOTES.md` 已删。
- [ ] `go test ./...` 全绿。
- [ ] CI 使用 `go/Dockerfile`（或等价）。

## Out of Scope

- 补齐 migration residual（WS/OpenAPI/增量 examine）。
- 前端现代化或 SPA。
- 把 `go/` 提升到仓库根。
- 复制根 public 缺的 fa-brands 到 go/web（当前无引用）。
