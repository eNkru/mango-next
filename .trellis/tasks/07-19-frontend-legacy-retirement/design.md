# Legacy frontend asset retirement design

## Strategy (executed)

1. Unmounted deferred product routes from `RegisterRoutes`.
2. Deleted page-local templates/scripts for deferred + migrated pages.
3. Deleted dead Go handlers and unused page data structs.
4. Removed shared legacy chrome (no remaining `renderLayout` consumer).
5. Kept React shell, built assets, PWA icons, OPDS.

## Routes removed

| Route | Status |
|-------|--------|
| `/admin/subscriptions` | removed |
| `/admin/downloads` | removed |
| `/download/plugins` | removed |
| form `POST /admin/user/edit*` | removed (React uses JSON APIs) |

Plugin/subscription **JSON APIs** retained for possible automation.

## Kept

- `views/react-shell.tmpl`
- `public/react/**`
- `public/img/icons/**` (manifest PWA)
- `favicon.ico`, `manifest.json`, `robots.txt`
- OPDS XML handlers

## Deleted (summary)

- All other `views/*.tmpl` (legacy layout + pages)
- `public/js/**`, `public/css/**`, `public/webfonts/**`
- UIkit decorative `public/img/*` (non-icon)
- Legacy page handlers: Home/Title legacy, plugin/subscription/download managers

## Rollback

Git revert of this branch/PR. No DB migration.
