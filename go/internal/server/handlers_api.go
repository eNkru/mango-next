package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/eNkru/mango-next/internal/config"
	"github.com/eNkru/mango-next/internal/library"
	"github.com/eNkru/mango-next/internal/plugin"
	"github.com/eNkru/mango-next/internal/queue"
	"github.com/eNkru/mango-next/internal/storage"
	"github.com/eNkru/mango-next/internal/upload"
	"github.com/go-chi/chi/v5"
)

// apiLogin mirrors Crystal POST /api/login (src/routes/api.cr).
// Unauthenticated; returns token as session_id and sets auth cookie.
func (s *Server) apiLogin(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, loginBodyLimit)
	if !loginAllowed(r) {
		w.WriteHeader(http.StatusForbidden)
		sendJSON(w, map[string]any{"success": false, "error": "login failed"})
		return
	}
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		recordLoginFailure(r)
		w.WriteHeader(http.StatusForbidden)
		sendJSON(w, map[string]any{"success": false, "error": "login failed"})
		return
	}
	token, err := s.Deps.Storage.VerifyUser(body.Username, body.Password)
	if err != nil || token == "" {
		if err != nil {
			log.Printf("API login error: %v", err)
		}
		recordLoginFailure(r)
		w.WriteHeader(http.StatusForbidden)
		sendJSON(w, map[string]any{"success": false, "error": "login failed"})
		return
	}
	clearLoginFailures(r)
	SetAuthTokenCookie(w, r, s.Deps.Config, token)
	isAdmin, _ := s.Deps.Storage.UsernameIsAdmin(body.Username)
	sendJSON(w, map[string]any{
		"success":    true,
		"session_id": token,
		"is_admin":   isAdmin,
	})
}

func (s *Server) apiLibrary(w http.ResponseWriter, r *http.Request) {
	username := GetUsername(r)
	lib := s.Deps.Library

	lib.RLock()
	titleIDs := make([]string, len(lib.TitleIDs))
	copy(titleIDs, lib.TitleIDs)
	lib.RUnlock()

	type titleResp struct {
		ID          string   `json:"id"`
		Name        string   `json:"name"`
		CoverURL    string   `json:"cover_url"`
		EntryCount  int      `json:"entry_count"`
		Tags        []string `json:"tags,omitempty"`
		Hidden      bool     `json:"hidden"`
		DisplayName string   `json:"display_name"`
	}

	var resp []titleResp
	for _, id := range titleIDs {
		lib.RLock()
		t, ok := lib.TitleHash[id]
		lib.RUnlock()
		if !ok {
			continue
		}
		st, _ := s.Deps.Storage.GetTitleHidden(t.ID)
		tags, _ := s.Deps.Storage.GetTitleTags(t.ID)
		_ = username
		resp = append(resp, titleResp{
			ID:   t.ID,
			Name: t.Name,
			CoverURL: fmt.Sprintf("%sapi/cover/%s/%s",
				s.Deps.Config.BaseURL, t.ID, firstEntryID(t)),
			EntryCount:  len(t.Entries),
			Tags:        tags,
			Hidden:      st == 1,
			DisplayName: t.Name,
		})
	}

	sendJSON(w, map[string]any{
		"success": true,
		"data":    resp,
	})
}

func firstEntryID(t *library.Title) string {
	// Prefer any deep entry (including nested volumes). Nested-only trees have
	// empty direct Entries; returning a sub-title ID here breaks /api/cover.
	for _, e := range t.DeepEntries() {
		return e.ID()
	}
	return ""
}

func countEntries(t *library.Title) int {
	return len(t.DeepEntries())
}

