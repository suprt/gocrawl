package downloader

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestServer(bodySize int, delay time.Duration, contentType string) *httptest.Server {
	body := make([]byte, bodySize)
	for i := range body {
		body[i] = byte(i % 256)
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Имитация сетевой задержки
		if delay > 0 {
			time.Sleep(delay)
		}

		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))
}

func benchmarkDownloader(b *testing.B, bodySize int, delay time.Duration, contentType string) {
	server := newTestServer(bodySize, delay, contentType)
	defer server.Close()
	client := New(Config{
		Timeout:      5 * time.Second,
		MaxBodySize:  10 * 1024 * 1024,
		MaxRedirects: 10,
	})
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		body, _, _, err := client.Download(ctx, server.URL)
		if err != nil {
			b.Fatal(err)
		}
		_ = body.Close()
	}
}

func BenchmarkDownloader_Download_1KB(b *testing.B) {
	benchmarkDownloader(b, 1024, 0, "text/html")
}
func BenchmarkDownloader_Download_100KB(b *testing.B) {
	benchmarkDownloader(b, 100*1024, 0, "text/html")
}
func BenchmarkDownloader_Download_1MB(b *testing.B) {
	benchmarkDownloader(b, 1024*1024, 0, "text/html")
}

func BenchmarkDownloader_Download_HTML(b *testing.B) {
	benchmarkDownloader(b, 100*1024, 0, "text/html")
}
func BenchmarkDownloader_Download_JSON(b *testing.B) {
	benchmarkDownloader(b, 100*1024, 0, "application/json")
}
func BenchmarkDownloader_Download_PNG(b *testing.B) {
	benchmarkDownloader(b, 100*1024, 0, "image/png")
}

func BenchmarkDownloader_Download_10msDelay(b *testing.B) {
	benchmarkDownloader(b, 100*1024, 10*time.Millisecond, "text/html")
}
func BenchmarkDownloader_Download_100msDelay(b *testing.B) {
	benchmarkDownloader(b, 100*1024, 100*time.Millisecond, "text/html")
}

func BenchmarkDownloader_Download_UTF8(b *testing.B) {
	benchmarkDownloader(b, 100*1024, 0, "text/html; charset=utf-8")
}
func BenchmarkDownloader_Download_Windows1251(b *testing.B) {
	benchmarkDownloader(b, 100*1024, 0, "text/html; charset=windows-1251")
}

func BenchmarkDownloader_Download_MaxBodySize(b *testing.B) {
	benchmarkDownloader(b, 20*1024*1024, 0, "text/html")
}
