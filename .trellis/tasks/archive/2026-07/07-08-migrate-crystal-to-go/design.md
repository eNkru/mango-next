# Design: Migrate Mango from Crystal to Go

## 1. 目标与边界

- 用 Go 重写 Mango，**行为/API/数据/前端/插件全兼容**。
- 平行新仓库（或本仓库新建 `go/` 目录，最终独立），复用现有 SQLite DB、插件目录、config 格式。
- 纯 Go 优先（无 cgo）：JS 引擎 goja、归档纯 Go 库、SQLite 用纯 Go 驱动，
  以保留静态编译与多架构 Docker 优势。

## 2. 技术选型（对齐 Crystal 依赖）

| 领域 | Crystal | Go 方案 | 备注 |
|------|---------|---------|------|
| Web 框架 | Kemal | `chi`（或 net/http + 轻量路由） | 67 条路由，需路径参数 `:tid` |
| Session | kemal-session | `gorilla/sessions` / `scs` | cookie+存储；接受升级后重登一次 |
| DB 驱动 | crystal-sqlite3 (cgo) | `modernc.org/sqlite`（纯 Go） | 无 cgo；兼容现有 .db 文件 |
| DB 迁移 | mg | 自写版本化迁移器 或 `golang-migrate` | 复刻 migration/ 的 13 个版本 |
| JS 引擎 | duktape (cgo) | `dop251/goja`（纯 Go） | **最大风险**，见 §5 |
| HTML 解析 | myhtml | `PuerkitoBio/goquery` | css/text/attribute 宝实函数依赖 |
| 归档 | Compress::Zip + libarchive(cgo) | `archive/zip` + `nwaples/rardecode` + `bodgit/sevenzip` | 纯 Go；边缘格式 §5 |
| 图片尺寸/缩略图 | image_size.cr | `disintegration/imaging` + std image | 宽200/高300 缩放规则 |
| 配置 | YAML + macro | `gopkg.in/yaml.v3` + struct tag + env | 22 项，env 覆盖 |
| CLI | clim | `spf13/cobra` | 主命令 + admin user 子命令 |
| 模板 | ECR | `html/template` | 10 个模板机械转写 |
| 静态资源打包 | baked_file_system | `embed.FS` | public/ 嵌入二进制 |
| 密码 | Crypto::Bcrypt | `golang.org/x/crypto/bcrypt` | 哈希格式兼容，可验证旧密码 |
| HTTP 代理 | http_proxy | net/http Transport Proxy | 插件 get/post |
| 表格输出 | tallboy | `olekukonko/tablewriter` | CLI list |

## 3. 目标包结构（Go）

```
cmd/mango/main.go            # cobra 入口 + admin 子命令
internal/config/             # 配置加载（yaml+env）
internal/storage/            # SQLite 访问 + 迁移 + users/progress/thumbnails
internal/library/            # 扫描、title/entry、dir/archive entry、cache
internal/archive/            # 统一归档抽象（zip + rardecode + sevenzip）
internal/thumbnail/          # 尺寸读取 + 缩略图生成
internal/plugin/             # goja 沙箱 + 宝实函数 + v1/v2 + 订阅/更新/下载
internal/queue/              # 下载队列
internal/server/             # 路由注册、handler、鉴权/CORS/静态中间件
internal/web/                # 模板渲染、OPDS XML
web/views/*.tmpl             # 由 .ecr 转写
web/public/                  # embed 的静态资源
migration/                   # Go 迁移文件
```

## 4. 数据流与契约

- **DB 契约**：直接打开现有 `~/mango.db`。迁移器读取 `schema_migrations`（或复刻 mg 的
  版本记录方式），从当前版本继续；对已是最新版的旧库应为 no-op。所有表结构与列类型
  必须逐一比对 `migration/*.cr` 的最终形态（users、titles、ids、thumbnails、tags、
  hidden、sort_title、relative_path、unavailable、md_account 等）。
- **鉴权契约**：登录读 users 表 bcrypt 校验；token 生成/校验、`disable_login` +
  `default_username`、`auth_proxy_header_name` 逻辑等价。
- **插件契约**：`info.json`（id/title/placeholder/wait_seconds/api_version/settings）、
  `storage.json`、目录布局不变；宝实函数签名与返回结构与 duktape 版一致。
- **API 契约**：67 条路由的响应 JSON 字段名/结构冻结，以 Crystal 版为准。

## 5. 关键风险与缓解

- **R-JS（高）**：goja vs duktape 的 ES 兼容性差异（正则、Date、字符串方法等）。
  缓解：先做 goja PoC 跑 2 个真实插件（1×v1、1×v2）；对 `mango.*` 宝实函数逐个写
  等价性单测；escape_js/eval_json 语义对齐。若个别插件失败，记录差异清单。
- **R-ARCHIVE（中）**：纯 Go rar/7z 库可能不支持加密/分卷/老 rar4。
  缓解：设计阶段用现有库样本测试 rardecode/sevenzip 覆盖率；不支持的格式在启动日志
  明确告警，作为已知限制（PRD OQ1）。
- **R-SQLITE（中）**：modernc.org/sqlite 与 crystal-sqlite3 在 PRAGMA、并发、
  FTS 等行为差异。缓解：保留 `PRAGMA foreign_keys=1`；用真实生产 DB 拷贝做读写回归。
- **R-CONCURRENCY（中）**：Crystal 用 MainFiber 串行化 DB。Go 侧用连接池 + 
  适当锁/单写协程复现串行写语义，避免 SQLite 写锁竞争。
- **R-SESSION（低）**：升级后 cookie 不兼容 → 接受重登一次（PRD OQ2）。

## 6. 兼容性 / 回滚

- Go 版与 Crystal 版指向**同一数据目录只读对拿**行为；写操作在验证期用 DB 副本。
- 回滚点：Docker 部署可随时切回 Crystal 镜像，因 DB/插件/config 未被破坏性改动。
- 迁移器对生产库只做「幂等/前向」操作，不删除或重命名现有列。

## 7. 验证策略（已确认：黑盒对比测试）

- 搭建对比脚本：同一批请求（含鉴权 cookie）分别打 Crystal 版（基线，端口 A）与
  Go 版（端口 B），对比状态码 + 规范化 JSON。
- 覆盖 67 条路由的代表性用例；插件走 search/list/select/nextPage/download 全流程。
- 前端页面做人工 + 快照核对（10 个模板）。
