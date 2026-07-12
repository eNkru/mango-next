package plugin

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hkalexling/mango-go/internal/queue"
)

// Downloader processes download queue jobs by fetching pages via the plugin's
// selectChapter/nextPage API and saving them as .cbz files in the library.
// Mirrors Crystal Plugin::Downloader (extends Queue::Downloader).
type Downloader struct {
	queue       *queue.Queue
	libraryPath string
	pluginDir   string
	httpClient  *http.Client
	downloading bool
}

// NewDownloader creates a Downloader instance.
func NewDownloader(q *queue.Queue, libraryPath, pluginDir string) *Downloader {
	return &Downloader{
		queue:       q,
		libraryPath: libraryPath,
		pluginDir:   pluginDir,
		// mirrors Crystal src/util/proxy.cr: respect HTTP(S)_PROXY / NO_PROXY
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		},
	}
}

// Start launches the background download loop. Blocks until ctx is cancelled.
func (d *Downloader) Start(ctx context.Context) {
	log.Println("Download queue processor started")

	// Run immediately, then poll every second (matching Crystal's sleep 1.second).
	d.processNext()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.processNext()
		case <-ctx.Done():
			log.Println("Download queue processor stopped")
			return
		}
	}
}

// processNext pops and processes the next downloadable job.
func (d *Downloader) processNext() {
	if d.downloading {
		return
	}

	job, err := d.queue.PopDownloadable()
	if err != nil {
		log.Printf("Error popping from download queue: %v", err)
		return
	}
	if job == nil {
		return
	}

	d.downloading = true
	go func() {
		defer func() {
			d.downloading = false
		}()
		d.downloadJob(job)
	}()
}

