# Implement: go-signature

## Ordered Checklist

### 1. ContentsSignature + tests

- [x] `signature.go` + ContentsSignature SHA1
- [x] `signature_test.go`
- [x] library tests pass

### 2. Unified API surface

- [x] FileSignature / DirSignature / DirectoryEntrySignature exported
- [x] FNV decision documented in signature.go

### 3. Optional NewTitle wire

- [x] `Title.ContentsSig` set in NewTitle
- [x] Skip full examine state machine

### 4. Quality gate

- [x] `go build/vet/test ./...` — 178+ tests

## Validation Commands

```bash
cd go
go test ./internal/library/ -count=1
go build ./... && go vet ./... && go test ./...
```

## Do Not Start Until

- [x] Algorithm decision locked (FNV + ContentsSignature SHA1)
- [x] design.md + implement.md
- [ ] user start **or** continue-through-implement (this session: implement after artifacts)
