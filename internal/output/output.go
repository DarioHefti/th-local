package output

import (
	"fmt"
	"os"

	"github.com/atotto/clipboard"
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

func PrintError(err error) {
	fmt.Fprintf(os.Stderr, "\x1b[31m✗ Error: %v\x1b[0m\n", err)
}

func PrintInfo(msg string) {
	fmt.Printf("\x1b[34mℹ %s\x1b[0m\n", msg)
}

func PrintPrompt(prompt string) {
	fmt.Fprintf(os.Stderr, "\x1b[35m--- Prompt Sent To LLM ---\x1b[0m\n%s\n", prompt)
}

func PrintSuccess(msg string) {
	fmt.Printf("\x1b[32m✓ %s\x1b[0m\n", msg)
}

func PrintAuthPrompt() {
	fmt.Printf("\x1b[33m🔐 Opening browser for authentication...\x1b[0m\n")
}

func PrintSetupRequired() {
	fmt.Printf("\x1b[33m🔧 First-time setup required!\x1b[0m\n\n")
}