func (s *Server) apiBook(w http.ResponseWriter, r *http.Request) {
	tid := chi.URLParam(r, "tid")
	lib := s.Deps.Library

	lib.RLock()
	t, ok := lib.TitleHash[tid]
	lib.RUnlock()

	if !ok {
		sendJSONError(w, "Title not found", http.StatusNotFound)
		return
	}

	type entryResp struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Pages    int    `json:"pages"`
		CoverURL string `json:"cover_url"`
		Progress int    `json:"progress"`
	}

	var entries []entryResp
	username := GetUsername(r)
	for _, e := range t.Entries {
		progress, _ := s.Deps.Storage.LoadProgress(username, t.ID, strPtr(e.ID()))
		entries = append(entries, entryResp{
			ID:       e.ID(),
			Name:     e.Name(),
			Pages:    e.PageCount(),
			CoverURL: fmt.Sprintf("%sapi/cover/%s/%s", s.Deps.Config.BaseURL, t.ID, e.ID()),
			Progress: progress,
		})
	}

	type subTitleResp struct {
		ID       string      `json:"id"`
		Name     string      `json:"name"`
		Entries  []entryResp `json:"entries"`
		CoverURL string      `json:"cover_url"`
	}

	var subTitles []subTitleResp
	lib.RLock()
	for _, subID := range t.TitleIDs {
		subT, subOk := lib.TitleHash[subID]
		if !subOk {
			continue
		}
		var subEntries []entryResp
		for _, e := range subT.Entries {
			progress, _ := s.Deps.Storage.LoadProgress(username, subT.ID, strPtr(e.ID()))
			subEntries = append(subEntries, entryResp{
				ID:       e.ID(),
				Name:     e.Name(),
				Pages:    e.PageCount(),
				CoverURL: fmt.Sprintf("%sapi/cover/%s/%s", s.Deps.Config.BaseURL, subT.ID, e.ID()),
				Progress: progress,
			})
		}
		subTitles = append(subTitles, subTitleResp{
			ID:      subT.ID,
			Name:    subT.Name,
			Entries: subEntries,
		})
	}
	lib.RUnlock()

	tags, _ := s.Deps.Storage.GetTitleTags(tid)
	hidden, _ := s.Deps.Storage.GetTitleHidden(tid)

	sendJSON(w, map[string]any{
		"success": true,
		"data": map[string]any{
			"id":           t.ID,
			"name":         t.Name,
			"display_name": t.Name,
			"cover_url":    fmt.Sprintf("%sapi/cover/%s/%s", s.Deps.Config.BaseURL, t.ID, firstEntryID(t)),
			"tags":         tags,
			"hidden":       hidden == 1,
			"entries":      entries,
			"titles":       subTitles,
		},
	})
}

