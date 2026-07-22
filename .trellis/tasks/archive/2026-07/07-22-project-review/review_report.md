# Comprehensive Project Architecture & Code Quality Review

**Date:** July 22, 2026  
**Target Project:** Mango Next (Go Backend + React/TypeScript Frontend)

---

## Executive Summary

Mango Next is a lightweight, modern web-based manga/comic server and reader. Overall, the codebase demonstrates **clean separation of concerns, high quality modularity, strong unit test coverage for core components, and modern frontend design principles**.

This review provides a structured evaluation across the Go backend and React frontend architectures, followed by prioritized optimization recommendations categorized by **Impact** and **Effort**.

---

## 1. Backend Architecture & Code Quality (Go)

### 1.1 Architecture & Package Structure
- **Strengths:**
  - Standard Go `internal/` package layout (`internal/storage`, `internal/library`, `internal/server`, `internal/tasks`, `internal/archive`, `internal/config`, `internal/upload`, `internal/rename`, `internal/thumbnail`).
  - Separation of domain logic (`library`, `archive`) from transport handlers (`server`).
  - Explicit dependency injection pattern via `server.Dependencies` struct passing singletons (`Storage`, `Library`, `Runner`, `Config`).

- **Areas for Improvement / Issues:**
  - **Overloaded Storage package (`internal/storage/storage.go`):** The `Storage` struct spans 1100+ lines in a single file, handling database connection, user CRUD, library data persistence, cover thumbnail extraction/caching, zip archive inspection, tag management, reading progress, and schema migrations.
  - **Direct Sqlite Connection Management & Single-Writer Bottleneck:** SQLite is configured with `db.SetMaxOpenConns(1)`. While necessary for standard SQLite single-writer safety, heavy concurrent API requests (e.g. image thumbnail generation + progress sync during rapid client scrolling) can suffer lock contention or database timeouts under high load. Enable WAL mode (`PRAGMA journal_mode=WAL`) explicitly.
  - **Inconsistent Error Handling & HTTP Statuses:** In `internal/server/handlers_api.go` and `browse_api.go`, errors are occasionally logged via standard `log.Printf` and sent back as standard `500 Internal Server Error` without structured error logging (e.g., using `slog` or contextual error codes).

### 1.2 Performance & Resource Management
- **Archive Handling (`internal/archive`):** Support for `.zip`, `.cbz`, `.rar`, `.cbr`, `.7z`, `.cb7` via custom archive wrappers.
- **Resource Leak Risks:** Archive extraction opens file handles inside zip/rar readers. Ensuring proper stream closing across image extraction endpoints (`/api/cover`, `/api/page`) is critical.
- **Concurrency & Background Tasks (`internal/tasks`):** Background Library Scanning (`Runner`) operates with clear job queues and mutexes. However, long-running scan jobs should report granular progress via Server-Sent Events (SSE) or WebSockets instead of polling `TaskInfo`.

---

## 2. Frontend Architecture & UI/UX (React + TypeScript)

### 1.1 Architecture & State Management
- **Strengths:**
  - Clean client-side page routing in `App.tsx` matching server-injected boot parameters (`readBoot()`).
  - Modularized page views (`HomePage`, `LibraryPage`, `TitleDetailPage`, `ReaderPage`, `AdminPage`).
  - Strong component isolation in reader (`ReaderPage`, `ReaderViewport`, `ReaderControls`, `ReaderTopBar`, `useReaderNavigation`, `useReaderProgress`).

- **Areas for Improvement / Issues:**
  - **Server-Driven Page Navigation Strategy:** Page navigation relies on full-page HTTP navigations/reloads in many places (changing boot state via URL reloads or standard link navigation) rather than smooth Client-Side Routing (e.g., HTML5 History API / React Router style without re-executing boot initialization).
  - **Global State vs Local State:** Preference handling (`useReaderPrefs`, UI filters) uses `localStorage` directly in separate hooks. Centralizing user settings and reader preferences into a global React Context / Store (e.g. Zustand or React Context) will prevent state synchronization drift across tabs or re-renders.
  - **CSS Styling Structure:** Styles are primarily centralized in monolithic CSS stylesheets rather than modular CSS modules / Tailwind / utility-first components, leading to potential style leakages across components.

### 1.2 UI/UX Experience & Visual Polish
- **Strengths:**
  - Interactive cover carousels (`ContinueCarousel`) with stacked 3D transformation cards.
  - Responsive Reader supporting Webtoon continuous scrolling, single-page, and double-page mode with RTL (Right-to-Left) toggle.
  - Dark/Light Theme adaptability and internationalization (`i18n`) support.

- **Areas for Improvement / Issues:**
  - **Reader Controls Auto-Hide Polish:** Mobile touch vs Desktop hover interactions for `ReaderTopBar` and `ReaderControls` occasionally collide with fast gestures or edge mouse movement.
  - **Empty / Loading / Error States Consistency:** StatePanels (`ErrorState`, `LoadingState`) exist, but skeleton loaders for manga poster rails (`PosterRail`) during async library fetching are missing, leading to abrupt layout shifts (CLS).
  - **Keyboard Navigation Accessibility:** Keybindings in the reader are well implemented, but focus outlines and ARIA live regions for screen readers in complex carousels need accessibility auditing.

---

## 3. Prioritized Recommendations Matrix

### 🚀 Quick Wins (High Impact, Low/Medium Effort)
1. **Enable SQLite WAL Mode & Busy Timeout (Backend):**
   - Add `PRAGMA journal_mode=WAL;` and `PRAGMA busy_timeout=5000;` in `internal/storage/storage.go`. Prevents "database locked" errors during simultaneous background scans and reader progress writes.
2. **Implement Skeleton Loaders for Poster Rails (Frontend):**
   - Add animated shimmer skeleton placeholders in `PosterRail.tsx` to eliminate Cumulative Layout Shift (CLS).
3. **Structured Logging Migration (Backend):**
   - Replace standard `log.Printf` with Go 1.21+ `log/slog` to structured context-aware JSON/Text logs.

### 🏗️ Deep Structural Improvements (High Impact, High Effort)
1. **Decompose `internal/storage` Package (Backend):**
   - Split `storage.go` into domain repositories: `user_repo.go`, `progress_repo.go`, `tag_repo.go`, `book_repo.go`.
2. **Client-Side SPA Router Integration (Frontend):**
   - Migrate from window navigation / server boot reloads to lightweight HTML5 History API client routing to eliminate flash-of-white-screen on page switches.
3. **Global UI Context & State Store (Frontend):**
   - Create unified `UserPrefsContext` to synchronize theme, language, and reader options across components cleanly.

---

## Conclusion & Action Plan

The project foundation is solid and maintainable. Implementing the Quick Wins first will dramatically boost SQLite concurrency stability and visual smoothness, setting up the codebase for deeper structural refactoring.
