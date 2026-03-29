package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
)

type Config struct {
	Level  slog.Level
	Output io.Writer
	Format string
}

type Logger struct {
	Logger *slog.Logger
}

func New(cfg Config) *Logger {
	if cfg.Output == nil {
		cfg.Output = os.Stderr
	}

	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(cfg.Output, &slog.HandlerOptions{
			Level: cfg.Level,
		})
	} else {
		handler = slog.NewTextHandler(cfg.Output, &slog.HandlerOptions{
			Level: cfg.Level,
		})
	}
	return &Logger{
		Logger: slog.New(handler),
	}
}

func Default() *Logger {
	return New(Config{
		Level:  slog.LevelInfo,
		Format: "text",
	})
}

func (l *Logger) Debug(msg string, args ...any) {
	l.Logger.Debug(msg, args...)
}

func (l *Logger) Info(msg string, args ...any) {
	l.Logger.Info(msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.Logger.Warn(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.Logger.Error(msg, args...)
}

func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.Logger.DebugContext(ctx, msg, args...)
}

func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.Logger.InfoContext(ctx, msg, args...)
}

func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.Logger.WarnContext(ctx, msg, args...)
}

func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.Logger.ErrorContext(ctx, msg, args...)
}
