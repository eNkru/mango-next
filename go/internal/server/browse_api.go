package server

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/eNkru/mango-next/internal/library"
	"github.com/go-chi/chi/v5"
)

type browseTitle struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	FileName    string   `json:"file_name"`
	SortName    string   `json:"sort_name"`
	CoverURL    string   `json:"cover_url"`
	EntryCount  int      `json:"entry_count"`
	Progress    float64  `json:"progress"`
	ModifiedAt  int64    `json:"modified_at"`
	Hidden      bool     `json:"hidden"`
	Tags        []string `json:"tags"`
}

type browseEntry struct {
	ID         string  `json:"id"`
	TitleID    string  `json:"title_id"`
	Name       string  `json:"name"`
	FileName   string  `json:"file_name"`
	SortName   string  `json:"sort_name"`
	CoverURL   string  `json:"cover_url"`
	Pages      int     `json:"pages"`
	Page       int     `json:"page"`
	Progress   float64 `json:"progress"`
	ModifiedAt int64   `json:"modified_at"`
}

func (s *Server) apiHome(w http.ResponseWriter, r *http.Request) {
	username := GetUsername(r)
	lib := s.Deps.Library
	lib.RLock()
	empty := len(lib.TitleIDs) == 0
	lib.RUnlock()

	hasProgress, _ := s.Deps.Storage.UserHasProgress(username)
	continued, _ := s.Deps.Storage.GetContinueReading(username)
	started, _ := s.Deps.Storage.GetStartReading(username)
	recent, _ := s.Deps.Storage.GetRecentlyAdded(username)

	const limit = 8
	continueItems := make([]browseEntry, 0, limit)
	for _, item := range continued {
		if len(continueItems) == limit {
			break
		}
		if entry, ok := s.browseEntry(item.TitleID, item.EntryID, username); ok {
			continueItems = append(continueItems, entry)
		}
	}
	startItems := make([]browseTitle, 0, limit)
	for _, item := range started {
		if len(startItems) == limit {
			break
		}
		if title, ok := s.browseTitle(item.TitleID, username); ok {
			startItems = append(startItems, title)
		}
	}
	recentItems := make([]browseTitle, 0, limit)
	for _, item := range recent {
		if len(recentItems) == limit {
			break
		}
		if title, ok := s.browseTitle(item.TitleID, username); ok {
			recentItems = append(recentItems, title)
		}
	}

	sendJSON(w, map[string]any{
		"success":          true,
		"new_user":         !hasProgress,
		"empty_library":    empty,
		"is_admin":         GetIsAdmin(r),
		"library_path":     s.Deps.Config.LibraryPath,
		"scan_interval":    s.Deps.Config.ScanIntervalMinutes,
		"continue_reading": continueItems,
		"start_reading":    startItems,
		"recently_added":   recentItems,
	})
}

func (s *Server) apiBrowseLibrary(w http.ResponseWriter, r *http.Request) {
	username := GetUsername(r)
	isAdmin := GetIsAdmin(r)
	showHidden := isAdmin && r.URL.Query().Get("show_hidden") == "1"
	lib := s.Deps.Library
	lib.RLock()
	ids := append([]string(nil), lib.TitleIDs...)
	lib.RUnlock()

	titles := make([]browseTitle, 0, len(ids))
	for _, id := range ids {
		item, ok := s.browseTitle(id, username)
		if !ok || (item.Hidden && !showHidden) {
			continue
		}
		titles = append(titles, item)
	}
	sendJSON(w, map[string]any{
		"success":     true,
		"is_admin":    isAdmin,
		"show_hidden": showHidden,
		"titles":      titles,
		// Keep the old response key while legacy API consumers coexist.
		"data": titles,
	})
}

func (s *Server) apiBrowseBook(w http.ResponseWriter, r *http.Request) {
	tid := chi.URLParam(r, "tid")
	username := GetUsername(r)
	lib := s.Deps.Library
	lib.RLock()
	title := lib.TitleHash[tid]
	lib.RUnlock()
	if title == nil {
		sendJSONError(w, "Title not found", http.StatusNotFound)
		return
	}

	entries := make([]browseEntry, 0, len(title.Entries))
	for _, entry := range title.Entries {
		entries = append(entries, s.makeBrowseEntry(title, entry, username))
	}
	children := make([]browseTitle, 0, len(title.TitleIDs))
	for _, id := range title.TitleIDs {
		if child, ok := s.browseTitle(id, username); ok {
			children = append(children, child)
		}
	}
	tags, _ := s.Deps.Storage.GetTitleTags(tid)
	if tags == nil {
		tags = []string{}
	}
	current, _ := s.browseTitle(tid, username)
	parents := s.browseParents(title.ParentID)
	legacyEntries := make([]map[string]any, 0, len(entries))
	for _, entry := range entries {
		legacyEntries = append(legacyEntries, map[string]any{
			"id": entry.ID, "name": entry.Name, "pages": entry.Pages,
			"cover_url": entry.CoverURL, "progress": entry.Page,
		})
	}
	legacyTitles := make([]map[string]any, 0, len(title.TitleIDs))
	for _, childID := range title.TitleIDs {
		lib.RLock()
		child := lib.TitleHash[childID]
		lib.RUnlock()
		if child == nil {
			continue
		}
		childEntries := make([]map[string]any, 0, len(child.Entries))
		for _, entry := range child.Entries {
			item := s.makeBrowseEntry(child, entry, username)
			childEntries = append(childEntries, map[string]any{
				"id": item.ID, "name": item.Name, "pages": item.Pages,
				"cover_url": item.CoverURL, "progress": item.Page,
			})
		}
		legacyTitles = append(legacyTitles, map[string]any{
			"id": child.ID, "name": child.Name, "entries": childEntries,
			"cover_url": s.titleCoverURL(child),
		})
	}

	sendJSON(w, map[string]any{
		"success":  true,
		"is_admin": GetIsAdmin(r),
		"title":    current,
		"parents":  parents,
		"tags":     tags,
		"titles":   children,
		"entries":  entries,
		// Preserve the legacy top-level data envelope and keys.
		"data": map[string]any{
			"id": current.ID, "name": current.Name, "display_name": current.Name,
			"cover_url": current.CoverURL, "tags": tags, "hidden": current.Hidden,
			"titles": legacyTitles, "entries": legacyEntries,
		},
	})
}

