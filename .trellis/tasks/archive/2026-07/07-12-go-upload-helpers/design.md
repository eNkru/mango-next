# Design: go-upload-helpers

## Architecture & Boundaries

```
Client multipart POST
  → chi: /api/admin/upload/{target}  (AdminMiddleware)
  → apiAdminUpload
      → parse form "file"
      → MIME check (SupportedImgTypes)
      → upload.New(cfg.UploadPath).Save("img", ext, body)
      → u.PathToURL(path)
      → library Title lookup + SetCoverURL (info.json)
  ← {"success": true|false, "error"?}

GET /uploads/...  (existing UploadHandler middleware, unchanged)
```

| Layer | Package | Responsibility |
|-------|---------|----------------|
| Helpers | `go/internal/upload` | 新建；Save / PathToURL / URLPrefix |
| Library | `go/internal/library` | 最小 `TitleInfo` 或 `Title.SetCoverURL` |
| HTTP | `go/internal/server` | 重写 `apiAdminUpload` stub |

## Contracts

### Upload.Save

| Input | Behavior |
|-------|----------|
| `subDir`, `ext`, `io.Reader` | `mkdir -p {dir}/{subDir}`；文件名 `{uuid_no_dash}{ext}`；写盘；返回完整路径 |
| 目录不可写 | return error |

对拍 Crystal：`random_str + ext`，`IO.copy`。

### Upload.PathToURL

| Input | Behavior |
|-------|----------|
| path 在 upload 根下 | `"/uploads/" + rel`（URL 路径用 `/`，对拍 `File.join` + `UPLOAD_URL_PREFIX`） |
| path 不在根下 | `( "", false )` + log warn |

实现注意：

- Crystal 用 `Path.each_part` + `File.same?` 逐段匹配根；Go 用
  `filepath.Abs` + `filepath.Rel` + `!strings.HasPrefix(rel, "..")` 更稳妥，
  **语义等价**即可（同源路径、symlink 边界与 Crystal 一致优先测试常规路径）。
- 不要把 Windows 反斜杠暴露进 URL。

### apiAdminUpload

| Case | Response body | HTTP status（对拍 Crystal） |
|------|---------------|---------------------------|
| target ≠ `cover` | `success:false`, error 含 unknown/Unkown target | **200**（Crystal rescue send_json） |
| 无 `file` part | `success:false` | **200** |
| MIME 不在白名单 | `success:false`, "must be either JPEG or PNG" 类 | **200** |
| tid 无效 / Title 不在 TitleHash | `success:false` | **200** |
| 成功 | `success:true`；info.json 已写；磁盘有文件 | **200** |

实现：优先 `sendJSON(w, map[string]any{"success": false, "error": msg})`，
**不要**对本 handler 用 `sendJSONError`（它会 `WriteHeader(4xx)`），除非明确
决定偏离 Crystal 并与其它 Go admin 错误统一。

Query：`tid` 必填；`eid` 可选 → 有 eid 时：

1. 在 `t.Entries` 中找 `e.ID() == eid`（新增小 helper `entryByID`）。
2. 用 `e.Name()` 作为 `entry_cover_url` 的 key（对拍 Crystal entry `.title`）。

Title 查找：

```go
lib.RLock()
t, ok := lib.TitleHash[tid]
lib.RUnlock()
```

与 `apiBook` 相同。**不**在本任务构建全树 Title 索引。

MIME：对拍 Crystal `MIME.from_filename?(filename)` → 用
`mime.TypeByExtension(filepath.Ext(filename))`（或等价注册表），**不要**用
`http.DetectContentType` 作唯一依据（内容嗅探与扩展名校验不一致）。

MIME 白名单（Crystal）：

```
image/jpeg, image/png, image/webp, image/apng, image/avif,
image/gif, image/svg+xml, image/jxl
```

### Minimal TitleInfo / SetCoverURL

**已确认**：`Title.Dir string` 已存在；`Entry.ID()` / `Name()` 已存在。

方案（推荐）：

```go
// library/title_info.go — preserve unknown keys via map[string]any merge
func LoadTitleInfoMap(dir string) (map[string]any, error)
func SaveTitleInfoMap(dir string, m map[string]any) error

func (t *Title) SetCoverURL(url string) error           // sets cover_url
func (t *Title) SetEntryCoverURL(entryName, url string) error // entry_cover_url[name]
```

- `info.json` 路径：`filepath.Join(t.Dir, "info.json")`。
- 写入的 `url` 必须是 `/uploads/...` 相对路径（**无** base_url）。
- 写文件：Crystal `to_pretty_json`；Go 推荐 `json.MarshalIndent` 便于 diff。
- **保留未知字段**：`map[string]any` 读改写；只动 `cover_url` /
  `entry_cover_url`；`entry_cover_url` 若缺失则初始化为 `map[string]any`。

## Compatibility

- 配置键 `upload_path` / `UPLOAD_PATH` 已存在，不改 schema。
- 不改 SQLite migration。
- 不改静态中间件行为；仅新增写盘侧。

## Trade-offs

| 选项 | 选择 | 原因 |
|------|------|------|
| 新包 `internal/upload` vs 塞进 `server` | 新包 | 对拍 Crystal 独立文件；可单测 |
| PathToURL 用 Abs+Rel vs 逐段 same? | Abs+Rel | 更惯用、可测；加测试覆盖同目录 |
| 完整 TitleInfo vs 最小 cover 字段 | 最小 + 保留未知键 | 本任务不吞掉进度数据 |
| 导出 randomStr 到 util | 包内私有 | 避免动 storage 公共 API |

## Rollback

- 删除 `internal/upload`、回退 `apiAdminUpload` stub、回退 TitleInfo 相关文件即可；
  无 DB 迁移。

## Open Implementation Notes

- [x] `Title.Dir`、`Entry.ID`/`Name` 已确认。
- Title 查找：复用 `TitleHash`，不新增全树 `GetTitle`（嵌套 tid 与 apiBook 同限）。
- 必做小 helper：`func entryByID(t *Title, eid string) Entry`（扫 `t.Entries`）。
