package crawler

import (
	"context"
	"io"
	"sync"
	"time"
)

type Storage interface {
	Save(r io.Reader, filename string) (string, error)
}
type Downloader interface {
	Download(ctx context.Context, url string) (io.ReadCloser, int, string, error)
}

type Namer interface {
	Name(rawURL string) string
	NameWithExtension(rawURL string, ext string) string
}

type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Error(format string, args ...interface{})
}

type ProgressBar interface {
	Add(n int) error
	Finish() error
}
type Config struct {
	Workers      int
	Timeout      time.Duration
	MaxRetries   int
	RateLimitMs  int // задержка между запросами в миллисекундах
	Logger       Logger
}

type Crawler struct {
	downloader Downloader
	storage    Storage
	namer      Namer
	config     Config
}

func New(downloader Downloader, storage Storage, namer Namer, cfg Config) *Crawler {
	return &Crawler{
		downloader: downloader,
		storage:    storage,
		namer:      namer,
		config:     cfg,
	}
}

func (c *Crawler) Run(ctx context.Context, urls []string, bar ProgressBar) ([]WorkerResult, []error, error) {
	if len(urls) == 0 {
		return []WorkerResult{}, []error{}, nil
	}

	jobs := make(chan Job, len(urls))
	results := make(chan WorkerResult, len(urls))
	errors := make(chan error, len(urls))
	retryJobs := make(chan Job, len(urls))

	// Отправляем все джобы в канал
	for i, url := range urls {
		jobs <- NewJob(url, i)
	}
	close(jobs)

	var wg sync.WaitGroup

	// Запускаем воркеры
	for i := 0; i < c.config.Workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			c.worker(ctx, workerID, jobs, results, errors, retryJobs)
		}(i)
	}

	// Горутина для обработки retryJobs
	var retryWg sync.WaitGroup
	if c.config.MaxRetries > 0 {
		retryWg.Add(1)
		go func() {
			defer retryWg.Done()
			c.retryWorker(ctx, results, errors, retryJobs)
		}()
	}

	// Ожидаем завершения всех воркеров
	wg.Wait()
	close(retryJobs)
	
	// Ожидаем завершения retryWorker
	retryWg.Wait()
	
	// Закрываем каналы результатов
	close(results)
	close(errors)

	var allResults []WorkerResult
	var allErrors []error

	// Собираем результаты
	for result := range results {
		allResults = append(allResults, result)
		if bar != nil {
			_ = bar.Add(1)
		}
	}
	for err := range errors {
		allErrors = append(allErrors, err)
		if bar != nil {
			_ = bar.Add(1)
		}
	}

	return allResults, allErrors, nil
}
