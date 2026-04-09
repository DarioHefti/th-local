package llm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindLocalModelPathPrefersExplicitPath(t *testing.T) {
	tmpDir := t.TempDir()
	explicitPath := filepath.Join(tmpDir, "explicit.gguf")
	envPath := filepath.Join(tmpDir, "env.gguf")

	if err := os.WriteFile(explicitPath, []byte("model"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	if err := os.WriteFile(envPath, []byte("model"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	resolved, err := findLocalModelPath(explicitPath, envPath, filepath.Join(tmpDir, "managed.gguf"), filepath.Join(tmpDir, "legacy.gguf"), filepath.Join(tmpDir, "cwd"), filepath.Join(tmpDir, "exe"))
	if err != nil {
		t.Fatalf("findLocalModelPath failed: %v", err)
	}

	if resolved != explicitPath {
		t.Fatalf("resolved path = %q, want %q", resolved, explicitPath)
	}
}

func TestFindLocalModelPathFallsBackToManagedModel(t *testing.T) {
	tmpDir := t.TempDir()
	managedPath := filepath.Join(tmpDir, "cache", "th", "models", "google_gemma-4-E2B-it-IQ2_M.gguf")
	if err := os.MkdirAll(filepath.Dir(managedPath), 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if err := os.WriteFile(managedPath, []byte("model"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	resolved, err := findLocalModelPath("", "", managedPath, filepath.Join(tmpDir, "legacy.gguf"), filepath.Join(tmpDir, "cwd"), filepath.Join(tmpDir, "exe"))
	if err != nil {
		t.Fatalf("findLocalModelPath failed: %v", err)
	}

	if resolved != managedPath {
		t.Fatalf("resolved path = %q, want %q", resolved, managedPath)
	}
}

func TestFindLocalModelPathFallsBackToWorkingDirectoryModel(t *testing.T) {
	tmpDir := t.TempDir()
	workingDir := filepath.Join(tmpDir, "cwd")
	modelDir := filepath.Join(workingDir, "model")
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	expectedPath := filepath.Join(modelDir, "google_gemma-4-E2B-it-IQ2_M.gguf")
	if err := os.WriteFile(expectedPath, []byte("model"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	resolved, err := findLocalModelPath("", "", filepath.Join(tmpDir, "missing-managed.gguf"), filepath.Join(tmpDir, "legacy.gguf"), workingDir, filepath.Join(tmpDir, "exe"))
	if err != nil {
		t.Fatalf("findLocalModelPath failed: %v", err)
	}

	if resolved != expectedPath {
		t.Fatalf("resolved path = %q, want %q", resolved, expectedPath)
	}
}

func TestFindLocalModelPathFallsBackToLegacyManagedModel(t *testing.T) {
	tmpDir := t.TempDir()
	legacyManagedPath := filepath.Join(tmpDir, ".cache", "th", "models", "google_gemma-4-E2B-it-IQ2_M.gguf")
	if err := os.MkdirAll(filepath.Dir(legacyManagedPath), 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if err := os.WriteFile(legacyManagedPath, []byte("model"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	resolved, err := findLocalModelPath("", "", filepath.Join(tmpDir, "missing-managed.gguf"), legacyManagedPath, filepath.Join(tmpDir, "cwd"), filepath.Join(tmpDir, "exe"))
	if err != nil {
		t.Fatalf("findLocalModelPath failed: %v", err)
	}

	if resolved != legacyManagedPath {
		t.Fatalf("resolved path = %q, want %q", resolved, legacyManagedPath)
	}
}

func TestFindLocalModelPathErrorMentionsManagedPath(t *testing.T) {
	tmpDir := t.TempDir()
	managedPath := filepath.Join(tmpDir, "cache", "th", "models", "google_gemma-4-E2B-it-IQ2_M.gguf")

	_, err := findLocalModelPath("", "", managedPath, filepath.Join(tmpDir, "legacy.gguf"), filepath.Join(tmpDir, "cwd"), filepath.Join(tmpDir, "exe"))
	if err == nil {
		t.Fatal("expected findLocalModelPath to fail")
	}

	if !strings.Contains(err.Error(), managedPath) {
		t.Fatalf("error %q does not mention managed path %q", err.Error(), managedPath)
	}
}

func TestCleanupCommandOutput(t *testing.T) {
	input := "```bash\nls -la\n```"
	if got := cleanupCommandOutput(input); got != "ls -la" {
		t.Fatalf("cleanupCommandOutput = %q, want %q", got, "ls -la")
	}
}

func TestFormatGemmaPrompt(t *testing.T) {
	prompt := formatGemmaPrompt("system prompt", "user prompt")
	if !strings.Contains(prompt, "<|turn>system") {
		t.Fatalf("prompt missing system turn marker: %q", prompt)
	}
	if !strings.Contains(prompt, "<|turn>user") {
		t.Fatalf("prompt missing user turn marker: %q", prompt)
	}
}

func TestBuildPrompt(t *testing.T) {
	got := BuildPrompt("system prompt", "user prompt")
	want := formatGemmaPrompt("system prompt", "user prompt")

	if got != want {
		t.Fatalf("BuildPrompt() = %q, want %q", got, want)
	}
}

func TestStripGemmaTurnMarkers(t *testing.T) {
	input := "echo test-mvp<turn|>\n<|turn>user"
	if got := stripGemmaTurnMarkers(input); got != "echo test-mvp" {
		t.Fatalf("stripGemmaTurnMarkers = %q, want %q", got, "echo test-mvp")
	}
}
