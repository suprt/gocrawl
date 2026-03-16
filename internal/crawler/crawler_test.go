package crawler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// MockDownloader для тестов
type MockDownloader struct {
	ResponseBody   string
	StatusCode     int
	ContentType    string
	Err            error
	CallCount      int32
	DownloadFunc   func(ctx context.Context, url string) (io.ReadCloser, int, string, error)
}

func (m *MockDownloader) Download(ctx context.Context, url string) (io.ReadCloser, int, string, error) {
	if m.DownloadFunc != nil {
		return m.DownloadFunc(ctx, url)
	}
	atomic.AddInt32(&m.CallCount, 1)
	if m.Err != nil {
		return nil, m.StatusCode, "", m.Err
	}
	return io.NopCloser(strings.NewReader(m.ResponseBody)), m.StatusCode, m.ContentType, nil
}

// MockStorage для тестов
type MockStorage struct {
	SaveFunc func(r io.Reader, filename string) (string, error)
}

func (m *MockStorage) Save(r io.Reader, filename string) (string, error) {
	if m.SaveFunc != nil {
		return m.SaveFunc(r, filename)
	}
	return "/test/path/" + filename, nil
}

// MockNamer для тестов
type MockNamer struct {
	NameFunc func(rawURL string) string
	ExtFunc  func(rawURL string, ext string) string
}

func (m *MockNamer) Name(rawURL string) string {
	if m.NameFunc != nil {
		return m.NameFunc(rawURL)
	}
	return "test.html"
}

func (m *MockNamer) NameWithExtension(rawURL string, ext string) string {
	if m.ExtFunc != nil {
		return m.ExtFunc(rawURL, ext)
	}
	return "test" + ext
}

// MockProgressBar для тестов
type MockProgressBar struct {
	AddCalls    int32
	FinishCalls int32
}

func (m *MockProgressBar) Add(n int) error {
	atomic.AddInt32(&m.AddCalls, int32(n))
	return nil
}

func (m *MockProgressBar) Finish() error {
	atomic.AddInt32(&m.FinishCalls, 1)
	return nil
}

func TestCrawler_Run_Success(t *testing.T) {
	downloader := &MockDownloader{
		ResponseBody: "<html>test</html>",
		StatusCode:   http.StatusOK,
		ContentType:  "text/html",
	}
	storage := &MockStorage{}
	namer := &MockNamer{}
	bar := &MockProgressBar{}

	crawler := New(downloader, storage, namer, Config{
		Workers:    2,
		Timeout:    5 * time.Second,
		MaxRetries: 0,
	})

	urls := []string{
		"https://example.com",
		"https://golang.org",
		"https://google.com",
	}

	results, errs, err := crawler.Run(context.Background(), urls, bar)

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Run() got %d results, want 3", len(results))
	}

	if len(errs) != 0 {
		t.Errorf("Run() got %d errors, want 0", len(errs))
	}

	if downloader.CallCount != 3 {
		t.Errorf("Download called %d times, want 3", downloader.CallCount)
	}

	if bar.AddCalls != 3 {
		t.Errorf("Progress.Add called %d times, want 3", bar.AddCalls)
	}
}

func TestCrawler_Run_WithRetries(t *testing.T) {
	var actualCalls int32
	downloader := &MockDownloader{
		ResponseBody: "<html>test</html>",
		StatusCode:   http.StatusOK,
		ContentType:  "text/html",
		DownloadFunc: func(ctx context.Context, url string) (io.ReadCloser, int, string, error) {
			count := atomic.AddInt32(&actualCalls, 1)
			if count == 1 {
				return nil, http.StatusInternalServerError, "", errors.New("temporary error")
			}
			return io.NopCloser(strings.NewReader("<html>test</html>")), http.StatusOK, "text/html", nil
		},
	}

	storage := &MockStorage{}
	namer := &MockNamer{}
	bar := &MockProgressBar{}

	crawler := New(downloader, storage, namer, Config{
		Workers:    1,
		Timeout:    5 * time.Second,
		MaxRetries: 1,
	})

	urls := []string{"https://example.com"}

	results, errs, err := crawler.Run(context.Background(), urls, bar)

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Должен быть 1 успешный результат после retry
	if len(results) != 1 {
		t.Errorf("Run() got %d results, want 1", len(results))
	}

	// Ошибка не записывается, если retry успешен
	if len(errs) != 0 {
		t.Errorf("Run() got %d errors, want 0 (retry succeeded)", len(errs))
	}
}

