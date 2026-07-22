# Journal - howard.ju (Part 1)

> AI development session journal
> Started: 2026-06-21

---

## 2026-07-08 — Phase 0 收尾

Task: `07-08-migrate-crystal-to-go` (in_progress)

**完成工作:**

1. `internal/archive` — 统一 archive.Reader 接口 + zip/cbz/rar/cbr/7z 读取。
   使用 nwaples/rardecode + bodgit/sevenzip，纯 Go 无 cgo。（8 tests）

2. `internal/plugin` — 完善 goja 沙箱：
   - 新增 `mango.storage(key, value?)` — 持久化 storage.json KV 存储
   - 新增 `mango.settings(key)` — 读取 info.json 的 settings 配置
   - 新增 `Plugin` 结构体：loadInfo + loadIndexJS + v1/v2 生命周期
     (SearchManga/ListChapters/SelectChapter/NextPage/NewChapters/CanSubscribe)
   - 17 tests（v1 + v2 完整流程、storage、settings、错误处理）

3. `internal/thumbnail` — 图片尺寸读取 + 缩略图生成：
   - DecodeConfig（JPEG/PNG/WebP）
   - Generate：纵向宽200、横向/方形高300，输出 JPEG
   - 9 tests（portrait/landscape/square/PNG/invalid/small）

4. Phase 0 验证: go build + vet + test — 43 tests 全通过

**Phase 1 已完成 (2026-07-08):**
- storage: users CRUD + bcrypt + token + thumbnails + tags + hidden + sort_title
- auth 中间件: token cookie/bearer + disable_login + auth_proxy_header + admin 中间件
- CLI: admin user add/delete/update/list (tablewriter)
- 89 tests 全通过

**Phase 2 — 库扫描（已完成）:**
- internal/library: Title/Entry/ArchiveEntry/DirEntry, scanner, natural sort, cache
- internal/tasks: background runner with scan/thumbnail ticker
- 128 tests, check 修复 3 个 bug

**Phase 3 — 插件完整（已完成）:**
- internal/queue: 下载队列（独立 SQLite DB）
- plugin/subscriptions: 订阅 CRUD + JSON 文件存储 + 过滤器
- plugin/updater: 后台订阅检查 + 推入队列
- plugin/downloader: 后台下载处理器（.cbz.part → rename）
- 170 tests, check 修复 4 个 bug（base64 解码、persistence、pages 类型、时间戳）

**Phase 4 已完成 (2026-07-09):**
- chi 路由框架 + CORS/log/upload 中间件
- 68 条路由全部注册（api/admin/main/reader/opds）
- API handlers: library、book、page、cover、progress、tags、sort、plugin、queue、admin 等
- Page handlers: home、library、title、reader、tags、login、admin、user、download、subscription
- OPDS XML: index + title 渲染
- progress storage: 表 + migration 14 + CRUD + continue/start/recently added
- 25+ 模板文件（layout + 10 pages + components + OPDS XML）
- public/ 静态资源 copy + embed.FS
- 170 tests 全通过，build + vet clean

**Phase 5 已完成 (2026-07-09):**
- Dockerfile: 纯 Go multi-stage build, scratch 基础镜像, CGO_ENABLED=0
- docker-compose.go.yml: Go 版 compose 对齐
- Makefile: go-build/go-static/go-test/go-check/go-run 目标
- README: Go 构建/测试/Docker 部署说明
- 二进制: 静态编译 36MB, 无任何外部依赖

**全部 Phase 0-5 完成!** Go 二进制可指向生产 DB+插件目录启动。
170 tests 全通过, build + vet clean。



## Session 1: Complete Go Migration Phase 4-5

**Date**: 2026-07-09
**Task**: Complete Go Migration Phase 4-5
**Branch**: `migrate-to-go`

### Summary

Implement Phase 4 (HTTP routes, API handlers, templates, OPDS, progress DB) and Phase 5 (Dockerfile, docker-compose, Makefile targets, README) of the Crystal-to-Go migration. 170 tests passing.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `f9efaa8` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 2: Remove Crystal, Go-only tree

**Date**: 2026-07-13
**Task**: Remove Crystal, Go-only tree
**Branch**: `main`

### Summary

Archived remove-crystal after Crystal source/build/docs cleanup merged (5a88df4). Repo is Go-only; working tree clean on main.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `5a88df4` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 3: Non-blocking library scan

