# Findings: UI 2 themes × 2 styles

**Task:** `07-11-ui-theme-style-review`  
**Date:** 2026-07-11  
**Method:** Code audit (views `comic-*` vs LESS / flat remap / `.uk-light` / `common.js`) — not browser screenshots  
**Axes:** comic/flat × light/dark (4 states)  
**Pass bar:** Consistency first (intentional 4-state look; no half-theme; comic ≠ flat)  
**Priority:** P0 unthemed/broken · P1 inconsistent · P2 polish/dead-code/docs

---

## Summary

| Priority | Count | Headline |
|----------|-------|----------|
| **P0** | 2 | flat+dark **login** background overridden; flat dark **body** background strategy hollow |
| **P1** | 13 | Login comic half-theme; form labels in comic dark; tags select2 flat-only; mobile chrome flat; FOUC; card progress / title select-bar half-theme |
| **P2** | 5 | Dead `mango-app-shell`; token duplication; dead comic selectors; FAB OK; no theming spec |

**Riskiest combination:** **flat+dark** (F01, F02).  
**Strongest areas:** FAB, desktop sidebar, admin card grid, most comic-btn / section headings (four intentional paths).

---

## Findings

| id | surface | states | priority | type | evidence | note |
|----|---------|--------|----------|------|----------|------|
| **F01** | Login | **flat+dark** | **P0** | bug | `comic-theme.less` ~1708–1710 `body:not(.comic-theme)… .comic-login-page { background:#E8DDD3 !important }` vs `mango.less` ~373–374 `.uk-light.login-page` | Flat remap **wins** over dark login → flat+dark login stays light beige (half-theme / wrong surface) |
| **F02** | Layout / global | **flat+dark** | **P0** | bug | `common.js` ~257–258 only sets `html` `rgb(20,20,20)`; `.uk-light` block does **not** set `body` bg; only dead `body.uk-light.mango-app-shell` has dark body | Flat dark relies on html inline; body may stay light → layered mismatch vs comic dark (`body.comic-theme-dark !important`) |
| **F03** | Login | comic+light, comic+dark | **P1** | inconsistency | `login.html.ecr` uses `login-card` not `comic-login-card*`; styles at `comic-theme.less` ~1263–1295 | Kirby login card CSS never mounted → comic login ≈ flat coral card |
| **F04** | Login | comic+dark | **P1** | inconsistency | `.comic-login-page` bright gradient `!important`, not under dark paper | Comic dark login stays loud light Kirby bg while cards may go dark → half set |
| **F05** | Login | all (mechanism) | **P1** | inconsistency | `login.html.ecr` only `setTheme()`; no style toggle UI | Style only from `localStorage` via `$(ready)`; no in-page comic/flat control |
| **F06** | Forms (user-edit etc.) | **comic+dark** | **P1** | bug | `.comic-form-label { color:@comic-dark-gray !important }`; dark block lacks label/input text override | Labels near-invisible on dark comic paper |
| **F07** | Tags (list pills) | comic vs flat | **P1** | inconsistency | `comic-tag-pill` comic-only; flat falls back to `.tag-pill` | Flat path OK; language split with title select2 (F08) |
| **F08** | Tags (title select2) | comic+* | **P1** | inconsistency | `tags.less` coral only; no comic scope | Tag editor always flat coral next to Kirby pills |
| **F09** | Chrome (mobile navbar) | comic+* | **P1** | inconsistency | `.mango-navbar` only in `mango.less`; no comic override | Mobile top bar always flat while desktop sidebar comic |
| **F10** | Chrome (mobile offcanvas) | comic+* | **P1** | inconsistency | `#mobile-nav` mango-only | Same as F09 for offcanvas |
| **F11** | Global FOUC | first paint | **P1** | bug | `common.js` bottom `setUIStyle()` may run before `body`; layout/login inline only `setTheme()` | First frame may miss `comic-theme` or only have `comic-theme-dark` → flash of flat language |
| **F12** | Admin UI style control | dark + style toggle | **P1** | inconsistency | `admin.js` `uiStyleChanged` → `setUIStyle` only | Usually OK (setUIStyle syncs dark classes) but not full theme recompute |
| **F13** | Reader chrome | comic vs flat | **P1** | inconsistency | `comic-reader-*` comic-scoped only; flat uses dark coral reader shell in mango.less | Flat intentional dark shell; no page-local theme/style FAB |
| **F14** | Home/library cards | comic+* | **P1** | inconsistency | Progress uses `card-progress-*`; `.comic-progress-*` unused | Kirby card chrome + coral progress on same card |
| **F15** | Title select-bar | comic+* | **P1** | inconsistency | `#select-bar` mango-only | Multi-select bar always warm flat orange on comic pages |
| **F16** | Dead CSS | all | **P2** | dead-code | `mango.less` `body.mango-app-shell` large block; **zero** view usage | Misleading dark-body comments; third unused visual system |
| **F17** | Tokens | all | **P2** | dead-code / docs | `@accent` in mango.less + tags.less; many hardcoded `#D96A4B` in flat remaps | Drift risk |
| **F18** | Dead comic selectors | comic | **P2** | dead-code | `.comic-login-card*`, `.comic-progress-*`, etc. unused or mismatched | Audit noise / false “already themed” |
| **F19** | FAB | all 4 | **P2** | taste / OK | mango.less + comic-theme.less full paths | Intentional four looks; related tasks own polish |
| **F20** | Docs | all | **P2** | docs | No `.trellis/spec` theming guide | Recurrence risk |