func (s *Server) apiPage(w http.ResponseWriter, r *http.Request) {
	tid := chi.URLParam(r, "tid")
	eid := chi.URLParam(r, "eid")
	pageStr := chi.URLParam(r, "page")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 0 {
		sendJSONError(w, "Invalid page number", http.StatusBadRequest)
		return
	}

	entry, err := s.findEntry(tid, eid)
	if err != nil {
		sendJSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	if page >= entry.PageCount() {
		sendJSONError(w, "Page out of range", http.StatusBadRequest)
		return
	}

	img, err := entry.ReadPage(page)
	if err != nil {
		log.Printf("Read page error: %v", err)
		sendJSONError(w, "Failed to read page", http.StatusInternalServerError)
		return
	}

	sendImage(w, img.Data, img.Mime)
}

func (s *Server) apiCover(w http.ResponseWriter, r *http.Request) {
	eid := chi.URLParam(r, "eid")

	img, err := s.Deps.Storage.GetThumbnail(eid)
	if err != nil {
		log.Printf("Get thumbnail error: %v", err)
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if img == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	sendImage(w, img.Data, img.Mime)
}

func (s *Server) apiSaveProgress(w http.ResponseWriter, r *http.Request) {
	username := GetUsername(r)
	tid := chi.URLParam(r, "tid")
	pageStr := chi.URLParam(r, "page")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		sendJSONError(w, "Invalid page", http.StatusBadRequest)
		return
	}

	eid := r.URL.Query().Get("eid")
	var entryID *string
	if eid != "" {
		entryID = &eid
	}

	if err := s.Deps.Storage.SaveProgress(username, tid, entryID, page); err != nil {
		log.Printf("Save progress error: %v", err)
		sendJSONError(w, "Failed to save progress", http.StatusInternalServerError)
		return
	}

	sendJSON(w, map[string]any{"success": true})
}

func (s *Server) apiBulkProgress(w http.ResponseWriter, r *http.Request) {
	username := GetUsername(r)
	action := chi.URLParam(r, "action")
	tid := chi.URLParam(r, "tid")

	var body struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	lib := s.Deps.Library
	lib.RLock()
	_, ok := lib.TitleHash[tid]
	lib.RUnlock()

	if !ok {
		sendJSONError(w, "Title not found", http.StatusNotFound)
		return
	}

	switch action {
	case "read":
		if err := s.Deps.Storage.BulkMarkRead(username, tid, body.IDs); err != nil {
			log.Printf("Bulk mark read error: %v", err)
			sendJSONError(w, "Failed", http.StatusInternalServerError)
			return
		}
	case "unread":
		if err := s.Deps.Storage.BulkMarkUnread(username, tid, body.IDs); err != nil {
			log.Printf("Bulk mark unread error: %v", err)
			sendJSONError(w, "Failed", http.StatusInternalServerError)
			return
		}
	default:
		sendJSONError(w, "Unknown action", http.StatusBadRequest)
		return
	}

	sendJSON(w, map[string]any{"success": true})
}

func (s *Server) apiDimensions(w http.ResponseWriter, r *http.Request) {
	tid := chi.URLParam(r, "tid")
	eid := chi.URLParam(r, "eid")

	entry, err := s.findEntry(tid, eid)
	if err != nil {
		sendJSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	sig := strconv.FormatUint(entry.Signature(), 10)
	if cached, ok, err := s.Deps.Storage.GetEntryDimensions(entry.ID(), sig); err == nil && ok {
		sendJSON(w, map[string]any{"success": true, "dimensions": cached})
		return
	}

	pageDims, err := library.ReadPageDimensions(entry)
	if err != nil {
		log.Printf("Read page dimensions error: %v", err)
		sendJSONError(w, "Failed to read dimensions", http.StatusInternalServerError)
		return
	}

	stored := make([]storage.PageDimension, len(pageDims))
	out := make([]map[string]any, len(pageDims))
	for i, d := range pageDims {
		stored[i] = storage.PageDimension{Width: d.Width, Height: d.Height}
		out[i] = map[string]any{"width": d.Width, "height": d.Height}
	}
	if err := s.Deps.Storage.SaveEntryDimensions(entry.ID(), sig, stored); err != nil {
		log.Printf("Save entry dimensions error: %v", err)
	}

	sendJSON(w, map[string]any{"success": true, "dimensions": out})
}

func (s *Server) apiDownload(w http.ResponseWriter, r *http.Request) {
	tid := chi.URLParam(r, "tid")
	eid := chi.URLParam(r, "eid")

	entry, err := s.findEntry(tid, eid)
	if err != nil {
		sendJSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	img, err := entry.ReadPage(0)
	if err != nil {
		sendJSONError(w, "Failed to read", http.StatusInternalServerError)
		return
	}

	sendAttachment(w, img.Data, entry.Name()+".cbz", "application/vnd.comicbook+zip")
}

func (s *Server) apiContinueReading(w http.ResponseWriter, r *http.Request) {
	username := GetUsername(r)
	items, err := s.Deps.Storage.GetContinueReading(username)
	if err != nil {
		log.Printf("Continue reading error: %v", err)
		sendJSON(w, map[string]any{"success": true, "data": []any{}})
		return
	}

	type itemResp struct {
		EntryID    string  `json:"entry_id"`
		TitleID    string  `json:"title_id"`
		Percentage float64 `json:"percentage"`
	}

	var resp []itemResp
	for _, item := range items {
		resp = append(resp, itemResp{
			EntryID:    item.EntryID,
			TitleID:    item.TitleID,
			Percentage: item.Percentage,
		})
	}

	sendJSON(w, map[string]any{"success": true, "data": resp})
}

func (s *Server) apiStartReading(w http.ResponseWriter, r *http.Request) {
	username := GetUsername(r)
	items, err := s.Deps.Storage.GetStartReading(username)
	if err != nil {
		log.Printf("Start reading error: %v", err)
		sendJSON(w, map[string]any{"success": true, "data": []any{}})
		return
	}

	type itemResp struct {
		TitleID string `json:"title_id"`
	}

	var resp []itemResp
	for _, item := range items {
		resp = append(resp, itemResp{TitleID: item.TitleID})
	}

	sendJSON(w, map[string]any{"success": true, "data": resp})
}

func (s *Server) apiRecentlyAdded(w http.ResponseWriter, r *http.Request) {
	items, err := s.Deps.Storage.GetRecentlyAdded("")
	if err != nil {
		log.Printf("Recently added error: %v", err)
		sendJSON(w, map[string]any{"success": true, "data": []any{}})
		return
	}

	type itemResp struct {
		TitleID string `json:"title_id"`
	}

	var resp []itemResp
	for _, item := range items {
		resp = append(resp, itemResp{TitleID: item.TitleID})
	}

	sendJSON(w, map[string]any{"success": true, "data": resp})
}

func (s *Server) apiGetSortOpt(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, map[string]any{
		"success": true,
		"data":    map[string]any{"method": "name", "ascend": true},
	})
}

func (s *Server) apiPutSortOpt(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, map[string]any{"success": true})
}

func (s *Server) apiGetTitleTags(w http.ResponseWriter, r *http.Request) {
	tid := chi.URLParam(r, "tid")
	tags, err := s.Deps.Storage.GetTitleTags(tid)
	if err != nil {
		sendJSONError(w, "Failed", http.StatusInternalServerError)
		return
	}
	if tags == nil {
		tags = []string{}
	}
	sendJSON(w, map[string]any{"success": true, "data": tags})
}

func (s *Server) apiListTags(w http.ResponseWriter, r *http.Request) {
	tags, err := s.Deps.Storage.ListTags()
	if err != nil {
		sendJSONError(w, "Failed", http.StatusInternalServerError)
		return
	}
	if tags == nil {
		tags = []string{}
	}
	sendJSON(w, map[string]any{"success": true, "data": tags})
}

func (s *Server) apiAdminScan(w http.ResponseWriter, r *http.Request) {
	go func() {
		if _, err := s.Deps.Library.Scan(); err != nil {
			log.Printf("Scan error: %v", err)
		}
	}()
	sendJSON(w, map[string]any{"success": true})
}

func (s *Server) apiAdminThumbnailProgress(w http.ResponseWriter, r *http.Request) {
	lib := s.Deps.Library
	progress, running := lib.ThumbnailStatus()
	sendJSON(w, map[string]any{
		"success":  true,
		"progress": progress,
		"running":  running,
	})
}

func (s *Server) apiAdminGenerateThumbnails(w http.ResponseWriter, r *http.Request) {
	go func() {
		if err := s.Deps.Library.GenerateThumbnails(); err != nil {
			log.Printf("Thumbnail generation error: %v", err)
		}
	}()
	sendJSON(w, map[string]any{"success": true})
}

func (s *Server) apiAdminListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.Deps.Storage.ListUsers()
	if err != nil {
		sendJSONError(w, "Failed to list users", http.StatusInternalServerError)
		return
	}
	out := make([]map[string]any, 0, len(users))
	for _, u := range users {
		out = append(out, map[string]any{
			"username": u.Username,
			"admin":    u.IsAdmin,
		})
	}
	sendJSON(w, map[string]any{
		"success":          true,
		"users":            out,
		"current_username": GetUsername(r),
	})
}

