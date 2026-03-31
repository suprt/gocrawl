package crawler

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// Добавляем задержку для имитации реального запроса
type mockDownloaderWithDelay struct {
	delay time.Duration
}

func (m *mockDownloaderWithDelay) Download(ctx context.Context, url string) (io.ReadCloser, int, string, error) {
	select {
	case <-time.After(m.delay):
		return io.NopCloser(strings.NewReader("<html>test</html>")), http.StatusOK, "text/html", nil
	case <-ctx.Done():
		return nil, 0, "", ctx.Err()
	}
}

type mockStorage struct{}

func (m *mockStorage) Save(r io.Reader, filename string) (string, error) {
	return "/test/path/" + filename, nil
}

type mockNamer struct{}

func (m mockNamer) Name(rawURL string) string {
	return "test.html"
}

func (m mockNamer) NameWithExtension(rawURL string, ext string) string {
	return "test" + ext
}

type mocBar struct{}

func (m mocBar) Add(n int) error {
	return nil
}

func (m mocBar) Finish() error {
	return nil
}

func generateURLs(n int) []string {
	urls := make([]string, n)
	for i := 0; i < n; i++ {
		urls[i] = "https://www.example.com/page" + string(rune(i))
	}
	return urls
}

func benchmarkCrawler(b *testing.B, workers, urls int, delay time.Duration) {
	crawler := New(&mockDownloaderWithDelay{delay: delay}, &mockStorage{}, &mockNamer{}, Config{
		Workers:     workers,
		Timeout:     5 * time.Second,
		MaxRetries:  0,
		RateLimitMs: 0,
	})
	urlList := generateURLs(urls)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		ctx := context.Background()
		b.StartTimer()
		_, _, _ = crawler.Run(ctx, urlList, &mocBar{})
	}
}

func BenchmarkCrawler_Run_100URLs_1workers(b *testing.B) {
	benchmarkCrawler(b, 1, 100, 10*time.Millisecond)
}
func BenchmarkCrawler_Run_100URLs_5workers(b *testing.B) {
	benchmarkCrawler(b, 5, 100, 10*time.Millisecond)
}
func BenchmarkCrawler_Run_100URLs_10workers(b *testing.B) {
	benchmarkCrawler(b, 10, 100, 10*time.Millisecond)
}
func BenchmarkCrawler_Run_100URLs_20workers(b *testing.B) {
	benchmarkCrawler(b, 20, 100, 10*time.Millisecond)
}
func BenchmarkCrawler_Run_100URLs_50workers(b *testing.B) {
	benchmarkCrawler(b, 50, 100, 10*time.Millisecond)
}
func BenchmarkCrawler_Run_100URLs_100workers(b *testing.B) {
	benchmarkCrawler(b, 100, 100, 10*time.Millisecond)
}

func BenchmarkCrawler_Run_10URLs_10fixedWorkers(b *testing.B) {
	benchmarkCrawler(b, 10, 10, 10*time.Millisecond)
}
func BenchmarkCrawler_Run_100URLs_10fixedWorkers(b *testing.B) {
	benchmarkCrawler(b, 10, 100, 10*time.Millisecond)
}
func BenchmarkCrawler_Run_1000URLs_10fixedWorkers(b *testing.B) {
	benchmarkCrawler(b, 10, 1000, 10*time.Millisecond)
}
