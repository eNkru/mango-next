# Optimize first book open load time

## Goal

缩短用户打开某一卷/条目（entry）阅读器时的等待时间。打掉最大头：`api/dimensions` 全量逐页读图 + 无缓存；不改阅读器前端等待整卷 dimensions 的交互形状。

## User value

进入阅读器时长时间停在「加载网络阅读器…」。优化后：

- **某卷第一次打开**：计算尺寸时更省（批量/复用归档打开，避免 N 次 open+list+全量读的灾难形态）
- **再次打开同一卷（内容未变）**：dimensions 从持久缓存返回，不再逐页读图
- 客户端行为与 API 响应形状保持兼容

## Confirmed facts (from code)

1. `/reader/{title}/{entry}/{page}` → `handleReader` 只渲染壳，不带尺寸。
2. `reader.js` `init()` 阻塞于 `GET api/dimensions/{tid}/{eid}`，完成后才 `loading=false`。
3. `apiDimensions` 对每页 `ReadPage` 全量读字节 + `DecodeConfig`；无 dimensions 缓存。
4. `ArchiveEntry.ReadPage` 每次 open → list → sort → 读页 → close。
5. `cache_enabled` / `cache_size_mbs` 在 Go 中未用于页面缓存。
6. Library 树缓存与封面缩略图与本路径无关。

## Product decisions

| Decision | Choice |
|----------|--------|
| 痛点定义 | 用户不确定「首次」精确含义 → 按最痛路径：卷冷路径 + 每次都慢一并优化 |
| 验收标准 | **行为可证**（缓存命中、无 N 次 open、signature 失效、功能不回归）；无硬性毫秒 SLA |
| 渐进首屏 | **不做**；不改 `reader.js` 等待整卷 dims |
| 计算时机 | **仅懒计算 + 持久化**；不在 scan/thumbnail 后台预热 |
| 前端范围 | 服务端为主；API JSON 形状保持 `{success, dimensions:[{width,height}]}` |

## Requirements

- **R1** 同一 entry、内容未变时，第二次及以后 `api/dimensions` 必须走持久缓存，不得对每一页完整 `ReadPage`。
- **R2** 缓存未命中时的首次计算：对同一归档不得做 N 次独立 open+list；应单次打开批量取尺寸（或等价更轻路径）。
- **R3** 失效键：至少绑定 entry 身份与内容签名（现有 `Signature()` / 路径身份）；变更后必须 miss 并重算。
- **R4** 响应兼容：每页 `width`/`height`；无法识别时 `0,0`（与现状一致）。
- **R5** `DirEntry` 与 `ArchiveEntry` 均需正确缓存与失效（实现细节可不同，契约相同）。
- **R6** 测试覆盖：缓存命中、失效重算、首次路径不 N 次 open（或可观测的批量 API）、关键回归。

## Acceptance Criteria

- [ ] 第二次 `GET api/dimensions` 走缓存（测试证明无 per-page full ReadPage）
- [ ] 首次 miss 计算不 N 次独立 archive open+list
- [ ] signature/内容变更后缓存失效并得到新尺寸
- [ ] 支持格式下 width/height 与优化前语义一致（含 0,0）
- [ ] 阅读器可进入、连续/分页可看、翻页与进度仍可用
- [ ] 无硬性毫秒 SLA

## Out of scope

- 渐进首屏 / 改 `reader.js` 加载时序
- scan/thumbnail 后台预计算 dimensions
- 完整 page image LRU（`cache_enabled` 接线）
- 插件下载、整库扫描速度、封面链路大改
- 硬性生产压测基线

## Open questions

- 无阻塞产品问题；技术选型写入 `design.md`。

## Notes

- Research: `research/book-open-path.md`
- Complex task → 需要 `design.md` + `implement.md`，用户审阅后 `task.py start`
