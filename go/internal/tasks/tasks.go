package tasks

import (
	"context"
	"log"
	"time"

	"github.com/eNkru/mango-next/internal/library"
)

// Runner manages background library scanning and thumbnail generation.
type Runner struct {
	lib                    *library.Library
	scanIntervalMinutes    int
	thumbnailIntervalHours int
}

// NewRunner creates a Runner attached to the given Library.
func NewRunner(lib *library.Library, scanIntervalMinutes, thumbnailIntervalHours int) *Runner {
	return &Runner{
		lib:                    lib,
		scanIntervalMinutes:    scanIntervalMinutes,
		thumbnailIntervalHours: thumbnailIntervalHours,
	}
}

// Start launches library background goroutines. Blocks until ctx is cancelled.
func (r *Runner) Start(ctx context.Context) {
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
