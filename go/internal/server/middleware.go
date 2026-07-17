package server

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Same-origin browser use only; do not advertise Access-Control-Allow-Origin: *.
		w.Header().Set("Access-Control-Allow-Methods", "HEAD,GET,PUT,POST,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With,X-HTTP-Method-Override, Content-Type, Cache-Control, Accept, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)
		log.Printf("%d %s %s %s", rw.statusCode, r.Method, r.URL.Path, time.Since(start))
	})
}

func underUploadRoot(absPath, absUpload string) bool {
	rel, err := filepath.Rel(absUpload, absPath)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func UploadHandler(uploadPath string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/uploads") {
				filePath := filepath.Join(uploadPath, strings.TrimPrefix(r.URL.Path, "/uploads/"))
				absPath, err := filepath.Abs(filePath)
				if err != nil {
					http.Error(w, "Bad request", http.StatusBadRequest)
					return
				}
				absUpload, err := filepath.Abs(uploadPath)
				if err != nil {
					http.Error(w, "Bad request", http.StatusBadRequest)
					return
				}
				if !underUploadRoot(absPath, absUpload) {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
				if _, err := os.Stat(absPath); os.IsNotExist(err) {
					http.Error(w, "Not found", http.StatusNotFound)
					return
				}
				http.ServeFile(w, r, absPath)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (s *Server) isStaticFile(path string) bool {
	staticDirs := []string{"/css", "/js", "/img", "/webfonts"}
	for _, dir := range staticDirs {
		if strings.HasPrefix(path, dir) {
			return true
		}
	}
	return path == "/favicon.ico" || path == "/robots.txt" || path == "/manifest.json"
}
