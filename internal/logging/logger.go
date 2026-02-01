package logging

import (
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
)

// Level represents a log level.
type Level string

const (
	LevelDebug  Level = "debug"
	LevelInfo   Level = "info"
	LevelWarn   Level = "warn"
	LevelError  Level = "error"
	LevelSilent Level = "silent"
)

// Format represents a log format.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// ParseLevel parses a log level string.
func ParseLevel(s string) Level {
	switch strings.ToLower(s) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	case "silent", "none", "off":
		return LevelSilent
	default:
		return LevelSilent
	}
}

// ParseFormat parses a log format string.
func ParseFormat(s string) Format {
	switch strings.ToLower(s) {
	case "json":
		return FormatJSON
	case "text", "human":
		return FormatText
	default:
		return FormatText
	}
}

// ToSlogLevel converts Level to slog.Level.
func (l Level) ToSlogLevel() slog.Level {
	switch l {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelError + 1 // Higher than error = silent
	}
}

// Logger wraps slog for pdf-cli logging.
type Logger struct {
	*slog.Logger
	level  Level
	format Format
}

// global is the global logger instance.
var global *Logger
var globalMu sync.RWMutex

// Init initializes the global logger with the given level and format.
func Init(level Level, format Format) {
	globalMu.Lock()
	defer globalMu.Unlock()
	global = New(level, format, os.Stderr)
}

// New creates a new logger with the given level, format, and output.
func New(level Level, format Format, w io.Writer) *Logger {
	if level == LevelSilent {
		// Create a logger that discards everything
		return &Logger{
			Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			level:  level,
			format: format,
		}
	}

	opts := &slog.HandlerOptions{
		Level: level.ToSlogLevel(),
	}

	var handler slog.Handler
	if format == FormatJSON {
		handler = slog.NewJSONHandler(w, opts)
	} else {
		handler = slog.NewTextHandler(w, opts)
	}

	return &Logger{
		Logger: slog.New(handler),
		level:  level,
		format: format,
	}
}

// Get returns the global logger, initializing with defaults if needed.
func Get() *Logger {
	globalMu.RLock()
	if global != nil {
		defer globalMu.RUnlock()
		return global
	}
	globalMu.RUnlock()

	globalMu.Lock()
	defer globalMu.Unlock()

	if global != nil {
		return global
	}

	global = New(LevelSilent, FormatText, os.Stderr)
	return global
}

// Reset resets the global logger (for testing).
func Reset() {
	globalMu.Lock()
	defer globalMu.Unlock()
	global = nil
}

// Level returns the logger's level.
func (l *Logger) Level() Level {
	return l.level
}

// Format returns the logger's format.
func (l *Logger) Format() Format {
	return l.format
}

// Debug logs at debug level.
func Debug(msg string, args ...any) {
	Get().Debug(msg, args...)
}

// Info logs at info level.
func Info(msg string, args ...any) {
	Get().Info(msg, args...)
}

// Warn logs at warn level.
func Warn(msg string, args ...any) {
	Get().Warn(msg, args...)
}

// Error logs at error level.
func Error(msg string, args ...any) {
	Get().Error(msg, args...)
}

// With returns a logger with additional attributes.
func With(args ...any) *Logger {
	l := Get()
	return &Logger{
		Logger: l.With(args...),
		level:  l.level,
		format: l.format,
	}
}
