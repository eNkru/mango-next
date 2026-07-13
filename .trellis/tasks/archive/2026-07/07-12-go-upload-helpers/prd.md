# Move upload.cr helpers to Go

## Goal

把 Crystal `src/upload.cr` 的上传助手（`save` / `path_to_url`）与唯一调用点
`POST /api/admin/upload/:target`（cover 上传）搬到 Go，使管理员上传封面图行为与
Crystal 版一致；上传文件仍由已有 `UploadHandler` 中间件通过 `/uploads/*` 对外服务。

## Confirmed Facts

- Crystal 源：`src/upload.cr`（60 行），`UPLOAD_URL_PREFIX = "/uploads"`（`src/util/util.cr`）。
- 唯一业务调用：`src/routes/api.cr:783-836` → `target=cover` 时：
  1. 解析 multipart `file`
  2. 校验 MIME ∈ `SUPPORTED_IMG_TYPES`（JPEG/PNG/WebP/…）
  3. `Upload.new(upload_path).save("img", ext, body)` → 随机文件名
  4. `path_to_url` → 写入 `Title.set_cover_url`（title 级或 entry 级）
- Go 已有：`config.UploadPath`、`UploadHandler`（读文件）、路由
  `r.Post("/upload/{target}", s.apiAdminUpload)`。
- Go 缺口：`apiAdminUpload` 是 stub；无 `Upload` 包；无 `TitleInfo` / `set_cover_url`
  （cover 持久化依赖 `info.json` 的 `cover_url` / `entry_cover_url`）。
- 随机文件名：Crystal `random_str` = UUID 去横线；Go 已有
  `storage.randomStr()`（未导出），本任务应用同算法（`uuid` 去 `-`）。

## Requirements

### R1 — Upload 包（行为对拍 Crystal）

- 新增 `go/internal/upload`（或等价包名）：
  - `New(dir string) (*Upload, error)`：目录不存在则创建（对拍 Crystal 构造逻辑）。
  - `Save(subDir, ext string, r io.Reader) (string, error)`：
    写入 `{dir}/{subDir}/{random_str}{ext}`，返回绝对/完整路径。
  - `PathToURL(path string) (string, bool)`：
    仅当 path 落在 upload 目录下时返回 `"/uploads/..." + 相对路径`；
    否则返回 false（对拍 Crystal warn + nil）。
- 常量：`URLPrefix = "/uploads"`；图片 MIME 白名单与 Crystal `SUPPORTED_IMG_TYPES` 对齐。

### R2 — 实现 `apiAdminUpload`

- 对拍 Crystal：`target=cover` 唯一合法 target；其它 target → error。
- Query：`tid` 必填；`eid` 可选（有则按 entry **Name()** 写 entry cover，
  对拍 Crystal `get_entry(eid).title`）。
- Multipart 字段名 `file`；无文件 / 非法 MIME / 未知 target →
  **HTTP 200** + `{"success":false,"error":"..."}`（对拍 Crystal：本 handler
  的 rescue 也走 `send_json`，**不用** `sendJSONError` 的 4xx WriteHeader，
  除非项目其它 admin API 已统一改成 4xx——则与现有 Go admin 错误风格对齐并
  在实现注释写明偏差）。
- 成功：**HTTP 200** + `{"success":true}`。
- Title 查找：与 `apiBook` 一致，用 `lib.TitleHash[tid]`（见 Risks）。

### R3 — 最小 cover 持久化（本任务必要依赖）

- 为实现 R2，需最小能力：
  - 按 `tid` 找到 `Title`（库已有路径/ID 查找则复用）。
  - 写 `info.json` 的 `cover_url` 或 `entry_cover_url[entry_name]`（对拍
    `Title#set_cover_url`）。
- **不**要求完整移植 `TitleInfo` 全部字段的读写生命周期；其它字段
  （progress/display_name/…）可只 round-trip 保留，不丢现有 JSON。

### R4 — 测试

- Upload 包：Save 落盘、PathToURL 正常/越界、子目录创建。
- Handler：cover 成功路径（可用 httptest + 临时 upload/library 目录）或
  至少 Upload + SetCover 的单元测试；非法 MIME / 缺 file 失败路径。

## Acceptance Criteria

- [x] `go/internal/upload` 存在且单测覆盖 Save / PathToURL。
- [x] `POST /api/admin/upload/cover?tid=...` 已实现（Save + PathToURL + SetCover*）；
      文件由既有 `/uploads/*` 中间件服务。
- [x] 非法 MIME / 未知 target / 缺 file 返回 HTTP 200 + `success:false`。
- [x] `go build/vet/test` 全绿 — **178 tests**（baseline 170+）。

## Out of Scope

- 完整 `TitleInfo`（progress / last_read / sort_by 等）功能迁移 → 归后续 library 对拍。
- 插件/队列相关 admin upload 扩展（Crystal 也只有 cover）。
- UI / 前端改动。
- rename / signature / routes 全量对拍（其它 child 任务）。

## Dependencies

- 依赖 parent：`07-12-crystal-go-migration`（范围说明）。
- 不依赖其它 sibling child 先完成。
- **被依赖**：`go-routes-coverage` / `go-gap-regression` 会假定本任务完成后
  upload 路由可用。

## Risks（规划审查 2026-07-12）

- **TitleHash 仅含顶层 Title**：`Library.Scan` 只把 `result.Titles`（顶层）放进
  `TitleHash`；嵌套子书在 parent 的 `TitleIDs` 里，但未必能按 tid 查到。
  现有 `apiBook` 同样只查 `TitleHash`。本任务**不扩大**为全树索引；cover
  上传与 `apiBook` 同可见性。全树 GetTitle 归 `go-routes-coverage` / library 对拍。
- **MIME 文案 vs 白名单**：Crystal 报错写 “JPEG or PNG”，实际校验完整
  `SUPPORTED_IMG_TYPES`。Go 校验跟白名单；文案可照抄 Crystal 或写更准确列表。
- **写入 info.json 的 URL**：存 `path_to_url` 结果（`/uploads/img/...`），**不含**
  `base_url`（Crystal `set_cover_url` 如此；读 cover 时再 join base_url）。

## Notes

- 实现注释保留 `// mirrors Crystal src/upload.cr` 等引用，便于 gap-report 对拍。
- `storage.randomStr` 未导出时，在 upload 包内用同算法私有 helper，避免跨包耦合。
- `Title.Dir` 字段已存在（`go/internal/library/title.go`）；`Entry.Name()` /
  `Entry.ID()` 已存在。缺的是按 eid 查找 entry 的小 helper（循环 `t.Entries`）。
