// Package server provides HTTP middleware and routing for the Mango server.
// This file contains authentication middleware that mirrors the Crystal
// AuthHandler (src/handlers/auth_handler.cr).
package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/eNkru/mango-next/internal/config"
	"github.com/eNkru/mango-next/internal/storage"
)

// Context keys for storing auth info in the request context.
type contextKey string

const (
	contextKeyUsername   contextKey = "username"
	contextKeyIsAdmin    contextKey = "is_admin"
	contextKeyAuthMethod contextKey = "auth_method" // "cookie", "bearer", "proxy", "disabled_login"
	cookieNamePrefix                = "mango-token-"
)

// AuthMiddleware returns an http.Handler that wraps the given handler with
// authentication checks, mirroring the Crystal AuthHandler#call.
//
// Routes that should skip authentication (login/logout/static) must be
// registered BEFORE this middleware or use a separate subrouter.
func AuthMiddleware(next http.Handler, st *storage.Storage) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// OPTIONS requests skip auth.
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		cfg := config.Current()

		// Attempt authentication via token (cookie or Bearer header).
		token := extractToken(r, cfg)
		username, err := st.VerifyToken(token)
		isAuthenticated := err == nil && username != ""

		if isAuthenticated {
			// Store auth info in context.
			isAdmin, _ := st.VerifyAdmin(token)
			ctx := context.WithValue(r.Context(), contextKeyUsername, username)
			ctx = context.WithValue(ctx, contextKeyIsAdmin, isAdmin)
			ctx = context.WithValue(ctx, contextKeyAuthMethod, "token")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Handle disable_login mode.
		if cfg.DisableLogin {
			if cfg.DefaultUsername == "" {
				http.Error(w, "Server misconfigured: default username not set", http.StatusInternalServerError)
				return
			}
			exists, err := st.UsernameExists(cfg.DefaultUsername)
			if err != nil || !exists {
				log.Printf("Default username %q does not exist", cfg.DefaultUsername)
				requireAuth(w, r)
				return
			}
			isAdmin, _ := st.UsernameIsAdmin(cfg.DefaultUsername)
			ctx := context.WithValue(r.Context(), contextKeyUsername, cfg.DefaultUsername)
			ctx = context.WithValue(ctx, contextKeyIsAdmin, isAdmin)
			ctx = context.WithValue(ctx, contextKeyAuthMethod, "disabled_login")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Handle auth proxy header.
		if cfg.AuthProxyHeaderName != "" {
			proxyUsername := r.Header.Get(cfg.AuthProxyHeaderName)
			if proxyUsername == "" || !userExists(st, proxyUsername) {
				log.Printf("Header %q unset or is not a valid username", cfg.AuthProxyHeaderName)
				requireAuth(w, r)
				return
			}
			isAdmin, _ := st.UsernameIsAdmin(proxyUsername)
			ctx := context.WithValue(r.Context(), contextKeyUsername, proxyUsername)
			ctx = context.WithValue(ctx, contextKeyIsAdmin, isAdmin)
			ctx = context.WithValue(ctx, contextKeyAuthMethod, "proxy")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		requireAuth(w, r)
	})
}

// AdminMiddleware returns an http.Handler that requires the user to be an
// admin. It must be used AFTER AuthMiddleware (or the auth info must be in
// the context).
func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAdmin, _ := r.Context().Value(contextKeyIsAdmin).(bool)
		if !isAdmin {
			http.Error(w, "HTTP 403: You are not authorized to visit this page", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// GetUsername extracts the authenticated username from the request context.
// Returns empty string if not authenticated.
func GetUsername(r *http.Request) string {
	u, _ := r.Context().Value(contextKeyUsername).(string)
	return u
}

// GetIsAdmin extracts the admin flag from the request context.
func GetIsAdmin(r *http.Request) bool {
	a, _ := r.Context().Value(contextKeyIsAdmin).(bool)
	return a
}

// SetAuthTokenCookie sets the auth token cookie on the response, matching
// the cookie name pattern from the Crystal server.
func SetAuthTokenCookie(w http.ResponseWriter, cfg *config.Config, token string) {
	cookieName := cookieNamePrefix + fmt.Sprintf("%d", cfg.Port)
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     cfg.BaseURL,
		HttpOnly: true,
		// MaxAge: 365*24*60*60 = 365 days, matching kemal-session timeout.
		MaxAge:   365 * 24 * 60 * 60,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearAuthTokenCookie clears the auth token cookie.
func ClearAuthTokenCookie(w http.ResponseWriter, cfg *config.Config) {
	cookieName := cookieNamePrefix + fmt.Sprintf("%d", cfg.Port)
	http.SetCookie(w, &http.Cookie{
		Name:   cookieName,
		Value:  "",
		Path:   cfg.BaseURL,
		MaxAge: -1,
	})
}

// extractToken tries to get the auth token from a cookie or the Authorization
// header, in that order.
func extractToken(r *http.Request, cfg *config.Config) string {
	// Try cookie first (matching kemal-session's cookie-per-port pattern).
	cookieName := cookieNamePrefix + fmt.Sprintf("%d", cfg.Port)
	if cookie, err := r.Cookie(cookieName); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// Fall back to legacy kemal-session cookie format.
	if cookie, err := r.Cookie("mango-sessid-" + fmt.Sprintf("%d", cfg.Port)); err == nil && cookie.Value != "" {
		legacyCookie := cookie.Value
		return legacyCookie
	}

	// Try Authorization: Bearer <token> header.
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	// Try Authorization: Basic (for OPDS). Returns base64 encoded "user:pass".
	// The caller must handle this separately for OPDS auth.
	if strings.HasPrefix(auth, "Basic ") {
		return auth
	}

	return ""
}

// userExists is a helper that checks if a username exists.
func userExists(st *storage.Storage, username string) bool {
	exists, err := st.UsernameExists(username)
	return err == nil && exists
}

// requireAuth sends a 401 for API routes or redirects to login for pages.
func requireAuth(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}
