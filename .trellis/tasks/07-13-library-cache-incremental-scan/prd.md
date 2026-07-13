# Load library from cache and skip unchanged scans

## Goal & User Value

1. **启动即有书**：进程起来后尽快从磁盘缓存恢复书库树，UI 不必等 10+ 分钟全量扫。
2. **重复扫要快**：周期扫/二次扫对**未变更**的 title 不重做昂贵的全树 `NewTitle`/打开归档，只校验签名；已扫过且未变的从缓存/内存复用。

## Confirmed Facts

- 当前 `ScanLibrary` 注释写明：`fresh scan without cache`；每次启动与 `scan_interval` 都是全量。
- `ContentsSignature` / `DirSignature` / `Title.ContentsSig` 已存在，但未用于跳过重扫。
- `Storage.SaveLibraryCache` / `LoadLibraryCache`（`library.yml.gz`）已有，**库树逻辑未接入**。
- Config 有 `library_cache_path`（默认 `~/mango/library.yml.gz`）；storage 实现实际用 `filepath.Dir(libraryPath)/library.yml.gz`，可能与 config 不一致（实现时对齐）。
- Crystal 曾有 `cache.cr`，Go 迁移时简化掉。
- 非阻塞扫库已合并：扫库不堵 HTTP，但首扫前内存树为空 → 无书。

## Decisions

1. 启动：`LoadCache` → 立刻 swap 进 `Library` → 后台 `Scan`（增量）。
2. 扫库：顶层目录对比签名；未变则复用缓存 Title 子树；变了/新增才 `NewTitle`。
3. 扫完：写回 `library.yml.gz`。
4. **OQ1：两者一起做**（启动加载 + 增量跳过）。

## Requirements

### R1 — 启动加载缓存
- 若缓存文件存在且可解析，在 HTTP 可用后尽快（或启动扫库前）填充 `TitleIDs`/`TitleHash`。
- 缓存损坏/缺失：行为与现在相同（空树 + 全量扫），不崩溃。

### R2 — 增量扫库
- 对每个顶层 title 路径：若缓存中有且 **DirSignature（或 ContentsSig）未变**，跳过深度重建，复用缓存节点。
- 路径消失 → 不进入新树（既有 MarkUnavailable 路径尽量保持）。
- 新路径 / 签名变 → 全量 `NewTitle`。

### R3 — 持久化
- 成功扫库后保存缓存。
- 缓存路径优先 `config.LibraryCachePath`（expand ~），与 Crystal/配置一致。

### R4 — 并发
- 保持：磁盘扫不长期持 `mu`；仅 swap/写缓存时短锁。
- 加载缓存的 swap 与 Scan 的 swap 序列化（`scanMu` 或同等）。

### R5 — 测试
- 缓存 round-trip；二次扫对未改库明显更轻（可用调用计数/时间或 mock）；损坏缓存回退。

## Acceptance Criteria

- [x] 有有效缓存时：启动后无需等完整 `ScanLibrary` 即可在内存中看到 title（`LoadFromCache` + 测试）。
- [x] 无变更二次 `Scan`：不重扫未变 title（`Reused`/`Rebuilt` 计数 + 测试）。
- [x] 改一个 title 的文件后：该 title 被重建，其它可跳过。
- [x] `go test ./...` 绿（195）。

## Out of Scope

- 缩略图 WebP/RIFF 修复。
- 真正的“只读 DB 路径表拼出完整 Entry 而不碰磁盘”（缓存仍是主路径；DB 只存 id/path/signature）。
- 分布式锁/多实例。

## Open Questions

- （无）OQ1 已确认：启动加载 + 增量跳过一并实现。
