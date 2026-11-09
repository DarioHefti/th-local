package llm

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// downloadProgress reports each whole percent on stderr, since the model takes minutes to fetch.
type downloadProgress struct {
	total, written, percent int64
}

func (p *downloadProgress) Write(data []byte) (int, error) {
	p.written += int64(len(data))
	if p.total > 0 {
		if percent := p.written * 100 / p.total; percent != p.percent {
			p.percent = percent
			fmt.Fprintf(os.Stderr, "\r%d%% (%d of %d MB)", percent, p.written>>20, p.total>>20)
		}
	}
	return len(data), nil
}

func downloadLocalModel(path, url string) error {
	if info, err := os.Stat(path); err == nil {
		if info.IsDir() || info.Size() == 0 {
			return fmt.Errorf("model path is not a nonempty file: %s", path)
		}
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("checking model path: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating model cache: %w", err)
	}
	client := &http.Client{Timeout: 30 * time.Minute}
	response, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("downloading model: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("downloading model: HTTP %d", response.StatusCode)
	}
	var magic [4]byte
	if _, err := io.ReadFull(response.Body, magic[:]); err != nil || string(magic[:]) != "GGUF" {
		return fmt.Errorf("download is not a GGUF model")
	}
	partial, err := os.CreateTemp(filepath.Dir(path), ".model-*.part")
	if err != nil {
		return fmt.Errorf("creating model download: %w", err)
	}
	defer os.Remove(partial.Name())
	defer partial.Close()
	if _, err := partial.Write(magic[:]); err != nil {
		return fmt.Errorf("writing model header: %w", err)
	}
	progress := &downloadProgress{total: response.ContentLength, written: int64(len(magic))}
	n, err := io.Copy(io.MultiWriter(partial, progress), response.Body)
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return fmt.Errorf("writing model download: %w", err)
	}
	if n == 0 || (response.ContentLength >= 0 && n+4 != response.ContentLength) {
		return fmt.Errorf("model download is truncated")
	}
	if err := partial.Close(); err != nil {
		return fmt.Errorf("closing model download: %w", err)
	}
	if err := os.Rename(partial.Name(), path); err != nil {
		return fmt.Errorf("saving model download: %w", err)
	}
	return nil
}
