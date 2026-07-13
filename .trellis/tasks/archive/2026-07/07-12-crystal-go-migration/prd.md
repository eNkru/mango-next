# Crystal to Go migration - copy missing functions

## Goal

把 Phase 0–5 (`f9efaa8`) 结束后仍在 `src/`(Crystal)中存在、但 `go/` 还没搬运
（或未完整行为对拍）的功能模块**全部搬运并补齐**，交付一份"广义对拍 + 回归 +
全量规范通过"的集成性补丁；同时**不做**UI 改造、**不引入**新数据库。完成后 Go
二进制可替换 Crystal 部署，所有现存的 `src/` Crystal 模块都有对应的 Go 实现（或
经过评审确认无对应需求）。

> 用户价值：迁移收官；任何功能在 Crystal 版能找到的，在 Go 版也能找到。

## Confirmed Facts（代码扫描）

### 现状

- 分支 `migrate-to-go`，最新 commit `035f6e6` "migratie to GO in progress"。
- Phase 0–5 在 `f9efaa8`:archive / library / plugin / queue / server / templates /
  Dockerfile，170 tests 全绿。
- 后端 spec: `.trellis/spec/backend/index.md` 写明 Go baseline 验证命令。
- 测试契约: `cd go/ && go build ./... && go vet ./... && go test ./...`,
 178+ tests（按改动增加）。

### Crystal 仍存但 Go 未搬 / 未对拍（候选清单）

| Crystal 文件 | 行数 | Go 状态证据 | 影响 |
|---|---|---|---|
| `src/rename.cr` | 151 | `grep Rename go/` → 0 命中 | 重命名 DSL |
| `src/util/signature.cr` | 96 | ✅ FNV file/dir + ContentsSignature SHA1（`07-12-go-signature`） | 扫描/识别 |
| `src/upload.cr` | 60 | ✅ `go/internal/upload` + `apiAdminUpload` + TitleInfo cover（07-12-go-upload-helpers） | 图床路径转换/写盘 |
| `src/handlers/static_handler.cr`（baked FS） | 31 | 用 `web.Public()` + `http.FS`，但 `requesting_static_file` 列表未对拍 Crystal 版本 | 静态分发 |
| `src/handlers/log_handler.cr`（elapsed_text） | 23 | `LoggingMiddleware` 简化为 `log.Printf` 单行打印；未对拍 µs / ms 二级精度 | 日志格式 |
| `src/handlers/auth_handler.cr`（Basic/Bearer+session） | 119 | `go/internal/server/auth.go` 含 token middleware，但 session bridge /Kemal session 未移植 | 鉴权路径 |
| `src/util/chapter_sort.cr` / `numeric_sort.cr` | 113 / 44 | ✅ `go/internal/library/sort.go` 覆盖 | — |
| `src/util/proxy.cr` | 44 | ❓ 未 grep 验证 | 插件 HTTP 代理 |
| `src/util/validation.cr` | 31 | ❓ 未 grep 验证 | 入参校验 |
| `src/util/web.cr` | 177 | ❓ 模板里引用的 `KEMAL_ENV` / `DEVELOPMENT` 未搬运 | layout 条件分支 |
| `src/routes/admin.cr` / `reader.cr` / `opds.cr` | 82 / 61 / 18 | `server.go` 已注册 68 路由，逐条 handler 对拍未做 | 路由覆盖完整性 |
| `src/main_fiber.cr`（DB fiber 串行化入口） | 34 | `cmd/mango/main.go` 用 sequential；并发 DB 是否需 channel 化未对照 | 并发模型差异 |
| `src/library/cache.cr` | 219 | Go 端 cache 散在 library/scanner；逐函数未对拍 | 缓存层迁移度 |

## Requirements

### R1 — 模块搬运

- 凡是 `src/` 中存在、且 `go/` 中无对应实现的，凡能被"现实路径"（路由 / 后台任务 /
  库扫描 / 插件 API / 模板)引用到的，全部按相同行为搬运到 `go/internal/` 下合适的包。
