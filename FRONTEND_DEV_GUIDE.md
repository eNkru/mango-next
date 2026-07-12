# Frontend Development Guide

Go embeds templates and static assets from `go/web/` (see `go/web/embed.go`). There is no separate Node/Gulp pipeline for the server binary.

## Key folders

| Folder | Purpose |
|---|---|
| **`go/web/views/`** | HTML templates (`html/template`, `*.tmpl`) |
| **`go/web/public/css/`** | Stylesheets (including LESS sources and compiled CSS) |
| **`go/web/public/js/`** | Client-side JavaScript |
| **`go/web/public/img/`**, `webfonts/` | Images and fonts |

## Stack

| Layer | Technology |
|---|---|
| **Templating** | Go `html/template` (`.tmpl`) |
| **CSS Framework** | UIkit 3.x |
| **Styles** | LESS sources may exist; commit compiled CSS used at runtime under `public/css/` |
| **JS** | jQuery, Alpine.js, Moment.js, Select2, etc. (local assets) |
| **Icons** | FontAwesome 5 |

## Quick start

1. Edit templates under `go/web/views/` (e.g. `home.tmpl`, `library.tmpl`, `reader.tmpl`).
2. Edit CSS/JS under `go/web/public/`.
3. Rebuild/run so embed picks up changes:

```bash
make run
# or
make build && ./mango
```

Shared chrome (navbar, theme) is split across partials such as `top.tmpl` / `bottom.tmpl` / `head.tmpl`.

## File map (views)

| File | Description |
|---|---|
| `home.tmpl` | Home |
| `library.tmpl` | Library browse |
| `title.tmpl` | Title detail |
| `reader.tmpl` | Reader |
| `login.tmpl` | Login |
| `admin.tmpl` | Admin |
| `tags.tmpl` / `tag.tmpl` | Tags |
| `user.tmpl` / `user-edit.tmpl` | Users |
| `download-manager.tmpl` | Downloads |
| `subscription-manager.tmpl` | Subscriptions |
| `missing-items.tmpl` | Missing titles/entries |
| `plugin-download.tmpl` | Plugin download UI |
| `opds/` | OPDS XML templates |

## Notes

- Do not reintroduce a root-level `public/` or Crystal ECR templates; the binary only embeds `go/web`.
- After UI changes, always rebuild — embed is compile-time.
