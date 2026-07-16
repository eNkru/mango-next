# Fix nested subtitle titles missing from library

## Goal

让 library 下 **多级嵌套目录** 中的卷（zip/rar 等）可被扫描、在 UI 中打开并阅读；修复子标题只记 ID、对象未进入 `TitleHash` 导致书页空白的问题。

## User value

用户可保持常见整理方式：

```text
library/
  Series Name/
    Part 7/
      Vol.15.zip
      Vol.16.rar
```

打开「Series Name」能看到子系列与卷，而不是空书 / 扫不到。

## Confirmed facts

1. `ScanLibrary` 只把一级目录当 top-level title；`NewTitle` 会递归子目录。
2. 子目录若含 archive 会生成子 `Title`，父级只 `append` 子 `Title.ID` 到 `TitleIDs`，**子 `*Title` 对象随后丢弃**（`title.go` `NewTitle`）。
3. `Library.applyTitles` 只把 top-level 列表写入 `TitleHash`；嵌套 title 不在 hash 中。
4. `handleTitle` / API 通过 `lib.TitleHash[subID]` 解析子书；找不到则 **静默跳过**。
5. 库缓存 `titlesToCache` 同样只序列化 top-level；嵌套 title 即使内存曾有也无法从 cache 恢复。
6. `DeepEntries` 注释写明：有 `TitleIDs` 时仍只返回直接 entries（嵌套卷不计入）。
7. 用户路径：
   - `library/JOJO (PART1-PART7)  DIGITAL COLORED VERSION/[JOJO的奇妙冒險PART.7…][…]/JOJO.Part7….Vol.15.zip`
   - 同目录 `…Vol.16.rar`
   - 扩展名合法；失败原因是嵌套 title 丢失，不是 rar/zip 不支持。

## Requirements

- **R1** 扫描后，所有有内容的 title（含任意深度嵌套）必须进入 `Library.TitleHash`。
- **R2** 顶层 `Library.TitleIDs` 仍只含 library 根下一级 title（书架只显示一级）。
- **R3** 打开父 title 页/API 时，能列出子 title 及其下 archive/dir entries。
- **R4** 嵌套 archive entries 可打开阅读（cover / page / dimensions 路径可用）。
- **R5** 库缓存 load/save 后嵌套 title 与 entries 不丢。
- **R6** 增量扫描：top-level `DirSignature` 命中复用时，其子树仍完整留在 `TitleHash`。
- **R7** `DeepEntries` 对嵌套子树递归汇总 entries（缩略图、进度等依赖方受益）。
- **R8** 回归测试覆盖至少两级嵌套 archive 结构；不回归现有单层/DirEntry 用例。

## Acceptance Criteria

- [ ] 合成库：`Series/Part/vol.zip` 扫描后 `TitleHash` 含 Series 与 Part；Part 的 entries 含 vol。
- [ ] `Library.TitleIDs` 仅含 Series（一级），不含 Part。
- [ ] `GET`/title 页等价路径能看到 Part 与卷（或单测模拟 `TitleHash` 解析成功）。
- [ ] cache save → load 后嵌套关系仍成立。
- [ ] 现有 `go test ./internal/library/... ./internal/server/...` 通过。
- [ ] 用户 JOJO 式两级路径可扫可读（手动或等价单测）。

## Out of scope

- 改用户目录命名规范
- 根目录直接放 zip 当书（仍要求至少一层目录）
- 新前端布局
- 排序算法大改（TitleIDs 按 ID 数字排序的历史问题可顺手修，非必须）

## Decisions (planning)

1. **内存：** `Title` 增加 `Children []*Title`；`NewTitle` 同时写 `TitleIDs` + `Children`。
2. **TitleHash：** `applyTitles` 递归放入全部深度；`TitleIDs` 仅一级。
3. **缓存：** 扁平 `titles[]` 含全部 title；load 时按 `title_ids` 重建 `Children`。
4. **DeepEntries：** 递归 `Children`。

## Open questions

（无阻塞项）

## Notes

- 分支：`fix/nested-title-hash`
- 任务：`.trellis/tasks/07-16-fix-nested-title-hash`
- 规划产物：`prd.md` + `design.md` + `implement.md` — 待 review 后 `task.py start`
