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

func TestFormatChatMLPrompt(t *testing.T) {
	prompt := formatChatMLPrompt("system prompt", "user prompt")
	if !strings.Contains(prompt, "<|im_start|>system") {
		t.Fatalf("prompt missing system marker: %q", prompt)
	}
	if !strings.Contains(prompt, "<|im_start|>user") {
		t.Fatalf("prompt missing user marker: %q", prompt)
	}
	if !strings.Contains(prompt, "<|im_start|>assistant") {
		t.Fatalf("prompt missing assistant marker: %q", prompt)
	}
}

func TestFormatLlama3Prompt(t *testing.T) {
	prompt := formatLlama3Prompt("system prompt", "user prompt")
	if !strings.Contains(prompt, "<|begin_of_text|>") {
		t.Fatalf("prompt missing begin_of_text: %q", prompt)
	}
	if !strings.Contains(prompt, "<|start_header_id|>system<|end_header_id|>") {
		t.Fatalf("prompt missing system header: %q", prompt)
	}
	if !strings.Contains(prompt, "<|start_header_id|>assistant<|end_header_id|>") {
		t.Fatalf("prompt missing assistant header: %q", prompt)
	}
}

func TestFormatRawPrompt(t *testing.T) {
	prompt := formatRawPrompt("system prompt", "user prompt")
	if !strings.Contains(prompt, "### System:") {
		t.Fatalf("prompt missing system section: %q", prompt)
	}
	if !strings.Contains(prompt, "### User:") {
		t.Fatalf("prompt missing user section: %q", prompt)
	}
	if !strings.Contains(prompt, "### Assistant:") {
		t.Fatalf("prompt missing assistant section: %q", prompt)
	}
}

func TestFormatPromptDispatch(t *testing.T) {
	sys, user := "sys", "usr"

	if FormatPrompt("gemma", sys, user) != formatGemmaPrompt(sys, user) {
		t.Fatal("FormatPrompt gemma mismatch")
	}
	if FormatPrompt("chatml", sys, user) != formatChatMLPrompt(sys, user) {
		t.Fatal("FormatPrompt chatml mismatch")
	}
	if FormatPrompt("llama3", sys, user) != formatLlama3Prompt(sys, user) {
		t.Fatal("FormatPrompt llama3 mismatch")
	}
	if FormatPrompt("raw", sys, user) != formatRawPrompt(sys, user) {
		t.Fatal("FormatPrompt raw mismatch")
	}
	if FormatPrompt("unknown", sys, user) != formatGemmaPrompt(sys, user) {
		t.Fatal("FormatPrompt should default to gemma for unknown format")
	}
}

func TestBuildPrompt(t *testing.T) {
	got := BuildPrompt("system prompt", "user prompt")
	want := formatGemmaPrompt("system prompt", "user prompt")

	if got != want {
		t.Fatalf("BuildPrompt() = %q, want %q", got, want)
	}
}

func TestBuildPromptWithFormat(t *testing.T) {
	got := BuildPromptWithFormat("chatml", "system prompt", "user prompt")
	want := formatChatMLPrompt("system prompt", "user prompt")

	if got != want {
		t.Fatalf("BuildPromptWithFormat(chatml) = %q, want %q", got, want)
	}
}

func TestStripGemmaTurnMarkers(t *testing.T) {
	input := "echo test-mvp<turn|>\n<|turn>user"
	if got := stripGemmaTurnMarkers(input); got != "echo test-mvp" {
		t.Fatalf("stripGemmaTurnMarkers = %q, want %q", got, "echo test-mvp")
	}
}

func TestStripTurnMarkersChatML(t *testing.T) {
	input := "ls -la<|im_end|>\n<|im_start|>user"
	if got := stripTurnMarkers(input); got != "ls -la" {
		t.Fatalf("stripTurnMarkers(chatml) = %q, want %q", got, "ls -la")
	}
}

func TestStripTurnMarkersLlama3(t *testing.T) {
	input := "git status<|eot_id|>"
	if got := stripTurnMarkers(input); got != "git status" {
		t.Fatalf("stripTurnMarkers(llama3) = %q, want %q", got, "git status")
	}
}

func TestStripTurnMarkersRaw(t *testing.T) {
	input := "find . -name '*.go'\n### User:"
	if got := stripTurnMarkers(input); got != "find . -name '*.go'" {
		t.Fatalf("stripTurnMarkers(raw) = %q, want %q", got, "find . -name '*.go'")
	}
}
