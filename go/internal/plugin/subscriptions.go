package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// FilterType mirrors Crystal's FilterType enum.
type FilterType int

const (
	FilterString  FilterType = iota
	FilterNumMin
	FilterNumMax
	FilterDateMin
	FilterDateMax
	FilterArray
)

// String returns the Crystal-compatible string representation.
func (ft FilterType) String() string {
	switch ft {
	case FilterString:
		return "string"
	case FilterNumMin:
		return "number-min"
	case FilterNumMax:
		return "number-max"
	case FilterDateMin:
		return "date-min"
	case FilterDateMax:
		return "date-max"
	case FilterArray:
		return "array"
	default:
		return "unknown"
	}
}

// FilterTypeFromString parses a Crystal FilterType string.
func FilterTypeFromString(s string) (FilterType, error) {
	switch s {
	case "string":
		return FilterString, nil
	case "number-min":
		return FilterNumMin, nil
	case "number-max":
		return FilterNumMax, nil
	case "date-min":
		return FilterDateMin, nil
	case "date-max":
		return FilterDateMax, nil
	case "array":
		return FilterArray, nil
	default:
		return 0, fmt.Errorf("unknown filter type: %s", s)
	}
}

// Filter mirrors Crystal's Filter struct.
type Filter struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
	Type  FilterType  `json:"type"`
}

// FilterJSON is the JSON-serializable representation of Filter.
type FilterJSON struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
	Type  string      `json:"type"`
}

// MarshalJSON implements custom JSON marshaling for Filter.
func (f Filter) MarshalJSON() ([]byte, error) {
	return json.Marshal(FilterJSON{
		Key:   f.Key,
		Value: f.Value,
		Type:  f.Type.String(),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for Filter.
func (f *Filter) UnmarshalJSON(data []byte) error {
	var fj FilterJSON
	if err := json.Unmarshal(data, &fj); err != nil {
		return err
	}
	ft, err := FilterTypeFromString(fj.Type)
	if err != nil {
		return err
	}
	f.Key = fj.Key
	f.Value = fj.Value
	f.Type = ft
	return nil
}

// Subscription mirrors Crystal's Subscription class.
type Subscription struct {
	ID          string   `json:"id"`
	PluginID    string   `json:"plugin_id"`
	MangaID     string   `json:"manga_id"`
	MangaTitle  string   `json:"manga_title"`
	Name        string   `json:"name"`
	CreatedAt   int64    `json:"created_at"`
	LastChecked int64    `json:"last_checked"`
	Filters     []Filter `json:"filters,omitempty"`
}

// NewSubscription creates a new subscription with a random UUID and current timestamps.
func NewSubscription(pluginID, mangaID, mangaTitle, name string) *Subscription {
	return &Subscription{
		ID:          uuid.New().String(),
		PluginID:    pluginID,
		MangaID:     mangaID,
		MangaTitle:  mangaTitle,
		Name:        name,
		CreatedAt:   time.Now().Unix(),
		LastChecked: time.Now().Unix(),
		Filters:     []Filter{},
	}
}

// SubscriptionList mirrors Crystal's SubscriptionList struct.
// It manages subscriptions stored in a subscriptions.json file within each plugin directory.
type SubscriptionList struct {
	Subscriptions []*Subscription `json:"subscriptions"`
	dir           string
	path          string
}

// LoadSubscriptionList loads subscriptions from the plugin directory.
func LoadSubscriptionList(pluginDir string) (*SubscriptionList, error) {
	sl := &SubscriptionList{
		Subscriptions: []*Subscription{},
		dir:           pluginDir,
		path:          filepath.Join(pluginDir, "subscriptions.json"),
	}

	data, err := os.ReadFile(sl.path)
	if err != nil {
		if os.IsNotExist(err) {
			return sl, nil
		}
		return nil, err
	}

	// Crystal stores subscriptions as a bare JSON array, not wrapped in an object.
	var subs []*Subscription
	if err := json.Unmarshal(data, &subs); err != nil {
		// Fallback: try as wrapped object (our own format)
		if err2 := json.Unmarshal(data, sl); err2 != nil {
			return nil, fmt.Errorf("parse subscriptions.json: %w (unwrap: %v)", err, err2)
		}
		return sl, nil
	}
	sl.Subscriptions = subs
	return sl, nil
}

// Save writes the subscription list to disk in Crystal-compatible format
// (bare JSON array).
func (sl *SubscriptionList) Save() error {
	data, err := json.MarshalIndent(sl.Subscriptions, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(sl.path, data, 0o644)
}

// Add appends a subscription and saves.
func (sl *SubscriptionList) Add(sub *Subscription) error {
	sl.Subscriptions = append(sl.Subscriptions, sub)
	return sl.Save()
}

// Remove removes a subscription by ID and saves.
func (sl *SubscriptionList) Remove(id string) error {
	filtered := make([]*Subscription, 0, len(sl.Subscriptions))
	for _, s := range sl.Subscriptions {
		if s.ID != id {
			filtered = append(filtered, s)
		}
	}
	sl.Subscriptions = filtered
	return sl.Save()
}

// Find returns a subscription by ID, or nil.
func (sl *SubscriptionList) Find(id string) *Subscription {
	for _, s := range sl.Subscriptions {
		if s.ID == id {
			return s
		}
	}
	return nil
}

// --- Plugin subscription methods ---

// Subscribe creates a new subscription for this plugin and saves it.
func (p *Plugin) Subscribe(mangaID, mangaTitle, name string) (*Subscription, error) {
	sub := NewSubscription(p.info.ID, mangaID, mangaTitle, name)
	list, err := LoadSubscriptionList(p.info.Dir)
	if err != nil {
		return nil, err
	}
	if err := list.Add(sub); err != nil {
		return nil, err
	}
	return sub, nil
}

// ListSubscriptions returns all subscriptions for this plugin.
func (p *Plugin) ListSubscriptions() ([]*Subscription, error) {
	list, err := LoadSubscriptionList(p.info.Dir)
	if err != nil {
		return nil, err
	}
	return list.Subscriptions, nil
}

// Unsubscribe removes a subscription by ID.
func (p *Plugin) Unsubscribe(id string) error {
	list, err := LoadSubscriptionList(p.info.Dir)
	if err != nil {
		return err
	}
	return list.Remove(id)
}

// CheckSubscription checks a single subscription for new chapters, returning
// the matching chapters. It does not enqueue — the caller (e.g. Updater)
// handles queue push and subscriptions.save.
// afterMs = sub.LastChecked * 1000 (seconds → ms, matching Crystal's new_chapters).
func (p *Plugin) CheckSubscription(sub *Subscription) ([]any, error) {
	result, err := p.NewChapters(sub.MangaID, sub.LastChecked*1000)
	if err != nil {
		return nil, fmt.Errorf("newChapters: %w", err)
	}

	chapters, ok := result.([]any)
	if !ok {
		return nil, fmt.Errorf("newChapters result is not an array")
	}

	// Apply subscription filters.
	var matching []any
	for _, ch := range chapters {
		chObj, ok := ch.(map[string]any)
		if !ok {
			continue
		}
		if sub.matchesChapter(chObj) {
			matching = append(matching, chObj)
		}
	}

	return matching, nil
}
