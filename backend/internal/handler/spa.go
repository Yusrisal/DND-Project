package handler

import (
	"net/http"
	"os"
	"path/filepath"
)

// spa serves the built frontend from dir, falling back to index.html so client
// side routes survive a refresh. In production this puts the API, the socket
// and the app on one origin, which is also why CORS is a dev-only concern.
func spa(dir string) http.Handler {
	files := http.FileServer(http.Dir(dir))
	index := filepath.Join(dir, "index.html")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(dir, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			files.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, index)
	})
}
