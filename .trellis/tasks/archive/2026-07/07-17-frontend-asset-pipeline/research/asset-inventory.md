# Research: Frontend asset inventory and provenance

- **Query**: 清点 `go/web/public` 下第三方 JS/CSS/font 与项目自有/派生文件，识别版本、许可证、上游包、模板消费者，并标记下载功能遗留代码。
- **Scope**: mixed
- **Date**: 2026-07-17

## Findings

### Runtime model

- `go/web/embed.go:9-13` embeds `views` and `public/*` at Go compile time. The embedded public filesystem is exposed through `fs.Sub` at `go/web/embed.go:19-24`.
- `go/internal/server/server.go:212-239` serves that embedded filesystem through `http.FileServer`; `go/internal/server/middleware.go:111-118` recognizes `/css`, `/js`, `/img`, `/webfonts`, favicon, robots, and manifest as static assets.
- Therefore every runtime asset listed below is a committed binary input, not a runtime-installed dependency. Rebuilding the Go binary is required after any asset change.

### Third-party JavaScript and CSS

| Committed file | Identified upstream/version | License evidence | Runtime consumer(s) | Classification / evidence |
|---|---|---|---|---|
| `go/web/public/js/jquery.min.js` | `jquery@3.2.1` | MIT; file header at line 1 and npm metadata | Global: `go/web/views/head.tmpl:41`; most project scripts assume `$` | Exact version is in the banner. npm package is deprecated but this task's PRD explicitly defers broad upgrades. |
| `go/web/public/js/jquery-ui.min.js` | `jquery-ui-dist@1.12.1` | MIT; banner lines 1-4 | `plugin-download.tmpl:224` via `jquery-ui.tmpl`; sorting/selecting plugin UI | Exact version and included modules are in the banner. |
| `go/web/public/js/alpine.min.js` | `alpinejs@2.8.0`, regular build | MIT; jsDelivr header lines 1-5 and npm metadata | Global: `head.tmpl:42`; templates use `x-data`, `x-show`, `x-text`, `@click` | Header identifies `/gh/alpinejs/alpine@2.8.0/dist/alpine.js`. |
| `go/web/public/js/alpine-ie11.min.js` | `alpinejs@2.8.0`, IE11 build plus polyfills | Alpine MIT plus embedded third-party polyfill notices; Polymer BSD-style notice visible at lines 10-17 | `head.tmpl:43` as `nomodule` | Obsolete compatibility path for a browser explicitly unsupported by `.trellis/tasks/archive/2026-07/07-17-frontend-browser-smoke/prd.md:25-26,66`. Removal affects this file and the `nomodule` template tag; it also removes bundled polyfill license obligations associated only with this artifact. |
| `go/web/public/js/uikit.min.js` | `uikit@3.5.9` | MIT; banner line 1 | Almost all layout pages via `bottom.tmpl:68` → `uikit.tmpl:2`; login and reader also include it | Exact version in banner and npm metadata. |
| `go/web/public/js/moment.min.js` | Moment.js `2.24.0` | MIT (upstream/npm) | `download-manager.tmpl:64`, `plugin-download.tmpl:225`, `subscription-manager.tmpl:110` through `moment.tmpl` | The minified file lacks a banner, but `strings` finds `2.24.0`; SHA-256 is `e22419e8…`. |
| `go/web/public/js/select2.min.js` | `select2@4.1.0-beta.1` | MIT; banner line 1 and npm metadata | `title.tmpl:195`; tag editor | Exact beta version in banner. |
| `go/web/public/css/select2.min.css` | Select2 4.1.0-beta.1 distribution CSS (strongly implied by paired JS and matching selectors) | MIT (upstream package) | `title.tmpl:193` | CSS has no banner; repository alone does not prove byte-for-byte package provenance. Pairing with the exact JS is the available evidence. |
| `go/web/public/js/jquery.inview.min.js` | Likely `jquery-inview` 1.1.x; exact version not encoded | ISC in npm metadata | `dots.tmpl:3` and separately `reader.tmpl:169`; used for visibility/lazy behavior | File has no banner/version. npm latest is 1.1.2 and requires jQuery 1.8+, but exact committed version remains unproven. |
| `go/web/public/js/dotdotdot.js` | `dotdotdot-js@4.0.11` | **CC-BY-NC-4.0**; file lines 1-10 and npm 4.0.11 metadata | `dots.tmpl:2`, used by `home.tmpl:216`, `library.tmpl:82`, `tag.tmpl:72`, `title.tmpl:192` | Exact version/license are explicit. This is the highest-impact license caveat: “NC” is not equivalent to the repository MIT license and needs distribution/use review. Upstream 4.2.0 changed to CC-BY-4.0, but an upgrade is outside the pipeline-first scope unless separately approved/tested. |

