# Implement: go-rename-dsl

1. [ ] `go/internal/rename/rename.go` — Parse + Render + postProcess
2. [ ] `rename_test.go` from `spec/rename_spec.cr`
3. [ ] `go test ./internal/rename/... && go test ./...`

## Validation

```bash
cd go && go test ./internal/rename/... && go build ./... && go vet ./... && go test ./...
```
