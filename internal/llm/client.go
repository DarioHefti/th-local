package llm

import (
	"context"
	"strings"

	"github.com/DarioHefti/th-local/internal/config"
)

type Generator interface {
	GetCommand(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

func NewGenerator(cfg *config.Config) (Generator, error) {
	return newLocalGenerator(cfg)
}

func cleanupCommandOutput(content string) string {
	command := strings.TrimSpace(content)
	if idx := strings.Index(command, "<turn|>"); idx >= 0 {
		command = command[:idx]
	}
	if idx := strings.Index(command, "<|turn>"); idx >= 0 {
		command = command[:idx]
	}
	command = strings.TrimPrefix(command, "```bash")
	command = strings.TrimPrefix(command, "```sh")
	command = strings.TrimPrefix(command, "```")
	command = strings.TrimSuffix(command, "```")
	command = strings.Trim(command, "`")
	return strings.TrimSpace(command)
}
