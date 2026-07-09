# Research: Phase 4 - HTTP Routes and Templates for Go Migration

- **Query**: Discover all Crystal routes, templates, public assets, middleware, and Go existing code for Phase 4 migration
- **Scope**: mixed (internal + external)
- **Date**: 2026-07-09

## Table of Contents

1. [Route Discovery (All 67 Routes)](#1-route-discovery)
2. [Template Discovery](#2-template-discovery)
3. [Public Assets](#3-public-assets)
4. [OPDS Route Detail](#4-opds-route-detail)
5. [Middleware](#5-middleware)
6. [Embed Strategy (baked_file_system)](#6-embed-strategy)
7. [Go Existing Auth Code](#7-go-existing-auth-code)
8. [Go Module Dependencies](#8-go-module-dependencies)
9. [Key Implementation Notes for Go Port](#9-key-implementation-notes)

---

## 1. Route Discovery

### 1.1 Route File: `src/routes/api.cr` — APIRouter (30 routes)

All routes are registered inside `struct APIRouter#initialize`. Uses Koa macros for OpenAPI docs.

| # | Method | Path | Handler Name | Description | Tags |
|---|--------|------|-------------|-------------|------|
| 1 | POST | `/api/login` | `post "/api/login"` | Authenticate user, return session cookie | users |
| 2 | GET | `/api/page/:tid/:eid/:page` | `get "/api/page/:tid/:eid/:page"` | Return a single page image for a manga entry | reader |
| 3 | GET | `/api/cover/:tid/:eid` | `get "/api/cover/:tid/:eid"` | Return cover image of an entry | library |
| 4 | GET | `/api/book/:tid` | `get "/api/book/:tid"` | Return a title/book with its entries and sub-titles | library |
| 5 | GET | `/api/sort_opt` | `get "/api/sort_opt"` | Get sort option for a title or library | library |
| 6 | PUT | `/api/sort_opt` | `put "/api/sort_opt"` | Update sort option for a title or library | library |
| 7 | GET | `/api/library` | `get "/api/library"` | Return the entire library with all titles/entries | library |
| 8 | GET | `/api/library/continue_reading` | `get "/api/library/continue_reading"` | Return continue reading entries | library |
| 9 | GET | `/api/library/start_reading` | `get "/api/library/start_reading"` | Return start reading titles | library |
| 10 | GET | `/api/library/recently_added` | `get "/api/library/recently_added"` | Return recently added items | library |
| 11 | POST | `/api/admin/scan` | `post "/api/admin/scan"` | Trigger a library scan | admin, library |
| 12 | GET | `/api/admin/thumbnail_progress` | `get "/api/admin/thumbnail_progress"` | Return thumbnail generation progress (0-1) | admin, library |
| 13 | POST | `/api/admin/generate_thumbnails` | `post "/api/admin/generate_thumbnails"` | Trigger thumbnail generation (spawn) | admin, library |
| 14 | DELETE | `/api/admin/user/delete/:username` | `delete "/api/admin/user/delete/:username"` | Delete a user by username | admin, users |
| 15 | PUT | `/api/progress/:tid/:page` | `put "/api/progress/:tid/:page"` | Update reading progress (entry via ?eid= or whole title) | progress |
| 16 | PUT | `/api/bulk_progress/:action/:tid` | `put "/api/bulk_progress/:action/:tid"` | Bulk mark read/unread for entries in a title | progress |
| 17 | PUT | `/api/admin/display_name/:tid/:name` | `put "/api/admin/display_name/:tid/:name"` | Set display name of a title or entry | admin, library |
| 18 | PUT | `/api/admin/sort_title/:tid` | `put "/api/admin/sort_title/:tid"` | Set sort title of a title or entry | admin, library |
| 19 | WS | `/api/admin/mangadex/queue` | `ws "/api/admin/mangadex/queue"` | WebSocket for live download queue updates | admin |
| 20 | GET | `/api/admin/mangadex/queue` | `get "/api/admin/mangadex/queue"` | Return current download queue | admin, downloader |
| 21 | POST | `/api/admin/mangadex/queue/:action` | `post "/api/admin/mangadex/queue/:action"` | Perform action on queue (delete/retry/pause/resume) | admin, downloader |
| 22 | POST | `/api/admin/upload/:target` | `post "/api/admin/upload/:target"` | Upload a file (currently only "cover") | admin |
| 23 | GET | `/api/admin/plugin` | `get "/api/admin/plugin"` | List available plugins | admin, downloader |
| 24 | GET | `/api/admin/plugin/info` | `get "/api/admin/plugin/info"` | Get plugin metadata/info | admin, downloader |
| 25 | GET | `/api/admin/plugin/search` | `get "/api/admin/plugin/search"` | Search manga from a plugin | admin, downloader |
| 26 | POST | `/api/admin/plugin/subscriptions` | `post "/api/admin/plugin/subscriptions"` | Create a new subscription | admin, downloader, subscription |
| 27 | GET | `/api/admin/plugin/subscriptions` | `get "/api/admin/plugin/subscriptions"` | List subscriptions for a plugin | admin, downloader, subscription |
| 28 | DELETE | `/api/admin/plugin/subscriptions` | `delete "/api/admin/plugin/subscriptions"` | Delete a subscription | admin, downloader, subscription |
| 29 | POST | `/api/admin/plugin/subscriptions/update` | `post "/api/admin/plugin/subscriptions/update"` | Check for updates on a subscription | admin, downloader, subscription |
| 30 | GET | `/api/admin/plugin/list` | `get "/api/admin/plugin/list"` | List chapters in a title from a plugin | admin, downloader |
| 31 | POST | `/api/admin/plugin/download` | `post "/api/admin/plugin/download"` | Add chapters to download queue | admin, downloader |
| 32 | GET | `/api/dimensions/:tid/:eid` | `get "/api/dimensions/:tid/:eid"` | Return image dimensions for all pages | reader |
| 33 | GET | `/api/download/:tid/:eid` | `get "/api/download/:tid/:eid"` | Download an entry as file attachment | library, reader |
| 34 | GET | `/api/tags/:tid` | `get "/api/tags/:tid"` | Get tags of a specific title | library, tags |
| 35 | GET | `/api/tags` | `get "/api/tags"` | Get all tags | library, tags |
| 36 | PUT | `/api/admin/tags/:tid/:tag` | `put "/api/admin/tags/:tid/:tag"` | Add a tag to a title | admin, library, tags |
| 37 | DELETE | `/api/admin/tags/:tid/:tag` | `delete "/api/admin/tags/:tid/:tag"` | Delete a tag from a title | admin, library, tags |
| 38 | GET | `/api/admin/titles/missing` | `get "/api/admin/titles/missing"` | List all missing titles | admin, library |
| 39 | GET | `/api/admin/entries/missing` | `get "/api/admin/entries/missing"` | List all missing entries | admin, library |
| 40 | DELETE | `/api/admin/titles/missing` | `delete "/api/admin/titles/missing"` | Delete all missing titles | admin, library |
| 41 | DELETE | `/api/admin/entries/missing` | `delete "/api/admin/entries/missing"` | Delete all missing entries | admin, library |
| 42 | DELETE | `/api/admin/titles/missing/:tid` | `delete "/api/admin/titles/missing/:tid"` | Delete a specific missing title | admin, library |
| 43 | DELETE | `/api/admin/entries/missing/:eid` | `delete "/api/admin/entries/missing/:eid"` | Delete a specific missing entry | admin, library |
| 44 | PUT | `/api/admin/hidden/:tid/:value` | `put "/api/admin/hidden/:tid/:value"` | Set hidden status of a title (0=visible, 1=hidden) | admin, library |
| 45 | GET | `/api/admin/hidden_titles` | `get "/api/admin/hidden_titles"` | Return list of hidden title IDs | admin, library |
| 46 | GET | `/openapi.json` | `get "/openapi.json"` | Return generated OpenAPI spec (from Koa) | none |

**Total in api.cr: 46 routes** (45 explicit + 1 websocket)

#### Key Implementation Notes for api.cr

- **Authentication**: All routes except `/api/login` require auth. Admin routes (`/api/admin/*`) require admin.
- **Response helpers**: `send_json(env, json)`, `send_text(env, text)`, `send_img(env, img)`, `send_attachment(env, path)` are macros in `src/util/web.cr`
- **CORS**: `cors` macro is called inside `send_json`, `send_img`, `send_text`, `send_attachment`
- **Error handling**: Every route wraps in `begin/rescue` — catches exceptions, logs them, returns `{"success": false, "error": msg}` for JSON routes, or sets `env.response.status_code = 500`
- **ETag support**: Routes `/api/page/:tid/:eid/:page`, `/api/cover/:tid/:eid`, `/api/dimensions/:tid/:eid` support `If-None-Match` headers and return 304
- **Cache control**: `"public, max-age=86400"` or `"no-cache, max-age=86400"` for DirEntry
- **Reader progress query**: `/api/progress/:tid/:page` accepts optional `?eid=` for per-entry progress
- **WebSocket queue**: WS at `/api/admin/mangadex/queue` sends JSON every N seconds (default 5)

### 1.2 Route File: `src/routes/main.cr` — MainRouter (10 routes)

| # | Method | Path | Handler Name | Description |
|---|--------|------|-------------|-------------|
| 47 | GET | `/login` | `get "/login"` | Render login page (src/views/login.html.ecr) |
| 48 | GET | `/logout` | `get "/logout"` | Clear session and redirect to /login |
| 49 | POST | `/login` | `post "/login"` | Process login form, redirect to / or callback |
| 50 | GET | `/library` | `get "/library"` | Render library page with sorted titles |
| 51 | GET | `/book/:title` | `get "/book/:title"` | Render title page with entries/sub-titles |
| 52 | GET | `/download/plugins` | `get "/download/plugins"` | Render plugin download page |
| 53 | GET | `/` | `get "/"` | Render home page with continue reading / recently added / start reading |
| 54 | GET | `/tags/:tag` | `get "/tags/:tag"` | Render a single tag page with matching titles |
| 55 | GET | `/tags` | `get "/tags"` | Render all tags page |
| 56 | GET | `/api` | `get "/api"` | Render API docs page (api.html.ecr with Redoc) |

#### Key Implementation Notes for main.cr

- Uses `layout "name"` macro which renders `src/views/name.html.ecr` inside `src/views/layout.html.ecr`
- Variables available in layout: `base_url`, `is_admin`, `page` (the page name variable)
- `/login` and `/logout` skip auth; all others require auth
- `/book/:title` has `title_id` as URL path parameter (not display name)
- Success/failure of login redirects, no JSON responses

### 1.3 Route File: `src/routes/reader.cr` — ReaderRouter (2 routes)

| # | Method | Path | Handler Name | Description |
|---|--------|------|-------------|-------------|
| 57 | GET | `/reader/:title/:entry` | `get "/reader/:title/:entry"` | Redirect to first page of entry (or reader-error) |
| 58 | GET | `/reader/:title/:entry/:page` | `get "/reader/:title/:entry/:page"` | Render reader page at specific page number |

#### Key Implementation Notes for reader.cr

- Route 57 (no page): loads progress, if finished starts from page 1, then redirects to `/reader/:title/:entry/:page_idx`
- Route 58 (with page): renders `reader.html.ecr` (standalone — NOT using layout macro, uses `render` directly)
- Reader page has its own `<html>` structure (not wrapped in layout)
- Data passed to template: `base_url`, `title`, `entry`, `page_idx`, `entries` (all sorted entries), `exit_url`, `next_entry_url`, `previous_entry_url`

### 1.4 Route File: `src/routes/opds.cr` — OPDSRouter (2 routes)

| # | Method | Path | Handler Name | Description |
|---|--------|------|-------------|-------------|
| 59 | GET | `/opds` | `get "/opds"` | Render OPDS index (all titles) |
| 60 | GET | `/opds/book/:title_id` | `get "/opds/book/:title_id"` | Render OPDS title page (entries + sub-titles) |

#### Key Implementation Notes for opds.cr

- Uses `render_xml "src/views/opds/index.xml.ecr"` macro
- Content type: `application/xml`
- Auth: Basic auth (handled by auth handler — OPDS routes get `require_basic_auth` when not authenticated)
- Template data: `base_url`, `titles` (all titles for index), `title` (single title for book)

### 1.5 Route File: `src/routes/admin.cr` — AdminRouter (7 routes)

| # | Method | Path | Handler Name | Description |
|---|--------|------|-------------|-------------|
| 61 | GET | `/admin` | `get "/admin"` | Render admin dashboard |
| 62 | GET | `/admin/user` | `get "/admin/user"` | Render user management page |
| 63 | GET | `/admin/user/edit` | `get "/admin/user/edit"` | Render user edit/create form |
| 64 | POST | `/admin/user/edit` | `post "/admin/user/edit"` | Create new user |
| 65 | POST | `/admin/user/edit/:original_username` | `post "/admin/user/edit/:original_username"` | Update existing user |
| 66 | GET | `/admin/downloads` | `get "/admin/downloads"` | Render download manager page |
| 67 | GET | `/admin/subscriptions` | `get "/admin/subscriptions"` | Render subscription manager page |
| 68 | GET | `/admin/missing` | `get "/admin/missing"` | Render missing items page |

**Total routes: 68** (but WS routes are listed separately, so approximately 67 HTTP routes as mentioned in PRD)

#### Key Implementation Notes for admin.cr

- All routes require admin access (enforced by AuthHandler)
- User CRUD: create via POST `/admin/user/edit`, update via POST `/admin/user/edit/:original_username`
- Uses sanitize library for user input in query params
- Forms redirect on success/failure rather than returning JSON

### 1.6 Server-level Routes (from `src/server.cr`)

| # | Method | Path | Description |
|---|--------|------|-------------|
| — | ANY | `/*` (404) | Custom 404 page using `layout "message"` |
| — | ANY | `/*` (500) | Custom 500 page in release mode |
| — | OPTIONS | `/api/*`, `/uploads/*`, `/img/*` | CORS preflight — returns empty 200 with CORS headers |

---

## 2. Template Discovery

### 2.1 Main Layout Templates

#### `src/views/layout.html.ecr` — Main Layout Wrapper
- **Used by**: `layout` macro (library, home, title, tag, tags, admin, user, user-edit, download-manager, subscription-manager, missing-items, message)
- **Data needed**: `base_url` (from Config), `is_admin` (bool), `page` (string — page name)
- **Renders**: Full HTML5 document with sidebar navigation, mobile top bar, footer controls (language toggle, theme toggle, UI style toggle, logout)
- **Components used**: `head` (in head section), `uikit` (before close body), `yield_content "script"` (for page-specific scripts)
- **Key JS**: Uses Alpine.js (`x-data`), jQuery, UIkit, FontAwesome. Sidebar collapse state stored in `localStorage`.

#### `src/views/message.html.ecr` — Error/Message Page
- **Used by**: 404/500 error pages, error fallback in `layout` macro
- **Data needed**: `message` (string — error message)
- **Renders**: Single centered paragraph with error message

### 2.2 Page Templates (10 pages)

#### `src/views/login.html.ecr` — Login Page
- **Data needed**: `base_url` (from Config)
- **Renders**: Standalone HTML (NOT wrapped in layout). Login form with username/password, password visibility toggle.
- **Components**: `head`, `uikit`
- **JS**: Inline script for toggle password, form submit loading state, input focus effects

#### `src/views/home.html.ecr` — Home Dashboard
- **Layout**: `layout "home"`
- **Data needed**:
  - `continue_reading` — Array of `{entry: Entry, percentage: Float64}`
  - `recently_added` — Array of `{entry: Entry, percentage: Float64, grouped_count: Int32}`
  - `start_reading` — Array of Title
  - `new_user` — Bool
  - `empty_library` — Bool
  - `titles` — Array of Title
  - `base_url`, `is_admin`
- **Components**: `card`, `entry-modal`, `dots`
- **States**: Empty (new user + empty library), new user + non-empty, returning user with content
- **JS**: `alert.js`, `title.js`, inline carousel scroll/drag

#### `src/views/library.html.ecr` — Library Browser
- **Layout**: `layout "library"`
- **Data needed**:
  - `titles` — Array of Title (sorted)
  - `percentage` — Array of Float64 (one per title)
  - `show_hidden` — Bool
  - `base_url`, `is_admin`
  - Sort options via `get_sort_opt` / `get_and_save_sort_opt`
- **Components**: `sort-form`, `card`, `dots`
- **JS**: `alert.js`, `title.js`, `search.js`, `sort-items.js`

#### `src/views/title.html.ecr` — Single Title/Book Page
- **Layout**: `layout "title"`
- **Data needed**:
  - `title` — Title object (with `id`, `display_name`, `title`, `sort_title_db`, `cover_url`, `parents`, `hidden?`, etc.)
  - `sorted_titles` — Array of Title (sub-titles)
  - `entries` — Array of Entry (sorted)
  - `percentage` — Array of Float64 (one per entry)
  - `title_percentage` — Array of Float64 (one per sub-title)
  - `title_percentage_map` — Hash of String => Float64 (id -> percentage)
  - `is_hidden` — Bool
  - `base_url`, `is_admin`
  - Sort options
- **Components**: `sort-form`, `card`, `entry-modal`, `dots`
- **JS**: `select2.min.js`, `tags.css`, `alert.js`, `title.js`, `search.js`, `sort-items.js`
- **Features**: Multi-select with bulk mark read/unread, tag management, edit modal (display name, sort title, cover upload, progress controls), hidden toggle

#### `src/views/reader.html.ecr` — Manga Reader
- **Standalone page** (NOT using layout — has its own `<html>` structure)
- **Data needed**:
  - `base_url`, `title` (Title object), `entry` (Entry object), `page_idx` (Int32)
  - `entries` (Array of Entry, sorted)
  - `exit_url`, `next_entry_url`, `previous_entry_url` (String or nil)
  - `MANGO_VERSION` (for template)
- **Components**: `head`, `uikit`
- **JS**: `jquery.inview.min.js`, `alert.js`, `reader.js`
- **Features**: Continuous/paged mode, fit-to-height/width/original, right-to-left, flip animation, page jump, preload, margin control, entry navigation
- **Inline JS**: Sets `base_url`, `page`, `tid`, `eid` as globals

#### `src/views/admin.html.ecr` — Admin Dashboard
- **Layout**: `layout "admin"`
- **Data needed**: `missing_count` (Int32), `base_url`, `is_admin`, `MANGO_VERSION`
- **Components**: None specific (inline Alpine.js `x-data="component()"`)
- **JS**: `alert.js`, `admin.js`
- **Features**: User management link, missing items link, scan button, thumbnail generation, theme/UI style selectors, version display

#### `src/views/user.html.ecr` — User Management List
- **Layout**: `layout "user"`
- **Data needed**: `users` (Array of [username, is_admin]), `username` (current logged-in user), `base_url`
- **JS**: `alert.js`, `user.js`

#### `src/views/user-edit.html.ecr` — User Create/Edit Form
- **Layout**: `layout "user-edit"`
- **Data needed**: `username` (String or nil — existing user), `admin` (Bool or nil), `error` (String or nil), `new_user` (Bool), `base_url`
- **JS**: `alert.js`, `user-edit.js`

#### `src/views/tags.html.ecr` — All Tags Page
- **Layout**: `layout "tags"`
- **Data needed**: `tags` (Array of {tag: String, encoded_tag: String, count: Int32}), `base_url`, `is_admin`
- **JS**: None (no custom JS block)

#### `src/views/tag.html.ecr` — Single Tag Page
- **Layout**: `layout "tag"`
- **Data needed**: `tag` (String), `titles` (Array of Title), `percentage` (Array of Float64), `show_hidden` (Bool), `base_url`, `is_admin`
- **Components**: `sort-form`, `card`, `dots`
- **JS**: `alert.js`, `title.js`, `search.js`, `sort-items.js`

### 2.3 Admin Sub-page Templates

#### `src/views/download-manager.html.ecr`
- **Layout**: `layout "download-manager"`
- **Data needed**: `base_url`, `is_admin`
- **JS/Data**: Fully client-side (Alpine.js `x-data="component()"`), fetches via API
- **Components**: `moment` (for time formatting)
- **JS**: `alert.js`, `download-manager.js`

#### `src/views/subscription-manager.html.ecr`
- **Layout**: `layout "subscription-manager"`
- **Data needed**: `base_url`, `is_admin`, `Config.current.plugin_path`
- **JS/Data**: Fully client-side (Alpine.js `x-data="component()"`), fetches via API
- **Components**: `moment`
- **JS**: `alert.js`, `subscription-manager.js`

#### `src/views/plugin-download.html.ecr`
- **Layout**: `layout "plugin-download"`
- **Data needed**: `base_url`, `is_admin`, `Config.current.plugin_path`
- **JS/Data**: Fully client-side (Alpine.js `x-data="component()"`), uses API
- **Components**: `jquery-ui`, `moment`
- **JS**: `alert.js`, `plugin-download.js`

#### `src/views/missing-items.html.ecr`
- **Layout**: `layout "missing-items"`
- **Data needed**: `base_url`, `is_admin`
- **JS/Data**: Fully client-side (Alpine.js `x-data="component()"`), uses API
- **JS**: `alert.js`, `missing-items.js`

### 2.4 OPDS XML Templates

#### `src/views/opds/index.xml.ecr`
- **Data needed**: `base_url`, `titles` (all)
- **Format**: Atom feed with OPDS profile
- **Structure**: `<feed>` with `<entry>` per title. Each entry links to `/opds/book/:id`

#### `src/views/opds/title.xml.ecr`
- **Data needed**: `base_url`, `title` (single)
- **Format**: Atom feed with OPDS profile
- **Structure**: Entry per sub-title (navigation) and entry per chapter (acquisition links)

### 2.5 Component Templates (8 files in `src/views/components/`)

| File | Purpose | Data Needed |
|------|---------|-------------|
| `head.html.ecr` | HTML `<head>` section with meta, CSS, fonts, title | `page` (string), `base_url` |
| `uikit.html.ecr` | UIkit JS initialization | — |
| `card.html.ecr` | Manga card (cover, title, progress, dropdown) | `item` (Title or Entry), `progress` (Float64), `base_url` |
| `dots.html.ecr` | Loading spinner script | — |
| `sort-form.html.ecr` | Sort dropdown (method + ascend) | `hash` (sort options), `base_url` |
| `entry-modal.html.ecr` | Modal for entry details (progress, actions) | `base_url` |
| `moment.html.ecr` | Moment.js library include | — |
| `jquery-ui.html.ecr` | jQuery UI library include (for drag-select) | — |

---

## 3. Public Assets

### 3.1 Directory Structure

```
public/
├── css/
│   ├── mango.css          (compiled from mango.less)
│   ├── mango.less
│   ├── comic-theme.css    (compiled from comic-theme.less)
│   ├── comic-theme.less
│   ├── uikit.less         (UIkit custom overrides)
│   ├── tags.css           (tag/select2 styling)
│   ├── tags.less
│   └── select2.min.css    (Select2 library)
├── js/
│   ├── admin.js
│   ├── alert.js
│   ├── alpine.min.js
│   ├── alpine-ie11.min.js
│   ├── common.js
│   ├── dots.js
│   ├── download-manager.js
│   ├── i18n.js
│   ├── jquery.min.js
│   ├── jquery-ui.min.js
│   ├── jquery.inview.min.js
│   ├── missing-items.js
│   ├── moment.min.js
│   ├── plugin-download.js
│   ├── reader.js
│   ├── search.js
│   ├── select2.min.js
│   ├── sort-items.js
│   ├── subscription-manager.js
│   ├── subscription.js
│   ├── title.js
│   ├── user.js
│   ├── user-edit.js
│   ├── uikit.min.js
│   └── uikit-icons.min.js
├── img/
│   ├── banner.png
│   ├── banner-paddings.png
│   ├── loading.gif
│   ├── icons/
│   │   ├── icon.png
│   │   ├── icon_x96.png
│   │   ├── icon_x192.png
│   │   └── icon_x512.png
│   ├── accordion-close.svg
│   ├── accordion-open.svg
│   ├── divider-icon.svg
│   ├── form-checkbox.svg
│   ├── form-checkbox-indeterminate.svg
│   ├── form-datalist.svg
│   ├── form-radio.svg
│   ├── form-select.svg
│   ├── list-bullet.svg
│   ├── nav-parent-close.svg
│   └── nav-parent-open.svg
├── webfonts/
│   ├── fa-solid-900.ttf
│   ├── fa-solid-900.woff
│   └── fa-solid-900.woff2
├── favicon.ico
├── manifest.json
└── robots.txt
```

### 3.2 Key Notes for Go Port

- Static files are served from `public/` directory in dev mode
- In production (release), they are embedded via `baked_file_system` (see §6)
- CSS files have LESS sources that are compiled via gulp; the Go port should serve the compiled `.css` files
- JS files are plain JS (no bundler); all loaded via `<script src>` tags
- Static URL paths: `/css/*`, `/js/*`, `/img/*`, `/webfonts/*`, `/favicon.ico`, `/robots.txt`, `/manifest.json`
- These are listed in `STATIC_DIRS` constant in `src/util/util.cr`
- MIME types for `.woff`, `.woff2`, `.ttf`, `.ico`, `.cbz`, `.cbr` are registered in `register_mime_types`

---

## 4. OPDS Route Detail

### 4.1 Route Registration (src/routes/opds.cr)

```crystal
struct OPDSRouter
  def initialize
    get "/opds" do |env|
      titles = Library.default.titles
      render_xml "src/views/opds/index.xml.ecr"
    end

    get "/opds/book/:title_id" do |env|
      begin
        title = Library.default.get_title(env.params.url["title_id"]).not_nil!
        render_xml "src/views/opds/title.xml.ecr"
      rescue e
        Logger.error e
        env.response.status_code = 404
      end
    end
  end
end
```

### 4.2 Template: `src/views/opds/index.xml.ecr`

```xml
<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <id>urn:mango:index</id>
  <link rel="self" href="<%= base_url %>opds/" type="application/atom+xml;profile=opds-catalog;kind=navigation" />
  <link rel="start" href="<%= base_url %>opds/" type="application/atom+xml;profile=opds-catalog;kind=navigation" />
  <title>资料库</title>
  <author>
    <name>Mango</name>
    <uri>https://github.com/hkalexling/Mango</uri>
  </author>
  <% titles.each do |t| %>
    <entry>
      <title><%= HTML.escape(t.display_name) %></title>
      <id>urn:mango:<%= t.id %></id>
      <link type="application/atom+xml;profile=opds-catalog;kind=navigation" rel="subsection" href="<%= base_url %>opds/book/<%= t.id %>" />
    </entry>
  <% end %>
</feed>
```

### 4.3 Template: `src/views/opds/title.xml.ecr`

```xml
<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <id>urn:mango:<%= title.id %></id>
  <link rel="self" href="<%= base_url %>opds/book/<%= title.id %>" type="application/atom+xml;profile=opds-catalog;kind=navigation" />
  <link rel="start" href="<%= base_url %>opds/" type="application/atom+xml;profile=opds-catalog;kind=navigation" />
  <title><%= HTML.escape(title.display_name) %></title>
  <author>
    <name>Mango</name>
    <uri>https://github.com/hkalexling/Mango</uri>
  </author>
  <% title.titles.each do |t| %>
    <!-- Sub-title entries (navigation links) -->
    <entry>
      <title><%= HTML.escape(t.display_name) %></title>
      <id>urn:mango:<%= t.id %></id>
      <link type="application/atom+xml;profile=opds-catalog;kind=navigation" rel="subsection" href="<%= base_url %>opds/book/<%= t.id %>" />
    </entry>
  <% end %>
  <% title.entries.each do |e| %>
    <% next if e.err_msg %>
    <entry>
      <title><%= HTML.escape(e.display_name) %></title>
      <id>urn:mango:<%= e.id %></id>
      <link rel="http://opds-spec.org/image" href="<%= e.cover_url %>" />
      <link rel="http://opds-spec.org/image/thumbnail" href="<%= e.cover_url %>" />
      <link rel="http://opds-spec.org/acquisition" href="<%= base_url %>api/download/<%= e.book.id %>/<%= e.id %>" title="阅读" type="<%= MIME.from_filename e.path %>" />
      <link type="text/html" rel="alternate" title="在Mango中阅读" href="<%= base_url %>reader/<%= e.book.id %>/<%= e.id %>" />
      <link type="text/html" rel="alternate" title="在Mango中打开" href="<%= base_url %>book/<%= e.book.id %>" />
    </entry>
  <% end %>
</feed>
```

### 4.4 OPDS Auth

- OPDS routes are NOT in the skip-auth list (which is `/login`, `/logout`, `/api/login`, static files)
- AuthHandler checks OPDS specially: if the path starts with `/opds`, it requires basic auth (`require_basic_auth`)
- Basic auth format: `Authorization: Basic base64(username:password)`
- The auth handler decodes this, verifies user, and stores token in session

### 4.5 OPDS MIME Types (from `src/util/util.cr`)

```crystal
".zip" => "application/zip",
".rar" => "application/x-rar-compressed",
".cbz" => "application/vnd.comicbook+zip",
".cbr" => "application/vnd.comicbook-rar",
```

---

## 5. Middleware

### 5.1 Middleware Pipeline (from `src/server.cr`)

```crystal
Kemal.config.logging = false    # disable default Kemal logger
use LogHandler.new                # custom logging
use AuthHandler.new               # authentication + admin check
use UploadHandler.new(Config.current.upload_path)  # serve uploads
# In release mode only:
#   serve_static false            # disable Kemal static serving
#   use StaticHandler.new         # custom baked_file_system static handler
```

### 5.2 LogHandler (`src/handlers/log_handler.cr`)

- Extends `Kemal::BaseLogHandler`
- Logs each request: status code, method, path, elapsed time
- Uses `Logger.debug` (custom logger, not stdout)

### 5.3 AuthHandler (`src/handlers/auth_handler.cr`)

**Skip list** (no auth required):
- `/login`, `/logout`, `/api/login`
- Static files (`requesting_static_file` checks `STATIC_DIRS`)

**Auth methods** (in order of precedence):
1. **Session cookie token** — `env.session.string? "token"`, validated via `Storage.default.verify_token`
2. **Authorization header** — `Bearer <session_id>` (looks up session for token) or `Basic <base64(user:pass)>` (authenticates and stores token in session)
3. **disable_login mode** — uses `Config.current.default_username`
4. **auth_proxy_header** — reads header name from config, checks if user exists

**Admin check** (after auth):
- Paths starting with `/admin`, `/api/admin`, `/download` require admin
- `is_admin?(env)` checks token admin status, or proxy/disable_login config
- Non-admin gets 403

**OPDS special handling**:
- If path starts with `/opds` and not authenticated, sends `WWW-Authenticate: Basic` + 401

### 5.4 CORS Handler (implicit, from `src/util/web.cr`)

Not a middleware class — CORS is handled inline via the `cors` macro:
```crystal
macro cors
  env.response.headers["Access-Control-Allow-Methods"] = "HEAD,GET,PUT,POST,DELETE,OPTIONS"
  env.response.headers["Access-Control-Allow-Headers"] = "X-Requested-With,X-HTTP-Method-Override, Content-Type, Cache-Control, Accept, Authorization"
  env.response.headers["Access-Control-Allow-Origin"] = "*"
end
```

Also, server.cr registers explicit OPTIONS handlers for `/api/*`, `/uploads/*`, `/img/*`:
```crystal
{% for path in %w(/api/* /uploads/* /img/*) %}
  options {{path}} do |env|
    cors
    halt env
  end
{% end %}
```

And static file serving also adds CORS:
```crystal
static_headers do |env, _path, _fileinfo|
  env.response.headers.add("Access-Control-Allow-Origin", "*")
end
```

### 5.5 UploadHandler (`src/handlers/upload_handler.cr`)

- Serves files from the configured upload directory at `/uploads/*` path
- Path traversal protection: verifies resolved path starts with upload directory
- GET only

### 5.6 StaticHandler (`src/handlers/static_handler.cr`)

- Used only in release mode (when `baked_file_system` embeds static files)
- Reads from baked FS instead of disk
- Serves with correct MIME type

### 5.7 Session Configuration (from `src/server.cr`)

```crystal
Kemal::Session.config do |c|
  c.timeout = 365.days
  c.secret = Config.current.session_secret  # or random if empty
  c.cookie_name = "mango-sessid-#{Config.current.port}"
  c.path = Config.current.base_url
end
```

Note: The Go port already uses a different cookie name pattern (`mango-token-<port>`) as seen in `auth.go`.

---

## 6. Embed Strategy

### 6.1 Crystal: baked_file_system (`src/handlers/static_handler.cr`)

```crystal
require "baked_file_system"

class FS
  extend BakedFileSystem
  {% if flag?(:release) %}
    {% if read_file? "#{__DIR__}/../../dist/favicon.ico" %}
      {% puts "baking ../../dist" %}
      bake_folder "../../dist"
    {% else %}
      {% puts "baking ../../public" %}
      bake_folder "../../public"
    {% end %}
  {% end %}
end

class StaticHandler < Kemal::Handler
  def call(env)
    if requesting_static_file env
      file = FS.get? env.request.path
      return call_next env if file.nil?
      io = IO::Memory.new
      IO.copy file, io
      file.close
      return send_file env, io.to_slice, MIME.from_filename file.path
    end
    call_next env
  end
end
```

Key points:
- In dev mode, Kemal's built-in static file serving is used (disk)
- In release mode, `serve_static false` is called and `StaticHandler` is used
- The baked FS reads from `../../dist/` first (for gulp build output), then falls back to `../../public/`

### 6.2 Go Counterpart: `embed.FS`

Use Go's `//go:embed` directive:

```go
import "embed"

//go:embed public/* public/css/* public/js/* public/img/* public/webfonts/*
var staticFS embed.FS
```

Then use `http.FS(staticFS)` or serve individual files via `staticFS.Open(path)`.

---

## 7. Go Existing Auth Code

### 7.1 `go/internal/server/auth.go`

**Package**: `server` (import path `github.com/hkalexling/mango-go/internal/server`)

**AuthMiddleware** — returns `func(http.Handler) http.Handler`:
1. OPTIONS requests skip auth
2. Extracts token: cookie (`mango-token-<port>`), then legacy cookie (`mango-sessid-<port>`), then `Authorization: Bearer <token>`, or `Authorization: Basic ...`
3. If valid token → sets context keys: `username`, `is_admin`, `auth_method`
4. If no token → checks `cfg.DisableLogin` mode (use default username), then `cfg.AuthProxyHeaderName`, then requires auth

**AdminMiddleware** — checks `is_admin` context key, returns 403 if not admin

**Helper functions**:
- `GetUsername(r *http.Request) string` — extract username from context
- `GetIsAdmin(r *http.Request) bool` — extract admin flag
- `SetAuthTokenCookie(w, cfg, token)` — set cookie named `mango-token-<port>`
- `ClearAuthTokenCookie(w, cfg)` — clear the auth cookie

**Context keys**:
```go
const (
  contextKeyUsername   contextKey = "username"
  contextKeyIsAdmin    contextKey = "is_admin"
  contextKeyAuthMethod contextKey = "auth_method"
  cookieNamePrefix                = "mango-token-"
)
```

### 7.2 `go/internal/server/auth_test.go`

Tests cover:
- `TestAuthMiddlewareValidToken` — cookie auth with valid token
- `TestAuthMiddlewareInvalidToken` — cookie auth with invalid token → 401
- `TestAuthMiddlewareBearerToken` — Bearer header auth
- `TestAuthMiddlewareAdminToken` — admin token passes AdminMiddleware
- `TestAuthMiddlewareAdminRejected` — non-admin fails AdminMiddleware → 403
- `TestAuthMiddlewareOptions` — OPTIONS bypasses auth
- `TestAuthMiddlewareDisabledLogin` — disable_login mode uses default username
- `TestAuthMiddlewareDisabledLoginMissingDefaultUser` — nonexistent default user → 401
- `TestAuthMiddlewareProxyHeader` — X-Auth-User header auth
- `TestAuthMiddlewareProxyHeaderInvalidUser` — invalid proxy user → 401
- `TestSetAndClearAuthTokenCookie` — cookie set/clear
- `TestExtractTokenFromCookie/Bearer/Precedence` — token extraction

### 7.3 What's Missing in Go

The Go `internal/server/` directory currently only has `auth.go` and `auth_test.go`. There is no:
- Router
- Route handlers
- Template rendering
- Static file serving
- CORS middleware
- OPDS handler
- Web server start/stop

---

## 8. Go Module Dependencies

### 8.1 `go/go.mod`

```
module github.com/hkalexling/mango-go
go 1.26.3
```

**Direct dependencies** (already available):
| Dependency | Purpose |
|---|---|
| `github.com/PuerkitoBio/goquery v1.12.0` | HTML parsing (for plugins) |
| `github.com/bodgit/sevenzip v1.6.4` | 7z archive support |
| `github.com/dop251/goja v0.0.0-20260701091749-b07b74453ea9` | JS engine (plugin runtime) |
| `github.com/google/uuid v1.6.0` | UUID generation |
| `github.com/nwaples/rardecode v1.1.3` | RAR archive support |
| `github.com/olekukonko/tablewriter v1.1.4` | CLI table formatting |
| `github.com/spf13/cobra v1.10.2` | CLI framework |
| `golang.org/x/crypto v0.53.0` | bcrypt password hashing |
| `golang.org/x/image v0.43.0` | Image processing (thumbnail generation) |
| `gopkg.in/yaml.v3 v3.0.1` | YAML config parsing |
| `modernc.org/sqlite v1.53.0` | Pure-Go SQLite driver |

### 8.2 Dependencies NEEDED for Phase 4 (not yet in go.mod)

Based on the design doc and Crystal requirements:

| Need | Recommended Go Library | Reason |
|---|---|---|
| HTTP Router with path params | `github.com/go-chi/chi/v5` | Matches Crystal route params (`:tid`, `:eid`) |
| Template rendering | Go stdlib `html/template` | Already available in std |
| Session management | `github.com/alexedwards/scs/v2` or gorilla/sessions | kemal-session replacement |
| CORS middleware | `github.com/go-chi/cors` or manual | CORS headers for API |

Note: chi is the recommended router in the design doc. It supports URL params via `chi.URLParam(r, "tid")`.

---

## 9. Key Implementation Notes for Go Port

### 9.1 Route Registration Pattern

The Crystal code uses inline route registration (routes are registered inside `#initialize` of each router struct). For Go, use a chi router:

```go
func RegisterRoutes(r chi.Router, deps *Dependencies) {
    r.Route("/api", func(r chi.Router) {
        r.Post("/login", deps.APILogin)
        r.Group(func(r chi.Router) {
            r.Use(deps.AuthMiddleware)
            r.Get("/page/{tid}/{eid}/{page}", deps.APIPage)
            // ... all other /api routes
        })
        r.Route("/admin", func(r chi.Router) {
            r.Use(deps.AdminMiddleware)
            r.Post("/scan", deps.APIScan)
            // ...
        })
    })
    r.Get("/", deps.Home)
    // ...
}
```

### 9.2 Response Helpers (from `src/util/web.cr`)

These Crystal macros/helpers need Go equivalents:

| Crystal | Go Equivalent |
|---|---|
| `send_json(env, json)` | `w.Header().Set("Content-Type", "application/json"); w.Write(json)` + CORS |
| `send_text(env, text)` | `w.Header().Set("Content-Type", "text/plain"); w.Write([]byte(text))` |
| `send_img(env, img)` | `w.Header().Set("Content-Type", img.mime); w.Write(img.data)` |
| `send_attachment(env, path)` | Set `Content-Disposition: attachment; filename=...` + serve file |
| `send_file(env, data, mime)` | Set content type + write bytes |
| `redirect(env, path)` | `http.Redirect(w, r, path, http.StatusFound)` |
| `render_xml(path)` | Execute XML template + set `Content-Type: application/xml` |
| `layout "name"` | Execute template inside layout template |
| `cors` | Set CORS headers |

### 9.3 User Context Access

Crystal uses `env.session` and `get_username env` macro. Go uses context:

```go
username := server.GetUsername(r)
isAdmin := server.GetIsAdmin(r)
```

### 9.4 Template Rendering Strategy

Crystal ECR templates need to be converted to Go `html/template` format. Key differences:
- `<%= expr %>` → `{{ .Expr }}`
- `<% code %>` → `{{/* code */}}` or use template functions
- `layout "name"` → Use `{{ template "layout" . }}` with `{{ block "content" . }}...{{ end }}`
- Components: Use `{{ template "component" . }}`
- `yield_content "script"` → Use `{{ block "scripts" . }}{{ end }}`

### 9.5 Template Data Structures

Each page will need a Go struct to hold template data:

```go
type LayoutData struct {
    BaseURL  string
    IsAdmin  bool
    PageName string
    Content  interface{} // nested template data
}

type HomeData struct {
    LayoutData
    ContinueReading []ContinueReadingItem
    RecentlyAdded   []RecentlyAddedItem
    StartReading    []*library.Title
    NewUser         bool
    EmptyLibrary    bool
}
```

### 9.6 Static File Route

The STATIC_DIRS paths from `src/util/util.cr`:
```crystal
STATIC_DIRS = %w(/css /js /img /webfonts /favicon.ico /robots.txt /manifest.json)
```

In Go, these can be served via:
```go
fileServer := http.FileServer(http.FS(staticFS))
r.Handle("/*", fileServer)  // or register specific prefixes
```

### 9.7 OPTIONS Preflight

Register OPTIONS handlers for `/api/*`, `/uploads/*`, `/img/*`:
```go
r.Options("/*", func(w http.ResponseWriter, r *http.Request) {
    setCORSHeaders(w)
    w.WriteHeader(http.StatusOK)
})
```

Or use chi's CORS middleware.

### 9.8 Upload URL Prefix

From `src/util/util.cr`:
```crystal
UPLOAD_URL_PREFIX = "/uploads"
```

The UploadHandler serves files from `Config.current.upload_path` at `/uploads/<path>`, with path traversal protection.

### 9.9 Reader Page Special Handling

The reader page at `/reader/:title/:entry/:page` is special:
- It's a standalone HTML page (NOT wrapped in layout)
- It renders `src/views/reader.html.ecr` directly via `render "src/views/reader.html.ecr"`
- In Go, this should render a template without the layout wrapper
- Route 57 (without page) redirects to route 58 (with page) after loading progress

### 9.10 Error Handling Pattern

Every Crystal route wraps in `begin/rescue`. The Go equivalent should follow the same pattern — each handler should catch panics or handle errors and return appropriate responses.

---

## Caveats / Not Found

- The exact `go.sum` file was not read (not needed for migration)
- The `public/` less files are compiled via gulp — Go just serves the compiled CSS
- The `redoc.standalone.js` mentioned in `api.html.ecr` was not found in `public/js/` — may be loaded from CDN or needs investigation
- The exact list of 67 routes vs 68 may vary — the WS route and OPTIONS routes are not standard HTTP routes
- Go module does not yet have chi or any session library — these need to be added
- No existing template files exist in the Go directory
