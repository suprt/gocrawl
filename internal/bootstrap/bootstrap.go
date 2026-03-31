package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/suprt/gocrawl/internal/config"
	"github.com/suprt/gocrawl/internal/crawler"
	"github.com/suprt/gocrawl/internal/downloader"
	"github.com/suprt/gocrawl/internal/logger"
	"github.com/suprt/gocrawl/internal/naming"
	"github.com/suprt/gocrawl/internal/parser"
	"github.com/suprt/gocrawl/internal/progress"
	"github.com/suprt/gocrawl/internal/storage"
)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

type App struct {
	Config   *config.Config
	Crawler  *crawler.Crawler
	Parser   *parser.Parser
	Progress crawler.ProgressBar
	Logger   Logger
}

func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	logLevel := slog.LevelInfo
	if cfg.Verbose {
		logLevel = slog.LevelDebug
	}
	log := logger.New(logger.Config{
		Level:  logLevel,
		Format: "text",
		Output: os.Stderr,
	})

	storageClient, err := storage.New(cfg.OutputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}
	downloaderClient := downloader.New(downloader.Config{
		Timeout:      cfg.Timeout,
		MaxBodySize:  10 * 1024 * 1024,
		MaxRedirects: 10,
	})

	namerClient := naming.NewReadableNamer(50)

	parserClient := parser.New(nil)

	crawlerClient := crawler.New(
		downloaderClient, storageClient, namerClient, crawler.Config{
			Workers:     cfg.Workers,
			Timeout:     cfg.Timeout,
			MaxRetries:  cfg.MaxRetries,
			RateLimitMs: cfg.RateLimitMs,
		},
	)

	var bar crawler.ProgressBar

	return &App{
		Config:   cfg,
		Crawler:  crawlerClient,
		Parser:   parserClient,
		Progress: bar,
		Logger:   log,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	a.Logger.Info("starting crawler",
		"workers", a.Config.Workers,
		"retries", a.Config.MaxRetries,
		"timeout", a.Config.Timeout.String())

	urls, err := a.Parser.Parse(parser.Config{
		FilePath:    a.Config.FilePath,
		URLs:        a.Config.URLs,
		SkipInvalid: true,
		Deduplicate: true,
	})

	if err != nil {
		a.Logger.Error("failed to parse URLs", "error", err)
		return fmt.Errorf("failed to parse URLs: %w", err)
	}
	a.Logger.Debug("parsed URLs", "count", len(urls))
	for i, url := range urls {
		a.Logger.Debug("URL", "index", i, "url", url)
	}

	// Инициализируем progress bar с правильным количеством URL только если показан прогресс
	if a.Config.ShowProgress {
		a.Progress = progress.New(len(urls), "Downloading")
		defer func(pb crawler.ProgressBar) {
			_ = pb.Finish()
		}(a.Progress)
	}
	results, errors, err := a.Crawler.Run(ctx, urls, a.Progress)
	if err != nil {
		a.Logger.Error("Crawler failed", "error", err)
		return fmt.Errorf("failed to crawl: %w", err)
	}

	for _, result := range results {
		a.Logger.Debug("Downloaded",
			"url", result.Job.URL,
			"path", result.FilePath,
			"status", result.StatusCode,
			"duration", result.Duration.String())
	}
	for _, err := range errors {
		a.Logger.Debug("Download failed", "error", err)
	}

	a.printStats(results, errors)

	return nil

}

func (a *App) printStats(results []crawler.WorkerResult, errors []error) {
	a.Logger.Info("Crawler finished",
		"success", len(results),
		"failed", len(errors))

	if len(errors) > 0 {
		errorCount := make(map[string]int)
		for _, err := range errors {
			errorCount[err.Error()]++
		}
		for errMsg, count := range errorCount {

			a.Logger.Error("Job failed", "error", errMsg, "count", count)
		}
	}
}

func WaitForSignal() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	// Обработка сигнала прерывания, при двойной отправке Ctrl+C - немедленный выход
	go func() {
		defer signal.Stop(sigCh)
		<-sigCh
		cancel()
		<-sigCh
		os.Exit(1)
	}()
	return ctx
}
