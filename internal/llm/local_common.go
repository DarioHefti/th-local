package llm

import (
	"fmt"
	"os"
	"strings"

	"github.com/DarioHefti/th-local/internal/config"
)

func resolveLocalModelPath() (string, error) {
	managedPath, err := config.DefaultManagedModelPath()
	if err != nil {
		return "", fmt.Errorf("cannot resolve the default model cache directory: %w", err)
	}
	legacyManagedPath, _ := config.LegacyManagedModelPath()

	if path, ok := findLocalModelPath(managedPath, legacyManagedPath); ok {
		return path, nil
	}
	fmt.Fprintf(os.Stderr, "First run: downloading %s to the user model cache...\n", config.DefaultLocalModelFile)
	if err := downloadLocalModel(managedPath, config.DefaultLocalModelURL); err != nil {
		return "", err
	}
	return managedPath, nil
}

func findLocalModelPath(candidates ...string) (string, bool) {
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, true
		}
	}
	return "", false
}

func BuildPrompt(systemPrompt, userPrompt string) string {
	// Match LiquidAI's text-only chat template with enable_thinking=true:
	// the model reasons after the opening tag and answers after </think>.
	prompt := "<|startoftext|>"
	if system := strings.TrimSpace(systemPrompt); system != "" {
		prompt += "<|im_start|>system\n" + system + "<|im_end|>\n"
	}
	return prompt + "<|im_start|>user\n" + strings.TrimSpace(userPrompt) + "<|im_end|>\n<|im_start|>assistant\n<think>"
}

var stopMarkers = []string{
	"<|im_end|>", "<|im_start|>",
}

const thinkEndMarker = "</think>"

// stripThinking returns the answer after the model's reasoning. Output without
// a closing tag ran out of tokens mid-thought, so it has no answer to return.
func stripThinking(content string) string {
	idx := strings.LastIndex(content, thinkEndMarker)
	if idx < 0 {
		return ""
	}
	return strings.TrimSpace(content[idx+len(thinkEndMarker):])
}

// streamThinking forwards generated text until the reasoning closes, so the
// answer is only ever shown once it has been cleaned up. The closing tag is a
// single token in LFM3.
func streamThinking(onThinking func(string)) func(string) {
	if onThinking == nil {
		return nil
	}
	done := false
	return func(piece string) {
		if piece == thinkEndMarker {
			done = true
		}
		if !done {
			onThinking(piece)
		}
	}
}

func stripTurnMarkers(content string) string {
	trimmed := strings.TrimSpace(content)
	for _, marker := range stopMarkers {
		if idx := strings.Index(trimmed, marker); idx >= 0 {
			trimmed = trimmed[:idx]
		}
	}
	return strings.TrimSpace(trimmed)
}