**Date**: 2026-07-13
**Task**: Non-blocking library scan
**Branch**: `fix/nonblocking-library-scan`

### Summary

Scan no longer holds RWMutex during disk walk; short lock swap only. Pushed fix/nonblocking-library-scan (9e45824). Archived task 07-13-nonblocking-library-scan.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `9e45824` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 4: Fix scratch Docker HOME, unify Dockerfile

**Date**: 2026-07-13
**Task**: Fix scratch Docker HOME, unify Dockerfile
**Branch**: `main`

### Summary

Added ENV HOME=/root to scratch Dockerfile so ~ expands correctly. Removed redundant go/Dockerfile, unified to single root Dockerfile. Updated spec with Docker conventions.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `30ffddd` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 5: Archive completed cover refresh and theme CSS tasks

**Date**: 2026-07-15
**Task**: Archive completed cover refresh and theme CSS tasks
**Branch**: `feature/fix-hide-show-toggle`

### Summary

Archived two completed tasks already merged via PR #29 and #30: cover-refresh-ui-block and theme-css-consolidation. Left 07-15-fix-hide-show-toggle and its dirty code changes untouched.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `38a3d40` | (see git log) |
| `3e4ade2` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 6: Fix hide/show title toggle

**Date**: 2026-07-15
**Task**: Fix hide/show title toggle
**Branch**: `main`

### Summary

Fixed hide/show title toggle on library and tag pages: wired GetHiddenTitleIDs filtering, admin show_hidden mode, tag card hide actions, and tests. Committed, pushed, merged via PR #31, archived task.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `1d7324f` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 7: Finish Flat Netflix restyle

**Date**: 2026-07-16
**Task**: Finish Flat Netflix restyle
**Branch**: `main`

### Summary

