# Implement: go-upload-helpers

## Ordered Checklist

### 1. Upload package + unit tests

- [x] 创建 `go/internal/upload/upload.go`
- [x] 创建 `go/internal/upload/upload_test.go`
- [x] 验证：`cd go && go test ./internal/upload/...`

### 2. Minimal TitleInfo + SetCoverURL

- [x] `title_info.go`：Load/Save，保留未知 JSON 键
- [x] `SetCoverURL` / `SetEntryCoverURL` / `EntryByID`
- [x] 单测 round-trip
- [x] `go test ./internal/library/...`

### 3. apiAdminUpload

- [x] 重写 stub：cover only、MIME 扩展名、TitleHash、HTTP 200 errors
- [ ] 可选：httptest 冒烟（未做；包/library 单测已覆盖主路径）

### 4. Quality gate

- [x] `go build ./... && go vet ./... && go test ./...` — **178 tests passed**
- [ ] 手工（可选）

### 5. Parent 回写

- [x] parent prd upload 行可标 ✅（实现完成）
- [ ] 全量 gap-report → `go-gap-regression`

## Validation Commands

```bash
cd go
go build ./...
go vet ./...
go test ./internal/upload/... ./internal/library/... ./internal/server/...
go test ./...
```

## Risky Files / Rollback Points

| File | Risk | Rollback |
|------|------|----------|
| `go/internal/server/handlers_api.go` | 改 stub 可能影响 admin 路由 | 恢复 stub |
| `go/internal/library/*` | info.json 写坏会丢字段 | 保留未知键 + 单测 |
| `go/internal/upload/*` | 新文件，可整删 | rm package |

## Review Gates

- 实现前：本 `prd.md` + `design.md` + `implement.md` 已审阅。
- 实现后：`trellis-check` 或本地 full `go test`。
- 启动实现：`python3 ./.trellis/scripts/task.py start 07-12-go-upload-helpers`
  （仅在用户明确批准后）。

## Do Not Start Until

- [x] prd / design / implement 已写入
- [ ] 用户 review 通过或明确说 “start”
