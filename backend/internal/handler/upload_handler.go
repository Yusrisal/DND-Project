package handler

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// maxUpload caps a single image. Battle maps are large, player tokens are not,
// so one generous limit covers both without a per-kind rule.
const maxUpload = 8 << 20 // 8 MiB

// extByType is the allowlist. The type is sniffed from the bytes, never taken
// from the filename or the Content-Type header, so a renamed .exe cannot land
// in a directory the server hands back out over HTTP.
var extByType = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

type UploadHandler struct {
	dir string
}

func NewUploadHandler(dir string) *UploadHandler {
	return &UploadHandler{dir: dir}
}

// Create accepts a multipart "file" field and answers with the public path to
// store in image_url. Uploads live outside the release directory so a deploy
// does not take the maps down with it.
func (h *UploadHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)

	file, _, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{
			Error: "expected a file field, at most 8 MB", Code: "bad_request", Field: "file"})
		return
	}
	defer func() { _ = file.Close() }()

	head := make([]byte, 512)
	n, err := io.ReadFull(file, head)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		writeError(w, err)
		return
	}
	ext, ok := extByType[detectImage(head[:n])]
	if !ok {
		writeJSON(w, http.StatusUnprocessableEntity, errorBody{
			Error: "only PNG, JPEG, GIF and WebP images are accepted",
			Code:  "validation_failed", Field: "file"})
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		writeError(w, err)
		return
	}

	if err := os.MkdirAll(h.dir, 0o755); err != nil {
		writeError(w, err)
		return
	}

	name := randomName() + ext
	dst, err := os.Create(filepath.Join(h.dir, name))
	if err != nil {
		writeError(w, err)
		return
	}
	if _, err := io.Copy(dst, file); err != nil {
		_ = dst.Close()
		_ = os.Remove(dst.Name())
		writeError(w, err)
		return
	}
	if err := dst.Close(); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"url": "/uploads/" + name})
}

// detectImage adds WebP to what http.DetectContentType knows, which reports it
// as application/octet-stream: "RIFF" .... "WEBP".
func detectImage(head []byte) string {
	if len(head) >= 12 && string(head[0:4]) == "RIFF" && string(head[8:12]) == "WEBP" {
		return "image/webp"
	}
	return http.DetectContentType(head)
}

func randomName() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf) // crypto/rand.Read never fails on supported platforms.
	return hex.EncodeToString(buf)
}

// uploads serves the stored files read-only. http.Dir rejects paths that climb
// out of the directory, and the handler is wrapped in g.user so the files are
// no more public than the campaign itself.
func uploads(dir string) http.Handler {
	return http.StripPrefix("/uploads/", http.FileServer(http.Dir(dir)))
}
