# React Vite foundation implementation plan

## 1. Redirect and clean the npm toolchain

- [x] Archive/close `07-17-frontend-asset-pipeline` as redirected.
- [x] Convert root `package.json` to React + Vite + TypeScript.
- [x] Keep lockfile-based `npm ci`.
- [x] Drop the old copy/LESS inventory as the primary build path while leaving
      legacy public files available for unmigrated templates.

## 2. Scaffold React app

- [x] Create `frontend/` (or equivalent) with Vite React-TS template structure.
- [x] Configure Vite `base` compatible with runtime BaseURL (document approach:
      build with relative base or inject BaseURL at runtime for asset tags via Go).
- [x] Output production assets to `go/web/public/react/`.
- [x] Add `typecheck` script.

## 3. Theme tokens and shell primitives

- [x] Add FOUC-safe theme boot in the Go HTML shell.
- [x] Port comic/flat tokens to React CSS.
- [x] Implement shell layout, alerts, confirm, loading/empty/error, baseUrl/api helpers.
- [x] Placeholder page using those primitives.

## 4. Go HTML shell + placeholder route

- [x] Add `react-shell.tmpl`.
- [x] Add admin-only placeholder route and handler.
- [x] Ensure shell injects BaseURL and page id for React boot.
- [x] Keep all existing template routes unchanged.

## 5. Make / Docker / docs

- [x] Wire Make build/static/run/check to React build/typecheck as appropriate.
- [x] Update Dockerfile Node stage for Vite build inputs (`frontend/`, config files).
- [x] Update README and FRONTEND_DEV_GUIDE for React foundation + coexistence.
- [x] Update theme spec notes if shell boot ownership changes.

## 6. Verify

```bash
npm ci
npm run typecheck
npm run build
make check
make test
make build
docker build -t mango-react-foundation .
```

- [x] Placeholder route works.
- [x] Unmigrated templates still work.
- [x] BaseURL-safe asset references.
- [x] No CDN font runtime references in React shell.
- [x] `npm run typecheck`, `npm run build`, `make check`, `make test`,
      `make build`, and `docker build -t mango-react-foundation .` passed.
