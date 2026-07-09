package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Info struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Placeholder string            `json:"placeholder"`
	WaitSeconds int               `json:"wait_seconds"`
	APIVersion  int               `json:"api_version"`
	Settings    map[string]string `json:"settings"`
	Dir         string
}

type Plugin struct {
	info *Info
	sbx  *Sandbox
}

func LoadPlugin(pluginDir, pluginID string) (*Plugin, error) {
	info, err := loadInfo(filepath.Join(pluginDir, pluginID))
	if err != nil {
		return nil, err
	}
	jsPath := filepath.Join(info.Dir, "index.js")
	js, err := os.ReadFile(jsPath)
	if err != nil {
		return nil, fmt.Errorf("read index.js: %w", err)
	}
	storagePath := filepath.Join(info.Dir, "storage.json")
	sbx, err := NewSandbox(storagePath, info.Dir)
	if err != nil {
		return nil, err
	}
	if _, err := sbx.Eval(string(js)); err != nil {
		return nil, fmt.Errorf("eval index.js: %w", err)
	}
	return &Plugin{info: info, sbx: sbx}, nil
}

func (p *Plugin) Info() *Info { return p.info }

func (p *Plugin) Eval(src string) (any, error) {
	v, err := p.sbx.Eval(src)
	if err != nil {
		return nil, err
	}
	return v.Export(), nil
}

func loadInfo(dir string) (*Info, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "info.json"))
	if err != nil {
		return nil, fmt.Errorf("read info.json: %w", err)
	}
	var info Info
	if err := json.Unmarshal(raw, &info); err != nil {
		return nil, fmt.Errorf("parse info.json: %w", err)
	}
	if info.ID == "" || info.Title == "" || info.Placeholder == "" {
		return nil, fmt.Errorf("info.json missing required fields (id, title, placeholder)")
	}
	if info.APIVersion == 0 {
		info.APIVersion = 1
	}
	info.Dir = dir
	return &info, nil
}

func escapeJS(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	return s
}

func evalJSON(sbx *Sandbox, src string) (any, error) {
	v, err := sbx.Eval(src)
	if err != nil {
		return nil, err
	}
	switch val := v.Export().(type) {
	case string:
		var result any
		if err := json.Unmarshal([]byte(val), &result); err != nil {
			return nil, fmt.Errorf("parse JSON result (string): %w", err)
		}
		return result, nil
	default:
		return val, nil
	}
}

func evalExists(sbx *Sandbox, fnName string) bool {
	_, err := sbx.Eval(fnName)
	return err == nil
}

func (p *Plugin) SearchManga(query string) (any, error) {
	if p.info.APIVersion == 1 {
		return nil, fmt.Errorf("manga searching is only available for plugins targeting API v2 or above")
	}
	return evalJSON(p.sbx, "searchManga('"+escapeJS(query)+"')")
}

func (p *Plugin) ListChapters(query string) (any, error) {
	return evalJSON(p.sbx, "listChapters('"+escapeJS(query)+"')")
}

func (p *Plugin) SelectChapter(id string) (any, error) {
	return evalJSON(p.sbx, "selectChapter('"+escapeJS(id)+"')")
}

func (p *Plugin) NextPage() (any, error) {
	return evalJSON(p.sbx, "nextPage()")
}

func (p *Plugin) NewChapters(mangaID string, afterMs int64) (any, error) {
	return evalJSON(p.sbx, fmt.Sprintf("newChapters('%s', %d)", escapeJS(mangaID), afterMs))
}

func (p *Plugin) CanSubscribe() bool {
	return p.info.APIVersion > 1 && evalExists(p.sbx, "newChapters")
}
