package server

import (
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	loginBodyLimit  = 1 << 20
	uploadBodyLimit = 32 << 20

	loginRateLimit  = 5
	loginRateWindow = time.Minute
)

func safeRedirectPath(callback string) string {
	callback = strings.TrimSpace(callback)
	if callback == "" {
		return "/"
	}
	if !strings.HasPrefix(callback, "/") || strings.HasPrefix(callback, "//") {
		return "/"
	}
	// Reject backslash forms that some clients normalize to protocol-relative URLs.
	if strings.Contains(callback, "\\") {
		return "/"
	}
	u, err := url.Parse(callback)
	if err != nil || u.IsAbs() || u.Host != "" || u.Scheme != "" {
		return "/"
	}
	if u.Path == "" || !strings.HasPrefix(u.Path, "/") || strings.HasPrefix(u.Path, "//") {
		return "/"
	}
	// Rebuild to drop unexpected components while keeping path/query/fragment.
	out := u.Path
	if u.RawQuery != "" {
		out += "?" + u.RawQuery
	}
	if u.Fragment != "" {
		out += "#" + u.Fragment
	}
	return out
}

func requestIsHTTPS(r *http.Request) bool {
	if r == nil {
		return false
	}
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

func clientIP(r *http.Request) string {
	if r == nil {
		return ""
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

type loginLimiter struct {
	mu      sync.Mutex
	failures map[string][]time.Time
	limit    int
	window   time.Duration
	now      func() time.Time
}

func newLoginLimiter(limit int, window time.Duration) *loginLimiter {
	return &loginLimiter{
		failures: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
		now:      time.Now,
	}
}

var defaultLoginLimiter = newLoginLimiter(loginRateLimit, loginRateWindow)

func (l *loginLimiter) pruneLocked(ip string, now time.Time) {
	cutoff := now.Add(-l.window)
	ts := l.failures[ip]
	i := 0
	for i < len(ts) && ts[i].Before(cutoff) {
		i++
	}
	if i > 0 {
		ts = append([]time.Time(nil), ts[i:]...)
	}
	if len(ts) == 0 {
		delete(l.failures, ip)
		return
	}
	l.failures[ip] = ts
}

func (l *loginLimiter) allowed(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	l.pruneLocked(ip, now)
	return len(l.failures[ip]) < l.limit
}

func (l *loginLimiter) recordFailure(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	l.pruneLocked(ip, now)
	l.failures[ip] = append(l.failures[ip], now)
}

func (l *loginLimiter) clear(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.failures, ip)
}

func loginAllowed(r *http.Request) bool {
	return defaultLoginLimiter.allowed(clientIP(r))
}

func recordLoginFailure(r *http.Request) {
	defaultLoginLimiter.recordFailure(clientIP(r))
}

func clearLoginFailures(r *http.Request) {
	defaultLoginLimiter.clear(clientIP(r))
}
