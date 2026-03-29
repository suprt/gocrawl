package bootstrap

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/suprt/gocrawl/internal/config"
	"github.com/suprt/gocrawl/internal/crawler"
	"github.com/suprt/gocrawl/internal/downloader"
	"github.com/suprt/gocrawl/internal/naming"
	"github.com/suprt/gocrawl/internal/parser"
	"github.com/suprt/gocrawl/internal/progress"
	"github.com/suprt/gocrawl/internal/storage"
)

// simpleLogger реализует интерфейс crawler.Logger
type simpleLogger struct {
	verbose bool
	out     io.Writer
	errOut  io.Writer
}

func (l *simpleLogger) Debug(format string, args ...interface{}) {
	if l.verbose {
		fmt.Fprintf(l.out, "[DEBUG] "+format+"\n", args...)
	}
}

func (l *simpleLogger) Info(format string, args ...interface{}) {
	fmt.Fprintf(l.out, "[INFO] "+format+"\n", args...)
}

func (l *simpleLogger) Error(format string, args ...interface{}) {
	fmt.Fprintf(l.errOut, "[ERROR] "+format+"\n", args...)
}

type App struct {
	Config     *config.Config
	Crawler    *crawler.Crawler
	Parser     *parser.Parser
	Progress   crawler.ProgressBar
	CancelFunc context.CancelFunc
}

func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

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

	// Создаём логгер
	logger := &simpleLogger{
		verbose: cfg.Verbose,
		out:     os.Stdout,
		errOut:  os.Stderr,
	}

	crawlerClient := crawler.New(
		downloaderClient, storageClient, namerClient, crawler.Config{
			Workers:     cfg.Workers,
			Timeout:     cfg.Timeout,
			MaxRetries:  cfg.MaxRetries,
			RateLimitMs: cfg.RateLimitMs,
			Logger:      logger,
		},
	)

	// Progress bar создаётся только один раз в Run(), здесь только сохраняем флаг
	var bar crawler.ProgressBar
	if cfg.ShowProgress {
		bar = &progress.NoopBar{} // заглушка, будет заменена в Run()
	}

	return &App{
		Config:   cfg,
		Crawler:  crawlerClient,
		Parser:   parserClient,
		Progress: bar,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	if a.Config.Verbose {
		fmt.Printf("Starting crawler with %d workers, %d retries, timeout=%v\n",
			a.Config.Workers, a.Config.MaxRetries, a.Config.Timeout)
	}

	urls, err := a.Parser.Parse(parser.Config{
		FilePath:    a.Config.FilePath,
		URLs:        a.Config.URLs,
		SkipInvalid: true,
		Deduplicate: true,
	})

	if err != nil {
		return fmt.Errorf("failed to parse URLs: %w", err)
	}

	if a.Config.Verbose {
		fmt.Printf("Parsed %d URL(s)\n", len(urls))
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
		return fmt.Errorf("failed to crawl: %w", err)
	}

	a.printStats(results, errors)

	return nil

}

func (a *App) printStats(results []crawler.WorkerResult, errors []error) {
	fmt.Printf("\nStatistics:\n")
	fmt.Printf("  Success: %d\n", len(results))
	fmt.Printf("  Failed:  %d\n", len(errors))

	if len(errors) > 0 {
		errorCount := make(map[string]int)
		for _, err := range errors {
			errorCount[err.Error()]++
		}

		fmt.Printf("\nError summary:\n")
		for errMsg, count := range errorCount {
			fmt.Printf("  %s: %d\n", errMsg, count)
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
