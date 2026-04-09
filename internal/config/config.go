package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/adrg/xdg"
)

const (
	ProviderLocal = "local"

	DefaultLocalModel     = "gemma-4-e2b-it"
	DefaultLocalContext   = 2048
	DefaultLocalGPULayers = 0
	DefaultLocalModelFile = "google_gemma-4-E2B-it-IQ2_M.gguf"
	DefaultLocalModelURL  = "https://huggingface.co/bartowski/google_gemma-4-E2B-it-GGUF/resolve/main/google_gemma-4-E2B-it-IQ2_M.gguf?download=true"
)

type Config struct {
	Provider         string `json:"provider"`
	Model            string `json:"model"`
	LocalModelPath   string `json:"local_model_path,omitempty"`
	LocalContextSize int    `json:"local_context_size,omitempty"`
	LocalThreads     int    `json:"local_threads,omitempty"`
	LocalGPULayers   int    `json:"local_gpu_layers,omitempty"`
}

var configDir = filepath.Join(xdg.ConfigHome, "th")
var configPath = filepath.Join(configDir, "config.json")
var userCacheDir = os.UserCacheDir
var userHomeDir = os.UserHomeDir

func Load() (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config not found")
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	cfg.applyDefaults()

	return &cfg, nil
}

func Save(cfg *Config) error {
	cfg.applyDefaults()

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}

func ConfigPath() string {
	return configPath
}

func ConfigDir() string {
	return configDir
}

func ModelDir() (string, error) {
	cacheDir, err := userCacheDir()
	if err != nil {
		return "", fmt.Errorf("resolving user cache directory: %w", err)
	}

	return filepath.Join(cacheDir, "th", "models"), nil
}

func DefaultManagedModelPath() (string, error) {
	modelDir, err := ModelDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(modelDir, DefaultLocalModelFile), nil
}

func LegacyManagedModelPath() (string, error) {
	homeDir, err := userHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving user home directory: %w", err)
	}

	return filepath.Join(homeDir, ".cache", "th", "models", DefaultLocalModelFile), nil
}

func IsConfigNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "config not found")
}

func (c *Config) applyDefaults() {
	originalProvider := c.Provider
	c.Provider = ProviderLocal
	if c.Model == "" || (originalProvider != "" && originalProvider != ProviderLocal) {
		c.Model = DefaultLocalModel
	}
	if c.LocalContextSize <= 0 {
		c.LocalContextSize = DefaultLocalContext
	}
	if c.LocalThreads <= 0 {
		c.LocalThreads = runtime.NumCPU()
		if c.LocalThreads <= 0 {
			c.LocalThreads = 4
		}
	}
	if c.LocalGPULayers < 0 {
		c.LocalGPULayers = DefaultLocalGPULayers
	}
}
