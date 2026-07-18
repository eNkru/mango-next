# React login implementation plan

## 1. Go auth redirect + shell mount

- [x] Update `requireAuth` page redirect to include safe `?callback=`
- [x] Update `handleLoginPage`:
  - [x] If valid session → redirect to safe callback or BaseURL home
  - [x] Else render React shell pageId `login`, pass sanitized callback in boot
- [x] Keep `POST /login` and `POST /api/login` unchanged except any shared
      helper extraction if needed
- [x] Add/adjust focused Go tests for login-page auth redirect and requireAuth
      callback

## 2. React login page

- [x] Add `LoginPage` (standalone layout, not admin AppShell)
- [x] Wire `POST /api/login` via `apiFetch`
- [x] Client-side safe callback helper aligned with Go intent
- [x] Password visibility toggle + loading submit + in-page error
- [x] Register pageId `login` in `App.tsx`
- [x] Add login page styles using React tokens

## 3. Verify

```bash
npm run typecheck
npm run build
make check
make test
make build
```

Manual:

- anonymous login success → home
- login with `?callback=/library` → library (BaseURL-aware)
- bad credentials → in-page error
- already logged-in GET `/login` → leaves login
- protected page when logged out → `/login?callback=...` then return
- dual-theme markers still apply on login card