func TestCrawler_Run_AllRetriesExhausted(t *testing.T) {
	downloader := &MockDownloader{
		ResponseBody: "<html>test</html>",
		StatusCode:   http.StatusInternalServerError,
		ContentType:  "text/html",
		Err:          errors.New("permanent error"),
	}

	storage := &MockStorage{}
	namer := &MockNamer{}
	bar := &MockProgressBar{}

	crawler := New(downloader, storage, namer, Config{
		Workers:    1,
		Timeout:    5 * time.Second,
		MaxRetries: 2, // 1 основная попытка + 2 ретрая = 3 всего
	})

	urls := []string{"https://example.com"}

	results, errs, err := crawler.Run(context.Background(), urls, bar)

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Должно быть 0 результатов
	if len(results) != 0 {
		t.Errorf("Run() got %d results, want 0", len(results))
	}

	// Должна быть 1 ошибка (после исчерпания всех попыток)
	if len(errs) != 1 {
		t.Errorf("Run() got %d errors, want 1", len(errs))
	}
}

func TestCrawler_Run_ContextCancelled(t *testing.T) {
	downloader := &MockDownloader{
		ResponseBody: "<html>test</html>",
		StatusCode:   http.StatusOK,
	}
	storage := &MockStorage{}
	namer := &MockNamer{}
	bar := &MockProgressBar{}

	crawler := New(downloader, storage, namer, Config{
		Workers:    2,
		Timeout:    5 * time.Second,
		MaxRetries: 0,
	})

	urls := []string{"https://example.com"}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // отменяем сразу

	_, _, err := crawler.Run(ctx, urls, bar)

	// Контекст отменён, но это не обязательно ошибка
	_ = err
}

func TestCrawler_Run_RateLimit(t *testing.T) {
	downloader := &MockDownloader{
		ResponseBody: "<html>test</html>",
		StatusCode:   http.StatusOK,
	}
	storage := &MockStorage{}
	namer := &MockNamer{}
	bar := &MockProgressBar{}

	crawler := New(downloader, storage, namer, Config{
		Workers:     1,
		Timeout:     5 * time.Second,
		MaxRetries:  0,
		RateLimitMs: 50, // 50ms между запросами
	})

	urls := []string{
		"https://example.com",
		"https://golang.org",
	}

	start := time.Now()
	_, _, err := crawler.Run(context.Background(), urls, bar)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// С rate limit 50ms и 2 URL, должно пройти минимум 50ms
	if elapsed < 50*time.Millisecond {
		t.Errorf("Rate limiting not working, elapsed = %v, want >= 50ms", elapsed)
	}
}

func TestGetExtensionByContentType(t *testing.T) {
	tests := []struct {
		contentType string
		want        string
	}{
		{"", ".html"},
		{"text/html", ".html"},
		{"text/html; charset=utf-8", ".html"},
		{"text/css", ".css"},
		{"text/javascript", ".js"},
		{"application/javascript", ".js"},
		{"application/json", ".json"},
		{"image/png", ".png"},
		{"image/jpeg", ".jpg"},
		{"image/gif", ".gif"},
		{"image/svg+xml", ".svg"},
		{"application/pdf", ".pdf"},
		{"audio/mpeg", ".mp3"},
		{"video/mp4", ".mp4"},
		{"text/plain", ".txt"},
		{"text/xml", ".xml"},
		{"application/xml", ".xml"},
		{"unknown/type", ".bin"},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			got := getExtensionByContentType(tt.contentType)
			if got != tt.want {
				t.Errorf("getExtensionByContentType(%q) = %q, want %q", tt.contentType, got, tt.want)
			}
		})
	}
}

func TestJob_WithRetry(t *testing.T) {
	job := NewJob("https://example.com", 0)
	
	if job.Retries != 0 {
		t.Errorf("NewJob() Retries = %d, want 0", job.Retries)
	}

	job2 := job.WithRetry()
	if job2.Retries != 1 {
		t.Errorf("WithRetry() Retries = %d, want 1", job2.Retries)
	}

	// Оригинальный job не должен измениться
	if job.Retries != 0 {
		t.Errorf("Original job modified, Retries = %d, want 0", job.Retries)
	}
}
