package server

import (
	"path/filepath"
	"testing"

	"github.com/eNkru/mango-next/internal/library"
	"github.com/eNkru/mango-next/internal/storage"
)

func newLibraryWithTitles(t *testing.T, dir string, titles ...*library.Title) *library.Library {
	t.Helper()
	lib := library.NewLibrary(filepath.Join(dir, "library"), nil, "")
	lib.TitleIDs = make([]string, 0, len(titles))
	lib.TitleHash = make(map[string]*library.Title, len(titles))
	for _, title := range titles {
		lib.TitleIDs = append(lib.TitleIDs, title.ID)
		lib.TitleHash[title.ID] = title
	}
	return lib
}

func seedTitle(t *testing.T, st *storage.Storage, id, path string) {
	t.Helper()
	_, err := st.DB().Exec(
		"INSERT INTO titles (id, path, signature, unavailable, hidden) VALUES (?, ?, ?, 0, 0)",
		id, path, "sig-"+id,
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBuildLibraryPageDataFiltersHiddenTitles(t *testing.T) {
	st, cfg, dir := setupTest(t)
	seedTitle(t, st, "visible", "path/visible")
	seedTitle(t, st, "hidden", "path/hidden")

	lib := newLibraryWithTitles(t, dir,
		&library.Title{ID: "visible", Name: "Visible Title"},
		&library.Title{ID: "hidden", Name: "Hidden Title"},
	)
	if err := st.SetTitleHidden("hidden", 1); err != nil {
		t.Fatal(err)
	}

	s := NewServer(&Dependencies{Config: cfg, Storage: st, Library: lib})

	// Default view: hidden titles are omitted.
	data := s.buildLibraryPageData(true, false)
	if data.ShowHidden {
		t.Fatal("default ShowHidden = true, want false")
	}
	if len(data.Titles) != 1 || data.Titles[0].ID != "visible" {
		t.Fatalf("default titles = %+v, want only visible", data.Titles)
	}

	// Admin show-hidden view: hidden titles appear with Hidden=true.
	data = s.buildLibraryPageData(true, true)
	if !data.ShowHidden {
		t.Fatal("show_hidden ShowHidden = false, want true")
	}
	if len(data.Titles) != 2 {
		t.Fatalf("show_hidden titles = %+v, want 2", data.Titles)
	}
	foundHidden := false
	for _, title := range data.Titles {
		if title.ID == "hidden" {
			foundHidden = true
			if !title.Hidden {
				t.Fatal("hidden title Hidden = false, want true")
			}
		}
	}
	if !foundHidden {
		t.Fatalf("show_hidden titles = %+v, missing hidden title", data.Titles)
	}

	// Non-admin callers cannot enable show-hidden mode.
	data = s.buildLibraryPageData(false, true)
	if data.ShowHidden {
		t.Fatal("non-admin ShowHidden = true, want false")
	}
	if len(data.Titles) != 1 || data.Titles[0].ID != "visible" {
		t.Fatalf("non-admin titles = %+v, want only visible", data.Titles)
	}
}

func TestBuildTagPageDataMarksHiddenTitles(t *testing.T) {
	st, cfg, dir := setupTest(t)
	seedTitle(t, st, "visible", "path/visible")
	seedTitle(t, st, "hidden", "path/hidden")

	lib := newLibraryWithTitles(t, dir,
		&library.Title{ID: "visible", Name: "Visible Title"},
		&library.Title{ID: "hidden", Name: "Hidden Title"},
	)
	if err := st.AddTag("visible", "action"); err != nil {
		t.Fatal(err)
	}
	if err := st.AddTag("hidden", "action"); err != nil {
		t.Fatal(err)
	}
	if err := st.SetTitleHidden("hidden", 1); err != nil {
		t.Fatal(err)
	}

	s := NewServer(&Dependencies{Config: cfg, Storage: st, Library: lib})

	// Default tag view excludes hidden titles.
	data, ok := s.buildTagPageData("action", true, false)
	if !ok {
		t.Fatal("default tag page data missing")
	}
	if data.ShowHidden {
		t.Fatal("default tag ShowHidden = true, want false")
	}
	if len(data.Titles) != 1 || data.Titles[0].ID != "visible" {
		t.Fatalf("default tag titles = %+v, want only visible", data.Titles)
	}

	// Admin show-hidden view includes and marks hidden titles.
	data, ok = s.buildTagPageData("action", true, true)
	if !ok {
		t.Fatal("show_hidden tag page data missing")
	}
	if !data.ShowHidden {
		t.Fatal("show_hidden tag ShowHidden = false, want true")
	}
	if len(data.Titles) != 2 {
		t.Fatalf("show_hidden tag titles = %+v, want 2", data.Titles)
	}
	for _, title := range data.Titles {
		if title.ID == "hidden" && !title.Hidden {
			t.Fatal("hidden tag title Hidden = false, want true")
		}
		if title.ID == "visible" && title.Hidden {
			t.Fatal("visible tag title Hidden = true, want false")
		}
	}

	// Non-admin show_hidden request is ignored.
	data, ok = s.buildTagPageData("action", false, true)
	if !ok {
		t.Fatal("non-admin tag page data missing")
	}
	if data.ShowHidden {
		t.Fatal("non-admin tag ShowHidden = true, want false")
	}
	if len(data.Titles) != 1 || data.Titles[0].ID != "visible" {
		t.Fatalf("non-admin tag titles = %+v, want only visible", data.Titles)
	}
}
