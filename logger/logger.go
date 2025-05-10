package logger

import (
	"context"
	"log/slog"
	"os"

	"go.uber.org/fx"
)

// Logger defines the interface for logging
type Logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
	Debug(ctx context.Context, msg string, args ...any)
}

// SlogLogger is an implementation of Logger using slog
type SlogLogger struct {
	logger *slog.Logger
}

// NewSlogLogger creates a new SlogLogger
func NewSlogLogger() Logger {
	return &SlogLogger{
		logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug, // Ensure DEBUG level
		})),
	}
}

// Info logs an info message
func (l *SlogLogger) Info(ctx context.Context, msg string, args ...any) {
	l.logger.InfoContext(ctx, msg, args...)
}

// Error logs an error message
func (l *SlogLogger) Error(ctx context.Context, msg string, args ...any) {
	l.logger.ErrorContext(ctx, msg, args...)
}

// Debug logs a debug message
func (l *SlogLogger) Debug(ctx context.Context, msg string, args ...any) {
	l.logger.DebugContext(ctx, msg, args...)
}

// ProvideLogger is an fx-compatible constructor for Logger
func ProvideLogger() fx.Option {
	return fx.Provide(NewSlogLogger)
}