### Font Awesome and framework-derived assets

| Files | Upstream/version | License | Consumer / generation evidence |
|---|---|---|---|
| `go/web/public/webfonts/fa-solid-900.{ttf,woff,woff2}`; `fa-brands-400.{eot,svg,ttf,woff,woff2}` | `@fortawesome/fontawesome-free@5.15.4` | Package composite license: CC-BY-4.0 AND OFL-1.1 AND MIT | `mango.less:5-7` imports FA LESS; `mango.css:12566,17136-17170` contains 5.15.4 banners and `@font-face`; templates use `fas`/`fab` classes. |
| `go/web/public/img/divider-icon.svg`, `form-*.svg`, `nav-parent-*.svg`, `list-bullet.svg`, `accordion-*.svg` | UIkit 3.5.9 theme internal images | MIT | Paths are declared in `go/web/public/css/uikit.less:35-45`; `mango.less:2` imports `uikit.less`, which imports `node_modules/uikit/src/less/uikit.theme.less`. These are framework-derived/copied runtime inputs. |
| `go/web/public/css/mango.css` | Derived from UIkit 3.5.9 + Font Awesome 5.15.4 + project `mango.less` and `_variables.less` | Mixed: project MIT + UIkit MIT + Font Awesome composite | Global stylesheet in `head.tmpl:35`. It is a generated/derived artifact, not the source of truth. |

License/provenance gaps visible in the current Font Awesome set:

- `mango.css:17144-17145` references `fa-solid-900.eot` and `fa-solid-900.svg`, but those files are absent. A later project-added `@font-face` at `mango.css:17168-17170` references existing WOFF/WOFF2 files, so modern browsers can still load solid icons.
- No third-party NOTICE/license bundle is present under `go/web/public`, and the root `LICENSE` only records Mango's MIT license. The package/license inventory must therefore carry notices for redistributed npm/font files.

### Project-owned JavaScript

These files have no third-party banner and implement Mango behavior. They should remain project source rather than be replaced by package copies.

| File | Template consumer(s) / role |
|---|---|
| `js/common.js` | Global through `head.tmpl:45`; theme/UI-style handling, helpers. It also injects Google Fonts at `common.js:189`. |
| `js/i18n.js` | Global through `head.tmpl:44`; local translations. |
| `js/alert.js` | Included by admin, user, reader, title/library, missing-items, subscription, plugin/download templates. |
| `js/admin.js` | `admin.tmpl:80`. |
| `js/dots.js` | `dots.tmpl:4`; adapter around dotdotdot/inview. |
| `js/search.js` | `library.tmpl:85`, `tag.tmpl:75`, `title.tmpl:198`. |
| `js/sort-items.js` | `library.tmpl:86`, `tag.tmpl:76`, `title.tmpl:199`. |
| `js/title.js` | `home.tmpl:218`, `library.tmpl:84`, `tag.tmpl:74`, `title.tmpl:197`. |
| `js/reader.js` | `reader.tmpl:172`. |
| `js/user.js` | `user.tmpl:38`. |
| `js/user-edit.js` | `user-edit.tmpl:74`. |
| `js/missing-items.js` | `missing-items.tmpl:46`. |
| `js/subscription-manager.js` | `subscription-manager.tmpl:112`. |
| `js/plugin-download.js` | `plugin-download.tmpl:227`; retained product/plugin flow. |
| `js/download-manager.js` | `download-manager.tmpl:66`; **unsupported retained code** (details below). |
| `js/subscription.js` | No template reference found; contains stale `/api/admin/mangadex/...` calls at lines 7, 23, 40, 58. It is presently unreachable from templates and should be represented as an unconsumed/stale asset in design evidence. |

### Unsupported retained download code

Per task scope, do not remove these in this task:

