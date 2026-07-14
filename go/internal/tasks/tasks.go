package tasks

import (
	"context"
	"log"
	"time"

	"github.com/eNkru/mango-next/internal/library"
	"github.com/eNkru/mango-next/internal/plugin"
	"github.com/eNkru/mango-next/internal/queue"
)

// Runner manages all background tasks: library scanning, thumbnail generation,
// plugin update checking, and download queue processing.
// Mirrors Crystal's spawn-based background jobs.
type Runner struct {
	lib *library.Library

	// Library tasks
	scanIntervalMinutes    int
	thumbnailIntervalHours int

	// Plugin tasks
	queue                     *queue.Queue
	pluginDir                 string
	pluginUpdateIntervalHours int
	libraryPath               string
}

// NewRunner creates a Runner attached to the given Library and plugin system.
func NewRunner(lib *library.Library, scanIntervalMinutes, thumbnailIntervalHours int) *Runner {
	return &Runner{
		lib:                    lib,
		scanIntervalMinutes:    scanIntervalMinutes,
		thumbnailIntervalHours: thumbnailIntervalHours,
	}
}

// SetPluginTasks configures the plugin update checking and download queue
// processing. Must be called before Start.
func (r *Runner) SetPluginTasks(q *queue.Queue, pluginDir, libraryPath string, pluginUpdateIntervalHours int) {
	r.queue = q
	r.pluginDir = pluginDir
	r.libraryPath = libraryPath
	r.pluginUpdateIntervalHours = pluginUpdateIntervalHours
}

// Start launches all background goroutines. Blocks until ctx is cancelled.
func (r *Runner) Start(ctx context.Context) {
	// --- Library scanner ---
	initialScanDone := make(chan struct{})
	go func() {
		r.runScan()
		close(initialScanDone)
		if r.scanIntervalMinutes >= 1 {
			ticker := time.NewTicker(time.Duration(r.scanIntervalMinutes) * time.Minute)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					r.runScan()
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	// --- Thumbnail generator ---
	if r.thumbnailIntervalHours >= 1 {
		go func() {
			select {
			case <-initialScanDone:
			case <-ctx.Done():
				return
			}
			r.runThumbnails()
			ticker := time.NewTicker(time.Duration(r.thumbnailIntervalHours) * time.Hour)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					r.runThumbnails()
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// --- Plugin update checker ---
	if r.queue != nil && r.pluginUpdateIntervalHours > 0 {
		go func() {
			updater := plugin.NewUpdater(r.pluginDir, r.queue, r.pluginUpdateIntervalHours)
			updater.Start(ctx)
		}()
	}

	// --- Download queue processor ---
	if r.queue != nil {
		go func() {
			downloader := plugin.NewDownloader(r.queue, r.libraryPath, r.pluginDir)
			downloader.Start(ctx)
		}()
	}

	// Block until cancelled
	<-ctx.Done()
	log.Println("Background tasks stopped")
}

func (r *Runner) runScan() {
	log.Println("Library scan starting...")
	start := time.Now()
	result, err := r.lib.Scan()
	if err != nil {
		log.Printf("Library scan error: %v", err)
		return
	}
	elapsed := time.Since(start)
	log.Printf("Library scan finished: %d titles, %d entries in %v",
		result.TitleCount, result.EntryCount, elapsed)
}

func (r *Runner) runThumbnails() {
	log.Println("Thumbnail generation starting...")
	start := time.Now()
	if err := r.lib.GenerateThumbnails(); err != nil {
		log.Printf("Thumbnail generation error: %v", err)
		return
	}
	log.Printf("Thumbnail generation finished in %v", time.Since(start))
}
