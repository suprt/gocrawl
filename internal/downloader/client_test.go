package downloader

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_Download_Success(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>Hello</body></html>"))
	}))
	defer testServer.Close()

	client := New(Config{
		Timeout:      5 * time.Second,
		MaxBodySize:  10 * 1024 * 1024,
		MaxRedirects: 10,
	})

	body, statusCode, contentType, err := client.Download(context.Background(), testServer.URL)
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}
	defer body.Close()

	if statusCode != http.StatusOK {
		t.Errorf("Download() statusCode = %d, want %d", statusCode, http.StatusOK)
	}

	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Download() contentType = %q, want %q", contentType, "text/html; charset=utf-8")
	}
}

func TestClient_Download_404(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer testServer.Close()

	client := New(Config{
		Timeout:      5 * time.Second,
		MaxBodySize:  10 * 1024 * 1024,
		MaxRedirects: 10,
	})

	body, statusCode, _, err := client.Download(context.Background(), testServer.URL)
	if err == nil {
		t.Error("Download() expected error for 404, got nil")
	}
	if body != nil {
		body.Close()
	}
	if statusCode != http.StatusNotFound {
		t.Errorf("Download() statusCode = %d, want %d", statusCode, http.StatusNotFound)
	}
}

func TestClient_Download_MaxRedirects(t *testing.T) {
	redirectCount := 0
	var serverURL string
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirectCount++
		http.Redirect(w, r, serverURL+"/redirect", http.StatusFound)
	}))
	serverURL = testServer.URL
	defer testServer.Close()

	client := New(Config{
		Timeout:      5 * time.Second,
		MaxBodySize:  10 * 1024 * 1024,
		MaxRedirects: 3,
	})

	_, _, _, err := client.Download(context.Background(), testServer.URL)
	if err == nil {
		t.Error("Download() expected error for too many redirects, got nil")
	}
}

func TestClient_Download_MaxBodySize(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("1234567890")) // 10 байт
	}))
	defer testServer.Close()

	client := New(Config{
		Timeout:      5 * time.Second,
		MaxBodySize:  5, // лимит 5 байт
		MaxRedirects: 10,
	})

	body, statusCode, _, err := client.Download(context.Background(), testServer.URL)
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}
	defer body.Close()

	if statusCode != http.StatusOK {
		t.Errorf("Download() statusCode = %d, want %d", statusCode, http.StatusOK)
	}

	// Читаем тело и проверяем, что оно обрезано
	buf := make([]byte, 100)
	n, _ := body.Read(buf)
	if n > 5 {
		t.Errorf("Download() read %d bytes, want <= 5", n)
	}
}

func TestClient_Download_ContextCancelled(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	client := New(Config{
		Timeout:      5 * time.Second,
		MaxBodySize:  10 * 1024 * 1024,
		MaxRedirects: 10,
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // отменяем сразу

	_, _, _, err := client.Download(ctx, testServer.URL)
	if err == nil {
		t.Error("Download() expected error for cancelled context, got nil")
	}
}

func TestClient_Download_ContentTypes(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
	}{
		{"HTML", "text/html"},
		{"CSS", "text/css"},
		{"JavaScript", "application/javascript"},
		{"JSON", "application/json"},
		{"PNG", "image/png"},
		{"JPEG", "image/jpeg"},
		{"PDF", "application/pdf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tt.contentType)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("data"))
			}))
			defer testServer.Close()

			client := New(Config{
				Timeout:      5 * time.Second,
				MaxBodySize:  10 * 1024 * 1024,
				MaxRedirects: 10,
			})

			_, _, contentType, err := client.Download(context.Background(), testServer.URL)
			if err != nil {
				t.Fatalf("Download() error = %v", err)
			}

			if contentType != tt.contentType {
				t.Errorf("Download() contentType = %q, want %q", contentType, tt.contentType)
			}
		})
	}
}

func TestNew_Defaults(t *testing.T) {
	client := New(Config{})

	if client.userAgent != "GoCrawl/1.0" {
		t.Errorf("New() userAgent = %q, want %q", client.userAgent, "GoCrawl/1.0")
	}
	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("New() timeout = %v, want %v", client.httpClient.Timeout, 30*time.Second)
	}
}
