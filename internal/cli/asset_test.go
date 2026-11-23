package cli

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var testAssetPNG = []byte("\x89PNG\r\n\x1a\n test image")

func TestWriteAssetWritesTheFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "output")

	path, err := WriteAsset(dir, "nano-image", []byte("image bytes"), "image/png")
	if err != nil {
		t.Fatalf("WriteAsset() error: %v", err)
	}

	if filepath.Dir(path) != dir {
		t.Errorf("path = %s, want it under %s", path, dir)
	}
	if filepath.Ext(path) != ".png" {
		t.Errorf("extension = %s, want .png", filepath.Ext(path))
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}
	if string(data) != "image bytes" {
		t.Errorf("data = %q, want image bytes", data)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o644 {
		t.Errorf("mode = %o, want 644", perm)
	}
}

func TestWriteAssetRejectsAnUnknownMediaType(t *testing.T) {
	_, err := WriteAsset(t.TempDir(), "nano-image", []byte("data"), "application/pdf")
	if err == nil || !strings.Contains(err.Error(), "cannot store") {
		t.Errorf("WriteAsset() error = %v, want a media-type rejection", err)
	}
}

func TestDownloadAssetUsesTheContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(testAssetPNG)
	}))
	defer server.Close()

	data, mediaType, err := DownloadAsset(context.Background(), server.URL, "")
	if err != nil {
		t.Fatalf("DownloadAsset() error: %v", err)
	}
	if string(data) != string(testAssetPNG) {
		t.Errorf("data = %q, want the response body", data)
	}
	if mediaType != "image/png" {
		t.Errorf("mediaType = %q, want image/png", mediaType)
	}
}

func TestDownloadAssetDetectsAGenericContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(testAssetPNG)
	}))
	defer server.Close()

	_, mediaType, err := DownloadAsset(context.Background(), server.URL, "")
	if err != nil {
		t.Fatalf("DownloadAsset() error: %v", err)
	}
	if mediaType != "image/png" {
		t.Errorf("mediaType = %q, want image/png", mediaType)
	}
}

func TestDownloadAssetUsesTheFallbackMediaType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write([]byte("unlabelled video"))
	}))
	defer server.Close()

	_, mediaType, err := DownloadAsset(context.Background(), server.URL, "video/mp4")
	if err != nil {
		t.Fatalf("DownloadAsset() error: %v", err)
	}
	if mediaType != "video/mp4" {
		t.Errorf("mediaType = %q, want video/mp4", mediaType)
	}
}

func TestDownloadAssetRejectsANon200Response(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "gone", http.StatusGone)
	}))
	defer server.Close()

	_, _, err := DownloadAsset(context.Background(), server.URL, "")
	if err == nil || !strings.Contains(err.Error(), "410") {
		t.Errorf("DownloadAsset() error = %v, want the status", err)
	}
}

func TestDownloadAssetRejectsAnOversizedAsset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("12345"))
	}))
	defer server.Close()

	_, _, err := downloadAsset(context.Background(), server.URL, "", 4)
	if err == nil || !strings.Contains(err.Error(), "byte cap") {
		t.Errorf("downloadAsset() error = %v, want the size cap", err)
	}
}

func TestDownloadAssetRejectsANonHTTPURL(t *testing.T) {
	_, _, err := DownloadAsset(context.Background(), "file:///tmp/image.png", "")
	if err == nil || !strings.Contains(err.Error(), "not an http URL") {
		t.Errorf("DownloadAsset() error = %v, want an http URL rejection", err)
	}
}
