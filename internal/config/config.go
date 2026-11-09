package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/adrg/xdg"
)

const (
	DefaultLocalContext   = 2048
	DefaultLocalGPULayers = 99
	DefaultLocalModelFile = "LFM3-1.7B-Q4_K_M.gguf"
	DefaultLocalModelURL  = "https://huggingface.co/LiquidAI/LFM3-1.7B-GGUF/resolve/main/LFM3-1.7B-Q4_K_M.gguf?download=true"
)

type Config struct {
	LocalContextSize int `json:"local_context_size,omitempty"`
	LocalThreads     int `json:"local_threads,omitempty"`
	LocalGPULayers   int `json:"local_gpu_layers,omitempty"`
}

var configPath = filepath.Join(xdg.ConfigHome, "th", "config.json")
var userCacheDir = os.UserCacheDir
var userHomeDir = os.UserHomeDir

// Load reads optional runtime tuning. A missing file means defaults, so th
// works without any setup.
func Load() (*Config, error) {
	var cfg Config
	data, err := os.ReadFile(configPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	if err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parsing config: %w", err)
		}
	}

	cfg.applyDefaults()

	return &cfg, nil
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

func (c *Config) applyDefaults() {
	if c.LocalContextSize <= 0 {
		c.LocalContextSize = DefaultLocalContext
	}
	if c.LocalThreads <= 0 {
		c.LocalThreads = runtime.NumCPU()
		if c.LocalThreads <= 0 {
			c.LocalThreads = 4
		}
	}
	if c.LocalGPULayers <= 0 {
		c.LocalGPULayers = DefaultLocalGPULayers
	}
}
