package plugin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type jsStore struct {
	path string
	mu   sync.RWMutex
	data map[string]string
}

func newStore(path string) (*jsStore, error) {
	s := &jsStore{
		path: path,
		data: make(map[string]string),
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); err == nil {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &s.data); err != nil {
				return nil, err
			}
		}
	}
	return s, nil
}

func (s *jsStore) get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	return v, ok
}

func (s *jsStore) set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return s.saveLocked()
}

func (s *jsStore) saveLocked() error {
	raw, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, raw, 0o644)
}
