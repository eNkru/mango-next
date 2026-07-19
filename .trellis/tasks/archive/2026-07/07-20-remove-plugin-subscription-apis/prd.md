# Remove plugin and subscription APIs

## Goal

Remove unused plugin download, subscription, and download-queue HTTP surfaces
and stop running plugin updater/downloader background tasks. Product will not
use these features.

## Requirements

- Unregister all `/api/admin/plugin*` and `/api/admin/queue*` routes.
- Delete corresponding handlers.
- Stop initializing download queue / plugin tasks at process start when nothing
  else needs them.
- Keep library scan, thumbnails, users, tags, progress, OPDS, React UI working.
- `go test ./...` and frontend typecheck stay green.

## Acceptance Criteria

- [ ] No plugin/subscription/queue admin HTTP routes registered.
- [ ] Server deps no longer require Queue/Plugins for normal serve.
- [ ] Main does not start plugin updater/downloader.
- [ ] Build and tests pass.

## Out of Scope

- React UI (already gone).
- Full config key deletion from YAML (can leave unused keys harmless).
- OPDS / core library APIs.
