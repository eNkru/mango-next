# Non-blocking library scan at startup

## Goal & User Value

启动后 HTTP 立刻可用；全量扫库不得长时间占用 `Library` 写锁导致首页/API 挂起数分钟。NAS 大库（~50 titles / 1000+ entries）上，用户打开 UI 应在秒级得到响应（可能短暂显示空库或旧状态），扫库在后台完成后再更新内存树。

## Confirmed Facts

- `Library.Scan()` 在整段 `ScanLibrary`（磁盘遍历 + 建 Title）期间持有 `mu.Lock()`（`library.go`）。
- 页面/API 用 `RLock()` 读 `TitleHash`；写锁期间读请求阻塞。
- `ScanLibrary` 本身不依赖 `lib.mu`；可先无锁构建 `ScanResult`，再短锁替换内存树。
- 后台任务：`tasks.Runner` 在 goroutine 里 `runScan()`；默认 `scan_interval_minutes: 5`，启动不同步阻塞 `srv.Start`。
- 用户生产日志：扫库 **12m26s**；`GET /` 等 **8m+** 后 `broken pipe` / 500。
- 缩略图 `GenerateThumbnails` 持 `RLock`，不阻塞其他 `RLock` 读者；本次以扫库写锁为主。

## Decisions

1. **分支**：`fix/nonblocking-library-scan`（已创建）。
2. **策略**：无锁执行 `ScanLibrary`，仅在替换 `TitleIDs`/`TitleHash` 时短写锁。
3. **首扫空库**：允许扫完前首页 `empty_library`/列表偏空；不强制做 “scanning” UI（可选 polish）。

## Requirements

### R1 — Scan 不长时间持写锁
- `Scan()` 磁盘扫描与 Title 构建不在 `mu.Lock()` 下执行。
- 仅原子替换内存索引时持写锁，持锁时间与 title 数量线性且应为毫秒～秒级，而非分钟级。

### R2 — HTTP 在扫库期间可响应
- 启动后 `GET /`、登录页等不因扫库等待数分钟。
- 扫库进行中读到的库状态可为扫前状态或空；扫完后一致。

### R3 — 行为正确性
- 扫完后 title/entry 计数与现逻辑一致；DB ID 持久化与 stale 标记仍由 `ScanLibrary` 路径完成。
- 现有 `go test ./...` 通过；补充锁持有时间或并发读不被长时间阻塞的测试（可测短锁语义）。

### R4 — 不扩大范围
- 不重做增量 examine / ContentsSignature 接线。
- 不强制修 WebP 缩略图 `riff` 错误（可另开任务）。

## Acceptance Criteria

- [x] `Library.Scan` 实现为：无锁 `ScanLibrary` + 短锁 swap。
- [x] 文档或测试说明：并发 `RLock` 读者在长扫期间不被写锁饿死数分钟。
- [x] `cd go && go test ./...` 全绿。
- [ ] 人工/日志：启动后立刻 `Starting server`，扫库日志仍后台进行；扫库中访问 `/` 不出现多分钟挂起（部署验证）。

## Out of Scope

- 增量扫库算法优化（减少 12 分钟本身）— 可 follow-up。
- 缩略图 WebP/RIFF 解码。
- 首页 “正在扫描” 横幅 UI（可选后续）。

## Open Questions

- 无阻塞规划的问题；实现前用户确认 design 即可。
