package llm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindLocalModelPathPrefersManagedModel(t *testing.T) {
	tmpDir := t.TempDir()
	managedPath := filepath.Join(tmpDir, "managed.gguf")
	legacyPath := filepath.Join(tmpDir, "legacy.gguf")
	for _, path := range []string{managedPath, legacyPath} {
		if err := os.WriteFile(path, []byte("model"), 0644); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}
	}

	if resolved, ok := findLocalModelPath(managedPath, legacyPath); !ok || resolved != managedPath {
		t.Fatalf("resolved = %q, %v, want %q", resolved, ok, managedPath)
	}
}

func TestFindLocalModelPathFallsBackToLegacyModel(t *testing.T) {
	tmpDir := t.TempDir()
	legacyPath := filepath.Join(tmpDir, "legacy.gguf")
	if err := os.WriteFile(legacyPath, []byte("model"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	if resolved, ok := findLocalModelPath(filepath.Join(tmpDir, "missing.gguf"), legacyPath); !ok || resolved != legacyPath {
		t.Fatalf("resolved = %q, %v, want %q", resolved, ok, legacyPath)
	}
}

func TestFindLocalModelPathReportsMissingModel(t *testing.T) {
	tmpDir := t.TempDir()
	if resolved, ok := findLocalModelPath(filepath.Join(tmpDir, "managed.gguf"), ""); ok {
		t.Fatalf("expected no model, got %q", resolved)
	}
}

func TestCleanupCommandOutput(t *testing.T) {
	input := "```bash\nls -la\n```"
	if got := cleanupCommandOutput(input); got != "ls -la" {
		t.Fatalf("cleanupCommandOutput = %q, want %q", got, "ls -la")
	}
}

func TestFormatLFM3Prompt(t *testing.T) {
	got := BuildPrompt(" system prompt ", " user prompt ")
	want := "<|startoftext|><|im_start|>system\nsystem prompt<|im_end|>\n<|im_start|>user\nuser prompt<|im_end|>\n<|im_start|>assistant\n<think>"
	if got != want {
		t.Fatalf("LFM3 prompt = %q, want %q", got, want)
	}
}

func TestFormatLFM3PromptWithoutSystem(t *testing.T) {
	got := BuildPrompt("", "user prompt")
	want := "<|startoftext|><|im_start|>user\nuser prompt<|im_end|>\n<|im_start|>assistant\n<think>"
	if got != want {
		t.Fatalf("LFM3 prompt = %q, want %q", got, want)
	}
}

func TestStripThinking(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"Use ls.</think>\nls -la", "ls -la"},
		{"Maybe </think> ls? Wait.</think>ls -la", "ls -la"},
		{"ran out of tokens mid-thought", ""},
	} {
		if got := stripThinking(tc.input); got != tc.want {
			t.Fatalf("stripThinking(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestStreamThinking(t *testing.T) {
	var shown strings.Builder
	stream := streamThinking(func(piece string) { shown.WriteString(piece) })
	for _, piece := range []string{"Use", " ls", "</think>", "ls -la"} {
		stream(piece)
	}
	if got := shown.String(); got != "Use ls" {
		t.Fatalf("streamed thinking = %q, want %q", got, "Use ls")
	}
	if streamThinking(nil) != nil {
		t.Fatal("streamThinking(nil) should not create a callback")
	}
}

func TestStripTurnMarkers(t *testing.T) {
	input := "ls -la<|im_end|>\n<|im_start|>user"
	if got := stripTurnMarkers(input); got != "ls -la" {
		t.Fatalf("stripTurnMarkers(chatml) = %q, want %q", got, "ls -la")
	}
}
