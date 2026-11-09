package output

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/atotto/clipboard"
	"golang.org/x/term"
)

func PrintCommand(cmd string, copyToClipboard bool) {
	fmt.Printf("\x1b[36m💡\x1b[0m %s\n", cmd)

	if copyToClipboard {
		if err := clipboard.WriteAll(cmd); err != nil {
			fmt.Fprintf(os.Stderr, "\x1b[33m⚠ Failed to copy to clipboard: %v\x1b[0m\n", err)
		} else {
			fmt.Printf("\x1b[32m📋 Copied to clipboard!\x1b[0m\n")
		}
	}
}

const thinkingFrameInterval = 300 * time.Millisecond

var thinkingFrames = []string{".  ", ".. ", "..."}

// ThinkingIndicator animates "thinking..." on stderr while the model reasons.
// Stdout carries only the command.
type ThinkingIndicator struct {
	out   io.Writer
	frame int
	next  time.Time
}

// NewThinkingIndicator returns an indicator that stays silent when stderr is not a terminal.
func NewThinkingIndicator() *ThinkingIndicator {
	if !term.IsTerminal(int(os.Stderr.Fd())) {
		return &ThinkingIndicator{}
	}
	return &ThinkingIndicator{out: os.Stderr}
}

// Tick runs for each generated token and advances the dots when a frame is due.
func (t *ThinkingIndicator) Tick(string) {
	if t.out == nil {
		return
	}
	now := time.Now()
	if now.Before(t.next) {
		return
	}
	t.next = now.Add(thinkingFrameInterval)
	fmt.Fprintf(t.out, "\r\x1b[2mthinking%s\x1b[0m", thinkingFrames[t.frame%len(thinkingFrames)])
	t.frame++
}

func (t *ThinkingIndicator) Clear() {
	if t.out != nil {
		fmt.Fprint(t.out, "\r\x1b[K")
	}
}

func PrintError(err error) {
	fmt.Fprintf(os.Stderr, "\x1b[31m✗ Error: %v\x1b[0m\n", err)
}

func PrintInfo(msg string) {
	fmt.Printf("\x1b[34mℹ %s\x1b[0m\n", msg)
}

func PrintPrompt(prompt string) {
	fmt.Fprintf(os.Stderr, "\x1b[35m--- Prompt Sent To LLM ---\x1b[0m\n%s\n", prompt)
}

func PrintAuthPrompt() {
	fmt.Printf("\x1b[33m🔐 Opening browser for authentication...\x1b[0m\n")
}
