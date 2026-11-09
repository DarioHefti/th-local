package output

import (
	"bytes"
	"testing"
	"time"
)

func TestThinkingIndicatorCyclesDots(t *testing.T) {
	var out bytes.Buffer
	indicator := &ThinkingIndicator{out: &out}
	for range 4 {
		indicator.next = time.Time{}
		indicator.Tick("token")
	}
	want := "\r\x1b[2mthinking.  \x1b[0m\r\x1b[2mthinking.. \x1b[0m\r\x1b[2mthinking...\x1b[0m\r\x1b[2mthinking.  \x1b[0m"
	if got := out.String(); got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestThinkingIndicatorWaitsForNextFrame(t *testing.T) {
	var out bytes.Buffer
	indicator := &ThinkingIndicator{out: &out}
	indicator.Tick("first")
	drawn := out.Len()
	indicator.Tick("second")
	if out.Len() != drawn {
		t.Fatalf("redrew before the next frame was due: %q", out.String())
	}
}

func TestThinkingIndicatorClearErasesLine(t *testing.T) {
	var out bytes.Buffer
	indicator := &ThinkingIndicator{out: &out}
	indicator.Clear()
	if got, want := out.String(), "\r\x1b[K"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestThinkingIndicatorIsSilentWithoutTerminal(t *testing.T) {
	indicator := &ThinkingIndicator{}
	indicator.Tick("token")
	indicator.Clear()
}
