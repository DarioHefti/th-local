//go:build cgo

package llm

import (
	"context"
	"fmt"
	"sync"

	"github.com/DarioHefti/th-local/internal/config"
	"github.com/DarioHefti/th-local/internal/llm/llamacpp"
)

const (
	localPredictTokens = 128
	localBatchSize     = 512
)

type localGenerator struct {
	modelPath    string
	contextSize  int
	threads      int
	promptFormat string

	mu     sync.Mutex
	engine *llamacpp.Model
}

func newLocalGenerator(cfg *config.Config) (Generator, error) {
	modelPath, err := resolveLocalModelPath(cfg.LocalModelPath)
	if err != nil {
		return nil, err
	}

	engine, err := llamacpp.New(modelPath, cfg.LocalContextSize, cfg.LocalThreads, cfg.LocalGPULayers)
	if err != nil {
		return nil, fmt.Errorf("loading local model: %w", err)
	}

	return &localGenerator{
		modelPath:    modelPath,
		contextSize:  cfg.LocalContextSize,
		threads:      cfg.LocalThreads,
		promptFormat: cfg.PromptFormat,
		engine:       engine,
	}, nil
}

func (g *localGenerator) GetCommand(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	prompt := FormatPrompt(g.promptFormat, systemPrompt, userPrompt)

	g.mu.Lock()
	defer g.mu.Unlock()

	response, err := g.engine.Predict(prompt, localPredictTokens)
	if err != nil {
		return "", fmt.Errorf("running local inference: %w", err)
	}

	return cleanupCommandOutput(stripTurnMarkers(response)), nil
}
