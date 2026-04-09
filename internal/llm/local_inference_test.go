//go:build cgo

package llm

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DarioHefti/th-local/internal/config"
)

func TestLocalGeneratorPerformsInference(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping local inference test in short mode")
	}

	if strings.TrimSpace(os.Getenv("TH_RUN_LOCAL_INFERENCE_TEST")) != "1" {
		t.Skip("skipping local inference test; set TH_RUN_LOCAL_INFERENCE_TEST=1 to enable it")
	}

	modelPath, err := resolveInferenceTestModelPath()
	if err != nil {
		t.Fatalf("resolveInferenceTestModelPath failed: %v", err)
	}
	if modelPath == "" {
		t.Skip("skipping local inference test; set TH_INFERENCE_MODEL_PATH or install the managed model first")
	}

	cfg := &config.Config{
		Provider:         config.ProviderLocal,
		Model:            config.DefaultLocalModel,
		LocalModelPath:   modelPath,
		LocalContextSize: 512,
		LocalThreads:     1,
		LocalGPULayers:   0,
	}

	generator, err := newLocalGenerator(cfg)
	if err != nil {
		t.Fatalf("newLocalGenerator failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	command, err := generator.GetCommand(
		ctx,
		"You are a shell command generator. Reply with exactly echo test-mvp and nothing else.",
		"Return the required shell command exactly.",
	)
	if err != nil {
		t.Fatalf("GetCommand failed: %v", err)
	}

	lower := strings.ToLower(command)
	if !strings.Contains(lower, "echo") || !strings.Contains(lower, "test-mvp") {
		t.Fatalf("unexpected command output: %q", command)
	}
}

func resolveInferenceTestModelPath() (string, error) {
	candidates := []string{
		strings.TrimSpace(os.Getenv("TH_INFERENCE_MODEL_PATH")),
		strings.TrimSpace(os.Getenv("TH_MODEL_PATH")),
	}

	if managedPath, err := config.DefaultManagedModelPath(); err == nil {
		candidates = append(candidates, managedPath)
	}
	if legacyManagedPath, err := config.LegacyManagedModelPath(); err == nil {
		candidates = append(candidates, legacyManagedPath)
	}

	compatibilityPath, err := filepath.Abs(filepath.Join("..", "..", "model", config.DefaultLocalModelFile))
	if err == nil {
		candidates = append(candidates, compatibilityPath)
	}

	for _, candidate := range candidates {
		if strings.TrimSpace(candidate) == "" {
			continue
		}

		info, statErr := os.Stat(candidate)
		if statErr == nil && !info.IsDir() {
			return candidate, nil
		}
	}

	return "", nil
}
