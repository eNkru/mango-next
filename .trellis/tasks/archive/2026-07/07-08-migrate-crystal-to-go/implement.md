# Implement: Migrate Mango from Crystal to Go

分阶段增量迁移。每个阶段独立可验证；先建骨架 + 关键风险 PoC，再逐块移植。
每阶段结束跑黑盒对比测试 + `go build ./... && go vet ./...`。

## Phase 0 — 骨架 & 高风险 PoC（已完工 ✓）

- [x] 初始化 Go module、目录结构（见 design §3）、cobra 入口空跑。
- [x] `internal/config`：加载现有 config.yml，22 项 + env 覆盖 + 优先级 + expand_paths + preprocess。
- [x] `internal/storage`：用 modernc.org/sqlite 打开现有 .db；实现迁移器骨架，
       对已最新的库 no-op；13 个迁移的最终 schema 全部转录。
- [x] **PoC-JS**：goja 沙箱完整实现，注入全部 `mango.*` 辅助函数
       （get/post/css/text/attribute/raise/storage/settings）；`Plugin` 结构体
       支持 v1/v2 生命周期（searchManga/listChapters/selectChapter/nextPage/newChapters/
       CanSubscribe）；43 个测试覆盖全部路径。
- [x] **PoC-Archive**：统一 `archive.Reader` 接口，覆盖 zip/cbz（stdlib）、
       rar/cbr（rardecode）、7z（sevenzip），含错误处理和测试。
- [x] **PoC-Thumbnail**：`internal/thumbnail` 包，尺寸读取 + 宽200/高300 缩放 +
       JPEG 输出，含 portrait/landscape/square/PNG 输入测试。
- [x] 验证：`go build ./... && go vet ./... && go test ./...` 全通过（43 tests）。

## Phase 1 — 数据与鉴权（已完工 ✓）

- [x] storage 完整：users CRUD（new/list/update/delete）、bcrypt 校验（兼容 Crystal 旧密码哈希）、
      token（UUID v4 无连字符）、thumbnails blob 读写、tags CRUD、hidden title、
      sort_title、admin 初始化（空库自动创建 admin+随机密码）。
- [x] 鉴权中间件（`internal/server/auth.go`）：token cookie + Bearer header 提取、
      disable_login+default_username、auth_proxy_header、admin 中间件（403）。
- [x] CLI `admin user add/delete/update/list`（cobra + tablewriter 表格输出）。
- [x] 验证：`go build ./... && go vet ./... && go test ./...` 全通过（89 tests）。
      Check 报告：PASS ✅，3 个 WARN 均为预期行为（重登、阅读进度属 Phase 2）。

## Phase 2 — 库与归档与图片

- [ ] `internal/archive`：统一抽象（zip 标准库 + rardecode + sevenzip）。
- [ ] `internal/thumbnail`：尺寸读取 + 宽200/高300 缩放，格式处理对齐。
- [ ] `internal/library`：扫描、title/entry、dir/archive entry、signature、cache、
      numeric/chapter sort。
- [ ] 后台任务：库扫描（scan_interval）、缩略图生成（interval），用 goroutine 复现。
- 验证：扫描现有库产生与 Crystal 版一致的 title/entry 计数；缩略图可生成。

## Phase 3 — 插件系统完整（已完工 ✓）

- [x] goja 沙箱完整宝实函数：get/post/css/text/attribute/raise/storage/settings（Phase 0）。
- [x] 插件 v1/v2：searchManga、listChapters、selectChapter、nextPage、newChapters（Phase 0）。
- [x] 订阅（subscriptions）：Subscription CRUD + Filter（String/NumMin/NumMax/DateMin/DateMax/Array），
      subscriptions.json 文件存储（兼容 Crystal 格式）。
- [x] 更新器（updater）：后台 goroutine 按 plugin_update_interval_hours 检查所有插件订阅，
      调用 newChapters + 过滤匹配 → 推入下载队列。
- [x] 下载器（downloader）：后台 goroutine 每秒 polling 队列，selectChapter → nextPage 逐页下载，
      4 次重试 + .cbz.part → rename 流程。
- [x] 下载队列（internal/queue/）：独立 SQLite DB（queue_db_path），JobStatus enum，
      Push/PopDownloadable/Reset/Delete/List/SetStatus 等完整操作。
- [x] 验证：`go build ./... && go vet ./... && go test ./...` 全通过（170 tests）。
      Check 修复 4 个 bug（base64 解码、LastChecked 持久化、pages 类型、时间戳）。

## Phase 4 — HTTP 路由与前端

- [ ] 路由框架（chi）+ CORS/log/静态/upload 中间件。
- [ ] 移植 67 条路由：api / main / reader / opds / admin。
- [ ] 10 个 .ecr → html/template；public/ 用 embed.FS；base_url 处理。
- [ ] OPDS XML 渲染。
- 验证：黑盒对比脚本覆盖 67 路由；10 个页面人工核对；阅读器/上传/后台可用。

## Phase 5 — 打包与收尾

- [ ] Dockerfile（纯 Go，多架构）+ docker-compose 对齐。
- [ ] 端到端：Go 二进制指向生产 DB+插件目录启动，无需数据迁移。
- [ ] 补齐 README/部署文档差异。
- 验证：全部 PRD 验收标准逐条勾选。

## 验证命令

```bash
go build ./... && go vet ./... && go test ./...
# 黑盒对比（Crystal 基线端口 A vs Go 端口 B）
./scripts/compare.sh    # 待建：批量请求 + 规范化 JSON diff
```

## 风险文件 / 回滚点

- 最高风险：`internal/plugin/*`（goja 兼容性）、`internal/archive/*`（纯 Go 格式覆盖）。
- 数据安全：验证期一律用生产 DB **副本**；迁移器只做前向幂等操作。
- 每阶段独立成 PR；Docker 可随时切回 Crystal 镜像回滚。

## Follow-up before task.py start

- 确认 OQ1（归档格式覆盖）与 OQ2（session 重登）在 Phase 0 PoC 后可闭环。
