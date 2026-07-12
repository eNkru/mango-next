package plugin

import (
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/hkalexling/mango-go/internal/queue"
)

func TestSandboxHTTPClientUsesProxyFromEnvironment(t *testing.T) {
	dir := t.TempDir()
	sb, err := NewSandbox(filepath.Join(dir, "store.json"), dir)
	if err != nil {
		t.Fatal(err)
	}
	tr, ok := sb.httpClient.Transport.(*http.Transport)
	if !ok || tr.Proxy == nil {
		t.Fatal("sandbox httpClient should use Transport with ProxyFromEnvironment")
	}
}

func TestDownloaderHTTPClientUsesProxyFromEnvironment(t *testing.T) {
	dir := t.TempDir()
	q, err := queue.NewQueue(filepath.Join(dir, "q.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer q.Close()
	d := NewDownloader(q, filepath.Join(dir, "lib"), dir)
	if d.httpClient.Timeout != 30*time.Second {
		t.Fatalf("timeout = %v", d.httpClient.Timeout)
	}
	tr, ok := d.httpClient.Transport.(*http.Transport)
	if !ok || tr.Proxy == nil {
		t.Fatal("downloader httpClient should use Transport with ProxyFromEnvironment")
	}
}