func (s *Server) apiAdminCreateUser(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Admin    bool   `json:"admin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := s.Deps.Storage.NewUser(body.Username, body.Password, body.Admin); err != nil {
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	sendJSON(w, map[string]any{"success": true})
}

func (s *Server) apiAdminUpdateUser(w http.ResponseWriter, r *http.Request) {
	original := chi.URLParam(r, "username")
	if original == "" {
		sendJSONError(w, "Missing username", http.StatusBadRequest)
		return
	}
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Admin    bool   `json:"admin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if body.Username == "" {
		body.Username = original
	}
	if err := s.Deps.Storage.UpdateUser(original, body.Username, body.Password, body.Admin); err != nil {
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	sendJSON(w, map[string]any{"success": true})
}

func (s *Server) apiAdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	currentUser := GetUsername(r)

	if username == currentUser {
		sendJSONError(w, "Cannot delete yourself", http.StatusBadRequest)
		return
	}

	if err := s.Deps.Storage.DeleteUser(username); err != nil {
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	sendJSON(w, map[string]any{"success": true})
}

func (s *Server) apiAdminSetDisplayName(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, map[string]any{"success": true})
}

func (s *Server) apiAdminSetSortTitle(w http.ResponseWriter, r *http.Request) {
	tid := chi.URLParam(r, "tid")
	var body struct {
		SortTitle string `json:"sort_title"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	st := body.SortTitle
	if st == "" {
		if err := s.Deps.Storage.SetTitleSortTitle(tid, nil); err != nil {
			sendJSONError(w, "Failed", http.StatusInternalServerError)
			return
		}
	} else {
		if err := s.Deps.Storage.SetTitleSortTitle(tid, &st); err != nil {
			sendJSONError(w, "Failed", http.StatusInternalServerError)
			return
		}
	}
	sendJSON(w, map[string]any{"success": true})
}

// apiAdminUpload mirrors Crystal POST /api/admin/upload/:target (src/routes/api.cr).
// Failures return HTTP 200 + {"success":false,"error":...} like Crystal send_json rescue.
func (s *Server) apiAdminUpload(w http.ResponseWriter, r *http.Request) {
	fail := func(msg string) {
		log.Printf("apiAdminUpload: %s", msg)
		sendJSON(w, map[string]any{"success": false, "error": msg})
	}

	target := chi.URLParam(r, "target")
	r.Body = http.MaxBytesReader(w, r.Body, uploadBodyLimit)
	if err := r.ParseMultipartForm(uploadBodyLimit); err != nil {
		fail("No file uploaded")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		fail("No part with name `file` found")
		return
	}
	defer file.Close()

	if header.Filename == "" {
		fail("No file uploaded")
		return
	}

	switch target {
	case "cover":
		tid := r.URL.Query().Get("tid")
		if tid == "" {
			fail("Title not found")
			return
		}
		eid := r.URL.Query().Get("eid")

		mimeType := upload.MIMEFromFilename(header.Filename)
		if !upload.IsSupportedImageMIME(mimeType) {
			fail("The uploaded image must be either JPEG or PNG")
			return
		}

		lib := s.Deps.Library
		lib.RLock()
		t, ok := lib.TitleHash[tid]
		lib.RUnlock()
		if !ok || t == nil {
			fail("Title not found")
			return
		}

		cfg := s.Deps.Config
		if cfg == nil {
			cfg = config.Current()
		}
		up, err := upload.New(cfg.UploadPath)
		if err != nil {
			fail(err.Error())
			return
		}
		ext := filepath.Ext(header.Filename)
		saved, err := up.Save("img", ext, file)
		if err != nil {
			fail(err.Error())
			return
		}
		urlPath, ok := up.PathToURL(saved)
		if !ok {
			fail("Failed to generate a public URL for the uploaded file")
			return
		}

		if eid == "" {
			if err := t.SetCoverURL(urlPath); err != nil {
				fail(err.Error())
				return
			}
		} else {
			entry := library.EntryByID(t, eid)
			if entry == nil {
				fail("Entry not found")
				return
			}
			if err := t.SetEntryCoverURL(entry.Name(), urlPath); err != nil {
				fail(err.Error())
				return
			}
		}
		sendJSON(w, map[string]any{"success": true})
	default:
		// Crystal typo "Unkown" preserved for behavior parity in message prefix.
		fail(fmt.Sprintf("Unkown upload target %s", target))
	}
}

func (s *Server) apiAdminListPlugins(w http.ResponseWriter, r *http.Request) {
	cfg := config.Current()
	pluginDir := cfg.PluginPath
	if pluginDir == "" {
		sendJSON(w, map[string]any{"success": true, "data": []any{}})
		return
	}

	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		sendJSON(w, map[string]any{"success": true, "data": []any{}})
		return
	}

	type pluginResp struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}

	var plugins []pluginResp
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		p, err := plugin.LoadPlugin(pluginDir, e.Name())
		if err != nil {
			continue
		}
		info := p.Info()
		plugins = append(plugins, pluginResp{
			ID:    info.ID,
			Title: info.Title,
		})
	}
	if plugins == nil {
		plugins = []pluginResp{}
	}
	sendJSON(w, map[string]any{"success": true, "data": plugins})
}

func (s *Server) apiAdminPluginInfo(w http.ResponseWriter, r *http.Request) {
	pluginID := r.URL.Query().Get("plugin")
	cfg := config.Current()
	p, err := plugin.LoadPlugin(cfg.PluginPath, pluginID)
	if err != nil {
		sendJSON(w, map[string]any{"success": true, "data": map[string]any{}})
		return
	}
	sendJSON(w, map[string]any{"success": true, "data": p.Info()})
}

func (s *Server) apiAdminPluginSearch(w http.ResponseWriter, r *http.Request) {
	pluginID := r.URL.Query().Get("plugin")
	query := r.URL.Query().Get("q")

	cfg := config.Current()
	p, err := plugin.LoadPlugin(cfg.PluginPath, pluginID)
	if err != nil {
		sendJSONError(w, "Plugin not found", http.StatusNotFound)
		return
	}

	result, err := p.SearchManga(query)
	if err != nil {
		sendJSON(w, map[string]any{"success": true, "data": []any{}})
		return
	}

	sendJSON(w, map[string]any{"success": true, "data": result})
}

func (s *Server) apiAdminCreateSubscription(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PluginID   string `json:"plugin_id"`
		MangaID    string `json:"manga_id"`
		MangaTitle string `json:"manga_title"`
		Name       string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendJSONError(w, "Invalid body", http.StatusBadRequest)
		return
	}

	cfg := config.Current()
	p, err := plugin.LoadPlugin(cfg.PluginPath, body.PluginID)
	if err != nil {
		sendJSONError(w, "Plugin not found", http.StatusNotFound)
		return
	}

	sub, err := p.Subscribe(body.MangaID, body.MangaTitle, body.Name)
	if err != nil {
		sendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, map[string]any{"success": true, "data": sub})
}

func (s *Server) apiAdminListSubscriptions(w http.ResponseWriter, r *http.Request) {
	pluginID := r.URL.Query().Get("plugin")
	cfg := config.Current()

	p, err := plugin.LoadPlugin(cfg.PluginPath, pluginID)
	if err != nil {
		sendJSON(w, map[string]any{"success": true, "data": []any{}})
		return
	}

	subs, err := p.ListSubscriptions()
	if err != nil {
		sendJSON(w, map[string]any{"success": true, "data": []any{}})
		return
	}

	sendJSON(w, map[string]any{"success": true, "data": subs})
}

func (s *Server) apiAdminDeleteSubscription(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PluginID string `json:"plugin_id"`
		ID       string `json:"id"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	cfg := config.Current()
	p, err := plugin.LoadPlugin(cfg.PluginPath, body.PluginID)
	if err != nil {
		sendJSONError(w, "Plugin not found", http.StatusNotFound)
		return
	}

	if err := p.Unsubscribe(body.ID); err != nil {
		sendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, map[string]any{"success": true})
}

func (s *Server) apiAdminUpdateSubscription(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PluginID string `json:"plugin_id"`
		ID       string `json:"id"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	cfg := config.Current()
	p, err := plugin.LoadPlugin(cfg.PluginPath, body.PluginID)
	if err != nil {
		sendJSONError(w, "Plugin not found", http.StatusNotFound)
		return
	}

	subs, _ := p.ListSubscriptions()
	_ = subs
	sendJSON(w, map[string]any{"success": true})
}

func (s *Server) apiAdminPluginList(w http.ResponseWriter, r *http.Request) {
	pluginID := r.URL.Query().Get("plugin")
	mangaID := r.URL.Query().Get("manga")

	cfg := config.Current()
	p, err := plugin.LoadPlugin(cfg.PluginPath, pluginID)
	if err != nil {
		sendJSONError(w, "Plugin not found", http.StatusNotFound)
		return
	}

	result, err := p.ListChapters(mangaID)
	if err != nil {
		sendJSON(w, map[string]any{"success": true, "data": []any{}})
		return
	}

	sendJSON(w, map[string]any{"success": true, "data": result})
}

func (s *Server) apiAdminPluginDownload(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PluginID string   `json:"plugin_id"`
		MangaID  string   `json:"manga_id"`
		Chapters []string `json:"chapters"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendJSONError(w, "Invalid body", http.StatusBadRequest)
		return
	}

	var jobs []queue.Job
	for _, ch := range body.Chapters {
		jobs = append(jobs, queue.Job{
			MangaID:         body.MangaID,
			PluginID:        body.PluginID,
			PluginChapterID: ch,
			Status:          queue.StatusPending,
		})
	}

	if len(jobs) > 0 {
		if _, err := s.Deps.Queue.Push(jobs); err != nil {
			sendJSONError(w, "Failed", http.StatusInternalServerError)
			return
		}
	}

	sendJSON(w, map[string]any{"success": true})
}

