# Fix thumbnail generation RIFF/WebP decode failure

## Goal

让漫画压缩包（zip/cbz 等）在后台生成封面缩略图时，对常见图片格式（至少 JPEG/PNG/GIF/WebP）能成功解码并生成缩略图，不再因误走 WebP 解码路径而失败。

## User value

用户打开书库时能看到封面缩略图；管理员触发 “生成缩略图” 后，日志中不再对正常可读的封面页出现 `riff: missing RIFF chunk header`。

## Confirmed facts (from codebase)

- 失败日志来自 `library.GenerateThumbnails`：`ReadPage(1)` 成功后调用 `thumbnail.Generate` 失败并打印  
  `Failed to generate thumbnail for %s: %v`（`go/internal/library/library.go`）。
- `thumbnail.Generate` / `DecodeConfig` 流程：
  1. `image.DecodeConfig` / `image.Decode`
  2. 失败则 fallback `webp.DecodeConfig` / `webp.Decode`（`go/internal/thumbnail/thumbnail.go`）
- 错误 `riff: missing RIFF chunk header` 来自 `golang.org/x/image/riff`，由 WebP 解码器在输入不是 RIFF/WebP 时返回。
- 生产路径中 `thumbnail` 包 import 了 `image/jpeg`（用于编码输出）和显式 `webp`；**没有**注册 `image/png`、`image/gif` 解码器。
- 因此当封面页是 PNG/GIF 时：`image.Decode*` 失败 → WebP fallback → 出现上述 RIFF 错误。
- 库层将 `.png` / `.gif` / `.webp` / `.avif` / `.jxl` 等标为 “支持的图片文件”，但缩略图解码能力与支持列表不一致。
- 单元测试在 `thumbnail_test.go` 中 import 了 `image/png`（仅测试包生效），因此 `TestGeneratePNGInput` 在测试进程里能过，**掩盖了生产二进制缺少 PNG 解码器注册的问题**。

## Requirements

1. 缩略图生成必须能解码至少以下格式的输入页：JPEG、PNG、GIF、WebP。
2. 解码失败时错误信息应反映真实原因（非法/不支持格式），避免非 WebP 输入被误报为 `riff: missing RIFF chunk header`。
3. 现有缩略图尺寸策略与输出 MIME（JPEG）行为保持不变。
4. 增加回归测试：在 **不依赖测试包副作用掩盖生产 import** 的前提下，证明 PNG（以及需要的其他格式）可被 `thumbnail.Generate` 成功处理。
5. 对无法解码的页继续跳过并打日志，不中断整批缩略图任务。

## Acceptance Criteria

- [ ] 使用 PNG 作为第 1 页的 zip/cbz（或等价内存数据）调用 `thumbnail.Generate` 成功，输出合法 JPEG 缩略图。
- [ ] JPEG / WebP 输入仍可生成缩略图。
- [ ] GIF 输入可生成缩略图（至少首帧 / 静态可用）。
- [ ] 非图片或损坏数据仍返回错误，且错误不再强制表现为 `riff: missing RIFF chunk header`（除非输入确实是损坏的 WebP/RIFF）。
- [ ] `go test` 覆盖 `thumbnail`（及相关必要用例）通过。
- [ ] 文档/行为上与 library 的 “支持图片扩展名” 至少在 JPEG/PNG/GIF/WebP 上对齐。

## Scope decision

- **In scope this round:** JPEG / PNG / GIF / WebP only（用户已确认：先做推荐方案）。
- **Explicitly deferred:** AVIF / JXL / SVG 完整解码与缩略图支持。

## Out of scope

- 为 AVIF / JXL / SVG 增加完整解码依赖与缩略图支持。
- 改变缩略图尺寸、质量、存储 schema。
- 修复除缩略图解码外的阅读器渲染问题。
- 用户书库中已损坏文件的自动修复。

## Behavioral decisions

1. **格式范围：** 仅 JPEG / PNG / GIF / WebP。
2. **封面页选择：** 仍只使用 page 1。解码失败则跳过该 entry 并打日志，**不**尝试后续页面。

## Open questions

（无阻塞项）

## Notes

- 分支：`fix/thumbnail-riff-decode`
- 任务目录：`.trellis/tasks/07-16-fix-thumbnail-riff-decode`
- 轻量 bugfix：PRD-only 即可（注册解码器 + 收紧 WebP fallback + 回归测试）。
- 实现前需用户 review 本 PRD 后 `task.py start`。
