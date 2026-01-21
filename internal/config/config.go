package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration.
type Config struct {
	Defaults DefaultsConfig `yaml:"defaults"`
	Compress CompressConfig `yaml:"compress"`
	Encrypt  EncryptConfig  `yaml:"encrypt"`
	OCR      OCRConfig      `yaml:"ocr"`
}

// DefaultsConfig holds default settings.
type DefaultsConfig struct {
	OutputFormat string `yaml:"output_format"` // json, csv, tsv, human
	Verbose      bool   `yaml:"verbose"`
	ShowProgress bool   `yaml:"show_progress"`
}

// CompressConfig holds compression settings.
type CompressConfig struct {
	Quality string `yaml:"quality"` // low, medium, high
}

// EncryptConfig holds encryption settings.
type EncryptConfig struct {
	Algorithm string `yaml:"algorithm"` // aes128, aes256
}

// OCRConfig holds OCR settings.
type OCRConfig struct {
	Language string `yaml:"language"` // eng, deu, fra, etc.
	Backend  string `yaml:"backend"`  // auto, native, wasm
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() *Config {
	return &Config{
		Defaults: DefaultsConfig{
			OutputFormat: "human",
			Verbose:      false,
			ShowProgress: true,
		},
		Compress: CompressConfig{
			Quality: "medium",
		},
		Encrypt: EncryptConfig{
			Algorithm: "aes256",
		},
		OCR: OCRConfig{
			Language: "eng",
			Backend:  "auto",
		},
	}
}

// ConfigPath returns the path to the config file.
func ConfigPath() string {
	// Check XDG_CONFIG_HOME first
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		// Fall back to ~/.config
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, "pdf-cli", "config.yaml")
}

// Load reads the config file and returns the configuration.
// If the file doesn't exist, returns default config.
// Environment variables can override config file values.
func Load() (*Config, error) {
	cfg := DefaultConfig()

	path := ConfigPath()
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
			// File doesn't exist, continue with defaults
		} else {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, err
			}
		}
	}

	// Always apply environment overrides
	applyEnvOverrides(cfg)

	return cfg, nil
}

// applyEnvOverrides applies environment variable overrides to the config.
func applyEnvOverrides(cfg *Config) {
	if env := os.Getenv("PDF_CLI_OUTPUT_FORMAT"); env != "" {
		cfg.Defaults.OutputFormat = env
	}
	if env := os.Getenv("PDF_CLI_VERBOSE"); env == "true" || env == "1" {
		cfg.Defaults.Verbose = true
	}
	if env := os.Getenv("PDF_CLI_OCR_LANGUAGE"); env != "" {
		cfg.OCR.Language = env
	}
	if env := os.Getenv("PDF_CLI_OCR_BACKEND"); env != "" {
		cfg.OCR.Backend = env
	}
}

// Save writes the config to the config file.
func Save(cfg *Config) error {
	path := ConfigPath()
	if path == "" {
		return nil
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// global holds the loaded configuration
var global *Config

// Get returns the global configuration, loading it if necessary.
func Get() *Config {
	if global == nil {
		var err error
		global, err = Load()
		if err != nil {
			// Fall back to defaults on error
			global = DefaultConfig()
		}
	}
	return global
}

// Reset clears the global config (useful for testing).
func Reset() {
	global = nil
}
