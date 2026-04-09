package cmd

import (
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

func TestValidateRootArgs_ConfigModeNoArgs(t *testing.T) {
	command := &cobra.Command{}
	command.Flags().Bool("config", false, "Run setup wizard")
	_ = command.Flags().Set("config", "true")

	if err := validateRootArgs(command, []string{}); err != nil {
		t.Fatalf("validateRootArgs returned error: %v", err)
	}
}

func TestValidateRootArgs_ConfigModeWithArgs(t *testing.T) {
	command := &cobra.Command{}
	command.Flags().Bool("config", false, "Run setup wizard")
	_ = command.Flags().Set("config", "true")

	if err := validateRootArgs(command, []string{"list files"}); err == nil {
		t.Fatal("expected validateRootArgs to fail when --config is used with query")
	}
}

func TestValidateRootArgs_RegularModeOneArg(t *testing.T) {
	command := &cobra.Command{}
	command.Flags().Bool("config", false, "Run setup wizard")

	if err := validateRootArgs(command, []string{"list files"}); err != nil {
		t.Fatalf("validateRootArgs returned error: %v", err)
	}
}

func TestValidateRootArgs_RegularModeTooManyArgs(t *testing.T) {
	command := &cobra.Command{}
	command.Flags().Bool("config", false, "Run setup wizard")

	if err := validateRootArgs(command, []string{"a", "b"}); err == nil {
		t.Fatal("expected validateRootArgs to fail with more than one query argument")
	}
}

func TestNormalizeArgs(t *testing.T) {
	got := normalizeArgs([]string{"--version", "-gt", "-log", "query"})
	want := []string{"version", "-gt", "--log", "query"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeArgs() = %#v, want %#v", got, want)
	}
}
