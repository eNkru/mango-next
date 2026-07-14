package plugin

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/eNkru/mango-next/internal/queue"
)

// Updater periodically checks all plugin subscriptions for new chapters and
// pushes new chapters to the download queue. Mirrors Crystal Plugin::Updater.
type Updater struct {
	pluginDir     string
	queue         *queue.Queue
	intervalHours int
}

// NewUpdater creates an Updater instance.
func NewUpdater(pluginDir string, q *queue.Queue, intervalHours int) *Updater {
	return &Updater{
		pluginDir:     pluginDir,
		queue:         q,
		intervalHours: intervalHours,
	}
}

// Start launches the background update loop. Returns immediately if
// intervalHours <= 0. Blocks until ctx is cancelled.
func (u *Updater) Start(ctx context.Context) {
	if u.intervalHours <= 0 {
		log.Println("Plugin update checker disabled (interval <= 0)")
		return
	}

	log.Printf("Plugin update checker started (interval: %d hours)", u.intervalHours)

	// Run once on startup, then on the interval.
	u.checkAll()

	ticker := time.NewTicker(time.Duration(u.intervalHours) * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			u.checkAll()
		case <-ctx.Done():
			log.Println("Plugin update checker stopped")
			return
		}
	}
}

// checkAll iterates over all available plugins and checks for updates.
func (u *Updater) checkAll() {
	pluginIDs := listPluginIDs(u.pluginDir)
	for _, pid := range pluginIDs {
		u.checkPlugin(pid)
	}
}

// checkPlugin checks all subscriptions for a single plugin.
// Matches Crystal Plugin::Updater#check_updates: loads the subscription list,
// checks each subscription, and saves all changes (including LastChecked) to disk.
func (u *Updater) checkPlugin(pluginID string) {
	log.Printf("Checking plugin %s for updates", pluginID)

	plugin, err := LoadPlugin(u.pluginDir, pluginID)
	if err != nil {
		log.Printf("Error loading plugin %s: %v", pluginID, err)
		return
	}

	// Skip v1 plugins — they don't support newChapters.
	if !plugin.CanSubscribe() {
		log.Printf("Plugin %s is targeting API version 1. Skipping update check", pluginID)
		return
	}

	// Load subscriptions directly so we can save the modified list (with updated LastChecked)
	// back to disk, matching Crystal's subscriptions.save call.
	list, err := LoadSubscriptionList(filepath.Join(u.pluginDir, pluginID))
	if err != nil {
		log.Printf("Error loading subscriptions for %s: %v", pluginID, err)
		return
	}

	for _, sub := range list.Subscriptions {
		u.checkSubscription(plugin, sub)
	}

	if err := list.Save(); err != nil {
		log.Printf("Error saving subscriptions for %s: %v", pluginID, err)
	}
}

