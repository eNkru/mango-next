package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/hkalexling/mango-go/internal/config"
	"github.com/hkalexling/mango-go/internal/library"
	"github.com/hkalexling/mango-go/internal/plugin"
	"github.com/hkalexling/mango-go/internal/queue"
	"github.com/hkalexling/mango-go/internal/storage"
	"github.com/hkalexling/mango-go/internal/tasks"
	"github.com/hkalexling/mango-go/web"
)

type Dependencies struct {
	Config    *config.Config
	Storage   *storage.Storage
	Library   *library.Library
	Queue     *queue.Queue
	Plugins   map[string]*plugin.Plugin
	Runner    *tasks.Runner
	Templates *TemplateManager
}

type Server struct {
	Router      *chi.Mux
	Deps        *Dependencies
	staticFS    http.FileSystem
}

func NewServer(deps *Dependencies) *Server {
	s := &Server{
		Router: chi.NewRouter(),
		Deps:   deps,
	}
	s.Router.Use(CORSMiddleware)
	s.Router.Use(LoggingMiddleware)
	s.Router.Use(UploadHandler(deps.Config.UploadPath))
	s.staticFS = http.FS(web.Public())
	return s
}

func (s *Server) Start(ctx context.Context) error {
	cfg := s.Deps.Config
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	log.Printf("Starting server on %s", addr)

	srv := &http.Server{
		Addr:    addr,
		Handler: s.Router,
	}

	go func() {
		<-ctx.Done()
		srv.Close()
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) RegisterRoutes() {
	r := s.Router
	deps := s.Deps
	cfg := deps.Config

	r.Get("/login", s.handleLoginPage)
	r.Post("/login", s.handleLogin)
	r.Get("/logout", s.handleLogout)

	r.Group(func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return AuthMiddleware(next, deps.Storage)
		})

		r.Get("/", s.handleHome)
		r.Get("/library", s.handleLibrary)
		r.Get("/book/{title}", s.handleTitle)
		r.Get("/tags", s.handleTags)
		r.Get("/tags/{tag}", s.handleTag)
		r.Get("/api", s.handleAPIDocs)

		r.Get("/reader/{title}/{entry}", s.handleReaderNoPage)
		r.Get("/reader/{title}/{entry}/{page}", s.handleReader)

		r.Get("/opds", s.handleOPDSIndex)
		r.Get("/opds/book/{title_id}", s.handleOPDSTitle)

		if cfg.PluginPath != "" {
			r.Get("/download/plugins", s.handlePluginDownload)
		}

		r.Route("/admin", func(r chi.Router) {
			r.Use(AdminMiddleware)
			r.Get("/", s.handleAdmin)
			r.Get("/user", s.handleUserList)
			r.Get("/user/edit", s.handleUserEdit)
			r.Post("/user/edit", s.handleUserEditPost)
			r.Post("/user/edit/{original_username}", s.handleUserEditPost)
			r.Get("/downloads", s.handleDownloadManager)
			r.Get("/subscriptions", s.handleSubscriptionManager)
			r.Get("/missing", s.handleMissingItems)
		})

		r.Route("/api", func(r chi.Router) {
			r.Get("/library", s.apiLibrary)
			r.Get("/library/continue_reading", s.apiContinueReading)
			r.Get("/library/start_reading", s.apiStartReading)
			r.Get("/library/recently_added", s.apiRecentlyAdded)
			r.Get("/book/{tid}", s.apiBook)
			r.Get("/sort_opt", s.apiGetSortOpt)
			r.Put("/sort_opt", s.apiPutSortOpt)
			r.Get("/page/{tid}/{eid}/{page}", s.apiPage)
			r.Get("/cover/{tid}/{eid}", s.apiCover)
			r.Get("/dimensions/{tid}/{eid}", s.apiDimensions)
			r.Get("/download/{tid}/{eid}", s.apiDownload)
			r.Put("/progress/{tid}/{page}", s.apiSaveProgress)
			r.Put("/bulk_progress/{action}/{tid}", s.apiBulkProgress)
			r.Get("/tags/{tid}", s.apiGetTitleTags)
			r.Get("/tags", s.apiListTags)

			r.Route("/admin", func(r chi.Router) {
				r.Use(AdminMiddleware)
				r.Post("/scan", s.apiAdminScan)
				r.Get("/thumbnail_progress", s.apiAdminThumbnailProgress)
				r.Post("/generate_thumbnails", s.apiAdminGenerateThumbnails)
				r.Delete("/user/delete/{username}", s.apiAdminDeleteUser)
				r.Put("/display_name/{tid}/{name}", s.apiAdminSetDisplayName)
				r.Put("/sort_title/{tid}", s.apiAdminSetSortTitle)
				r.Post("/upload/{target}", s.apiAdminUpload)
				r.Get("/plugin", s.apiAdminListPlugins)
				r.Get("/plugin/info", s.apiAdminPluginInfo)
				r.Get("/plugin/search", s.apiAdminPluginSearch)
				r.Post("/plugin/subscriptions", s.apiAdminCreateSubscription)
				r.Get("/plugin/subscriptions", s.apiAdminListSubscriptions)
				r.Delete("/plugin/subscriptions", s.apiAdminDeleteSubscription)
				r.Post("/plugin/subscriptions/update", s.apiAdminUpdateSubscription)
				r.Get("/plugin/list", s.apiAdminPluginList)
				r.Post("/plugin/download", s.apiAdminPluginDownload)
				r.Get("/queue", s.apiAdminQueue)
				r.Post("/queue/{action}", s.apiAdminQueueAction)
				r.Put("/tags/{tid}/{tag}", s.apiAdminAddTag)
				r.Delete("/tags/{tid}/{tag}", s.apiAdminDeleteTag)
				r.Get("/titles/missing", s.apiAdminMissingTitles)
				r.Get("/entries/missing", s.apiAdminMissingEntries)
				r.Delete("/titles/missing", s.apiAdminDeleteMissingTitles)
				r.Delete("/entries/missing", s.apiAdminDeleteMissingEntries)
				r.Delete("/titles/missing/{tid}", s.apiAdminDeleteMissingTitle)
				r.Delete("/entries/missing/{eid}", s.apiAdminDeleteMissingEntry)
				r.Put("/hidden/{tid}/{value}", s.apiAdminSetHidden)
				r.Get("/hidden_titles", s.apiAdminHiddenTitles)
			})
		})
	})

	s.servePublic()
}

func (s *Server) servePublic() {
	r := s.Router
	fileServer := http.FileServer(s.staticFS)

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "" || path == "/" {
			return
		}
		if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/admin") ||
			strings.HasPrefix(path, "/reader") || strings.HasPrefix(path, "/opds") ||
			strings.HasPrefix(path, "/download") || strings.HasPrefix(path, "/uploads") {
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}

func (s *Server) renderLayout(w http.ResponseWriter, page string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// The layout template references page content via {{ template "content" . }}
	if err := s.Deps.Templates.Render(w, "views/layout", data); err != nil {
		log.Printf("Template render error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func (s *Server) renderPage(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.Deps.Templates.Render(w, name, data); err != nil {
		log.Printf("Template render error %s: %v", name, err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}
