# Design: go-rename-dsl

## Overview

Mechanical port of Crystal AST:

```
Rule = (Group | String | Pattern)*
Group = '[' (Pattern | String)* ']'   // omit whole group if any Pattern empty
Pattern = '{' Variable ('|' Variable)* '}'  // first key present in hash
Variable = id
```

Render post_process (Crystal):

- if result `== ".."` → `"_"`
- rstrip spaces and `.`
- replace `/ ? < > \ : * | " ^` with `_`

## Package

`go/internal/rename`

```go
type Rule struct { /* nodes */ }
func Parse(s string) (*Rule, error)
func (r *Rule) Render(vars map[string]string) string
```

## Compatibility

No DB, no routes. Pure CPU.

## Tests

Mirror `spec/rename_spec.cr` cases one-for-one.
