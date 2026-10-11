package handler

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// The point of the handler is that the kind comes from the bytes, not from the
// filename, so the cases are named after what is inside the file.
func TestUploadCreate(t *testing.T) {
	png := []byte("\x89PNG\r\n\x1a\n" + "rest of the file")
	webp := append([]byte("RIFF\x00\x00\x00\x00WEBP"), []byte("payload")...)

	cases := []struct {
		name     string
		filename string
		content  []byte
		status   int
		wantExt  string
	}{
		{"png", "map.png", png, http.StatusCreated, ".png"},
		{"webp", "map.webp", webp, http.StatusCreated, ".webp"},
		{"script renamed to png", "map.png", []byte("#!/bin/sh\nrm -rf /\n"), http.StatusUnprocessableEntity, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()

			body := &bytes.Buffer{}
			form := multipart.NewWriter(body)
			part, err := form.CreateFormFile("file", tc.filename)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := part.Write(tc.content); err != nil {
				t.Fatal(err)
			}
			if err := form.Close(); err != nil {
				t.Fatal(err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/uploads", body)
			req.Header.Set("Content-Type", form.FormDataContentType())
			rec := httptest.NewRecorder()

			NewUploadHandler(dir).Create(rec, req)

			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tc.status, rec.Body.String())
			}
			if tc.status != http.StatusCreated {
				if entries, _ := os.ReadDir(dir); len(entries) != 0 {
					t.Fatalf("rejected upload still wrote %d file(s)", len(entries))
				}
				return
			}

			var out struct {
				URL string `json:"url"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
				t.Fatal(err)
			}
			if filepath.Ext(out.URL) != tc.wantExt {
				t.Fatalf("url = %q, want extension %q", out.URL, tc.wantExt)
			}
			saved, err := os.ReadFile(filepath.Join(dir, filepath.Base(out.URL)))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(saved, tc.content) {
				t.Fatal("saved bytes differ from what was uploaded")
			}
		})
	}
}
