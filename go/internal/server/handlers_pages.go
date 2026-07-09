package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	s.renderPage(w, "views/login", map[string]any{
		"BaseURL":  s.Deps.Config.BaseURL,
		"PageName": "login",
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	token, err := s.Deps.Storage.VerifyUser(username, password)
	if err != nil {
		log.Printf("Login error: %v", err)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if token == "" {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	SetAuthTokenCookie(w, s.Deps.Config, token)

	callback := r.FormValue("callback")
	if callback == "" {
		callback = "/"
	}
	http.Redirect(w, r, callback, http.StatusFound)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	ClearAuthTokenCookie(w, s.Deps.Config)
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	data := HomePageData{
		LayoutData: LayoutData{
			BaseURL:  s.Deps.Config.BaseURL,
			IsAdmin:  GetIsAdmin(r),
			PageName: "home",
			Version:  "0.1.0",
		},
		ConfigLibraryPath:  s.Deps.Config.LibraryPath,
		ConfigPath:         s.Deps.Config.DBPath,
		ScanIntervalMinutes: s.Deps.Config.ScanIntervalMinutes,
	}
	s.renderLayout(w, "home", data)
}

func (s *Server) handleLibrary(w http.ResponseWriter, r *http.Request) {
	isAdmin := GetIsAdmin(r)

	// Build library title list
	lib := s.Deps.Library
	lib.RLock()
	titleIDs := make([]string, len(lib.TitleIDs))
	copy(titleIDs, lib.TitleIDs)
	lib.RUnlock()

	var titles []LibraryTitle
	for _, id := range titleIDs {
		lib.RLock()
		t, ok := lib.TitleHash[id]
		lib.RUnlock()
		if !ok {
			continue
		}
		titles = append(titles, LibraryTitle{
			ID:         t.ID,
			Name:       t.Name,
			CoverURL:   fmt.Sprintf("%sapi/cover/%s/%s", s.Deps.Config.BaseURL, t.ID, firstEntryID(t)),
			EntryCount: len(t.DeepEntries()),
			Hidden:     false, // needs storage lookup
		})
	}

	data := LibraryPageData{
		LayoutData: LayoutData{
			BaseURL:  s.Deps.Config.BaseURL,
			IsAdmin:  isAdmin,
			PageName: "library",
			Version:  "0.1.0",
		},
		Titles:     titles,
		Percentage: make([]float64, len(titles)),
		ShowHidden: false,
	}
	s.renderLayout(w, "library", data)
}

func (s *Server) handleTitle(w http.ResponseWriter, r *http.Request) {
	ld := LayoutData{
		BaseURL:  s.Deps.Config.BaseURL,
		IsAdmin:  GetIsAdmin(r),
		PageName: "title",
		Version:  "0.1.0",
	}
	s.renderLayout(w, "title", ld)
}

func (s *Server) handleTags(w http.ResponseWriter, r *http.Request) {
	isAdmin := GetIsAdmin(r)

	tags, err := s.Deps.Storage.ListTags()
	if err != nil {
		http.Error(w, "Failed", http.StatusInternalServerError)
		return
	}

	var tagList []TagInfo
	for _, tag := range tags {
		titleIDs, err := s.Deps.Storage.GetTagTitles(tag, isAdmin)
		if err != nil {
			continue
		}
		tagList = append(tagList, TagInfo{
			Tag:        tag,
			EncodedTag: tag,
			Count:      len(titleIDs),
		})
	}

	data := TagsPageData{
		LayoutData: LayoutData{
			BaseURL:  s.Deps.Config.BaseURL,
			IsAdmin:  isAdmin,
			PageName: "tags",
			Version:  "0.1.0",
		},
		Tags: tagList,
	}

	s.renderLayout(w, "tags", data)
}

func (s *Server) handleTag(w http.ResponseWriter, r *http.Request) {
	ld := LayoutData{
		BaseURL:  s.Deps.Config.BaseURL,
		IsAdmin:  GetIsAdmin(r),
		PageName: "tag",
		Version:  "0.1.0",
	}
	s.renderLayout(w, "tag", ld)
}

func (s *Server) handleAPIDocs(w http.ResponseWriter, r *http.Request) {
	ld := LayoutData{
		BaseURL:  s.Deps.Config.BaseURL,
		IsAdmin:  GetIsAdmin(r),
		PageName: "api",
		Version:  "0.1.0",
	}
	s.renderPage(w, "views/api", ld)
}

func (s *Server) handleReaderNoPage(w http.ResponseWriter, r *http.Request) {
	titleID := chi.URLParam(r, "title")
	entryID := chi.URLParam(r, "entry")

	_, err := s.findEntry(titleID, entryID)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/reader/%s/%s/1",
		titleID, entryID), http.StatusFound)
}

func (s *Server) handleReader(w http.ResponseWriter, r *http.Request) {
	titleID := chi.URLParam(r, "title")
	entryID := chi.URLParam(r, "entry")
	pageStr := chi.URLParam(r, "page")

	page := 0
	fmt.Sscanf(pageStr, "%d", &page)

	entry, err := s.findEntry(titleID, entryID)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	lib := s.Deps.Library
	lib.RLock()
	t, ok := lib.TitleHash[titleID]
	lib.RUnlock()

	if !ok {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	var entries []EntryDetail
	for _, e := range t.Entries {
		entries = append(entries, EntryDetail{
			ID:   e.ID(),
			Name: e.Name(),
		})
	}

	var nextEntryURL, prevEntryURL string
	for i, e := range t.Entries {
		if e.ID() == entryID {
			if i+1 < len(t.Entries) {
				nextEntryURL = fmt.Sprintf("/reader/%s/%s/1", titleID, t.Entries[i+1].ID())
			}
			if i > 0 {
				prevEntryURL = fmt.Sprintf("/reader/%s/%s/1", titleID, t.Entries[i-1].ID())
			}
			break
		}
	}

	data := ReaderPageData{
		BaseURL:          s.Deps.Config.BaseURL,
		Title:            TitleDetail{ID: t.ID, Name: t.Name},
		Entry:            EntryDetail{ID: entryID, Name: entry.Name()},
		PageIdx:          page,
		Entries:          entries,
		ExitURL:          fmt.Sprintf("/book/%s", titleID),
		NextEntryURL:     nextEntryURL,
		PreviousEntryURL: prevEntryURL,
		Version:          "0.1.0",
	}

	s.renderPage(w, "views/reader", data)
}

func (s *Server) handlePluginDownload(w http.ResponseWriter, r *http.Request) {
	ld := LayoutData{
		BaseURL:    s.Deps.Config.BaseURL,
		IsAdmin:    GetIsAdmin(r),
		PageName:   "plugin-download",
		Version:    "0.1.0",
		PluginPath: s.Deps.Config.PluginPath,
	}
	s.renderLayout(w, "plugin-download", ld)
}

func (s *Server) handleAdmin(w http.ResponseWriter, r *http.Request) {
	ld := AdminPageData{
		LayoutData: LayoutData{
			BaseURL:  s.Deps.Config.BaseURL,
			IsAdmin:  true,
			PageName: "admin",
			Version:  "0.1.0",
		},
	}
	s.renderLayout(w, "admin", ld)
}

func (s *Server) handleUserList(w http.ResponseWriter, r *http.Request) {
	users, err := s.Deps.Storage.ListUsers()
	if err != nil {
		http.Error(w, "Failed", http.StatusInternalServerError)
		return
	}

	var userPairs [][2]string
	for _, u := range users {
		adminStr := "0"
		if u.IsAdmin {
			adminStr = "1"
		}
		userPairs = append(userPairs, [2]string{u.Username, adminStr})
	}

	data := UserPageData{
		LayoutData: LayoutData{
			BaseURL:  s.Deps.Config.BaseURL,
			IsAdmin:  true,
			PageName: "user",
			Version:  "0.1.0",
		},
		Users:    userPairs,
		Username: GetUsername(r),
	}

	s.renderLayout(w, "user", data)
}

func (s *Server) handleUserEdit(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")

	data := UserEditPageData{
		LayoutData: LayoutData{
			BaseURL:  s.Deps.Config.BaseURL,
			IsAdmin:  true,
			PageName: "user-edit",
			Version:  "0.1.0",
		},
		NewUser:  username == "",
		Username: username,
	}

	if username != "" {
		exists, err := s.Deps.Storage.UsernameExists(username)
		if err == nil && exists {
			isAdmin, _ := s.Deps.Storage.UsernameIsAdmin(username)
			data.Admin = isAdmin
		}
	}

	s.renderLayout(w, "user-edit", data)
}

func (s *Server) handleUserEditPost(w http.ResponseWriter, r *http.Request) {
	originalUsername := chi.URLParam(r, "original_username")
	username := r.FormValue("username")
	password := r.FormValue("password")
	admin := r.FormValue("admin") == "1"

	if originalUsername == "" {
		if err := s.Deps.Storage.NewUser(username, password, admin); err != nil {
			http.Redirect(w, r, "/admin/user/edit?error="+err.Error(), http.StatusFound)
			return
		}
	} else {
		if err := s.Deps.Storage.UpdateUser(originalUsername, username, password, admin); err != nil {
			http.Redirect(w, r, "/admin/user/edit?username="+originalUsername+"&error="+err.Error(), http.StatusFound)
			return
		}
	}

	http.Redirect(w, r, "/admin/user", http.StatusFound)
}

func (s *Server) handleDownloadManager(w http.ResponseWriter, r *http.Request) {
	ld := LayoutData{
		BaseURL:    s.Deps.Config.BaseURL,
		IsAdmin:    true,
		PageName:   "download-manager",
		Version:    "0.1.0",
		PluginPath: s.Deps.Config.PluginPath,
	}
	s.renderLayout(w, "download-manager", ld)
}

func (s *Server) handleSubscriptionManager(w http.ResponseWriter, r *http.Request) {
	ld := LayoutData{
		BaseURL:    s.Deps.Config.BaseURL,
		IsAdmin:    true,
		PageName:   "subscription-manager",
		Version:    "0.1.0",
		PluginPath: s.Deps.Config.PluginPath,
	}
	s.renderLayout(w, "subscription-manager", ld)
}

func (s *Server) handleMissingItems(w http.ResponseWriter, r *http.Request) {
	ld := LayoutData{
		BaseURL:  s.Deps.Config.BaseURL,
		IsAdmin:  true,
		PageName: "missing-items",
		Version:  "0.1.0",
	}
	s.renderLayout(w, "missing-items", ld)
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
