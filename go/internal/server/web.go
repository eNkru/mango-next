package server

import (
	"encoding/json"
	"html/template"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"
)

// viteManifestEntry mirrors the relevant fields of a Vite manifest entry for the
// HTML input (index.html): the entry JS file and its CSS dependencies. Asset
// filenames are content-hashed (vite.config.ts), so the Go shell resolves them
// from the manifest at startup instead of hardcoding assets/main.js|css.
type viteManifestEntry struct {
	File string   `json:"file"`
	CSS  []string `json:"css"`
}

// viteManifest maps the manifest's entry keys (e.g. "index.html") to entries.
type viteManifest map[string]viteManifestEntry

// reactAssets holds the content-hashed entry asset paths (relative to /react/)
// resolved from the Vite manifest. Empty strings mean the manifest was missing
// or malformed — the shell renders with no asset URLs, surfacing at boot.
type reactAssets struct {
	mainJS  string
	mainCSS string
}

// loadReactAssets reads the Vite manifest from the embedded public FS and
// resolves the entry JS + CSS paths (relative to /react/, e.g.
// "assets/index-<hash>.js"). Returns zero values if the manifest is missing
// or malformed; callers should treat that as a build-not-run error.
func loadReactAssets(publicFS fs.FS) reactAssets {
	raw, err := fs.ReadFile(publicFS, "react/manifest.json")
	if err != nil {
		slog.Error("react manifest missing; run npm run build", "err", err)
		return reactAssets{}
	}
	var manifest viteManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		slog.Error("react manifest parse", "err", err)
		return reactAssets{}
	}
	entry, ok := manifest["index.html"]
	if !ok || entry.File == "" {
		slog.Error("react manifest has no index.html entry")
		return reactAssets{}
	}
	assets := reactAssets{mainJS: entry.File}
	if len(entry.CSS) > 0 {
		assets.mainCSS = entry.CSS[0]
	}
	return assets
}

type TemplateManager struct {
	templates *template.Template
}

func NewTemplateManager(viewsFS fs.FS) (*TemplateManager, error) {
	funcMap := template.FuncMap{
		"slice": func(items []any) []any { return items },
		"seq": func(n int) []int {
			s := make([]int, n)
			for i := range s {
				s[i] = i
			}
			return s
		},
		"add":  func(a, b int) int { return a + b },
		"sub":  func(a, b int) int { return a - b },
		"html": func(s string) template.HTML { return template.HTML(s) },
		"url":  func(s string) template.URL { return template.URL(s) },
		"js":   func(s string) template.JS { return template.JS(s) },
	}

	tmpl := template.New("").Funcs(funcMap)

	err := fs.WalkDir(viewsFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".tmpl") {
			return nil
		}
		name := strings.TrimSuffix(path, ".tmpl")
		name = filepath.ToSlash(name)
		data, err := fs.ReadFile(viewsFS, path)
		if err != nil {
			return err
		}
		_, err = tmpl.New(name).Parse(string(data))
		return err
	})
	if err != nil {
		return nil, err
	}

	return &TemplateManager{templates: tmpl}, nil
}

func (tm *TemplateManager) Render(w interface{ Write([]byte) (int, error) }, name string, data any) error {
	return tm.templates.ExecuteTemplate(w, name, data)
}

func (tm *TemplateManager) Lookup(name string) *template.Template {
	return tm.templates.Lookup(name)
}

// ReactShellData is the Go HTML shell payload for migrated React routes.
type ReactShellData struct {
	BaseURL  string
	PageName string
	// BootJSON is raw JSON embedded in #mango-boot (already serialized).
	BootJSON template.JS
	// MainJS / MainCSS are the content-hashed entry asset paths (relative to
	// /react/, e.g. "assets/index-<hash>.js"), resolved from the Vite manifest
	// at startup. Empty when the manifest is missing/malformed (build-not-run).
	MainJS  string
	MainCSS string
}

// LayoutData is shared metadata for internal page-data helpers still used by
// visibility filters (library/tag), not for full HTML layouts.
type LayoutData struct {
	BaseURL  string
	IsAdmin  bool
	PageName string
	Version  string
}

type LibraryPageData struct {
	LayoutData
	Titles     []LibraryTitle
	Percentage []float64
	ShowHidden bool
}

type LibraryTitle struct {
	ID         string
	Name       string
	CoverURL   string
	EntryCount int
	Hidden     bool
}

type TagPageData struct {
	LayoutData
	Tag        string
	Titles     []LibraryTitle
	ShowHidden bool
}
