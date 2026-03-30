package crawler

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type WorkerResult struct {
	Job        Job
	FilePath   string
	Duration   time.Duration
	StatusCode int
}

func (c *Crawler) worker(ctx context.Context, id int, jobs <-chan Job, results chan<- WorkerResult, errors chan<- error, retryJobs chan Job, bar ProgressBar) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}
			// Rate limiting
			if c.config.RateLimitMs > 0 {
				time.Sleep(time.Duration(c.config.RateLimitMs) * time.Millisecond)
			}

			result, err := c.process(ctx, id, job)
			if err != nil {
				// Если есть ретраи — отправляем в retryJobs, иначе сразу в errors
				if job.Retries < c.config.MaxRetries {
					select {
					case <-ctx.Done():
						return
					case retryJobs <- job.WithRetry():
					}
				} else {
					// Ретраи исчерпаны или отключены — пишем ошибку
					select {
					case <-ctx.Done():
						return
					case errors <- fmt.Errorf("worker %d failed job %d after %d attempts: %w", id, job.Index, job.Retries+1, err):
					}
					if bar != nil {
						_ = bar.Add(1)
					}
				}
				continue
			}

			select {
			case <-ctx.Done():
				return
			case results <- result:
			}
			if bar != nil {
				_ = bar.Add(1)
			}
		}
	}

}

func (c *Crawler) process(ctx context.Context, workerID int, job Job) (WorkerResult, error) {
	start := time.Now()

	body, statusCode, contentType, err := c.downloader.Download(ctx, job.URL)
	if err != nil {
		return WorkerResult{}, fmt.Errorf("downloading %s failed: %s", job.URL, err)
	}
	defer func() {
		_ = body.Close()
	}()

	ext := getExtensionByContentType(contentType)
	filename := c.namer.NameWithExtension(job.URL, ext)

	filepath, err := c.storage.Save(body, filename)
	if err != nil {
		return WorkerResult{}, fmt.Errorf("saving %s failed: %s", job.URL, err)
	}

	return WorkerResult{
		Job:        job,
		FilePath:   filepath,
		Duration:   time.Since(start),
		StatusCode: statusCode,
	}, nil

}

// getExtensionByContentType определяет расширение файла по Content-Type
func getExtensionByContentType(contentType string) string {
	switch {
	case contentType == "":
		return ".html"
	case strings.Contains(contentType, "text/html"):
		return ".html"
	case strings.Contains(contentType, "text/css"):
		return ".css"
	case strings.Contains(contentType, "text/javascript"):
		return ".js"
	case strings.Contains(contentType, "application/javascript"):
		return ".js"
	case strings.Contains(contentType, "application/json"):
		return ".json"
	case strings.Contains(contentType, "image/png"):
		return ".png"
	case strings.Contains(contentType, "image/jpeg"):
		return ".jpg"
	case strings.Contains(contentType, "image/gif"):
		return ".gif"
	case strings.Contains(contentType, "image/svg"):
		return ".svg"
	case strings.Contains(contentType, "image/webp"):
		return ".webp"
	case strings.Contains(contentType, "application/pdf"):
		return ".pdf"
	case strings.Contains(contentType, "audio/"):
		return ".mp3"
	case strings.Contains(contentType, "video/"):
		return ".mp4"
	case strings.Contains(contentType, "text/plain"):
		return ".txt"
	case strings.Contains(contentType, "text/xml"):
		return ".xml"
	case strings.Contains(contentType, "application/xml"):
		return ".xml"
	default:
		return ".bin"
	}
}

func (c *Crawler) retryWorker(ctx context.Context, results chan<- WorkerResult, errors chan<- error, retryJobs chan Job, bar ProgressBar) {
	for job := range retryJobs {
		var result WorkerResult
		var err error
		// attempts показывает текущую попытку (начиная с job.Retries)
		attempts := job.Retries

		// Повторяем попытки пока не исчерпаем лимит
		// job.Retries = сколько раз уже пытались, MaxRetries = сколько ещё можно
		for attempts <= c.config.MaxRetries {
			// Экспоненциальная задержка перед попыткой
			// При attempts=0 или 1 задержка = 1s, при attempts=2 = 2s, при attempts=3 = 4s
			delay := time.Second
			if attempts > 1 {
				delay = time.Second * time.Duration(1<<uint(attempts-1))
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(delay):
			}

			result, err = c.process(ctx, -1, job)
			if err == nil {
				// Успех — отправляем результат
				select {
				case <-ctx.Done():
					return
				case results <- result:
				}
				if bar != nil {
					_ = bar.Add(1)
				}
				break
			}

			attempts++
		}

		// Если ошибка и все попытки исчерпаны — пишем ошибку
		if err != nil {
			select {
			case <-ctx.Done():
				return
			case errors <- fmt.Errorf("job %d failed after %d attempts: %w", job.Index, attempts, err):
			}
			if bar != nil {
				_ = bar.Add(1)
			}
		}
	}
}
