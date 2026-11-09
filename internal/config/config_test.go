package config

import (
	"os"
	"path/filepath"
	"testing"
)

func useConfigPath(t *testing.T, name string) string {
	t.Helper()
	orig := configPath
	configPath = filepath.Join(t.TempDir(), name)
	t.Cleanup(func() { configPath = orig })
	return configPath
}

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

func TestLoadMissingFileUsesDefaults(t *testing.T) {
	useConfigPath(t, "nonexistent.json")

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.LocalContextSize != DefaultLocalContext {
		t.Fatalf("LocalContextSize: got %d, want %d", loaded.LocalContextSize, DefaultLocalContext)
	}
	if loaded.LocalThreads <= 0 {
		t.Fatalf("LocalThreads: got %d, want positive value", loaded.LocalThreads)
	}
	if loaded.LocalGPULayers != DefaultLocalGPULayers {
		t.Fatalf("LocalGPULayers: got %d, want %d", loaded.LocalGPULayers, DefaultLocalGPULayers)
	}
}

func TestLoadKeepsTuning(t *testing.T) {
	path := useConfigPath(t, "config.json")
	if err := os.WriteFile(path, []byte(`{"local_context_size":4096,"local_threads":8,"local_gpu_layers":10}`), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.LocalContextSize != 4096 || loaded.LocalThreads != 8 || loaded.LocalGPULayers != 10 {
		t.Fatalf("tuning was not preserved: %+v", loaded)
	}
}

func TestLoadIgnoresRetiredModelSettings(t *testing.T) {
	path := useConfigPath(t, "config.json")
	if err := os.WriteFile(path, []byte(`{"provider":"legacy","model":"legacy-model","local_model_path":"/tmp/other.gguf"}`), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.LocalContextSize != DefaultLocalContext || loaded.LocalGPULayers != DefaultLocalGPULayers {
		t.Fatalf("defaults were not applied: %+v", loaded)
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	path := useConfigPath(t, "invalid.json")
	if err := os.WriteFile(path, []byte("not json"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	if _, err := Load(); err == nil {
		t.Error("Load should fail for invalid JSON")
	}
}
