package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultManagedModelPath(t *testing.T) {
	origUserCacheDir := userCacheDir
	userCacheDir = func() (string, error) {
		return filepath.Join("cache-root"), nil
	}
	defer func() {
		userCacheDir = origUserCacheDir
	}()

	got, err := DefaultManagedModelPath()
	if err != nil {
		t.Fatalf("DefaultManagedModelPath failed: %v", err)
	}

	want := filepath.Join("cache-root", "th", "models", DefaultLocalModelFile)
	if got != want {
		t.Fatalf("DefaultManagedModelPath = %q, want %q", got, want)
	}
}

func TestLegacyManagedModelPath(t *testing.T) {
	origUserHomeDir := userHomeDir
	userHomeDir = func() (string, error) {
		return filepath.Join("home-root"), nil
	}
	defer func() {
		userHomeDir = origUserHomeDir
	}()

	got, err := LegacyManagedModelPath()
	if err != nil {
		t.Fatalf("LegacyManagedModelPath failed: %v", err)
	}

	want := filepath.Join("home-root", ".cache", "th", "models", DefaultLocalModelFile)
	if got != want {
		t.Fatalf("LegacyManagedModelPath = %q, want %q", got, want)
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()

	origConfigDir := configDir
	origConfigPath := configPath
	configDir = tmpDir
	configPath = filepath.Join(tmpDir, "config.json")
	defer func() {
		configDir = origConfigDir
		configPath = origConfigPath
	}()

	cfg := &Config{
		Provider:         ProviderLocal,
		Model:            DefaultLocalModel,
		LocalModelPath:   filepath.Join(tmpDir, "google_gemma-4-E2B-it-IQ2_M.gguf"),
		LocalContextSize: 4096,
		LocalThreads:     8,
		LocalGPULayers:   0,
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Provider != cfg.Provider {
		t.Errorf("Provider: got %q, want %q", loaded.Provider, cfg.Provider)
	}
	if loaded.Model != cfg.Model {
		t.Errorf("Model: got %q, want %q", loaded.Model, cfg.Model)
	}
	if loaded.LocalModelPath != cfg.LocalModelPath {
		t.Errorf("LocalModelPath: got %q, want %q", loaded.LocalModelPath, cfg.LocalModelPath)
	}
	if loaded.LocalContextSize != cfg.LocalContextSize {
		t.Errorf("LocalContextSize: got %d, want %d", loaded.LocalContextSize, cfg.LocalContextSize)
	}
	if loaded.LocalThreads != cfg.LocalThreads {
		t.Errorf("LocalThreads: got %d, want %d", loaded.LocalThreads, cfg.LocalThreads)
	}
}

func TestLoadLegacyNonLocalConfigMigratesToLocalDefaults(t *testing.T) {
	tmpDir := t.TempDir()

	origConfigDir := configDir
	origConfigPath := configPath
	configDir = tmpDir
	configPath = filepath.Join(tmpDir, "config.json")
	defer func() {
		configDir = origConfigDir
		configPath = origConfigPath
	}()

	if err := os.WriteFile(configPath, []byte(`{"provider":"legacy","model":"legacy-model"}`), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Provider != ProviderLocal {
		t.Fatalf("Provider: got %q, want %q", loaded.Provider, ProviderLocal)
	}
	if loaded.Model != DefaultLocalModel {
		t.Fatalf("Model: got %q, want %q", loaded.Model, DefaultLocalModel)
	}
	if loaded.LocalContextSize != DefaultLocalContext {
		t.Fatalf("LocalContextSize: got %d, want %d", loaded.LocalContextSize, DefaultLocalContext)
	}
}

func TestLoadLocalConfigAppliesDefaults(t *testing.T) {
	tmpDir := t.TempDir()

	origConfigDir := configDir
	origConfigPath := configPath
	configDir = tmpDir
	configPath = filepath.Join(tmpDir, "config.json")
	defer func() {
		configDir = origConfigDir
		configPath = origConfigPath
	}()

	if err := os.WriteFile(configPath, []byte(`{"provider":"local"}`), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Model != DefaultLocalModel {
		t.Fatalf("Model: got %q, want %q", loaded.Model, DefaultLocalModel)
	}
	if loaded.LocalContextSize != DefaultLocalContext {
		t.Fatalf("LocalContextSize: got %d, want %d", loaded.LocalContextSize, DefaultLocalContext)
	}
	if loaded.LocalThreads <= 0 {
		t.Fatalf("LocalThreads: got %d, want positive value", loaded.LocalThreads)
	}
	if loaded.PromptFormat != DefaultPromptFormat {
		t.Fatalf("PromptFormat: got %q, want %q", loaded.PromptFormat, DefaultPromptFormat)
	}
}

func TestPromptFormatPreservedOnSaveLoad(t *testing.T) {
	tmpDir := t.TempDir()

	origConfigDir := configDir
	origConfigPath := configPath
	configDir = tmpDir
	configPath = filepath.Join(tmpDir, "config.json")
	defer func() {
		configDir = origConfigDir
		configPath = origConfigPath
	}()

	cfg := &Config{
		Provider:       ProviderLocal,
		Model:          DefaultLocalModel,
		LocalModelPath: filepath.Join(tmpDir, "custom-model.gguf"),
		PromptFormat:   PromptFormatChatML,
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.PromptFormat != PromptFormatChatML {
		t.Fatalf("PromptFormat: got %q, want %q", loaded.PromptFormat, PromptFormatChatML)
	}
	if loaded.LocalModelPath != cfg.LocalModelPath {
		t.Fatalf("LocalModelPath: got %q, want %q", loaded.LocalModelPath, cfg.LocalModelPath)
	}
}

func TestIsValidPromptFormat(t *testing.T) {
	for _, f := range ValidPromptFormats {
		if !IsValidPromptFormat(f) {
			t.Errorf("IsValidPromptFormat(%q) = false, want true", f)
		}
	}
	if IsValidPromptFormat("nonexistent") {
		t.Error("IsValidPromptFormat(\"nonexistent\") = true, want false")
	}
	if IsValidPromptFormat("") {
		t.Error("IsValidPromptFormat(\"\") = true, want false")
	}
}

func TestLoadNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	origConfigDir := configDir
	origConfigPath := configPath
	configDir = tmpDir
	configPath = filepath.Join(tmpDir, "nonexistent.json")
	defer func() {
		configDir = origConfigDir
		configPath = origConfigPath
	}()

	_, err := Load()
	if err == nil {
		t.Error("Load should fail for nonexistent file")
	}
	if !IsConfigNotFound(err) {
		t.Errorf("IsConfigNotFound should return true, got false for: %v", err)
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()

	origConfigDir := configDir
	origConfigPath := configPath
	configDir = tmpDir
	configPath = filepath.Join(tmpDir, "invalid.json")
	defer func() {
		configDir = origConfigDir
		configPath = origConfigPath
	}()

	os.WriteFile(configPath, []byte("not json"), 0644)

	_, err := Load()
	if err == nil {
		t.Error("Load should fail for invalid JSON")
	}
}
