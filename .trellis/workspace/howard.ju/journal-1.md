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
