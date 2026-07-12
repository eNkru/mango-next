# PRD: Migrate Mango from Crystal to Go

## Goal & User Value

将 Mango（漫画服务器 + Web 阅读器，当前 Crystal 实现，~7500 行源码）迁移到 Go，
换取更成熟的构建/部署生态、更易维护的依赖、更广的贡献者基础，同时对最终用户保持
**行为、API、数据、前端、插件生态完全兼容**——用户升级到 Go 版无需迁移数据、无需修改插件。

## Confirmed Facts（来自代码检查）

- **语言与规模**：Crystal，`src/` 下 ~7510 行；最大文件 `routes/api.cr`(1497)、
  `library/title.cr`(780)、`storage.cr`(761)、`plugin/plugin.cr`(544)。
- **Web 框架**：Kemal + kemal-session（session 管理）。
- **HTTP 路由**：共 67 条，分布在 `routes/{api,main,reader,opds,admin}.cr`。
  含 `/api/*`、`/reader/*`、`/opds`、静态资源、鉴权。
- **数据库**：SQLite（`crystal-sqlite3`），用 `mg` 做迁移；`migration/` 下 13 个
  版本化迁移（users.1 → hidden.13），表含 users/entries/titles/thumbnails/tags 等。
  DB 内还存 bcrypt 密码、session token、缩略图二进制、阅读进度。
- **JS 插件运行时**：`plugin/plugin.cr` 用 `duktape`（cgo C 库）在进程内跑第三方
  `index.js`，注入宝实函数 `mango.get/post/css/text/attribute/storage/settings/raise`。
  插件 API 有 v1/v2 两个版本；v2 支持 `searchManga`/`newChapters`/settings/订阅。
  HTML 解析用 `myhtml`（css/text/attribute 依赖它）。
- **压缩包处理**：`archive.cr`——zip/cbz 走 Crystal 标准库 `Compress::Zip`，
  其它格式（cbr/rar/7z 等）走 `archive.cr`（libarchive，cgo）。
- **图片处理**：`image_size.cr`——读取尺寸 + 生成缩略图（宽 200 或高 300 缩放）。
- **配置**：`config.cr`，YAML 文件，22 个配置项；支持环境变量覆盖；
  优先级 config 文件 > 环境变量 > 默认值。默认路径 `~/.config/mango/config.yml`。
- **并发模型**：Crystal fiber（`spawn` + `MainFiber` 串行化 DB 访问）；
  后台任务有：库扫描（scan_interval）、缩略图生成、下载队列、插件订阅更新。
- **CLI**：`clim` 实现主命令 + `admin user add/delete/update/list` 子命令。
- **前端**：`src/views/` 下 10 个 `.ecr` 模板 + `public/`（UIkit 3.5.9、
  FontAwesome、自写 JS/CSS）；`gulpfile.js` 用 gulp 编译 less/压缩。
  Crystal 用 `baked_file_system` 把静态资源打进二进制。
- **其它依赖**：`http_proxy`（插件代理）、`sanitize`、`koa`/`open_api`（API 文档）、
  `tallboy`（CLI 表格）、bcrypt、UUID。

## Decisions（已与用户确认）

1. **迁移策略**：分阶段增量迁移（先搭 Go 骨架 → 按模块逐块移植 → 对比行为）。
2. **JS 插件引擎**：goja（纯 Go）。完整保留所有宝实函数；现有 `index.js` 插件无需改动。
   接受 goja 与 duktape 的 ES 特性差异需逐插件回归测试的风险。
3. **共存与验证**：平行新仓库 + 兼容旧数据。Go 版复用现有 SQLite DB 文件与 schema、
   相同插件目录、相同配置格式；两版可指向同一数据目录对拿行为，最终用 Go 二进制替换部署。
4. **前端**：直接移植现有 `.ecr` 模板（机械转成 Go `html/template`），
   `public/` 静态资源原样复用，UIkit 前端保持不变。

## Requirements

### R1 — 功能对等
- 复现全部 67 条 HTTP 路由，路径、方法、请求/响应结构与现版一致。
- 复现 CLI（主命令 + `admin user` 子命令）。
- 复现后台任务：库扫描、缩略图生成、下载队列、插件订阅更新。

### R2 — 数据兼容
- 直接打开现有 SQLite DB，无需数据迁移。
- schema 与 `migration/` 产出的最终表结构一致；实现等价迁移机制供全新安装使用。
- bcrypt 密码校验、session token、阅读进度、缩略图 blob 读写全部兼容。

### R3 — 插件兼容
- goja 沙箱注入 `mango.get/post/css/text/attribute/storage/settings/raise`，语义等价。
- 支持插件 API v1 与 v2（含 searchManga/newChapters/settings/订阅）。
- `info.json`/`storage.json`/插件目录布局不变；现有社区插件可直接运行。

### R4 — 归档与图片
- 支持 zip/cbz + libarchive 覆盖的格式（cbr/rar/7z 等）。
- 缩略图生成规则等价（宽 200 或高 300 缩放，格式处理一致）。

### R5 — 配置兼容
- 读取现有 `config.yml`，22 个配置项全部支持，环境变量覆盖与优先级一致。

### R6 — 前端对等
- 10 个模板全部移植，页面渲染与现版视觉/交互一致；静态资源与 base_url 行为一致。

## Acceptance Criteria

- [ ] Go 二进制指向现有生产 DB 与插件目录可正常启动，无需任何数据迁移。
- [ ] 67 条路由逐条通过对比测试（对同一请求，Go 版与 Crystal 版响应结构一致）。
- [ ] 至少 2 个真实社区插件（1 个 v1、1 个 v2）在 goja 下 search/list/select/下载全流程通过。
- [ ] 现有用户可用原密码登录（bcrypt 兼容），阅读进度正确显示。
- [ ] 全部 10 个前端页面渲染正确，阅读器翻页、上传、管理后台功能可用。
- [ ] zip/cbz 与至少一种 libarchive 格式（如 cbr）可正常读取与生成缩略图。
- [ ] CLI `admin user add/delete/update/list` 行为等价。
- [ ] Docker 镜像可构建并运行（保留多架构能力，goja 纯 Go 无 cgo 负担）。

## Out of Scope

- 前端框架现代化（不改成 SPA，保持 UIkit）。
- 重新设计数据库 schema 或 REST API。
- 新功能开发（仅做等价迁移）。
- MangaDex 集成的协议变更（如仍在用则等价移植，不重构）。

## Open Questions（待规划中或实现前澄清）

- OQ1：libarchive 的处理——纯 Go 库（如 `mholt/archives`）能否覆盖现有全部格式？
  cbr/rar 加密/分卷等边缘格式需在设计阶段做兼容性验证（见 design.md 风险项）。
- OQ2：session 存储——kemal-session 的 cookie/存储格式是否需与现版二进制兼容
  （即老用户升级后是否需重新登录）？倾向可接受「升级后重新登录一次」。