| Retained item | Evidence / status |
|---|---|
| Route | `go/internal/server/server.go:146` registers `/downloads`; `server.go:136` registers `/download/plugins`. |
| Page handler | `go/internal/server/handlers_pages.go:658-666` renders `download-manager`. |
| Template | `go/web/views/download-manager.tmpl`; loads Moment, alert, and `download-manager.js` at lines 64-66. |
| Script | `go/web/public/js/download-manager.js`; calls nonexistent MangaDex queue HTTP/WebSocket endpoints (`:10-12`, `:37`, `:56`, `:83`). |
| Contract evidence | `.trellis/tasks/archive/2026-07/07-17-frontend-queue-contract/prd.md:3-7,14-24` records that the feature is disabled and the UI/API contract is unrepaired. |
| Translation strings | Reside in project `i18n.js`; retained with the unsupported page. |

The ordinary entry download endpoint (`go/internal/server/server.go:162`, OPDS link in `views/opds/title.tmpl:31`) is distinct from the unsupported download-manager UI and must not be conflated with it.

### Project-owned and application-derived CSS

| Source / output | Role |
|---|---|
| `_variables.less` | Shared Mango theme tokens and font stacks. |
| `mango.less` → `mango.css` | Main entry; imports UIkit, Font Awesome, then project styles. Runtime consumer: global `head.tmpl:35`. |
| `comic-theme.less` → `comic-theme.css` | Comic skin entry. Runtime consumer: global `head.tmpl:36`. |
| `flat-theme.less` → `flat-theme.css` | Flat skin entry. Runtime consumer: global `head.tmpl:37`. |
| `tags.less` → `tags.css` | Select2/tag overrides. Runtime consumer: `title.tmpl:194`. |
| `uikit.less` | Project customization/import wrapper for npm UIkit source; it is not directly served by a template. |

### Other public assets

- App-owned: `favicon.ico`, `manifest.json`, `robots.txt`, `img/icons/*`, including `mango-mark.svg` used by `top.tmpl:14,60`; PWA PNG icons used by `manifest.json:6-18`.
- `banner.png`, `banner-paddings.png`, and `loading.gif` have no current template/CSS/JS consumer found. Repository history shows they predate the Go migration, but exact authorship/license cannot be established from current files.
- `manifest.json:6,11,16,22` uses root-absolute icon/start URLs rather than `.BaseURL`; this is an existing asset-consumer fact relevant to BaseURL validation, not a dependency-pipeline behavior.

## External References

- [jquery 3.2.1 npm metadata](https://registry.npmjs.org/jquery/3.2.1) — version, MIT license, deprecation status.
- [jquery-ui-dist 1.12.1](https://registry.npmjs.org/jquery-ui-dist/1.12.1) — MIT license and distribution package.
- [UIkit 3.5.9](https://registry.npmjs.org/uikit/3.5.9) — MIT license and package paths.
- [Alpine.js 2.8.0](https://registry.npmjs.org/alpinejs/2.8.0) — MIT license and regular/IE11 build provenance.
- [Select2 4.1.0-beta.1](https://registry.npmjs.org/select2/4.1.0-beta.1) — MIT license and dist paths.
- [Font Awesome Free 5.15.4](https://registry.npmjs.org/@fortawesome/fontawesome-free/5.15.4) — composite license and package version.
- [jquery-inview npm metadata](https://registry.npmjs.org/jquery-inview) — ISC license; exact local version unresolved.
- [dotdotdot-js npm metadata](https://registry.npmjs.org/dotdotdot-js) — confirms 4.0.11 CC-BY-NC-4.0 and later 4.2.0 license change.

## Related Specs

- `.trellis/spec/frontend/ui-theme-layout.md:39-50` — names theme LESS/CSS pairs and explicitly permits hand synchronization today.
- `.trellis/tasks/07-17-frontend-asset-pipeline/prd.md` — pipeline and scope contract.
- `.trellis/tasks/archive/2026-07/07-17-frontend-queue-contract/prd.md` — unsupported retained download-manager contract.

## Caveats / Not Found

- Exact provenance/version of `jquery.inview.min.js` is not encoded and was not proven byte-for-byte.
- `select2.min.css` is strongly associated with the paired 4.1.0-beta.1 JS, but lacks its own version banner.
- The exact historical source/license of unreferenced PNG/GIF assets is not recorded in the current repository.
- Current committed files do not contain local Bangers/Fredoka font binaries or their license texts.
