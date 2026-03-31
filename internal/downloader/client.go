package downloader

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
)

type Client struct {
	httpClient  *http.Client
	userAgent   string
	maxBodySize int64
}

type Config struct {
	Timeout      time.Duration
	UserAgent    string
	MaxBodySize  int64
	MaxRedirects int
}

type limitReadCloser struct {
	io.Reader
	io.Closer
}

func New(cfg Config) *Client {
	var userAgent = []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = userAgent[rand.Intn(len(userAgent))]
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.MaxRedirects <= 0 {
		cfg.MaxRedirects = 10
	}

	c := &Client{
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= cfg.MaxRedirects {
					return fmt.Errorf("stopped after %d redirects", cfg.MaxRedirects)
				}
				return nil
			},
		},
		userAgent:   cfg.UserAgent,
		maxBodySize: cfg.MaxBodySize,
	}

	// Устанавливаем лимит размера тела по умолчанию (100 МБ), если не задан
	if c.maxBodySize == 0 {
		c.maxBodySize = 100 * 1024 * 1024 // 100 МБ
	}

	return c
}

func (c *Client) Download(ctx context.Context, url string) (io.ReadCloser, int, string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, http.StatusInternalServerError, "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, http.StatusInternalServerError, "", fmt.Errorf("error downloading %s: %w", url, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err := resp.Body.Close()
		if err != nil {
			return nil, http.StatusInternalServerError, "", err
		}
		return nil, resp.StatusCode, "", fmt.Errorf("bad status code: %s", resp.Status)
	}

	contentType := resp.Header.Get("Content-Type")

	buf := bufio.NewReader(resp.Body)
	peek, err := buf.Peek(1024)
	if err != nil && err != io.EOF {
		// io.EOF — нормальное состояние, если тело ответа короче 1024 байт
		_ = resp.Body.Close()
		return nil, resp.StatusCode, "", fmt.Errorf("error peeking %s: %w", url, err)
	}

	encoding, name, _ := charset.DetermineEncoding(peek, contentType)

	var reader io.Reader = buf
	if name != "utf-8" {
		reader = transform.NewReader(buf, encoding.NewDecoder())
	}

	var body io.ReadCloser
	if c.maxBodySize > 0 {
		body = &limitReadCloser{
			Reader: io.LimitReader(reader, c.maxBodySize),
			Closer: resp.Body,
		}
	} else {
		body = &struct {
			io.Reader
			io.Closer
		}{
			Reader: reader,
			Closer: resp.Body,
		}
	}

	return body, resp.StatusCode, contentType, nil
}

func (c *Client) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}
