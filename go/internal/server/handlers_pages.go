package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hkalexling/mango-go/internal/library"
	"github.com/hkalexling/mango-go/internal/storage"
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

// handleHome mirrors Crystal GET / (src/routes/main.cr):
//   new_user = !titles.any? { load_percentage(username) > 0 }
//   empty_library = titles.size == 0
//   plus continue_reading / start_reading / recently_added sections.
func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	username := GetUsername(r)
	base := s.Deps.Config.BaseURL
	lib := s.Deps.Library
	st := s.Deps.Storage

	lib.RLock()
	titleCount := len(lib.TitleIDs)
	lib.RUnlock()
	emptyLibrary := titleCount == 0

	// Crystal: new_user if no title has load_percentage > 0 for this user.
	hasProgress, err := st.UserHasProgress(username)
	if err != nil {
		log.Printf("home UserHasProgress: %v", err)
	}
	newUser := !hasProgress

	continueReading, err := st.GetContinueReading(username)
	if err != nil {
		log.Printf("home GetContinueReading: %v", err)
		continueReading = nil
	}
	continueReading = s.enrichContinueReading(base, continueReading)

	startReading, err := st.GetStartReading(username)
	if err != nil {
		log.Printf("home GetStartReading: %v", err)
		startReading = nil
	}
	startReading = s.enrichStartReading(base, startReading)

	recentlyAdded, err := st.GetRecentlyAdded(username)
	if err != nil {
		log.Printf("home GetRecentlyAdded: %v", err)
		recentlyAdded = nil
	}
	recentlyAdded = s.enrichRecentlyAdded(base, recentlyAdded)

	// Cap sections (Crystal ENTRIES_IN_HOME_SECTIONS = 8)
	const homeLimit = 8
	if len(continueReading) > homeLimit {
		continueReading = continueReading[:homeLimit]
	}
	if len(startReading) > homeLimit {
		startReading = startReading[:homeLimit]
	}
	if len(recentlyAdded) > homeLimit {
		recentlyAdded = recentlyAdded[:homeLimit]
	}

	data := HomePageData{
		LayoutData: LayoutData{
			BaseURL:  base,
			IsAdmin:  GetIsAdmin(r),
			PageName: "home",
			Version:  "2.0.0",
		},
		ContinueReading:     continueReading,
		StartReading:        startReading,
		RecentlyAdded:       recentlyAdded,
		NewUser:             newUser,
		EmptyLibrary:        emptyLibrary,
		ConfigLibraryPath:   s.Deps.Config.LibraryPath,
		ConfigPath:          s.Deps.Config.DBPath,
		ScanIntervalMinutes: s.Deps.Config.ScanIntervalMinutes,
	}
	s.renderLayout(w, "home", data)
}

func (s *Server) enrichContinueReading(base string, items []storage.ContinueReadingItem) []storage.ContinueReadingItem {
	lib := s.Deps.Library
	out := make([]storage.ContinueReadingItem, 0, len(items))
	for _, it := range items {
		lib.RLock()
		t, ok := lib.TitleHash[it.TitleID]
		lib.RUnlock()
		if !ok || t == nil {
			continue
		}
		it.TitleName = t.Name
		coverEID := firstEntryID(t)
		if it.EntryID != "" {
			if e := library.EntryByID(t, it.EntryID); e != nil {
				it.EntryName = e.Name()
				coverEID = e.ID()
				if pages := e.PageCount(); pages > 0 && it.Page > 0 {
					pct := float64(it.Page) / float64(pages) * 100
					if pct > 100 {
						pct = 100
					}
					it.Percentage = pct
				}
			}
		}
		if it.EntryName == "" {
			it.EntryName = t.Name
		}
		if coverEID != "" {
			it.CoverURL = fmt.Sprintf("%sapi/cover/%s/%s", base, t.ID, coverEID)
		}
		out = append(out, it)
	}
	return out
}

func (s *Server) enrichStartReading(base string, items []storage.StartReadingItem) []storage.StartReadingItem {
	lib := s.Deps.Library
	out := make([]storage.StartReadingItem, 0, len(items))
	for _, it := range items {
		lib.RLock()
		t, ok := lib.TitleHash[it.TitleID]
		lib.RUnlock()
		if !ok || t == nil {
			continue
		}
		it.TitleName = t.Name
		if eid := firstEntryID(t); eid != "" {
			it.CoverURL = fmt.Sprintf("%sapi/cover/%s/%s", base, t.ID, eid)
		}
		out = append(out, it)
	}
	return out
}

