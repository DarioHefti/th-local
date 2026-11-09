package cmd

import (
	"reflect"
	"testing"
)

func TestNormalizeArgs(t *testing.T) {
	got := normalizeArgs([]string{"--version", "-gt", "-log", "query"})
	want := []string{"version", "-gt", "--log", "query"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeArgs() = %#v, want %#v", got, want)
	}
}
