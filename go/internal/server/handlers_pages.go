package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	cfg := s.Deps.Config
	if token := extractToken(r, cfg); token != "" {
		if username, err := s.Deps.Storage.VerifyToken(token); err == nil && username != "" {
			http.Redirect(w, r, s.resolvePostLoginRedirect(r.URL.Query().Get("callback")), http.StatusFound)
			return
		}
	}

	extra := map[string]any{"isAdmin": false}
	if raw := strings.TrimSpace(r.URL.Query().Get("callback")); raw != "" {
		extra["callback"] = safeRedirectPath(raw)
	}
	s.renderReactShell(w, "login", "login", extra)
}

// resolvePostLoginRedirect maps a raw callback to a BaseURL-aware app path.
func (s *Server) resolvePostLoginRedirect(callback string) string {
	cb := safeRedirectPath(callback)
	if cb == "/" {
		return s.appPath("")
	}
	if s.Deps.Config != nil && s.Deps.Config.BaseURL != "/" && strings.HasPrefix(cb, "/") && !strings.HasPrefix(cb, s.Deps.Config.BaseURL) {
		return strings.TrimSuffix(s.Deps.Config.BaseURL, "/") + cb
	}
	return cb
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	loginPath := s.appPath("login")
	r.Body = http.MaxBytesReader(w, r.Body, loginBodyLimit)
	if !loginAllowed(r) {
		http.Redirect(w, r, loginPath, http.StatusFound)
		return
	}
	if err := r.ParseForm(); err != nil {
		recordLoginFailure(r)
		http.Redirect(w, r, loginPath, http.StatusFound)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	token, err := s.Deps.Storage.VerifyUser(username, password)
	if err != nil || token == "" {
		if err != nil {
			slog.Error("Login error", "err", err)
		}
		recordLoginFailure(r)
		http.Redirect(w, r, loginPath, http.StatusFound)
		return
	}

	clearLoginFailures(r)
	SetAuthTokenCookie(w, r, s.Deps.Config, token)
	http.Redirect(w, r, s.resolvePostLoginRedirect(r.FormValue("callback")), http.StatusFound)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	cfg := s.Deps.Config
	if token := extractToken(r, cfg); token != "" {
		_ = s.Deps.Storage.Logout(token)
	}
	ClearAuthTokenCookie(w, r, cfg)
	http.Redirect(w, r, s.appPath("login"), http.StatusFound)
}

// handleHome mirrors Crystal GET / (src/routes/main.cr):
//
//	new_user = !titles.any? { load_percentage(username) > 0 }
//	empty_library = titles.size == 0
//	plus continue_reading / start_reading / recently_added sections.
func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	s.renderReactShell(w, "home", "home", map[string]any{"isAdmin": GetIsAdmin(r)})
}

// buildLibraryPageData assembles library page data with hidden-title filtering.
// showHidden is only effective for admins.
func (s *Server) buildLibraryPageData(isAdmin bool, showHidden bool) LibraryPageData {
	showHidden = isAdmin && showHidden

	lib := s.Deps.Library
	lib.RLock()
	titleIDs := make([]string, len(lib.TitleIDs))
	copy(titleIDs, lib.TitleIDs)
	lib.RUnlock()

	hiddenIDs, err := s.Deps.Storage.GetHiddenTitleIDs()
	if err != nil {
		slog.Error("library GetHiddenTitleIDs", "err", err)
		hiddenIDs = nil
	}
	hiddenSet := make(map[string]bool, len(hiddenIDs))
	for _, id := range hiddenIDs {
		hiddenSet[id] = true
	}

	var titles []LibraryTitle
	for _, id := range titleIDs {
		lib.RLock()
		t, ok := lib.TitleHash[id]
		lib.RUnlock()
		if !ok {
			continue
		}
		isHidden := hiddenSet[t.ID]
		if isHidden && !showHidden {
			continue
		}
		titles = append(titles, LibraryTitle{
			ID:         t.ID,
			Name:       t.Name,
			CoverURL:   fmt.Sprintf("%sapi/cover/%s/%s", s.Deps.Config.BaseURL, t.ID, firstEntryID(t)),
			EntryCount: len(t.DeepEntries()),
			Hidden:     isHidden,
		})
	}

	return LibraryPageData{
		LayoutData: LayoutData{
			BaseURL:  s.Deps.Config.BaseURL,
			IsAdmin:  isAdmin,
			PageName: "library",
			Version:  "2.0.0",
		},
		Titles:     titles,
		Percentage: make([]float64, len(titles)),
		ShowHidden: showHidden,
	}
}

func (s *Server) handleLibrary(w http.ResponseWriter, r *http.Request) {
	s.renderReactShell(w, "library", "library", map[string]any{
		"isAdmin": GetIsAdmin(r),
	})
}