func (s *Server) enrichRecentlyAdded(base string, items []storage.RecentlyAddedItem) []storage.RecentlyAddedItem {
	lib := s.Deps.Library
	out := make([]storage.RecentlyAddedItem, 0, len(items))
	for _, it := range items {
		lib.RLock()
		t, ok := lib.TitleHash[it.TitleID]
		lib.RUnlock()
		if !ok || t == nil {
			continue
		}
		it.TitleName = t.Name
		it.EntryName = t.Name
		if len(t.Entries) > 0 {
			it.EntryID = t.Entries[0].ID()
			it.EntryName = t.Entries[0].Name()
		}
		if eid := firstEntryID(t); eid != "" {
			it.CoverURL = fmt.Sprintf("%sapi/cover/%s/%s", base, t.ID, eid)
		}
		out = append(out, it)
	}
	return out
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
			BaseURL:             s.Deps.Config.BaseURL,
			IsAdmin:             isAdmin,
			PageName:            "library",
			Version:             "2.0.0",
		},
		Titles:     titles,
		Percentage: make([]float64, len(titles)),
		ShowHidden: false,
	}
	s.renderLayout(w, "library", data)
}

func (s *Server) handleTitle(w http.ResponseWriter, r *http.Request) {
	titleID := chi.URLParam(r, "title")
	username := GetUsername(r)
	lib := s.Deps.Library

	lib.RLock()
	t, ok := lib.TitleHash[titleID]
	lib.RUnlock()

	if !ok {
		http.Error(w, "Title not found", http.StatusNotFound)
		return
	}

	// Build sorted sub-titles
	var sortedTitles []TitleDetail
	lib.RLock()
	for _, subID := range t.TitleIDs {
		subT, subOk := lib.TitleHash[subID]
		if !subOk {
			continue
		}
		sortedTitles = append(sortedTitles, TitleDetail{
			ID:       subT.ID,
			Name:     subT.Name,
			CoverURL: fmt.Sprintf("%sapi/cover/%s/%s", s.Deps.Config.BaseURL, subT.ID, firstEntryID(subT)),
		})
	}

	// Build entries
	var entries []EntryDetail
	for _, e := range t.Entries {
		entries = append(entries, EntryDetail{
			ID:        e.ID(),
			Name:      e.Name(),
			PageCount: e.PageCount(),
			CoverURL:  fmt.Sprintf("%sapi/cover/%s/%s", s.Deps.Config.BaseURL, t.ID, e.ID()),
		})
	}

	// Build parent IDs for breadcrumb
	var parentIDs []string
	if t.ParentID != "" {
		parentIDs = append(parentIDs, t.ParentID)
	}

	// Compute percentages
	percentage := make([]float64, len(entries))
	for i, entry := range entries {
		prog, _ := s.Deps.Storage.LoadProgress(username, t.ID, strPtr(entry.ID))
		if entry.PageCount > 0 {
			percentage[i] = float64(prog) / float64(entry.PageCount) * 100
		}
	}

	titlePercentage := make([]float64, len(sortedTitles))
	for i, st := range sortedTitles {
		subT, subOk := lib.TitleHash[st.ID]
		if !subOk {
			continue
		}
		prog, _ := s.Deps.Storage.LoadProgress(username, st.ID, nil)
		totalPages := 0
		for _, e := range subT.DeepEntries() {
			totalPages += e.PageCount()
		}
		if totalPages > 0 {
			titlePercentage[i] = float64(prog) / float64(totalPages) * 100
		}
	}
	lib.RUnlock()

	hidden, _ := s.Deps.Storage.GetTitleHidden(t.ID)

	data := TitlePageData{
		LayoutData: LayoutData{
			BaseURL:  s.Deps.Config.BaseURL,
			IsAdmin:  GetIsAdmin(r),
			PageName: "title",
			Version:  "2.0.0",
		},
		Title: TitleDetail{
			ID:        t.ID,
			Name:      t.Name,
			CoverURL:  fmt.Sprintf("%sapi/cover/%s/%s", s.Deps.Config.BaseURL, t.ID, firstEntryID(t)),
			ParentIDs: parentIDs,
			Hidden:    hidden == 1,
		},
		SortedTitles:    sortedTitles,
		Entries:         entries,
		Percentage:      percentage,
		TitlePercentage: titlePercentage,
		IsHidden:        hidden == 1,
	}

	s.renderLayout(w, "title", data)
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
			BaseURL:             s.Deps.Config.BaseURL,
			IsAdmin:             isAdmin,
			PageName:            "tags",
			Version:             "2.0.0",
		},
		Tags: tagList,
	}

	s.renderLayout(w, "tags", data)
}

