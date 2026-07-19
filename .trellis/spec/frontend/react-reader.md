# React Immersive Reader

## 1. Scope / Trigger

Apply when changing the React reader page (`pageId: reader`), reader prefs,
navigation/progress math, or immersive chrome. Reader is **not** wrapped in
`AppShell`.

## 2. Signatures / Layout

```text
frontend/src/pages/reader/
  ReaderPage.tsx
  ReaderViewport.tsx
  ReaderTopBar.tsx
  ReaderControls.tsx
  useReaderBootstrap.ts
  useReaderNavigation.ts
  useReaderProgress.ts
  useReaderPrefs.ts
  readerMath.ts
  types.ts
```

Boot fields (from `#mango-boot`): `tid`, `eid`, `page` (one-based), plus shared
`baseUrl` / `pageId`.

## 3. Contracts

### Prefs (`localStorage`, namespaced only)

| Key | Default |
|---|---|
| `mango.reader.mode` | `continuous` \| `paged` |
| `mango.reader.margin` | `30` |
| `mango.reader.fitType` | `vert` \| `horz` \| `original` |
| `mango.reader.preloadLookahead` | `0..5` (default 3) |
| `mango.reader.enableFlipAnimation` | `true` |
| `mango.reader.enableRightToLeft` | `false` |

Do **not** read legacy unprefixed keys (`mode`, `margin`, …).

### Navigation / progress

- Single page source of truth in `useReaderNavigation`; URL via
  `history.replaceState` (`baseUrl('reader/{tid}/{eid}/{page}')`).
- Page images: `baseUrl(readerPageImagePath(tid, eid, page))` with **1-based**
  page index (matches Go `/api/page` and `ReadPage`).
- Progress throttle (`shouldSaveProgress`): save on first/last page, when
  `|page - lastSaved| >= 5`, or long-page title (avg height/width > 2).
- Next-entry / exit: `complete()` saves `pages` then `location.replace`.

### Chrome

- Default full-viewport black reader; top bar hidden.
- Show on top-edge hit (~36px), control open, or Escape; animate ~160ms; idle
  hide ~1.8s when pointer leaves bar and controls closed.
- Language selector in top bar; control panel for mode/fit/margin/RTL/preload/
  page jump/entry jump/prev-next/exit.

## 4. Validation & Error Matrix

| Condition | UI |
|---|---|
| Bootstrap loading | gate `LoadingState` |
| Bootstrap 4xx/5xx or empty pages | gate `ErrorState` + library link |
| Last page + zone next | open controls (legacy) |
| Continuous mode | keyboard flip ignored; strip + footer next/exit |

## 5. Good/Base/Bad Cases

- Good: deep link `/reader/t/e/5` opens page 5; flip updates URL without reload;
  prefs survive reload under `mango.reader.*`.
- Base: continuous mode default; bar hidden until edge hover.
- Bad: wrap reader in `AppShell` (breaks immersive layout and double chrome).

## 6. Tests Required

- Pure unit tests for `readerMath` (clamp, throttle, long-page, RTL direction).
- Frontend `npm run typecheck` + `npm run build`.
- Manual smoke: title-detail → reader, mid-page deep link, modes, next/exit,
  auto-hide bar, language switch, BaseURL if non-root.

## 7. Wrong vs Correct

#### Wrong

```ts
// 0-based conversion against live API
`api/page/${tid}/${eid}/${page - 1}`
```

#### Correct

```ts
readerPageImagePath(tid, eid, page) // 1-based
```

#### Wrong

```tsx
case 'reader':
  return <AppShell title="Reader"><ReaderPage /></AppShell>;
```

#### Correct

```tsx
case 'reader':
  return <ReaderPage />; // immersive shell only
```
