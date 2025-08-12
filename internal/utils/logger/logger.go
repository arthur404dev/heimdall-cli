package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
)

var (
	// Global logger instance
	Log *slog.Logger

	// Log level
	level = new(slog.LevelVar)
)

func init() {
	// Set default log level
	level.Set(slog.LevelInfo)

	// Create default logger with text handler for console
	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Customize time format
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Value = slog.StringValue(t.Format("15:04:05.000"))
				}
			}
			// Shorten source file paths
			if a.Key == slog.SourceKey {
				if src, ok := a.Value.Any().(*slog.Source); ok {
					a.Value = slog.StringValue(fmt.Sprintf("%s:%d", filepath.Base(src.File), src.Line))
				}
			}
			return a
		},
	}

	handler := slog.NewTextHandler(os.Stderr, opts)
	Log = slog.New(handler)
	slog.SetDefault(Log)
}

// SetLevel sets the global log level
func SetLevel(l slog.Level) {
	level.Set(l)
}

// SetDebug enables debug logging
func SetDebug(debug bool) {
	if debug {
		level.Set(slog.LevelDebug)
	} else {
		level.Set(slog.LevelInfo)
	}
}

// SetVerbose enables verbose logging
func SetVerbose(verbose bool) {
	if verbose {
		level.Set(slog.LevelDebug)
	}
}

// SetupFileLogging sets up file logging in addition to console
func SetupFileLogging(logFile string) error {
	if logFile == "" {
		logFile = filepath.Join(paths.HeimdallStateDir, "heimdall.log")
	}

	// Ensure log directory exists
	if err := paths.EnsureParentDir(logFile); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Create JSON handler for file
	fileOpts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	}
	fileHandler := slog.NewJSONHandler(file, fileOpts)

	// Create text handler for console
	consoleOpts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Value = slog.StringValue(t.Format("15:04:05.000"))
				}
			}
			return a
		},
	}
	consoleHandler := slog.NewTextHandler(os.Stderr, consoleOpts)

	// Create multi-handler that writes to both
	multiHandler := &MultiHandler{
		handlers: []slog.Handler{consoleHandler, fileHandler},
	}

	Log = slog.New(multiHandler)
	slog.SetDefault(Log)

	return nil
}

// MultiHandler writes to multiple handlers
type MultiHandler struct {
	handlers []slog.Handler
}

func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, handler := range h.handlers {
		if err := handler.Handle(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: handlers}
}

func (h *MultiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return &MultiHandler{handlers: handlers}
}

// Helper functions for common logging patterns

// Debug logs a debug message
func Debug(msg string, args ...any) {
	Log.Debug(msg, args...)
}

// Info logs an info message
func Info(msg string, args ...any) {
	Log.Info(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	Log.Warn(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...any) {
	Log.Error(msg, args...)
}

// Fatal logs an error message and exits
func Fatal(msg string, args ...any) {
	Log.Error(msg, args...)
	os.Exit(1)
}
