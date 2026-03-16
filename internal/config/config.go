package config

import (
	"errors"
	"flag"
	"os"
	"strconv"
	"time"
)

type Config struct {
	FilePath     string
	URLs         []string
	OutputDir    string
	Workers      int
	Timeout      time.Duration
	MaxRetries   int
	ShowProgress bool
	RateLimitMs  int
	Verbose      bool
}

func Load() (*Config, error) {
	filePath := flag.String("file", "", "File with URLs (one per line)")
	outputDir := flag.String("output", "./downloads", "Output directory")
	workers := flag.Int("workers", 5, "Number of workers")
	timeout := flag.Duration("timeout", 30*time.Second, "Timeout for each download")
	maxRetries := flag.Int("retries", 5, "Maximum number of retries for failed downloads")
	showProgress := flag.Bool("progress", true, "Show progress bar")
	rateLimitMs := flag.Int("rate-limit", 0, "Rate limit in milliseconds between requests")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")

	flag.Parse()

	// URL из аргументов командной строки (после flag.Parse())
	urls := flag.Args()

	if envWorkers := os.Getenv("GOCRAWL_WORKERS"); envWorkers != "" {
		if w, err := strconv.Atoi(envWorkers); err == nil {
			if w > 0 {
				*workers = w
			}
		}
	}

	if envTimeout := os.Getenv("GOCRAWL_TIMEOUT"); envTimeout != "" {
		if t, err := time.ParseDuration(envTimeout); err == nil {
			if t > 0 {
				*timeout = t
			}
		}
	}

	if envRetries := os.Getenv("GOCRAWL_RETRIES"); envRetries != "" {
		if r, err := strconv.Atoi(envRetries); err == nil {
			if r < 0 {
				r = 0
			}
			*maxRetries = r
		}
	}

	if *filePath == "" && len(urls) == 0 {
		return nil, errors.New("no URLs provided: use -file or provide URLs as arguments")
	}

	return &Config{
		FilePath:     *filePath,
		URLs:         urls,
		OutputDir:    *outputDir,
		Workers:      *workers,
		Timeout:      *timeout,
		MaxRetries:   *maxRetries,
		ShowProgress: *showProgress,
		RateLimitMs:  *rateLimitMs,
		Verbose:      *verbose,
	}, nil
}
