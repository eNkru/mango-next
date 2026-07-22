# Design: Client-Side SPA Router

## Architecture

```
main.tsx
  I18nProvider
    BootProvider (sticky session from #mango-boot)
      BrowserRouter basename={baseUrl without trailing slash rules}
        App routes
```

| Piece | Role |
|-------|------|
| `BootProvider` | Read `readBoot()` once; expose `baseUrl`, `isAdmin`, `version`, helpers. Does **not** re-read pageId from DOM on every nav. |
| `react-router` routes | Map path → page component; params for `titleId`, `tag`, `tid`, `eid`, `page`, `username`. |
| `AppLink` / `navigate` | Internal helper: SPA for in-app paths; raw `<a>` for logout/download/external. |
| Go | Unchanged: each path first load still returns React shell + fresh `#mango-boot`. |

## Routes (under BaseURL)

| Path | Component | Notes |
|------|-----------|--------|
| `/` | HomePage | |
| `/login` | LoginPage | may stay full load after submit |
| `/library` | LibraryPage | query `show_hidden` |
| `/book/:titleId` | TitleDetailPage | |
| `/tags` | TagsIndexPage | |
| `/tags/:tag` | TagDetailPage | |
| `/admin` | AdminPage | |
| `/admin/user` | UserListPage | |
| `/admin/user/edit` | UserEditPage | query username |
| `/admin/missing` | MissingItemsPage | |
| `/reader/:tid/:eid` | ReaderPage | optional `/:page` |
| `/reader/:tid/:eid/:page` | ReaderPage | 1-based page |

`basename` = configured base path (from boot), matching existing `baseUrl()` join rules.

## Reader

- Route renders `ReaderPage` only (no `AppShell`).
- Replace `window.location.replace` / hard entry links with `navigate(...)` where SPA.
- Page number updates: prefer `navigate(..., { replace: true })` aligned with current `useReaderNavigation` replaceState so history is not flooded.
- Exit to library/title: SPA navigate.

## Link conversion priority

1. `AppShell` topbar + brand  
2. `PosterCard`, Continue carousel continue button  
3. Admin cards, user list/edit links  
4. Tags list/detail, title breadcrumb  
5. Reader entry from title detail / jump select / completeAndGo  

Keep hard: `logout`, `api/download/*`.

## Compatibility

- First paint: Go injects boot; Router matches current URL.
- Vite dev: no mango-boot → BootProvider uses `bootFromPathname` defaults + URL.
- No API contract changes.

## Trade-offs

| Choice | Why |
|--------|-----|
| react-router | User preference; standard nested routes / params |
| Sticky boot | Avoids re-fetch; admin flag stable until full reload (login already full-nav) |
| Reader in SPA | User preference for nicer enter/exit |

## Rollback

Remove dependency and restore `<a href>` + `App` switch. Git revert.