// checkSubscription checks a single subscription for new chapters.
func (u *Updater) checkSubscription(plugin *Plugin, sub *Subscription) {
	log.Printf("Checking subscription %s for updates", sub.Name)

	// Call newChapters to find chapters published after last_checked.
	// The Crystal version does: newChapters(manga_id, last_checked * 1000)
	// where afterMs = last_checked * 1000 (convert seconds to ms for JS Date).
	result, err := plugin.NewChapters(sub.MangaID, sub.LastChecked*1000)
	if err != nil {
		log.Printf("Error checking subscription %s: %v", sub.Name, err)
		return
	}

	chapters, ok := result.([]any)
	if !ok {
		log.Printf("newChapters returned non-array result for %s", sub.Name)
		return
	}

	var matching []map[string]any
	for _, ch := range chapters {
		chObj, ok := ch.(map[string]any)
		if !ok {
			continue
		}

		// Apply filters if any.
		if !sub.matchesChapter(chObj) {
			continue
		}

		matching = append(matching, chObj)
	}

	if len(matching) == 0 {
		log.Printf("No new chapters found for subscription %s", sub.Name)
		sub.LastChecked = time.Now().Unix()
		return
	}

	log.Printf("Found %d new chapters for subscription %s. Pushing to download queue",
		len(matching), sub.Name)

	// Build queue jobs.
	jobs := make([]queue.Job, 0, len(matching))
	for _, ch := range matching {
		chID := fmt.Sprintf("%v", ch["id"])
		chTitle := fmt.Sprintf("%v", ch["title"])

		// Encode chapter ID with base64 (matching Crystal's Base64.encode)
		encodedID := base64.StdEncoding.EncodeToString([]byte(chID))
		jobID := plugin.info.ID + "-" + encodedID

		jobs = append(jobs, queue.Job{
			ID:         jobID,
			MangaID:    sub.MangaID,
			Title:      chTitle,
			MangaTitle: sub.MangaTitle,
			Status:     queue.StatusPending,
			Time:       time.Now(),
		})
	}

	inserted, err := u.queue.Push(jobs)
	if err != nil {
		log.Printf("Error pushing to download queue: %v", err)
		return
	}

	log.Printf("%d/%d new chapters added to the download queue. Plugin ID %s, subscription name %s",
		inserted, len(matching), plugin.info.ID, sub.Name)

	if inserted != len(matching) {
		log.Printf("Failed to add %d chapters to download queue",
			len(matching)-inserted)
	}

	sub.LastChecked = time.Now().Unix()
}

// listPluginIDs scans the plugin directory and returns all valid plugin IDs.
func listPluginIDs(pluginDir string) []string {
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		log.Printf("Error reading plugin directory %s: %v", pluginDir, err)
		return nil
	}

	var ids []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		infoPath := filepath.Join(pluginDir, entry.Name(), "info.json")
		if _, err := os.Stat(infoPath); err == nil {
			ids = append(ids, entry.Name())
		}
	}
	return ids
}

// matchesChapter applies the subscription's filters to a chapter object.
func (s *Subscription) matchesChapter(chapter map[string]any) bool {
	if len(s.Filters) == 0 {
		return true
	}
	for _, f := range s.Filters {
		if !f.matchChapter(chapter) {
			return false
		}
	}
	return true
}

// matchChapter applies a single filter to a chapter object.
func (f Filter) matchChapter(chapter map[string]any) bool {
	if f.Value == nil {
		return true
	}

	rawVal, ok := chapter[f.Key]
	if !ok {
		return false
	}
	rawStr := fmt.Sprintf("%v", rawVal)

	switch f.Type {
	case FilterString:
		valStr := fmt.Sprintf("%v", f.Value)
		return rawStr == valStr
	case FilterNumMin, FilterDateMin:
		// Numeric comparison: raw >= value
		return compareNumeric(rawStr, fmt.Sprintf("%v", f.Value)) >= 0
	case FilterNumMax, FilterDateMax:
		// Numeric comparison: raw <= value
		return compareNumeric(rawStr, fmt.Sprintf("%v", f.Value)) <= 0
	case FilterArray:
		valStr := fmt.Sprintf("%v", f.Value)
		if valStr == "all" {
			return true
		}
		// Check if the raw value (comma-separated) contains the filter value
		rawLower := toLower(rawStr)
		valLower := toLower(valStr)
		for _, part := range splitTrim(rawLower, ",") {
			if part == valLower {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func compareNumeric(a, b string) int {
	// Parse as float for comparison.
	var fa, fb float64
	if _, err := fmt.Sscanf(a, "%f", &fa); err != nil {
		return 0
	}
	if _, err := fmt.Sscanf(b, "%f", &fb); err != nil {
		return 0
	}
	if fa < fb {
		return -1
	}
	if fa > fb {
		return 1
	}
	return 0
}

func splitTrim(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if string(s[i]) == sep {
			part := trimSpace(s[start:i])
			if part != "" {
				result = append(result, part)
			}
			start = i + 1
		}
	}
	part := trimSpace(s[start:])
	if part != "" {
		result = append(result, part)
	}
	return result
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		b[i] = c
	}
	return string(b)
}
