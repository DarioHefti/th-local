package llm

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
)

func TestDownloadLocalModel(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		_, _ = w.Write([]byte("GGUFmodel-data"))
	}))
	defer server.Close()
	path := filepath.Join(t.TempDir(), "models", "test.gguf")
	if err := downloadLocalModel(path, server.URL); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "GGUFmodel-data" {
		t.Fatalf("downloaded data = %q, error = %v", data, err)
	}
	if err := downloadLocalModel(path, server.URL); err != nil {
		t.Fatal(err)
	}
	if requests.Load() != 1 {
		t.Fatalf("existing model was downloaded again: %d requests", requests.Load())
	}
}

func TestDownloadLocalModelRejectsBadResponses(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
	}{
		{"not-found", 404, "not found"},
		{"html", 200, "<html>not a model</html>"},
		{"header-only", 200, "GGUF"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			path := filepath.Join(t.TempDir(), "test.gguf")
			if err := downloadLocalModel(path, server.URL); err == nil {
				t.Fatal("bad download was accepted")
			}
			entries, err := os.ReadDir(filepath.Dir(path))
			if err != nil || len(entries) != 0 {
				t.Fatalf("failed download left files behind: %v, %v", entries, err)
			}
		})
	}
}
