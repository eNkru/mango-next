package server

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"
)

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
}

// LayoutData is shared metadata for internal page-data helpers still used by
// visibility filters (library/tag), not for full HTML layouts.
type LayoutData struct {
	BaseURL    string
	IsAdmin    bool
	PageName   string
	Version    string
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