func (s *Server) browseTitle(id, username string) (browseTitle, bool) {
	lib := s.Deps.Library
	lib.RLock()
	title := lib.TitleHash[id]
	lib.RUnlock()
	if title == nil {
		return browseTitle{}, false
	}
	hidden, _ := s.Deps.Storage.GetTitleHidden(id)
	sortName, _ := s.Deps.Storage.GetTitleSortTitle(id)
	tags, _ := s.Deps.Storage.GetTitleTags(id)
	return browseTitle{
		ID: id, Name: title.Name, DisplayName: title.Name, FileName: filepath.Base(title.Dir),
		SortName: stringValue(sortName), CoverURL: s.titleCoverURL(title),
		EntryCount: len(title.DeepEntries()), Progress: s.titleProgress(title, username),
		ModifiedAt: title.Mtime.Unix(), Hidden: hidden == 1, Tags: tags,
	}, true
}

func (s *Server) browseEntry(tid, eid, username string) (browseEntry, bool) {
	entry, err := s.findEntry(tid, eid)
	if err != nil || entry.Book() == nil {
		return browseEntry{}, false
	}
	return s.makeBrowseEntry(entry.Book(), entry, username), true
}

func (s *Server) makeBrowseEntry(title *library.Title, entry library.Entry, username string) browseEntry {
	page, _ := s.Deps.Storage.LoadProgress(username, title.ID, strPtr(entry.ID()))
	sortName, _ := s.Deps.Storage.GetEntrySortTitle(entry.ID())
	return browseEntry{
		ID: entry.ID(), TitleID: title.ID, Name: entry.Name(),
		FileName: entryFileName(entry), SortName: stringValue(sortName),
		CoverURL: s.entryCoverURL(title, entry), Pages: entry.PageCount(), Page: page,
		Progress: progressPercent(page, entry.PageCount()), ModifiedAt: entry.Mtime().Unix(),
	}
}

func (s *Server) titleProgress(title *library.Title, username string) float64 {
	var pages, read int
	for _, entry := range title.DeepEntries() {
		page, _ := s.Deps.Storage.LoadProgress(username, entry.Book().ID, strPtr(entry.ID()))
		pages += entry.PageCount()
		if page < 0 || page > entry.PageCount() {
			read += entry.PageCount()
		} else {
			read += page
		}
	}
	return progressPercent(read, pages)
}

func progressPercent(page, pages int) float64 {
	if pages <= 0 || page == 0 {
		return 0
	}
	if page < 0 || page >= pages {
		return 100
	}
	return float64(page) / float64(pages) * 100
}

func (s *Server) browseParents(parentID string) []browseTitle {
	reverse := make([]browseTitle, 0)
	for parentID != "" {
		item, ok := s.browseTitle(parentID, "")
		if !ok {
			break
		}
		reverse = append(reverse, item)
		s.Deps.Library.RLock()
		parent := s.Deps.Library.TitleHash[parentID]
		s.Deps.Library.RUnlock()
		if parent == nil {
			break
		}
		parentID = parent.ParentID
	}
	for left, right := 0, len(reverse)-1; left < right; left, right = left+1, right-1 {
		reverse[left], reverse[right] = reverse[right], reverse[left]
	}
	return reverse
}

func (s *Server) titleCoverURL(title *library.Title) string {
	if custom := title.CoverURL(); custom != "" {
		return s.assetURL(custom)
	}
	if eid := firstEntryID(title); eid != "" {
		return fmt.Sprintf("%sapi/cover/%s/%s", s.Deps.Config.BaseURL, title.ID, eid)
	}
	return ""
}

func (s *Server) entryCoverURL(title *library.Title, entry library.Entry) string {
	if custom := title.EntryCoverURL(entry); custom != "" {
		return s.assetURL(custom)
	}
	return fmt.Sprintf("%sapi/cover/%s/%s", s.Deps.Config.BaseURL, title.ID, entry.ID())
}

func (s *Server) assetURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	return strings.TrimSuffix(s.Deps.Config.BaseURL, "/") + "/" + strings.TrimPrefix(path, "/")
}

func entryFileName(entry library.Entry) string {
	name := filepath.Base(entry.Path())
	if _, ok := entry.(*library.ArchiveEntry); ok {
		name = strings.TrimSuffix(name, filepath.Ext(name))
	}
	return name
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
