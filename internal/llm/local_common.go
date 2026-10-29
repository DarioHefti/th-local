package llm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DarioHefti/th-local/internal/config"
)

func resolveLocalModelPath(explicitPath string) (string, error) {
	envPath := strings.TrimSpace(os.Getenv("TH_MODEL_PATH"))
	for _, override := range []string{explicitPath, envPath} {
		if strings.TrimSpace(override) != "" {
			absolutePath, err := filepath.Abs(override)
			if err != nil {
				return "", err
			}
			info, err := os.Stat(absolutePath)
			if err != nil || info.IsDir() {
				return "", fmt.Errorf("configured model not found at %s; provide an existing %s or clear the override to download the default model", absolutePath, config.DefaultLocalModelFile)
			}
			return absolutePath, nil
		}
	}
	managedPath, _ := config.DefaultManagedModelPath()
	legacyManagedPath, _ := config.LegacyManagedModelPath()

	cwd, _ := os.Getwd()
	executablePath, _ := os.Executable()
	executableDir := filepath.Dir(executablePath)

	if path, err := findLocalModelPath("", "", managedPath, legacyManagedPath, cwd, executableDir); err == nil {
		return path, nil
	}
	if managedPath == "" {
		return "", fmt.Errorf("cannot resolve the default model cache directory")
	}
	fmt.Fprintf(os.Stderr, "Downloading %s to the user model cache...\n", config.DefaultLocalModelFile)
	if err := downloadLocalModel(managedPath, config.DefaultLocalModelURL); err != nil {
		return "", err
	}
	return managedPath, nil
}

func findLocalModelPath(explicitPath, envPath, managedPath, legacyManagedPath, cwd, executableDir string) (string, error) {
	candidates := []string{
		explicitPath,
		envPath,
		managedPath,
		legacyManagedPath,
		filepath.Join(cwd, "model", config.DefaultLocalModelFile),
		filepath.Join(executableDir, "model", config.DefaultLocalModelFile),
	}

	for _, candidate := range candidates {
		if strings.TrimSpace(candidate) == "" {
			continue
		}

		absolutePath, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}

		info, statErr := os.Stat(absolutePath)
		if statErr == nil && !info.IsDir() {
			return absolutePath, nil
		}
	}

	compatibilityPath := filepath.Join("model", config.DefaultLocalModelFile)
	if strings.TrimSpace(managedPath) == "" {
		return "", fmt.Errorf("local model not found; set local_model_path in config or TH_MODEL_PATH, or place %s in the managed th model directory or %s", config.DefaultLocalModelFile, compatibilityPath)
	}

	return "", fmt.Errorf("local model not found; set local_model_path in config or TH_MODEL_PATH, or install %s at %s (compatibility fallback: %s)", config.DefaultLocalModelFile, managedPath, compatibilityPath)
}

func BuildPrompt(systemPrompt, userPrompt string) string {
	// Match LiquidAI's text-only chat template with enable_thinking=false:
	// shell commands need a direct answer, not a reasoning preamble.
	prompt := "<|startoftext|>"
	if system := strings.TrimSpace(systemPrompt); system != "" {
		prompt += "<|im_start|>system\n" + system + "<|im_end|>\n"
	}
	return prompt + "<|im_start|>user\n" + strings.TrimSpace(userPrompt) + "<|im_end|>\n<|im_start|>assistant\n<think></think>"
}

var stopMarkers = []string{
	"<|im_end|>", "<|im_start|>",
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