// downloadJob downloads a single queue job.
func (d *Downloader) downloadJob(job *queue.Job) {
	log.Printf("Downloading job %s: %s", job.ID, job.Title)

	if err := d.queue.SetStatus(queue.StatusDownloading, job.ID); err != nil {
		log.Printf("Error setting status to Downloading: %v", err)
		return
	}

	// Load the plugin.
	plugin, err := LoadPlugin(d.pluginDir, job.PluginID)
	if err != nil {
		d.failJob(job, fmt.Sprintf("Failed to load plugin %s: %v", job.PluginID, err))
		return
	}

	// Call selectChapter to get chapter details and pages.
	result, err := plugin.SelectChapter(job.PluginChapterID)
	if err != nil {
		d.failJob(job, fmt.Sprintf("selectChapter failed: %v", err))
		return
	}

	chapter, ok := result.(map[string]any)
	if !ok {
		d.failJob(job, "selectChapter did not return a chapter object")
		return
	}

	// Crystal's assert_chapter_type validates obj["pages"].as_i — pages is an
	// integer count (not an array). Individual pages are fetched via nextPage().
	pagesCountRaw, ok := chapter["pages"]
	if !ok {
		d.failJob(job, "chapter has no pages field")
		return
	}
	var pageCount int
	switch v := pagesCountRaw.(type) {
	case float64:
		pageCount = int(v)
	case int64:
		pageCount = int(v)
	case int:
		pageCount = v
	default:
		d.failJob(job, fmt.Sprintf("pages has unexpected type %T (expected number)", pagesCountRaw))
		return
	}

	if err := d.queue.SetPages(pageCount, job.ID); err != nil {
		log.Printf("Error setting pages: %v", err)
	}

	mangaTitle := sanitizeFilename(job.MangaTitle)
	chapterTitle := sanitizeFilename(fmt.Sprintf("%v", chapter["title"]))

	// Create manga directory.
	mangaDir := filepath.Join(d.libraryPath, mangaTitle)
	if err := os.MkdirAll(mangaDir, 0o755); err != nil {
		d.failJob(job, fmt.Sprintf("Failed to create manga dir: %v", err))
		return
	}

	// Create .cbz.part file.
	zipPartPath := filepath.Join(mangaDir, chapterTitle+".cbz.part")
	zipFinalPath := filepath.Join(mangaDir, chapterTitle+".cbz")

	failCount := 0

	// Create zip writer.
	zipFile, err := os.Create(zipPartPath)
	if err != nil {
		d.failJob(job, fmt.Sprintf("Failed to create zip file: %v", err))
		return
	}

	zipWriter := zip.NewWriter(zipFile)

	// Download each page.
	pageIndex := 1
	for {
		// Get next page from plugin.
		pageResult, err := plugin.NextPage()
		if err != nil {
			d.failJob(job, fmt.Sprintf("nextPage failed: %v", err))
			zipWriter.Close()
			zipFile.Close()
			os.Remove(zipPartPath)
			return
		}

		if pageResult == nil {
			break
		}

		page, ok := pageResult.(map[string]any)
		if !ok {
			d.failJob(job, "nextPage result is not a map")
			zipWriter.Close()
			zipFile.Close()
			os.Remove(zipPartPath)
			return
		}

		// Check if job still exists (not cancelled).
		exists, err := d.queue.Exists(job.ID)
		if err != nil || !exists {
			log.Printf("Download cancelled for %s", job.ID)
			zipWriter.Close()
			zipFile.Close()
			os.Remove(zipPartPath)
			return
		}

		fn := sanitizeFilename(fmt.Sprintf("%v", page["filename"]))
		url := fmt.Sprintf("%v", page["url"])

		// Check for custom headers.
		headers := map[string]string{}
		if h, ok := page["headers"]; ok {
			if hMap, ok := h.(map[string]any); ok {
				for k, v := range hMap {
					headers[k] = fmt.Sprintf("%v", v)
				}
			}
		}

		success := false
		tries := 4

		for tries >= 0 {
			// Wait according to plugin's wait_seconds.
			if plugin.info.WaitSeconds > 0 {
				time.Sleep(time.Duration(plugin.info.WaitSeconds) * time.Second)
			}

			log.Printf("Downloading %s", url)
			tries--

			err := d.downloadPageToZip(zipWriter, fn, url, headers)
			if err == nil {
				d.queue.AddSuccess(job.ID)
				log.Printf("[success] %s", url)
				success = true
				break
			}

			d.queue.AddFail(job.ID)
			failCount++
			msg := fmt.Sprintf("Failed to download page %s. Error: %v", url, err)
			d.queue.AddMessage(msg, job.ID)
			log.Printf("[failed] %s - %v", url, err)
		}

		if !success {
			log.Printf("All retries exhausted for page %d", pageIndex)
		}

		pageIndex++
	}

	// Close the zip file.
	if err := zipWriter.Close(); err != nil {
		d.failJob(job, fmt.Sprintf("Failed to close zip: %v", err))
		zipFile.Close()
		os.Remove(zipPartPath)
		return
	}
	zipFile.Close()

	// Check job still exists.
	exists, err := d.queue.Exists(job.ID)
	if err != nil || !exists {
		log.Printf("Download cancelled for %s after completion", job.ID)
		os.Remove(zipPartPath)
		return
	}

	// Rename .part to final.
	if err := os.Rename(zipPartPath, zipFinalPath); err != nil {
		d.failJob(job, fmt.Sprintf("Failed to rename zip file: %v", err))
		return
	}

	log.Printf("cbz file created at %s", zipFinalPath)

	// Validate the archive.
	if err := validateZip(zipFinalPath); err != nil {
		d.queue.AddMessage(fmt.Sprintf("The downloaded archive is corrupted. Error: %v", err), job.ID)
		d.queue.SetStatus(queue.StatusError, job.ID)
	} else if failCount > 0 {
		d.queue.SetStatus(queue.StatusMissingPages, job.ID)
	} else {
		d.queue.SetStatus(queue.StatusCompleted, job.ID)
	}
}

// downloadPageToZip downloads a page and adds it to the zip writer.
func (d *Downloader) downloadPageToZip(zw *zip.Writer, filename, url string, headers map[string]string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	w, err := zw.Create(filename)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, resp.Body)
	return err
}

// failJob sets a job to error status with a message.
func (d *Downloader) failJob(job *queue.Job, msg string) {
	log.Printf("Download failed for %s: %s", job.ID, msg)
	d.queue.AddMessage(msg, job.ID)
	d.queue.SetStatus(queue.StatusError, job.ID)
}

// sanitizeFilename removes path separators and other problematic characters,
// matching Crystal's sanitize_filename helper.
func sanitizeFilename(name string) string {
	// Replace path separators
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	// Remove null bytes
	name = strings.ReplaceAll(name, "\x00", "")
	// Trim whitespace
	name = strings.TrimSpace(name)
	if name == "" {
		name = "untitled"
	}
	return name
}

// validateZip checks if a zip file is valid by opening and reading its entries.
func validateZip(path string) error {
	r, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer r.Close()

	if len(r.File) == 0 {
		return fmt.Errorf("empty archive")
	}

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("corrupt entry %s: %w", f.Name, err)
		}
		// Read a small amount to verify the data is accessible.
		_, err = io.ReadAll(io.LimitReader(rc, 1024))
		rc.Close()
		if err != nil {
			return fmt.Errorf("unreadable entry %s: %w", f.Name, err)
		}
	}

	return nil
}
