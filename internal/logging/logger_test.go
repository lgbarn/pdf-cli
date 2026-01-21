package logging

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
	}{
		{"debug", LevelDebug},
		{"DEBUG", LevelDebug},
		{"Debug", LevelDebug},
		{"info", LevelInfo},
		{"INFO", LevelInfo},
		{"warn", LevelWarn},
		{"WARN", LevelWarn},
		{"warning", LevelWarn},
		{"WARNING", LevelWarn},
		{"error", LevelError},
		{"ERROR", LevelError},
		{"silent", LevelSilent},
		{"SILENT", LevelSilent},
		{"none", LevelSilent},
		{"off", LevelSilent},
		{"invalid", LevelSilent},
		{"", LevelSilent},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseLevel(tt.input)
			if got != tt.expected {
				t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected Format
	}{
		{"json", FormatJSON},
		{"JSON", FormatJSON},
		{"Json", FormatJSON},
		{"text", FormatText},
		{"TEXT", FormatText},
		{"human", FormatText},
		{"HUMAN", FormatText},
		{"invalid", FormatText},
		{"", FormatText},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseFormat(tt.input)
			if got != tt.expected {
				t.Errorf("ParseFormat(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestLevelToSlogLevel(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{LevelSilent, "ERROR+1"},
	}

	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			got := tt.level.ToSlogLevel()
			// Just verify the levels are correctly ordered
			if tt.level == LevelSilent {
				// Silent should be higher than error
				if got <= LevelError.ToSlogLevel() {
					t.Errorf("Silent level should be higher than error")
				}
			}
		})
	}
}

func TestLoggerText(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LevelDebug, FormatText, &buf)

	logger.Info("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Expected 'test message' in output, got: %s", output)
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("Expected 'key=value' in output, got: %s", output)
	}
}

func TestLoggerJSON(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LevelDebug, FormatJSON, &buf)

	logger.Info("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, `"msg":"test message"`) {
		t.Errorf("Expected JSON msg field, got: %s", output)
	}
	if !strings.Contains(output, `"key":"value"`) {
		t.Errorf("Expected JSON key field, got: %s", output)
	}
}

func TestLoggerSilent(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LevelSilent, FormatText, &buf)

	logger.Debug("should not appear")
	logger.Info("should not appear")
	logger.Warn("should not appear")
	logger.Error("should not appear")

	if buf.Len() > 0 {
		t.Errorf("Silent logger should produce no output, got: %s", buf.String())
	}
}

func TestLoggerLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LevelWarn, FormatText, &buf)

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	output := buf.String()
	if strings.Contains(output, "debug message") {
		t.Error("Debug should be filtered")
	}
	if strings.Contains(output, "info message") {
		t.Error("Info should be filtered")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("Warn should appear")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Error should appear")
	}
}

func TestLoggerLevelFilteringAtError(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LevelError, FormatText, &buf)

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	output := buf.String()
	if strings.Contains(output, "debug message") {
		t.Error("Debug should be filtered")
	}
	if strings.Contains(output, "info message") {
		t.Error("Info should be filtered")
	}
	if strings.Contains(output, "warn message") {
		t.Error("Warn should be filtered at error level")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Error should appear")
	}
}

func TestGlobalLogger(t *testing.T) {
	Reset()

	// First call should initialize with defaults
	l := Get()
	if l == nil {
		t.Fatal("Get() should return a logger")
	}

	// Should be silent by default
	if l.Level() != LevelSilent {
		t.Errorf("Default level should be silent, got: %v", l.Level())
	}

	// Subsequent calls should return the same instance
	l2 := Get()
	if l != l2 {
		t.Error("Get() should return the same instance")
	}

	Reset()
}

func TestInit(t *testing.T) {
	Reset()

	Init(LevelDebug, FormatJSON)
	l := Get()

	if l.Level() != LevelDebug {
		t.Errorf("Expected level debug, got: %v", l.Level())
	}
	if l.Format() != FormatJSON {
		t.Errorf("Expected format json, got: %v", l.Format())
	}

	Reset()
}

func TestWith(t *testing.T) {
	Reset()

	var buf bytes.Buffer
	Init(LevelDebug, FormatText)
	global.Logger = New(LevelDebug, FormatText, &buf).Logger

	logger := With("operation", "test")
	logger.Info("message")

	output := buf.String()
	if !strings.Contains(output, "operation=test") {
		t.Errorf("Expected 'operation=test' in output, got: %s", output)
	}

	Reset()
}

func TestGlobalHelpers(t *testing.T) {
	Reset()

	var buf bytes.Buffer
	logger := New(LevelDebug, FormatText, &buf)
	global = logger

	Debug("debug msg", "k1", "v1")
	Info("info msg", "k2", "v2")
	Warn("warn msg", "k3", "v3")
	Error("error msg", "k4", "v4")

	output := buf.String()
	if !strings.Contains(output, "debug msg") {
		t.Error("Debug message should appear")
	}
	if !strings.Contains(output, "info msg") {
		t.Error("Info message should appear")
	}
	if !strings.Contains(output, "warn msg") {
		t.Error("Warn message should appear")
	}
	if !strings.Contains(output, "error msg") {
		t.Error("Error message should appear")
	}

	Reset()
}

func TestLoggerAccessors(t *testing.T) {
	logger := New(LevelInfo, FormatJSON, &bytes.Buffer{})

	if logger.Level() != LevelInfo {
		t.Errorf("Level() = %v, want %v", logger.Level(), LevelInfo)
	}
	if logger.Format() != FormatJSON {
		t.Errorf("Format() = %v, want %v", logger.Format(), FormatJSON)
	}
}

func TestLoggerWithMultipleAttributes(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LevelDebug, FormatText, &buf)

	logger.Info("test", "key1", "value1", "key2", 42, "key3", true)

	output := buf.String()
	if !strings.Contains(output, "key1=value1") {
		t.Errorf("Expected 'key1=value1' in output, got: %s", output)
	}
	if !strings.Contains(output, "key2=42") {
		t.Errorf("Expected 'key2=42' in output, got: %s", output)
	}
	if !strings.Contains(output, "key3=true") {
		t.Errorf("Expected 'key3=true' in output, got: %s", output)
	}
}
