# React Vite frontend migration design

## Architecture

```text
Browser
  ├─ Migrated route: Go HTML shell + React bundle
  └─ Unmigrated route: existing Go templates + legacy public assets
           ↓
Go HTTP Server
  ├─ Auth / admin middleware
  ├─ BaseURL mount
  ├─ JSON APIs
  ├─ SQLite / library / queue / tasks
  └─ go:embed of views + public assets
```

Go remains the only long-running process. Vite is build-time only.

## Build and embed contract

1. Root `package.json` owns React/Vite/TypeScript and build scripts.
2. `npm ci` installs the lockfile.
3. `npm run build` produces hashed JS/CSS into a declared public output path
   under `go/web/public/`.
4. Go serves and embeds those outputs.
5. Make `build` / `static` / `run` and Docker Node stage generate React assets
   before `go build`.

The redirected asset-pipeline work contributes npm lockfile discipline, Make and
Docker ordering, local font packaging, and offline constraints. It does not keep
the old jQuery/LESS copy inventory as the final frontend architecture.

## Routing and coexistence

- Unmigrated routes keep current handlers and templates.
- Migrated routes keep their Go paths, e.g. `/admin/missing`.
- Go returns a React HTML shell for migrated routes after auth/admin checks.
- The shell injects `BaseURL`, page identity, and any tiny boot config needed by
  React.
- React mounts into a root element and renders only the page for that path in
  the first delivery. Client-side multi-page SPA ownership is deferred.
- Navigation links may point from React pages to unmigrated Go pages and vice
  versa.

## Theme strategy

- Preserve comic/flat and light/dark behavior.
- Keep FOUC-safe boot markers on `html` before paint.
- Port existing visual tokens into React-owned CSS files.
- Host local fonts already required by the comic theme.
- Do not adopt MUI/Ant Design or another full component library in the first
  delivery.

## Pilot API boundary (completed)

Missing-items pilot and later children landed real JSON contracts for browse,
tags, users, login, and reader bootstrap. See:

- `.trellis/spec/backend/react-browse-api.md`
- `.trellis/spec/backend/react-reader-api.md`
- `.trellis/spec/frontend/react-reader.md`

## Remaining migration boundaries

| Surface | Status |
|---------|--------|
| Core browse/read/admin React surfaces | Done via children |
| `/admin/subscriptions` | **Deferred** — product will not migrate |
| `/download/plugins` | **Deferred** — product will not migrate |
| `/opds` | Keep Go XML |
| `/admin/downloads` | Disabled product; out of scope |
| Legacy tmpl/js retirement | Optional cleanup only |

## Security and BaseURL

- Reuse Go session cookie auth and admin middleware.
- React requests stay same-origin under `BaseURL`.
- Static asset URLs and API clients must prefix `BaseURL`.
- No public CDN runtime dependencies.

## Rollback

- Foundation can ship alone if no migrated route is switched yet.
- Pilot rollback is route-level: restore the Go template handler for
  `/admin/missing` without removing the React build toolchain.
- Do not hard-cut the entire UI until later children land.

## Validation shape

```bash
npm ci
npm run build
npm run typecheck   # or equivalent
make check
make test
make build
docker build -t mango-react-check .
```

Additional checks:

- `/admin/missing` loads the React shell under `/` and `/mango/`.
- Unmigrated routes still render Go templates.
- Missing list/delete APIs return real data and consistent errors.
- No Google Fonts or other CDN runtime requests from the React shell.
