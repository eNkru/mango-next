# Crystal → Go migration gap report

**Date:** 2026-07-12  
**Branch:** `migrate-to-go`  
**Tests:** `cd go && go build ./... && go vet ./... && go test ./...` → **192 passed**

## Legend

| Mark | Meaning |
|------|---------|
| ✅ | Ported / covered |
| 🔧 | Partial / intentional algorithm difference |
| ⊘ | Crystal-only or unused at runtime |
| ❌ | Still missing |

## Core modules

| Crystal | Go | Status | Notes |
|---------|-----|--------|-------|
| `config.cr` | `internal/config` | ✅ | YAML + env |
| `storage.cr` + migrations | `internal/storage` | ✅ | progress migration 14 |
| `archive.cr` | `internal/archive` | ✅ | zip/rar/7z pure Go |
| `library/*` | `internal/library` | ✅ | scan, sort, title/entry |
| `util/signature.cr` | `library/signature.go` | 🔧 | FNV not inode; **ContentsSignature** SHA1 ✅ |
| `util/chapter_sort` / numeric | `library/sort.go` | ✅ | |
| `util/proxy.cr` | plugin clients | ✅ | ProxyFromEnvironment |
| `util/validation.cr` | storage + validateZip | ✅ | |
| `util/web.cr` | server middleware | ✅ | macros → helpers |
| `upload.cr` | `internal/upload` + apiAdminUpload | ✅ | cover + info.json |
| `rename.cr` | `internal/rename` | ✅ | DSL + tests; no runtime wire |
| `queue.cr` | `internal/queue` | ✅ | |
| `plugin/*` | `internal/plugin` | ✅ | goja sandbox, subs, DL |
| `handlers/*` | server middleware/auth | ✅ | |
| `routes/*` | `RegisterRoutes` | 🔧 | Core + plugin re-enabled; see routes |
| `main_fiber.cr` | sequential + tasks | 🔧 | no fiber DB serialize |
| templates `.ecr` | `web/views/*.tmpl` | ✅ | |
| Docker/Makefile | go Dockerfile + make | ✅ | |

## Routes snapshot

| Area | Status |
|------|--------|
| Login/logout pages | ✅ |
| **POST /api/login** | ✅ (added) |
| Library/book/reader/tags/home | ✅ |
| OPDS | ✅ |
| Admin users/missing/scan/thumbs/upload | ✅ |
| Admin downloads/subscriptions pages | ✅ re-enabled |
| Plugin/queue admin API | ✅ re-enabled (handlers existed) |
| `/download/plugins` | ✅ re-enabled |

## Intentional differences

1. **Signature algorithm:** path+mtime+size FNV vs inode/CRC32; IDs stable via path-only DB fallback.
2. **Session:** token cookie `mango-token-{port}` vs Kemal session id (API returns token as `session_id`).
3. **Full Title#examine rescan tree:** ContentsSig stored; full examine state machine not fully ported (full scan still works).
4. **rename DSL:** library only; downloader still uses sanitizeFilename.

## Child tasks completed

| Child | Commit (approx) |
|-------|-----------------|
| go-upload-helpers | `75629ae` |
| go-signature | `9fe27a3` |
| go-rename-dsl | `777fa9b` |
| go-util-completion | `7f86a12` |
| go-routes-coverage | `037b554` |
| go-gap-regression | this report |

## Regression command

```bash
cd go && go build ./... && go vet ./... && go test ./...
# 192 tests, 2026-07-12
```

## Residual follow-ups (optional)

- Wire ContentsSignature into incremental examine (skip full rescan)
- Port full Crystal `sanitize_filename` Unicode rules
- Revisit mangadex path aliases if clients depend on old URLs
- End-to-end Docker smoke against production DB
