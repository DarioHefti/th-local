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
	managedPath, _ := config.DefaultManagedModelPath()
	legacyManagedPath, _ := config.LegacyManagedModelPath()

	cwd, _ := os.Getwd()
	executablePath, _ := os.Executable()
	executableDir := filepath.Dir(executablePath)

	return findLocalModelPath(explicitPath, envPath, managedPath, legacyManagedPath, cwd, executableDir)
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

func formatGemmaPrompt(systemPrompt, userPrompt string) string {
	return fmt.Sprintf("<bos><|turn>system\n%s<turn|>\n<|turn>user\n%s<turn|>\n<|turn>model\n", strings.TrimSpace(systemPrompt), strings.TrimSpace(userPrompt))
}

func formatChatMLPrompt(systemPrompt, userPrompt string) string {
	return fmt.Sprintf("<|im_start|>system\n%s<|im_end|>\n<|im_start|>user\n%s<|im_end|>\n<|im_start|>assistant\n", strings.TrimSpace(systemPrompt), strings.TrimSpace(userPrompt))
}

func formatLlama3Prompt(systemPrompt, userPrompt string) string {
	return fmt.Sprintf("<|begin_of_text|><|start_header_id|>system<|end_header_id|>\n\n%s<|eot_id|><|start_header_id|>user<|end_header_id|>\n\n%s<|eot_id|><|start_header_id|>assistant<|end_header_id|>\n\n", strings.TrimSpace(systemPrompt), strings.TrimSpace(userPrompt))
}

func formatRawPrompt(systemPrompt, userPrompt string) string {
	return fmt.Sprintf("### System:\n%s\n\n### User:\n%s\n\n### Assistant:\n", strings.TrimSpace(systemPrompt), strings.TrimSpace(userPrompt))
}

func FormatPrompt(format, systemPrompt, userPrompt string) string {
	switch format {
	case config.PromptFormatChatML:
		return formatChatMLPrompt(systemPrompt, userPrompt)
	case config.PromptFormatLlama3:
		return formatLlama3Prompt(systemPrompt, userPrompt)
	case config.PromptFormatRaw:
		return formatRawPrompt(systemPrompt, userPrompt)
	default:
		return formatGemmaPrompt(systemPrompt, userPrompt)
	}
}

func BuildPrompt(systemPrompt, userPrompt string) string {
	return formatGemmaPrompt(systemPrompt, userPrompt)
}

func BuildPromptWithFormat(format, systemPrompt, userPrompt string) string {
	return FormatPrompt(format, systemPrompt, userPrompt)
}

var stopMarkers = []string{
	"<turn|>", "<|turn>user", "<|turn>system", "<|turn>model",
	"<|im_end|>", "<|im_start|>",
	"<|eot_id|>", "<|start_header_id|>", "<|end_header_id|>",
	"### User:", "### System:", "### Assistant:",
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

func stripGemmaTurnMarkers(content string) string {
	return stripTurnMarkers(content)
}