func (s *Server) handleTag(w http.ResponseWriter, r *http.Request) {
	tag := chi.URLParam(r, "tag")
	showHidden := r.URL.Query().Get("show_hidden") == "1"

	titleIDs, err := s.Deps.Storage.GetTagTitles(tag, showHidden)
	if err != nil || len(titleIDs) == 0 {
		http.Error(w, "Tag not found", http.StatusNotFound)
		return
	}

	lib := s.Deps.Library
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
		})
	}
	lib.RUnlock()

	data := TagPageData{
		LayoutData: LayoutData{
			BaseURL:  s.Deps.Config.BaseURL,
			IsAdmin:  GetIsAdmin(r),
			PageName: "tag",
			Version:  "2.0.0",
		},
		Tag:        tag,
		Titles:     titles,
		ShowHidden: showHidden,
	}

	s.renderLayout(w, "tag", data)
}

func (s *Server) handleAPIDocs(w http.ResponseWriter, r *http.Request) {
	ld := LayoutData{
		BaseURL:  s.Deps.Config.BaseURL,
		IsAdmin:  GetIsAdmin(r),
		PageName: "api",
		Version:  "2.0.0",
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
		PageName:         "reader",
		Title:            TitleDetail{ID: t.ID, Name: t.Name},
		Entry:            EntryDetail{ID: entryID, Name: entry.Name()},
		PageIdx:          page,
		Entries:          entries,
		ExitURL:          fmt.Sprintf("/book/%s", titleID),
		NextEntryURL:     nextEntryURL,
		PreviousEntryURL: prevEntryURL,
		Version:          "2.0.0",
	}

	s.renderPage(w, "views/reader", data)
}

func (s *Server) handlePluginDownload(w http.ResponseWriter, r *http.Request) {
	ld := LayoutData{
		BaseURL:             s.Deps.Config.BaseURL,
		IsAdmin:             GetIsAdmin(r),
		PageName:            "plugin-download",
		Version:             "2.0.0",
		PluginPath:          s.Deps.Config.PluginPath,
	}
	s.renderLayout(w, "plugin-download", ld)
}

func (s *Server) handleAdmin(w http.ResponseWriter, r *http.Request) {
	ld := AdminPageData{
		LayoutData: LayoutData{
			BaseURL:             s.Deps.Config.BaseURL,
			IsAdmin:             true,
			PageName:            "admin",
			Version:             "2.0.0",
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
			BaseURL:             s.Deps.Config.BaseURL,
			IsAdmin:             true,
			PageName:            "user",
			Version:             "2.0.0",
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
			BaseURL:             s.Deps.Config.BaseURL,
			IsAdmin:             true,
			PageName:            "user-edit",
			Version:             "2.0.0",
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
		BaseURL:             s.Deps.Config.BaseURL,
		IsAdmin:             true,
		PageName:            "download-manager",
		Version:             "2.0.0",
		PluginPath:          s.Deps.Config.PluginPath,
	}
	s.renderLayout(w, "download-manager", ld)
}

func (s *Server) handleSubscriptionManager(w http.ResponseWriter, r *http.Request) {
	ld := LayoutData{
		BaseURL:             s.Deps.Config.BaseURL,
		IsAdmin:             true,
		PageName:            "subscription-manager",
		Version:             "2.0.0",
		PluginPath:          s.Deps.Config.PluginPath,
	}
	s.renderLayout(w, "subscription-manager", ld)
}

func (s *Server) handleMissingItems(w http.ResponseWriter, r *http.Request) {
	ld := LayoutData{
		BaseURL:             s.Deps.Config.BaseURL,
		IsAdmin:             true,
		PageName:            "missing-items",
		Version:             "2.0.0",
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