func (s *Server) handleTitle(w http.ResponseWriter, r *http.Request) {
	titleID := chi.URLParam(r, "title")
	s.renderReactShell(w, "title-detail", "title", map[string]any{
		"isAdmin": GetIsAdmin(r),
		"titleId": titleID,
	})
}

func (s *Server) handleTags(w http.ResponseWriter, r *http.Request) {
	s.renderReactShell(w, "tags-index", "tags", map[string]any{
		"isAdmin": GetIsAdmin(r),
	})
}

// buildTagPageData assembles tag page data with hidden-title filtering.
// Returns false when the tag has no visible titles for the requested mode.
func (s *Server) buildTagPageData(tag string, isAdmin bool, showHidden bool) (TagPageData, bool) {
	showHidden = isAdmin && showHidden

	titleIDs, err := s.Deps.Storage.GetTagTitles(tag, showHidden)
	if err != nil || len(titleIDs) == 0 {
		return TagPageData{}, false
	}

	hiddenIDs, err := s.Deps.Storage.GetHiddenTitleIDs()
	if err != nil {
		slog.Error("tag GetHiddenTitleIDs", "err", err)
		hiddenIDs = nil
	}
	hiddenSet := make(map[string]bool, len(hiddenIDs))
	for _, id := range hiddenIDs {
		hiddenSet[id] = true
	}

	lib := s.Deps.Library
	if lib == nil {
		return TagPageData{}, false
	}
	lib.RLock()
	var titles []LibraryTitle
	for _, id := range titleIDs {
		t, ok := lib.TitleHash[id]
		if !ok {
			continue
		}
		titles = append(titles, LibraryTitle{
			ID:         t.ID,
			Name:       t.Name,
			CoverURL:   fmt.Sprintf("%sapi/cover/%s/%s", s.Deps.Config.BaseURL, t.ID, firstEntryID(t)),
			EntryCount: len(t.Entries),
			Hidden:     hiddenSet[t.ID],
		})
	}
	lib.RUnlock()

	if len(titles) == 0 {
		return TagPageData{}, false
	}

	return TagPageData{
		LayoutData: LayoutData{
			BaseURL:  s.Deps.Config.BaseURL,
			IsAdmin:  isAdmin,
			PageName: "tag",
			Version:  "2.0.0",
		},
		Tag:        tag,
		Titles:     titles,
		ShowHidden: showHidden,
	}, true
}

func (s *Server) handleTag(w http.ResponseWriter, r *http.Request) {
	tag := chi.URLParam(r, "tag")
	// Validate existence with the same visibility rules before mounting React.
	isAdmin := GetIsAdmin(r)
	showHidden := r.URL.Query().Get("show_hidden") == "1"
	if _, ok := s.buildTagPageData(tag, isAdmin, showHidden); !ok {
		http.Error(w, "Tag not found", http.StatusNotFound)
		return
	}
	s.renderReactShell(w, "tag-detail", "tag", map[string]any{
		"tag":        tag,
		"showHidden": isAdmin && showHidden,
		"isAdmin":    isAdmin,
	})
}

func (s *Server) handleReaderNoPage(w http.ResponseWriter, r *http.Request) {
	titleID := chi.URLParam(r, "title")
	entryID := chi.URLParam(r, "entry")

	_, err := s.findEntry(titleID, entryID)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, s.appPath(fmt.Sprintf("reader/%s/%s/1", titleID, entryID)), http.StatusFound)
}

func (s *Server) handleReader(w http.ResponseWriter, r *http.Request) {
	titleID := chi.URLParam(r, "title")
	entryID := chi.URLParam(r, "entry")
	pageStr := chi.URLParam(r, "page")

	page := 1
	fmt.Sscanf(pageStr, "%d", &page)
	if page < 1 {
		page = 1
	}

	// Missing/corrupt entry surfaces as React error state after shell boot.
	s.renderReactShell(w, "reader", "reader", map[string]any{
		"isAdmin": GetIsAdmin(r),
		"tid":     titleID,
		"eid":     entryID,
		"page":    page,
	})
}

func (s *Server) handleAdmin(w http.ResponseWriter, r *http.Request) {
	s.renderReactShell(w, "admin", "admin", map[string]any{
		"isAdmin": true,
	})
}

func (s *Server) handleUserList(w http.ResponseWriter, r *http.Request) {
	s.renderReactShell(w, "user-list", "user-list", nil)
}

func (s *Server) handleUserEdit(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	extra := map[string]any{}
	if username != "" {
		extra["username"] = username
	}
	s.renderReactShell(w, "user-edit", "user-edit", extra)
}

