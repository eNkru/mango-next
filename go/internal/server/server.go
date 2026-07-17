package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/eNkru/mango-next/internal/config"
	"github.com/eNkru/mango-next/internal/library"
	"github.com/eNkru/mango-next/internal/plugin"
	"github.com/eNkru/mango-next/internal/queue"
	"github.com/eNkru/mango-next/internal/storage"
	"github.com/eNkru/mango-next/internal/tasks"
	"github.com/eNkru/mango-next/web"
	"github.com/go-chi/chi/v5"
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
	Router   *chi.Mux
	Deps     *Dependencies
	staticFS http.FileSystem
}

func NewServer(deps *Dependencies) *Server {
	s := &Server{
		Router: chi.NewRouter(),
		Deps:   deps,
	}
	s.Router.Use(SecurityHeadersMiddleware)
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
	if cfg.AuthProxyHeaderName != "" {
		log.Printf("WARNING: auth_proxy_header_name is set to %q. Do not expose this process directly; put it behind a reverse proxy that strips or overwrites that header for every request.", cfg.AuthProxyHeaderName)
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           s.Router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP server shutdown: %v", err)
		}
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// baseMountPath returns the chi mount path for BaseURL ("/" → "", "/mango/" → "/mango").
func baseMountPath(baseURL string) string {
	p := strings.TrimSuffix(baseURL, "/")
	if p == "" || p == "/" {
		return ""
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}

// appPath joins BaseURL with a relative path segment (no leading slash required).
func (s *Server) appPath(rel string) string {
	base := "/"
	if s.Deps != nil && s.Deps.Config != nil && s.Deps.Config.BaseURL != "" {
		base = s.Deps.Config.BaseURL
	}
	rel = strings.TrimPrefix(rel, "/")
	if rel == "" {
		return base
	}
	return strings.TrimSuffix(base, "/") + "/" + rel
}

func (s *Server) RegisterRoutes() {
	deps := s.Deps
	mount := baseMountPath(deps.Config.BaseURL)

	registerApp := func(r chi.Router) {
		r.Get("/login", s.handleLoginPage)
		r.Post("/login", s.handleLogin)
		r.Get("/logout", s.handleLogout)
		// Crystal POST /api/login is unauthenticated (session/token mint).
		r.Post("/api/login", s.apiLogin)

		r.Group(func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				return AuthMiddleware(next, deps.Storage)
			})

			r.Get("/", s.handleHome)
			r.Get("/library", s.handleLibrary)
			r.Get("/book/{title}", s.handleTitle)
			r.Get("/tags", s.handleTags)
			r.Get("/tags/{tag}", s.handleTag)

			r.Get("/reader/{title}/{entry}", s.handleReaderNoPage)
			r.Get("/reader/{title}/{entry}/{page}", s.handleReader)

			r.Get("/opds", s.handleOPDSIndex)
			r.Get("/opds/book/{title_id}", s.handleOPDSTitle)

			if deps.Config.PluginPath != "" {
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
				// Placeholder route for the React + Vite foundation shell.
				r.Get("/react-preview", s.handleReactPreview)
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
					r.Get("/users", s.apiAdminListUsers)
					r.Post("/users", s.apiAdminCreateUser)
					r.Put("/users/{username}", s.apiAdminUpdateUser)
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

		s.servePublic(r)
	}

	if mount == "" {
		registerApp(s.Router)
		return
	}
	s.Router.Route(mount, registerApp)
}

func (s *Server) servePublic(r chi.Router) {
	fileServer := http.FileServer(s.staticFS)
	base := ""
	if s.Deps != nil && s.Deps.Config != nil {
		base = baseMountPath(s.Deps.Config.BaseURL)
	}

	r.Get("/*", func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path
		if base != "" {
			path = strings.TrimPrefix(path, base)
			if path == "" {
				path = "/"
			}
		}
		if path == "" || path == "/" {
			return
		}
		if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/admin") ||
			strings.HasPrefix(path, "/reader") || strings.HasPrefix(path, "/opds") ||
			strings.HasPrefix(path, "/download") || strings.HasPrefix(path, "/uploads") {
			return
		}
		// Strip mount prefix for embedded FS lookup.
		r2 := req.Clone(req.Context())
		r2.URL.Path = path
		fileServer.ServeHTTP(w, r2)
	})
}

func (s *Server) renderLayout(w http.ResponseWriter, page string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.Deps.Templates.Render(w, "views/"+page, data); err != nil {
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