func (s *Server) apiAdminQueue(w http.ResponseWriter, r *http.Request) {
	jobs, err := s.Deps.Queue.List()
	if err != nil {
		sendJSON(w, map[string]any{"success": true, "data": []any{}})
		return
	}

	sendJSON(w, map[string]any{"success": true, "data": jobs})
}

func (s *Server) apiAdminQueueAction(w http.ResponseWriter, r *http.Request) {
	action := chi.URLParam(r, "action")

	var body struct {
		ID string `json:"id"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	switch action {
	case "delete":
		if err := s.Deps.Queue.Delete(body.ID); err != nil {
			sendJSONError(w, "Failed", http.StatusInternalServerError)
			return
		}
	case "retry":
		if err := s.Deps.Queue.Reset(body.ID); err != nil {
			sendJSONError(w, "Failed", http.StatusInternalServerError)
			return
		}
	default:
		sendJSONError(w, "Unknown action", http.StatusBadRequest)
		return
	}

	sendJSON(w, map[string]any{"success": true})
}

func (s *Server) apiAdminAddTag(w http.ResponseWriter, r *http.Request) {
	tid := chi.URLParam(r, "tid")
	tag := chi.URLParam(r, "tag")
	tag = strings.ToLower(strings.TrimSpace(tag))
	if tag == "" {
		sendJSONError(w, "Invalid tag", http.StatusBadRequest)
		return
	}
	if err := s.Deps.Storage.AddTag(tid, tag); err != nil {
		sendJSONError(w, "Failed", http.StatusInternalServerError)
		return
	}
	sendJSON(w, map[string]any{"success": true})
}

func (s *Server) apiAdminDeleteTag(w http.ResponseWriter, r *http.Request) {
	tid := chi.URLParam(r, "tid")
	tag := chi.URLParam(r, "tag")
	if err := s.Deps.Storage.DeleteTag(tid, tag); err != nil {
		sendJSONError(w, "Failed", http.StatusInternalServerError)
		return
	}
	sendJSON(w, map[string]any{"success": true})
}

func (s *Server) apiAdminMissingTitles(w http.ResponseWriter, r *http.Request) {
	items, err := s.Deps.Storage.ListMissingTitles()
	if err != nil {
		sendJSONError(w, "Failed to list missing titles", http.StatusInternalServerError)
		return
	}
	// Match the legacy browser client shape: { titles: [...] }.
	sendJSON(w, map[string]any{"success": true, "titles": items})
}

func (s *Server) apiAdminMissingEntries(w http.ResponseWriter, r *http.Request) {
	items, err := s.Deps.Storage.ListMissingEntries()
	if err != nil {
		sendJSONError(w, "Failed to list missing entries", http.StatusInternalServerError)
		return
	}
	sendJSON(w, map[string]any{"success": true, "entries": items})
}

func (s *Server) apiAdminDeleteMissingTitles(w http.ResponseWriter, r *http.Request) {
	if err := s.Deps.Storage.DeleteAllMissingTitles(); err != nil {
		sendJSONError(w, "Failed to delete missing titles", http.StatusInternalServerError)
		return
	}
	sendJSON(w, map[string]any{"success": true})
}

func (s *Server) apiAdminDeleteMissingEntries(w http.ResponseWriter, r *http.Request) {
	if err := s.Deps.Storage.DeleteAllMissingEntries(); err != nil {
		sendJSONError(w, "Failed to delete missing entries", http.StatusInternalServerError)
		return
	}
	sendJSON(w, map[string]any{"success": true})
}

func (s *Server) apiAdminDeleteMissingTitle(w http.ResponseWriter, r *http.Request) {
	tid := chi.URLParam(r, "tid")
	if tid == "" {
		sendJSONError(w, "Missing title id", http.StatusBadRequest)
		return
	}
	if err := s.Deps.Storage.DeleteMissingTitle(tid); err != nil {
		sendJSONError(w, "Failed to delete missing title", http.StatusInternalServerError)
		return
	}
	sendJSON(w, map[string]any{"success": true})
}

func (s *Server) apiAdminDeleteMissingEntry(w http.ResponseWriter, r *http.Request) {
	eid := chi.URLParam(r, "eid")
	if eid == "" {
		sendJSONError(w, "Missing entry id", http.StatusBadRequest)
		return
	}
	if err := s.Deps.Storage.DeleteMissingEntry(eid); err != nil {
		sendJSONError(w, "Failed to delete missing entry", http.StatusInternalServerError)
		return
	}
	sendJSON(w, map[string]any{"success": true})
}

func (s *Server) apiAdminSetHidden(w http.ResponseWriter, r *http.Request) {
	tid := chi.URLParam(r, "tid")
	valueStr := chi.URLParam(r, "value")
	value, _ := strconv.Atoi(valueStr)
	if err := s.Deps.Storage.SetTitleHidden(tid, value); err != nil {
		sendJSONError(w, "Failed", http.StatusInternalServerError)
		return
	}
	sendJSON(w, map[string]any{"success": true})
}

func (s *Server) apiAdminHiddenTitles(w http.ResponseWriter, r *http.Request) {
	ids, err := s.Deps.Storage.GetHiddenTitleIDs()
	if err != nil {
		sendJSON(w, map[string]any{"success": true, "data": []any{}})
		return
	}
	if ids == nil {
		ids = []string{}
	}
	sendJSON(w, map[string]any{"success": true, "data": ids})
}

func (s *Server) findEntry(tid, eid string) (library.Entry, error) {
	lib := s.Deps.Library
	lib.RLock()
	t, ok := lib.TitleHash[tid]
	lib.RUnlock()

	if !ok {
		return nil, fmt.Errorf("title not found")
	}

	for _, e := range t.Entries {
		if e.ID() == eid {
			return e, nil
		}
	}

	for _, subID := range t.TitleIDs {
		lib.RLock()
		subT, subOk := lib.TitleHash[subID]
		lib.RUnlock()
		if !subOk {
			continue
		}
		for _, e := range subT.Entries {
			if e.ID() == eid {
				return e, nil
			}
		}
	}

	return nil, fmt.Errorf("entry not found")
}

func strPtr(s string) *string {
	return &s
}