---

## Surfaces OK (code)

| Surface | Notes |
|---------|--------|
| Utility FAB | Flat + `.uk-light` + comic/comic-dark |
| Desktop sidebar | Flat `.app-sidebar` / comic `.comic-sidebar` |
| Admin card grid | comic / flat remap / flat-dark paths |
| Buttons & section headings | `comic-btn`, `comic-section-*`, `page-heading-comic` covered |
| Tables / empty / search (admin-ish) | Mostly remapped |
| Reader page canvas | Intentionally dark shell |
| Tags list (flat) | `.tag-pill` + `.uk-light` |
| Toggle mechanism | FAB + admin wired to classes |

---

## Four-state scorecard

| State | Overall | Main breaks |
|-------|---------|-------------|
| comic+light | Mostly intentional | Login card dead (F03); mobile chrome flat; select2/progress coral |
| comic+dark | Mostly intentional | Form labels (F06); login bright bg (F04); same half-themes |
| flat+light | Intentional coral | Distinguishable from comic |
| flat+dark | **Highest risk** | **F01 login**, **F02 body bg**; rest remaps better |

---

## Prioritized backlog (for later fix tasks)

1. **P0 F01** — Flat+dark login background specificity  
2. **P0 F02** — Flat dark body/html background policy (non-shell)  
3. **P1 F03+F04** — Login comic structure + dark comic bg alignment  
4. **P1 F06 + F08 + F14** — Comic-dark forms; title tags; card progress half-theme  
5. **P1 F11 + F09/F10** — Early theme classes (anti-FOUC); mobile chrome comic  

**Cleanup later:** F16 shell dead code, F17 tokens, F18 dead comic CSS, F20 theming spec.

---

## False positives (do not re-file)

- Flat missing `comic-sidebar` / `comic-reader-*` / `comic-tag-pill` remap when mango baseline exists  
- FAB as P0 (already intentional)  
- “comic-* in views undefined” — used classes exist in comic-theme.less  
- `setTheme` only adding `comic-theme-dark` without `comic-theme` — selectors use OR; FOUC is real issue (F11), not total loss of comic rules in dark

---

## Related tasks (do not duplicate scope)

- `07-11-float-utility-cluster` — FAB chrome actions  
- `07-11-github-float-theme-align` — GitHub float (legacy control)  
- `07-09-fix-reader-blank` — reader blank bug (separate)

---

## Out of scope for this task

Implementation of fixes. This document is the deliverable.
