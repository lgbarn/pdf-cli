package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Defaults.OutputFormat != "human" {
		t.Errorf("Expected output_format 'human', got %s", cfg.Defaults.OutputFormat)
	}
	if cfg.Defaults.ShowProgress != true {
		t.Error("Expected show_progress true")
	}
	if cfg.OCR.Language != "eng" {
		t.Errorf("Expected OCR language 'eng', got %s", cfg.OCR.Language)
	}
	if cfg.OCR.Backend != "auto" {
		t.Errorf("Expected OCR backend 'auto', got %s", cfg.OCR.Backend)
	}
}

func TestLoadNonExistent(t *testing.T) {
	// Set a non-existent config path
	os.Setenv("XDG_CONFIG_HOME", "/nonexistent/path")
	defer os.Unsetenv("XDG_CONFIG_HOME")
	Reset()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Should return defaults
	if cfg.Defaults.OutputFormat != "human" {
		t.Errorf("Expected default output_format, got %s", cfg.Defaults.OutputFormat)
	}
}

func TestLoadWithEnvOverride(t *testing.T) {
	os.Setenv("XDG_CONFIG_HOME", "/nonexistent/path")
	os.Setenv("PDF_CLI_OUTPUT_FORMAT", "json")
	os.Setenv("PDF_CLI_VERBOSE", "true")
	os.Setenv("PDF_CLI_OCR_LANGUAGE", "deu")
	defer func() {
		os.Unsetenv("XDG_CONFIG_HOME")
		os.Unsetenv("PDF_CLI_OUTPUT_FORMAT")
		os.Unsetenv("PDF_CLI_VERBOSE")
		os.Unsetenv("PDF_CLI_OCR_LANGUAGE")
	}()
	Reset()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Defaults.OutputFormat != "json" {
		t.Errorf("Expected output_format 'json', got %s", cfg.Defaults.OutputFormat)
	}
	if !cfg.Defaults.Verbose {
		t.Error("Expected verbose to be true")
	}
	if cfg.OCR.Language != "deu" {
		t.Errorf("Expected OCR language 'deu', got %s", cfg.OCR.Language)
	}
}

func TestLoadWithEnvOverrideNumericVerbose(t *testing.T) {
	os.Setenv("XDG_CONFIG_HOME", "/nonexistent/path")
	os.Setenv("PDF_CLI_VERBOSE", "1")
	defer func() {
		os.Unsetenv("XDG_CONFIG_HOME")
		os.Unsetenv("PDF_CLI_VERBOSE")
	}()
	Reset()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if !cfg.Defaults.Verbose {
		t.Error("Expected verbose to be true with PDF_CLI_VERBOSE=1")
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Create temp directory for config
	tmpDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")
	Reset()

	// Create custom config
	cfg := DefaultConfig()
	cfg.Defaults.OutputFormat = "csv"
	cfg.OCR.Language = "fra"

	// Save it
	if err := Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	path := filepath.Join(tmpDir, "pdf-cli", "config.yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("Config file should exist after save")
	}

	// Load it back
	Reset()
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Defaults.OutputFormat != "csv" {
		t.Errorf("Expected output_format 'csv', got %s", loaded.Defaults.OutputFormat)
	}
	if loaded.OCR.Language != "fra" {
		t.Errorf("Expected OCR language 'fra', got %s", loaded.OCR.Language)
	}
}

func TestGet(t *testing.T) {
	os.Setenv("XDG_CONFIG_HOME", "/nonexistent/path")
	defer os.Unsetenv("XDG_CONFIG_HOME")
	Reset()

	cfg1 := Get()
	cfg2 := Get()

	// Should return the same instance
	if cfg1 != cfg2 {
		t.Error("Get should return the same config instance")
	}
}

func TestConfigPath(t *testing.T) {
	// Test with XDG_CONFIG_HOME set
	os.Setenv("XDG_CONFIG_HOME", "/custom/config")
	defer os.Unsetenv("XDG_CONFIG_HOME")

	path := ConfigPath()
	expected := "/custom/config/pdf-cli/config.yaml"
	if path != expected {
		t.Errorf("ConfigPath() = %s, want %s", path, expected)
	}
}

func TestConfigPathDefaultsToHome(t *testing.T) {
	// Unset XDG_CONFIG_HOME to test fallback
	os.Unsetenv("XDG_CONFIG_HOME")

	path := ConfigPath()
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Could not get home directory")
	}

	expected := filepath.Join(home, ".config", "pdf-cli", "config.yaml")
	if path != expected {
		t.Errorf("ConfigPath() = %s, want %s", path, expected)
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")
	Reset()

	// Create invalid YAML file
	configDir := filepath.Join(tmpDir, "pdf-cli")
	if err := os.MkdirAll(configDir, 0750); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("invalid: yaml: ["), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := Load()
	if err == nil {
		t.Error("Load should fail with invalid YAML")
	}
}

func TestLoadWithEnvBackendOverride(t *testing.T) {
	os.Setenv("XDG_CONFIG_HOME", "/nonexistent/path")
	os.Setenv("PDF_CLI_OCR_BACKEND", "wasm")
	defer func() {
		os.Unsetenv("XDG_CONFIG_HOME")
		os.Unsetenv("PDF_CLI_OCR_BACKEND")
	}()
	Reset()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.OCR.Backend != "wasm" {
		t.Errorf("Expected OCR backend 'wasm', got %s", cfg.OCR.Backend)
	}
}

func TestDefaultConfigValues(t *testing.T) {
	cfg := DefaultConfig()

	// Test all default values
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"Defaults.OutputFormat", cfg.Defaults.OutputFormat, "human"},
		{"Compress.Quality", cfg.Compress.Quality, "medium"},
		{"Encrypt.Algorithm", cfg.Encrypt.Algorithm, "aes256"},
		{"OCR.Language", cfg.OCR.Language, "eng"},
		{"OCR.Backend", cfg.OCR.Backend, "auto"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %s, want %s", tt.name, tt.got, tt.expected)
			}
		})
	}

	// Test boolean defaults
	if cfg.Defaults.Verbose != false {
		t.Error("Expected Defaults.Verbose to be false")
	}
	if cfg.Defaults.ShowProgress != true {
		t.Error("Expected Defaults.ShowProgress to be true")
	}
}

func TestReset(t *testing.T) {
	os.Setenv("XDG_CONFIG_HOME", "/nonexistent/path")
	defer os.Unsetenv("XDG_CONFIG_HOME")
	Reset()

	cfg1 := Get()
	Reset()
	cfg2 := Get()

	// After reset, Get should return a new instance
	if cfg1 == cfg2 {
		t.Error("After Reset, Get should return a new config instance")
	}
}
