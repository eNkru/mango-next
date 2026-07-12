package server

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/hkalexling/mango-go/internal/storage"
)

type TemplateManager struct {
	templates *template.Template
}

func NewTemplateManager(viewsFS fs.FS) (*TemplateManager, error) {
	funcMap := template.FuncMap{
		"slice": func(items []any) []any { return items },
		"seq":   func(n int) []int { s := make([]int, n); for i := range s { s[i] = i }; return s },
		"add":   func(a, b int) int { return a + b },
		"sub":   func(a, b int) int { return a - b },
		"html":  func(s string) template.HTML { return template.HTML(s) },
		"url":   func(s string) template.URL { return template.URL(s) },
		"js":    func(s string) template.JS { return template.JS(s) },
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

type LayoutData struct {
	BaseURL    string
	IsAdmin    bool
	PageName   string
	Version    string
	PluginPath string
}

type HomePageData struct {
	LayoutData
	ContinueReading   []storage.ContinueReadingItem
	RecentlyAdded     []storage.RecentlyAddedItem
	StartReading      []storage.StartReadingItem
	NewUser           bool
	EmptyLibrary      bool
	ConfigLibraryPath string
	ConfigPath        string
	ScanIntervalMinutes int
}

type LibraryPageData struct {
	LayoutData
	Titles     []LibraryTitle
	Percentage []float64
	ShowHidden bool
}

type LibraryTitle struct {
	ID           string
	Name         string
	CoverURL     string
	EntryCount   int
	Hidden       bool
}

type TitlePageData struct {
	LayoutData
	Title            TitleDetail
	SortedTitles     []TitleDetail
	Entries          []EntryDetail
	Percentage       []float64
	TitlePercentage  []float64
	IsHidden         bool
}

type TitleDetail struct {
	ID          string
	Name        string
	CoverURL    string
	ParentIDs   []string
	Hidden      bool
}

type EntryDetail struct {
	ID       string
	Name     string
	PageCount int
	CoverURL string
	MimeType string
}

type ReaderPageData struct {
	BaseURL          string
	PageName         string
	Title            TitleDetail
	Entry            EntryDetail
	PageIdx          int
	Entries          []EntryDetail
	ExitURL          string
	NextEntryURL     string
	PreviousEntryURL string
	Version          string
}

type ReaderErrorPageData struct {
	BaseURL      string
	PageName     string
	EntryName    string
	EntryError   string
	ExitURL      string
	NextEntryURL string
	Version      string
}

type AdminPageData struct {
	LayoutData
	MissingCount int
}

type UserPageData struct {
	LayoutData
	Users    [][2]string
	Username string
}

type UserEditPageData struct {
	LayoutData
	Username string
	Admin    bool
	Error    string
	NewUser  bool
}

type TagPageData struct {
	LayoutData
	Tag      string
	Titles   []LibraryTitle
	ShowHidden bool
}

type TagsPageData struct {
	LayoutData
	Tags []TagInfo
}

type TagInfo struct {
	Tag        string
	EncodedTag string
	Count      int
}

type OPDSTitleEntry struct {
	ID   string
	Name string
}

type OPDSIndexPageData struct {
	BaseURL string
	Titles  []OPDSTitleEntry
}

type OPDSTitlePageData struct {
	BaseURL    string
	Title      TitleDetail
	SubTitles  []TitleDetail
	Entries    []EntryDetail
}