Archived 07-15-flat-netflix-restyle after Netflix-inspired Flat UI restyle landed (PR #33 merge + follow-up polish). User happy with Flat; next will brainstorm Comic layout copy of Flat with comic visual feel kept.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `51ec5b1` | (see git log) |
| `58849df` | (see git log) |
| `a40900f` | (see git log) |
| `64fd106` | (see git log) |
| `3d1065c` | (see git log) |
| `c8c6960` | (see git log) |
| `11bb5eb` | (see git log) |
| `05e2543` | (see git log) |
| `a603fe2` | (see git log) |
| `4a9ea43` | (see git log) |
| `61a9265` | (see git log) |
| `b08bc87` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 8: Comic Netflix layout

**Date**: 2026-07-16
**Task**: Comic Netflix layout
**Branch**: `ui/comic-netflix-layout`

### Summary

Implemented Comic layout aligned to Flat top-bar shell (shared topbar DOM, hide sidebar, full-width, rails/library poster fixes, global sharp corners). Spec ui-theme-layout.md. Branch ui/comic-netflix-layout.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `259c5b5` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 9: Fix thumbnail RIFF/WebP decode failure

**Date**: 2026-07-16
**Task**: Fix thumbnail RIFF/WebP decode failure
**Branch**: `fix/thumbnail-riff-decode`

### Summary

Registered PNG/GIF/WebP decoders in thumbnail package, removed WebP fallback that produced riff: missing RIFF chunk header on non-WebP pages, added regression tests with non-masking fixtures, and documented decoder registration contracts in library-background-jobs spec.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `6350380` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 10: Fix nested titles missing from library TitleHash

**Date**: 2026-07-16
**Task**: Fix nested titles missing from library TitleHash
**Branch**: `fix/nested-title-hash`

### Summary

Nested Series/Part/vol archives now retain Children, land in TitleHash, persist via cache v2, DeepEntries recurses, firstEntryID uses first deep entry for series covers, with regression tests and library-background-jobs spec updates.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `d5c5ade` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 11: Project-wide review wrap-up

**Date**: 2026-07-17
**Task**: Project-wide review wrap-up
**Branch**: `chore/project-wide-review`

### Summary

Completed project-wide review deliverables (review.md + four child PRDs), committed on chore/project-wide-review, archived 07-17-project-wide-review. Next: plan auth-http-hardening design/implement.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `2807b4b` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 12: Auth and HTTP hardening

**Date**: 2026-07-17
**Task**: Auth and HTTP hardening
**Branch**: `fix/auth-http-hardening`

### Summary

Implemented and verified auth/HTTP hardening: logout token revoke, safe login redirect, Secure cookies, login IP rate limit, timeouts+Shutdown, CORS tighten, security headers, body limits, upload path containment; docs+spec updated. Committed on fix/auth-http-hardening.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `f3c1def` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 13: Config deploy docs cleanup

**Date**: 2026-07-17
**Task**: Config deploy docs cleanup
**Branch**: `fix/config-deploy-docs-cleanup`

### Summary

Implemented BaseURL mount, wired log_level/download_timeout/cache_enabled, fixed Compose/env/Makefile, removed broken API ReDoc and Crystal spec/, updated README and backend Trellis index. Quality gate passed.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `ad02a53` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 14: React login migration

**Date**: 2026-07-18
**Task**: React login migration
**Branch**: `main`

### Summary

Migrated GET /login to React shell with POST /api/login, safe callback redirects via requireAuth, already-authenticated bounce, dual-theme login card; archived 07-18-frontend-react-login after PR merge.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `9ec992d` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 15: Finish React browse migration

**Date**: 2026-07-19
**Task**: Finish React browse migration
**Branch**: `main`

### Summary

Completed and archived frontend-react-browse: home/library/title React migration already committed (488bcad) and merged via PR #44. Working tree was clean; archived task and recorded session.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `488bcad` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 16: React reader migration

**Date**: 2026-07-19
**Task**: React reader migration
**Branch**: `feature/react-reader`

### Summary

Migrated /reader to React immersive shell with GET /api/reader bootstrap, continuous/paged modes, progress throttle, and mango.reader.* prefs. Specs updated for 1-based page API. Pushed feature/react-reader.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `e822a72` | (see git log) |
| `fbe6f9b` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 17: React admin home migration

**Date**: 2026-07-19
**Task**: React admin home migration
**Branch**: `feature/react-admin`

### Summary

Migrated /admin to React AppShell with users/missing links, async scan job + scan_progress polling, thumbnail progress, and global theme/UI-style controls. Specs updated. Browser smoke deferred. Pushed feature/react-admin.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `ac65317` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 18: Legacy frontend asset retirement

**Date**: 2026-07-20
**Task**: Legacy frontend asset retirement
**Branch**: `feature/legacy-retirement`

### Summary

Disabled subscriptions/downloads/plugin-download routes; deleted jQuery/Alpine/UIkit templates, vendor JS/CSS/webfonts, and dead legacy handlers. Kept react-shell, public/react, PWA icons, OPDS. go test 246 + frontend typecheck green. Pushed feature/legacy-retirement.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `3238dc7` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 19: Remove plugin/subscription APIs

**Date**: 2026-07-20
**Task**: Remove plugin/subscription APIs
**Branch**: `feature/remove-plugin-apis`

### Summary

Removed plugin/queue packages, admin plugin/subscription/queue HTTP APIs, background updater/downloader, and related config keys. go test 185 green. Pushed feature/remove-plugin-apis.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `04c4f37` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 20: Close React Vite frontend migration parent

**Date**: 2026-07-20
**Task**: Close React Vite frontend migration parent
**Branch**: `feature/remove-plugin-apis`

### Summary

Archived parent 07-17-frontend-react-vite after 10/10 children. Product React UI complete; legacy chrome and plugin/subscription/queue removed.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `0df3d30` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 21: 主页继续阅读布局优化

**Date**: 2026-07-20
**Task**: 主页继续阅读布局优化
**Branch**: `main`

### Summary

将首页继续阅读从宽卡网格改为首项大卡+紧凑竖列表（默认3行可展开），与开始阅读poster轨差异化；更新ui-theme-layout spec；build通过并归档任务。

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `4502f11` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 22: UI consistency audit and fix

**Date**: 2026-07-21
**Task**: UI consistency audit and fix
**Branch**: `feat/ui-consistency`

### Summary

Unified React shell: PosterCard/BrowseToolbar contracts, full-page i18n, ErrorState onRetry, design tokens (danger/reader/comic buttons), form field markup, removed react-preview scaffold; specs updated; typecheck/build/go server tests pass.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `c18bb42` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 23: Restore icons UI/UX after React rewrite

**Date**: 2026-07-21
**Task**: Restore icons UI/UX after React rewrite
**Branch**: `feat/restore-icons-ui-ux`

### Summary

Brainstormed and implemented system-wide icon restoration with lucide-react: Icon wrapper, semantic map, brand mark, density-mixed icons across shell/browse/pages/reader. Specs updated. PR #51 opened.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `1597345` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 24: Unify comic UI font stack

**Date**: 2026-07-21
**Task**: Unify comic UI font stack
**Branch**: `main`

### Summary

Self-hosted Fredoka (400/700 WOFF2) for comic theme; full comic body uses --mango-font-comic with system CJK fallbacks; removed heading-only split and --mango-font-sound; Reader stays on body font; updated ui-theme-layout font contract.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `48d0660` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 25: Vite dev proxy + wordmark style

**Date**: 2026-07-21
**Task**: Vite dev proxy + wordmark style
**Branch**: `main`

### Summary

Added Vite dev-only /api+/img proxy and npm run server; styled Mango wordmark (uppercase, theme-aware hard offset shadow); rebuilt embed assets; PR #53 merged.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `f678645` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 26: Continue-reading stack deck + home rails

**Date**: 2026-07-21
**Task**: Continue-reading stack deck + home rails
**Branch**: `feat/continue-read-carousel`

### Summary

Shipped circular stacked continue-reading deck, poster-rail arrows without scrollbars, Vite URL pageId fallback for multi-page HMR; PR #54; archived 07-21-continue-read-carousel.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `4503d8d` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 27: Continue-stack focus polish

**Date**: 2026-07-21
**Task**: Continue-stack focus polish
**Branch**: `polish/continue-stack-focus`

### Summary

Solid 3px accent border on active continue card; viewport-scaled stack peeks; PR #56.

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `69dad29` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 28: Poster rail skeleton + SQLite WAL

**Date**: 2026-07-22
**Task**: Poster rail skeleton + SQLite WAL
**Branch**: `main`

### Summary

Home PosterRail 加载骨架（shimmer + CLS）已落地并接 HomePage；SQLite Open 启用 WAL 与 busy_timeout=5000；spec 更新 component-guidelines；分支 feat/poster-rail-skeleton-sqlite-wal 已合并 PR #59。

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `b422da8` | (see git log) |
| `cbc7b64` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 29: Structured logging with slog

**Date**: 2026-07-22
**Task**: Structured logging with slog
**Branch**: `feat/structured-logging`

### Summary

server 包 log.Print* 迁至 slog；ApplyLogLevel 真级别过滤 + Text handler + stdlib 桥接；access log 结构化字段；spec structured-logging.md；分支 feat/structured-logging 已推送。

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `776d3bb` | (see git log) |
| `114d279` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 30: Decompose storage package

**Date**: 2026-07-22
**Task**: Decompose storage package
**Branch**: `feat/decompose-storage`

### Summary

机械拆分 storage.go 为 user/thumbnail/tag/title/identity/missing/library_cache；storage.go ~124 行；API 不变；go test ./... 通过。

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `de73081` | (see git log) |
| `0ee11a3` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 31: Client SPA router

**Date**: 2026-07-22
**Task**: Client SPA router
**Branch**: `feat/client-spa-router`

### Summary

集成 react-router-dom；BootProvider 常驻会话；壳层/浏览/管理/阅读器 SPA 导航；logout/download 仍整页；typecheck+build 通过。

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `4b894b1` | (see git log) |
| `ac3640a` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 32: Global UI prefs Zustand stores

**Date**: 2026-07-22
**Task**: Global UI prefs Zustand stores
**Branch**: `feat/global-ui-context`

### Summary

引入 zustand：themeStore + readerPrefsStore；AppShell/useReaderPrefs 接入；跨 tab storage 同步；I18n/Boot 留 follow-up zustand-i18n-boot。

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `07fc350` | (see git log) |
| `d43323e` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 33: Zustand i18n and boot migration

**Date**: 2026-07-22
**Task**: Zustand i18n and boot migration
**Branch**: `feat/zustand-i18n-boot`

### Summary

I18n/Boot 迁 Zustand；main 去掉 Provider；语言跨 tab storage 同步；useI18n/useBoot API 不变。

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `2db16eb` | (see git log) |
| `bf01390` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete


## Session 34: Close project-review task

**Date**: 2026-07-22
**Task**: Close project-review task
**Branch**: `main`

### Summary

勾选 07-22-project-review 验收项并归档；交付物 review_report.md 已就绪，7 个子任务均已 done。

### Main Changes

(Add details)

### Git Commits

| Hash | Message |
|------|---------|
| `deabb4d` | (see git log) |

### Testing

- [OK] (Add test results)

### Status

[OK] **Completed**

### Next Steps

- None - task complete
