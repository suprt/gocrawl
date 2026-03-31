package crawler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/suprt/gocrawl/internal/downloader"
	"github.com/suprt/gocrawl/internal/naming"
	"github.com/suprt/gocrawl/internal/storage"
)

func TestCrawler_Integration_RealHTTP(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html><body>Test page " + string(rune(requestCount)) + "</body></html>"))
	}))
	defer server.Close()

	tmpDir := t.TempDir()

	storageClient, err := storage.New(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	downloaderClient := downloader.New(downloader.Config{
		Timeout:      5 * time.Second,
		MaxBodySize:  10 * 1024 * 1024,
		MaxRedirects: 10,
	})

	namerClient := naming.NewReadableNamer(50)

	c := New(downloaderClient, storageClient, namerClient, Config{
		Workers:     2,
		Timeout:     5 * time.Second,
		MaxRetries:  1,
		RateLimitMs: 0,
	})

	urls := []string{
		server.URL,
		server.URL + "/page2",
		server.URL + "/page3",
	}

	results, errors, err := c.Run(context.Background(), urls, nil)
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	if len(errors) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(errors))
	}

	if requestCount != 3 {
		t.Errorf("Expected 3 requests to server, got %d", requestCount)
	}

	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read output directory: %v", err)
	}

	if len(files) != 3 {
		t.Errorf("Expected 3 files in output directory, got %d", len(files))
	}

	if len(files) > 0 {
		content, err := os.ReadFile(filepath.Join(tmpDir, files[0].Name()))
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		if len(content) == 0 {
			t.Error("File is empty")
		}
	}
}

func TestCrawler_Integration_WithRetry(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	attemptCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html><body>Success after retry</body></html>"))
	}))
	defer server.Close()

	tmpDir := t.TempDir()

	storageClient, _ := storage.New(tmpDir)
	downloaderClient := downloader.New(downloader.Config{
		Timeout:      5 * time.Second,
		MaxBodySize:  10 * 1024 * 1024,
		MaxRedirects: 10,
	})
	namerClient := naming.NewReadableNamer(50)

	c := New(downloaderClient, storageClient, namerClient, Config{
		Workers:     1,
		Timeout:     5 * time.Second,
		MaxRetries:  2,
		RateLimitMs: 0,
	})

	results, errors, err := c.Run(context.Background(), []string{server.URL}, nil)
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result after retry, got %d", len(results))
	}

	if len(errors) != 0 {
		t.Errorf("Expected 0 errors after retry, got %d", len(errors))
	}

	if attemptCount != 2 {
		t.Errorf("Expected 2 attempts (1 fail + 1 success), got %d", attemptCount)
	}
}

func TestCrawler_Integration_404Error(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	tmpDir := t.TempDir()

	storageClient, _ := storage.New(tmpDir)
	downloaderClient := downloader.New(downloader.Config{
		Timeout:      5 * time.Second,
		MaxBodySize:  10 * 1024 * 1024,
		MaxRedirects: 10,
	})
	namerClient := naming.NewReadableNamer(50)

	c := New(downloaderClient, storageClient, namerClient, Config{
		Workers:     1,
		Timeout:     5 * time.Second,
		MaxRetries:  0,
		RateLimitMs: 0,
	})

	results, errors, err := c.Run(context.Background(), []string{server.URL}, nil)
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results for 404, got %d", len(results))
	}

	if len(errors) != 1 {
		t.Errorf("Expected 1 error for 404, got %d", len(errors))
	}
}

func TestCrawler_Integration_RealExternalHTTP(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping external HTTP test in short mode")
	}

	tmpDir := t.TempDir()

	storageClient, err := storage.New(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	downloaderClient := downloader.New(downloader.Config{
		Timeout:      10 * time.Second,
		MaxBodySize:  10 * 1024 * 1024,
		MaxRedirects: 10,
	})

	namerClient := naming.NewReadableNamer(50)

	c := New(downloaderClient, storageClient, namerClient, Config{
		Workers:     1,
		Timeout:     10 * time.Second,
		MaxRetries:  1,
		RateLimitMs: 0,
	})

	urls := []string{"https://www.google.com"}

	results, errors, err := c.Run(context.Background(), urls, nil)
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if len(errors) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(errors))
	}

	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read output directory: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 file in output directory, got %d", len(files))
	}

	if len(files) > 0 {
		info, err := files[0].Info()
		if err != nil {
			t.Fatalf("Failed to get file info: %v", err)
		}

		if info.Size() == 0 {
			t.Error("Downloaded file is empty")
		}

		if info.Size() < 1024 {
			t.Errorf("Expected file >1KB, got %d bytes", info.Size())
		}
	}
}
