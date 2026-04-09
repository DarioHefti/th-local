//go:build !cgo

package llm

import (
	"fmt"

	"github.com/DarioHefti/th-local/internal/config"
)

func newLocalGenerator(cfg *config.Config) (Generator, error) {
	_ = cfg
	return nil, fmt.Errorf("local provider requires cgo-enabled builds with the vendored llama.cpp bridge")
}
