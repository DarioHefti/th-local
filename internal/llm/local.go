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
	localPredictTokens = 1024
)

type localGenerator struct {
	modelPath   string
	contextSize int
	threads     int

	mu     sync.Mutex
	engine *llamacpp.Model
}

func newLocalGenerator(cfg *config.Config) (Generator, error) {
	modelPath, err := resolveLocalModelPath()
	if err != nil {
		return nil, err
	}

	return newLocalGeneratorFromPath(modelPath, cfg)
}

func newLocalGeneratorFromPath(modelPath string, cfg *config.Config) (Generator, error) {
	engine, err := llamacpp.New(modelPath, cfg.LocalContextSize, cfg.LocalThreads, cfg.LocalGPULayers)
	if err != nil {
		return nil, fmt.Errorf("loading local model: %w", err)
	}

	return &localGenerator{
		modelPath:   modelPath,
		contextSize: cfg.LocalContextSize,
		threads:     cfg.LocalThreads,
		engine:      engine,
	}, nil
}

func (g *localGenerator) GetCommand(ctx context.Context, systemPrompt, userPrompt string, onThinking func(string)) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	prompt := BuildPrompt(systemPrompt, userPrompt)

	g.mu.Lock()
	defer g.mu.Unlock()

	response, err := g.engine.PredictStream(prompt, localPredictTokens, streamThinking(onThinking))
	if err != nil {
		return "", fmt.Errorf("running local inference: %w", err)
	}

	return cleanupCommandOutput(stripThinking(stripTurnMarkers(response))), nil
}
