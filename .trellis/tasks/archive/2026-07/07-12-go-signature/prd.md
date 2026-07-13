# Move util/signature.cr to Go

## Goal

对齐 Crystal `src/util/signature.cr` 与库扫描 ID/变更检测契约：文件/目录
signature、contents signature（rescan）、directory entry signature。在 Go 端
形成可测试、可引用的 API，并补上当前缺失的 **contents signature / examine-rescan**
路径。

## Confirmed Facts

### Crystal（`src/util/signature.cr`）

| API | 算法 | 用途 |
|-----|------|------|
| `File.signature(path) → UInt64` | 支持 archive/image → **inode**；否则 0 | entry/title ID 稳定键 |
| `Dir.signature(dir) → UInt64` | 自身 inode + 子项 signature，**sort 后 CRC32** | `Title.signature` → DB title id |
| `Dir.contents_signature(dir, cache) → String` | 支持文件名列表（递归）**SHA1 hex** | 判断是否需要 rescan |
| `Dir.directory_entry_signature(dir, cache) → String` | 有序图片的 `File.signature` 字符串拼 SHA1 | `DirEntry` 存在性/ID |

调用点：`title.cr`、`archive_entry.cr`、`dir_entry.cr`、`library.cr` examine。

### Go 现状（`go/internal/library/title.go`）

| API | 现状 |
|-----|------|
| `fileSignature` | ✅ 存在；**非 inode**：FNV64(path + mtime + size) |
| `dirSignature` | ✅ 存在；WalkDir FNV(rel + mtime)，**非** CRC32(inodes) |
| `dirEntrySignature` | ✅ 存在；FNV(mtime+size of files)，**非** SHA1(inode strings) |
| `contents_signature` | ❌ **无** |
| examine/rescan 用 contents sig | ❌ **无**（scanner 偏 full scan） |

注释已写明：Go 用 path+mtime+size 替代 inode（跨平台/可移植）。

### 数据层

- `storage.GetOrCreateTitleID/EntryID(path, sig uint64)` 已存在；sig 存在 DB 为字符串。
- **改算法会改变新扫描产生的 sig**；对已有 DB：path 命中仍可复用 id，sig 更新行为取决于 storage 实现（实现前读 `GetOrCreate*`）。

## Requirements

### R1 — 统一 Signature API 面

- 在 `go/internal/library`（或 `internal/signature` 若需解耦）导出清晰命名，例如：
  - `FileSignature(path string) (uint64, error)`
  - `DirSignature(dirname string) (uint64, error)`
  - `ContentsSignature(dirname string, cache map[string]string) (string, error)`
  - `DirectoryEntrySignature(dirname string, cache map[string]string) (string, error)`（或保持现有 `dirEntrySignature` 并文档化）
- 现有 `fileSignature` / `dirSignature` / `dirEntrySignature` 调用点改为走统一 API（行为可渐进）。

### R2 — 补齐 ContentsSignature

- 行为对拍 Crystal 意图：
  - 跳过 `.` 前缀；
  - 目录递归；文件仅统计 supported archive/image **basename** 列表（Crystal 用 `fn` 文件名）；
  - 结果 SHA1 hex；支持 cache map。
- 供后续 library examine/rescan 使用；本任务至少：**API + 单测**。若 wire 进 `Scan` 成本可控，可一并做最小「内容变了才重扫子树」；否则明确留给 routes/library 后续，本任务只交付 API。

### R3 — 算法策略（决策见下，写入 design）

**推荐（默认）**：保留 Go 现有 **path+mtime+size / FNV** 作为 file/dir/entry 的稳定策略（文档写清与 Crystal inode/CRC32 的差异与原因），**严格实现 ContentsSignature 的 SHA1 文件名语义**（与 Crystal 可对拍、与平台无关）。

**备选**：尝试 `syscall.Stat_t.Ino` 做 file signature（Unix only），CRC32 拼目录——更贴近 Crystal，但 macOS/Linux/Docker 卷、复制文件行为差异大，且与已写入 DB 的 FNV sig 不兼容。

### R4 — 测试

- ContentsSignature：增删文件 → hash 变；无关文件（非 archive/image）不进 hash。
- File/Dir signature：同内容路径改 mtime/size → hash 变（在推荐策略下）。
- 回归：`go test ./internal/library/...` 及全量 `go test ./...`。

## Acceptance Criteria

- [ ] 有文档化的 Signature API；ContentsSignature 已实现 + 单测。
- [ ] 现有 title/entry 创建仍通过测试；注明与 Crystal inode 算法差异。
- [ ] `cd go && go build ./... && go vet ./... && go test ./...` 全绿。
- [ ] parent gap 表 signature 行可更新为 ✅ 或「API 齐 / rescan wire 可选」。

## Out of Scope

- 完整 library examine 状态机（deleted_title_ids 等）全量移植——可在 R2 wire 最小钩子，完整 examine 归 routes/library 对拍。
- rename DSL、upload、util 其它文件。
- 强制迁移已有 SQLite 中的 signature 列到新算法（若选 inode）。

## Dependencies

- parent：`07-12-crystal-go-migration`
- 不依赖 upload-helpers（已完成归档）
- **被依赖**：gap-regression；library rescan 质量

## Decisions

1. **算法（已确认 2026-07-12）**：保留 Go **FNV(path+mtime+size)** 作为
   file/dir/entry 的 uint64 signature；**不**切换到 Unix inode/CRC32。
2. **ContentsSignature**：严格对拍 Crystal 意图（supported 文件名递归列表 → SHA1 hex + cache）。
3. **ID 稳定性**：依赖 `GetOrCreateTitleID/EntryID` 的 **path-only 回退**（sig 变则 UPDATE signature，id 不变）。已验证 storage 实现支持。

## Notes

- Crystal：inode 在 rename 时往往稳定；Go FNV 含 path，**rename 会变 sig**，但
  path 命中仍保留 id（见 storage step 2）。
- 完整 `Title#examine` 状态机可只交付 API；最小 wire 可选（见 design）。