func (s *Server) handleMissingItems(w http.ResponseWriter, r *http.Request) {
	s.renderReactShell(w, "missing-items", "missing-items", nil)
}

func (s *Server) renderReactShell(w http.ResponseWriter, pageID, pageName string, extra map[string]any) {
	boot := map[string]any{
		"baseUrl":  s.Deps.Config.BaseURL,
		"pageId":   pageID,
		"pageName": pageName,
		"isAdmin":  true,
		"version":  "2.0.0",
	}
	for k, v := range extra {
		boot[k] = v
	}
	raw, err := json.Marshal(boot)
	if err != nil {
		http.Error(w, "boot config error", http.StatusInternalServerError)
		return
	}
	s.renderPage(w, "views/react-shell", ReactShellData{
		BaseURL:  s.Deps.Config.BaseURL,
		PageName: pageName,
		BootJSON: template.JS(raw),
	})
}

func (s *Server) handleOPDSIndex(w http.ResponseWriter, r *http.Request) {
	lib := s.Deps.Library
	lib.RLock()
	titleIDs := make([]string, len(lib.TitleIDs))
	copy(titleIDs, lib.TitleIDs)
	lib.RUnlock()

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <id>urn:mango:index</id>
  <link rel="self" href="%sopds/" type="application/atom+xml;profile=opds-catalog;kind=navigation"/>
  <link rel="start" href="%sopds/" type="application/atom+xml;profile=opds-catalog;kind=navigation"/>
  <title>Mango Library</title>
  <author><name>Mango</name><uri>https://github.com/hkalexling/Mango</uri></author>`,
		s.Deps.Config.BaseURL, s.Deps.Config.BaseURL)

	for _, id := range titleIDs {
		lib.RLock()
		t, ok := lib.TitleHash[id]
		lib.RUnlock()
		if !ok {
			continue
		}
		fmt.Fprintf(w, `
  <entry>
    <title>%s</title>
    <id>urn:mango:%s</id>
    <link type="application/atom+xml;profile=opds-catalog;kind=navigation" rel="subsection" href="%sopds/book/%s"/>
  </entry>`,
			t.Name, t.ID, s.Deps.Config.BaseURL, t.ID)
	}

	fmt.Fprint(w, "\n</feed>")
}

func (s *Server) handleOPDSTitle(w http.ResponseWriter, r *http.Request) {
	titleID := chi.URLParam(r, "title_id")

	lib := s.Deps.Library
	lib.RLock()
	t, ok := lib.TitleHash[titleID]
	lib.RUnlock()

	if !ok {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <id>urn:mango:%s</id>
  <link rel="self" href="%sopds/book/%s" type="application/atom+xml;profile=opds-catalog;kind=navigation"/>
  <link rel="start" href="%sopds/" type="application/atom+xml;profile=opds-catalog;kind=navigation"/>
  <title>%s</title>
  <author><name>Mango</name><uri>https://github.com/hkalexling/Mango</uri></author>`,
		t.ID, s.Deps.Config.BaseURL, t.ID, s.Deps.Config.BaseURL, t.Name)

	lib.RLock()
	for _, subID := range t.TitleIDs {
		subT, subOk := lib.TitleHash[subID]
		if !subOk {
			continue
		}
		fmt.Fprintf(w, `
  <entry>
    <title>%s</title>
    <id>urn:mango:%s</id>
    <link type="application/atom+xml;profile=opds-catalog;kind=navigation" rel="subsection" href="%sopds/book/%s"/>
  </entry>`,
			subT.Name, subT.ID, s.Deps.Config.BaseURL, subT.ID)
	}

	for _, e := range t.Entries {
		fmt.Fprintf(w, `
  <entry>
    <title>%s</title>
    <id>urn:mango:%s</id>
    <link rel="http://opds-spec.org/image" href="%sapi/cover/%s/%s"/>
    <link rel="http://opds-spec.org/image/thumbnail" href="%sapi/cover/%s/%s"/>
    <link rel="http://opds-spec.org/acquisition" href="%sapi/download/%s/%s" type="application/vnd.comicbook+zip"/>
    <link type="text/html" rel="alternate" title="Read in Mango" href="%sreader/%s/%s/1"/>
    <link type="text/html" rel="alternate" title="Open in Mango" href="%sbook/%s"/>
  </entry>`,
			e.Name(), e.ID(),
			s.Deps.Config.BaseURL, t.ID, e.ID(),
			s.Deps.Config.BaseURL, t.ID, e.ID(),
			s.Deps.Config.BaseURL, t.ID, e.ID(),
			s.Deps.Config.BaseURL, t.ID, e.ID(),
			s.Deps.Config.BaseURL, t.ID)
	}
	lib.RUnlock()

	fmt.Fprint(w, "\n</feed>")
}