- 行为契约统一通过**单测对拍**（对每一函数列一组用例：spec → Go 实现 → 期望相同输出）。
- 已存在的 Go 实现若与 Crystal 实现有行为差异，按 Crystal 版合约修正，并在实现
  上方用注释保留原文引用 `// mirrors Crystal src/<file>:<line>`。

### R2 — 路由/出口对拍

- 对拍 `src/routes/admin.cr`、`reader.cr`、`opds.cr` 与 `go/internal/server/showroutes`
  注册的路由，确保每个 Crystal handler 都有一个 Go 实现（方法、路径、参数语义一致）。
- 缺失的 handler 补齐；行为差异的（返回 JSON 结构、状态码）修正。

### R3 — 模板 / 前端引用常量对拍

- 模板中有引用的 Crystal 常量（如 `@@layout`、asset path）逐一在 Go 模板中确认
  等价；若不存在的，删除引用或补函。

### R4 — 后台任务合约

- 列出 Crystal 端 `src/main_fiber.cr` / `config.cr` 启动期启动的后台任务，
  对拍 `go/cmd/mango/main.go` + `go/internal/tasks/` 注册情况，记录差异。

### R5 — 行为回归

- 全跑：`cd go/ && go build ./... && go vet ./... && go test ./...`。
- Crystal 端（仅作为参考，不要求 main 跑通）：`crystal spec` 返回绿或全部跳过。

### R6 — 一份 gap 表

- 交付 `.trellis/tasks/07-12-crystal-go-migration/gap-report.md`：列出每个 Crystal
 模块 → Go 模块 的状态（已搬 ✅、行为差异 🔧、未搬但无引用 ⊘、未搬需补 ❌）。

## Acceptance Criteria

- [ ] 每条"❌ 未搬"的 Crystal 模块都有对应 Go 实现（旧函数被 Crystal-spec 测试驱动）。
- [ ] 全量 `go build ./... && go vet ./... && go test ./...` 绿；新测试数 ≥ baseline (170)。
- [ ] `go/internal/server/` 路由注册数 ≥ Crystal 路由总数（67），且每个路由都有对应的 Go handler。
- [ ] `gap-report.md` 列出全部 src/ 模块 → go/ 模块 的对照表，覆盖率 = 100%。
- [ ] PR commit 表达清楚: "(module list) — copy from Crystal to Go"。
- [ ] `migrate-to-go` 分支 `git push` PR 可审查且 CI 通过。

## Out of Scope

- 重写 UI 样式或前端组件
- 重写 Dockerfile / docker-compose
- 引入新的数据库 / ORM 抽象
- goja 与 duktape 在 v2 插件 API 上的差异（如某插件需特殊回归脚本）

## Child Task Map

| Child | Slug | Focus |
|-------|------|--------|
| Move rename.cr DSL | `07-12-go-rename-dsl` | Rename DSL 移植 |
| Move util/signature.cr | `07-12-go-signature` | inode/CRC/SHA signature |
| Move upload.cr helpers | `07-12-go-upload-helpers` | ✅ 已完成并归档（commit `75629ae`） |
| Routes coverage | `07-12-go-routes-coverage` | admin/reader/opds 对拍 |
| util completion | `07-12-go-util-completion` | proxy/validation/web |
| gap + regression | `07-12-go-gap-regression` | gap-report + 全量回归 |

推荐实现顺序：upload-helpers → signature → rename → util → routes → gap-regression。

## Risks & Open Questions

- ⚠️ `main.cr` + `main_fiber.cr` 串行化 DB 的并发模式被 Go sequential 取代，是否
  会让高并发请求数据库死锁待实测。
- ⚠️ Crystal `baked_file_system` 与 Go `embed.FS` 在 build 时的差异（macOS vs Linux
  文件 mtime、symlink）。
- ⚠️ 第三方插件可能在 v2 API（goja 与 duktape）行为差异，需要保留至少一个集成测试
  fixture。
- ✅ 已拆 parent + 6 child。
