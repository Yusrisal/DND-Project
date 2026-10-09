package middleware

import (
	"net/http"
	"slices"
	"strings"
)

// CORS allows only the configured origins. Credentials are enabled, because the
// session lives in a cookie, which is also why the allowlist must stay explicit:
// a wildcard origin plus credentials is exactly the combination browsers refuse.
func CORS(origins []string) Middleware {
	allowed := make([]string, 0, len(origins))
	for _, o := range origins {
		if o = strings.TrimSpace(o); o != "" {
			allowed = append(allowed, o)
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && slices.Contains(allowed, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
