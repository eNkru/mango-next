# Fix P0 flat+dark theme surfaces

## Goal

Fix the two **P0** findings from `07-11-ui-theme-style-review`: flat+dark login page stays light beige, and flat dark lacks a real body background strategy.

## User value

Users on flat + dark should not see a light login page or mixed light body / dark chrome. Dark flat must look intentional.

## Source findings

| id | issue | from |
|----|--------|------|
| **F01** | Flat remap `.comic-login-page { background:#E8DDD3 !important }` beats `.uk-light.login-page` | findings.md |
| **F02** | Flat dark only sets `html` inline `rgb(20,20,20)`; body dark bg only on dead `mango-app-shell` | findings.md |

## Confirmed facts

1. Login body: `class="login-page comic-login-page"` (`login.html.ecr`).
2. Flat light remap: `body:not(.comic-theme):not(.comic-theme-dark) .comic-login-page` → `#E8DDD3 !important` (`comic-theme.less` ~1708).
3. Dark login intent: `.uk-light.login-page { background: @dark-bg-base }` (`mango.less` ~373–374, `@dark-bg-base: #121212`).
4. Comic dark body: `body.comic-theme-dark { background-color: … !important }` — works.
5. `setTheme` / `setUIStyle` set `html` inline bg only for flat dark (`common.js`).

## Decisions

| Decision | Choice | Notes |
|----------|--------|--------|
| F01 approach | **Add flat-dark override** for `.comic-login-page` under existing `body.uk-light:not(.comic-theme):not(.comic-theme-dark)` | Keep light flat remap; dark wins with equal/higher specificity + `!important` |
| F01 color | **`#121212`** (match `@dark-bg-base`) | comic-theme.less compiles separately; no mango vars |
| F02 approach | **CSS `body` dark bg for flat** under same flat-dark selector; keep JS `html` bg for flat dark | Minimal; no new theme class system |
| F02 color align | **Use `@dark-bg-base` / `#121212`** for body; align JS html to `#121212` | Was `rgb(20,20,20)` ≈ `#141414` |
| Scope | **F01 + F02 only** | No login comic structure (F03/F04), no shell purge (F16) |

## Requirements

- R1: flat+dark login page background is dark (`#121212` / `@dark-bg-base`), not `#E8DDD3`.
- R2: flat+light login still uses light beige (`#E8DDD3` / `@login-bg`).
- R3: comic login background behavior unchanged by this task (Kirby gradient still P1/F04).
- R4: flat+dark non-login pages: `body` has intentional dark background without `mango-app-shell`.
- R5: comic+dark body still uses `comic-theme-dark` path (not broken by flat body rule).
- R6: Prefer CSS class scopes already in use; no new theme axes.
- R7: Recompile CSS via project gulp if needed so runtime CSS matches LESS.

## Acceptance criteria

1. With `ui-style=flat` and theme dark: login body background is dark base, not light beige.
2. With `ui-style=flat` and theme light: login body background remains light beige.
3. With comic style: no regression to login Kirby bg / comic dark paper (relative to pre-change).
4. flat+dark app shell pages: body background dark without relying on `mango-app-shell`.
5. comic+dark: body still dark paper via `comic-theme-dark`.
6. Toggling theme/style updates backgrounds without full page redesign.

## Out of scope

- P1 login comic card structure (F03/F04)
- FOUC early class (F11)
- Dead `mango-app-shell` purge (F16) beyond not depending on it
- Token unification (F17)
- Reader / tags / mobile chrome

## Complexity

Lightweight: small LESS + optional one-line JS color align; PRD-only sufficient.

## Related

- Parent audit: `07-11-ui-theme-style-review` / `findings.md` F01, F02
