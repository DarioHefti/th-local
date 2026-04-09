package detect

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeShellName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: `/usr/bin/bash`, want: "bash"},
		{input: `C:\\Program Files\\PowerShell\\7\\pwsh.exe`, want: "pwsh"},
		{input: `C:\\Windows\\System32\\cmd.exe`, want: "cmd"},
		{input: `powershell_ise.exe`, want: "powershell"},
		{input: `unknown-process`, want: ""},
	}

	for _, tt := range tests {
		if got := normalizeShellName(tt.input); got != tt.want {
			t.Fatalf("normalizeShellName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestDetectWindowsShellFromEnvironment(t *testing.T) {
	t.Run("powershell core from module path", func(t *testing.T) {
		if got := detectWindowsShellFromEnvironment(`C:\\Users\\heda\\Documents\\PowerShell\\Modules;C:\\Program Files\\PowerShell\\7\\Modules`, `C:\\Windows\\System32\\cmd.exe`); got != "pwsh" {
			t.Fatalf("detectWindowsShellFromEnvironment() = %q, want %q", got, "pwsh")
		}
	})

	t.Run("cmd from comspec", func(t *testing.T) {
		if got := detectWindowsShellFromEnvironment("", `C:\\Windows\\System32\\cmd.exe`); got != "cmd" {
			t.Fatalf("detectWindowsShellFromEnvironment() = %q, want %q", got, "cmd")
		}
	})
}

func TestResolveShell(t *testing.T) {
	t.Run("prefers explicit override", func(t *testing.T) {
		got := resolveShell("windows", "pwsh.exe", "fish", `C:\\Program Files\\PowerShell\\7\\pwsh.exe`, "", `C:\\Windows\\System32\\cmd.exe`)
		if got != "fish" {
			t.Fatalf("resolveShell() = %q, want %q", got, "fish")
		}
	})

	t.Run("prefers parent shell over cmd hints", func(t *testing.T) {
		got := resolveShell("windows", "pwsh.exe", "", "", "", `C:\\Windows\\System32\\cmd.exe`)
		if got != "pwsh" {
			t.Fatalf("resolveShell() = %q, want %q", got, "pwsh")
		}
	})

	t.Run("falls back to powershell on windows", func(t *testing.T) {
		got := resolveShell("windows", "", "", "", "", "")
		if got != "powershell" {
			t.Fatalf("resolveShell() = %q, want %q", got, "powershell")
		}
	})

	t.Run("falls back to bash on non-windows", func(t *testing.T) {
		got := resolveShell("linux", "", "", "", "", "")
		if got != "bash" {
			t.Fatalf("resolveShell() = %q, want %q", got, "bash")
		}
	})
}

func TestDetectShellFromProcessNames(t *testing.T) {
	processNames := []string{"OpenConsole.exe", "Code.exe", "pwsh.exe"}
	if got := detectShellFromProcessNames(processNames); got != "pwsh" {
		t.Fatalf("detectShellFromProcessNames() = %q, want %q", got, "pwsh")
	}
}

func TestDetectPowerShellShell(t *testing.T) {
	if got := detectPowerShellShell(`C:\\Users\\heda\\Documents\\WindowsPowerShell\\Modules;C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\Modules`); got != "powershell" {
		t.Fatalf("detectPowerShellShell() = %q, want %q", got, "powershell")
	}

	if got := detectPowerShellShell(`C:\\Users\\heda\\Documents\\PowerShell\\Modules;C:\\Program Files\\PowerShell\\7\\Modules`); got != "pwsh" {
		t.Fatalf("detectPowerShellShell() = %q, want %q", got, "pwsh")
	}
}

func TestEnvironmentSystemPrompt(t *testing.T) {
	env := &Environment{
		OS:           "linux",
		Shell:        "bash",
		ShellVersion: "5.1.0",
		CWD:          "/test/dir",
	}

	prompt := env.SystemPrompt(PromptOptions{})

	if !strings.Contains(prompt, "linux") {
		t.Error("prompt should contain OS")
	}
	if !strings.Contains(prompt, "bash") {
		t.Error("prompt should contain Shell")
	}
	if strings.Contains(prompt, "Workspace:") {
		t.Error("prompt should not contain workspace info by default")
	}
	if strings.Contains(prompt, "```") {
		t.Error("prompt should not contain code fences")
	}
	if strings.Contains(prompt, "Git branch") {
		t.Error("prompt should not contain git info by default")
	}
}

func TestEnvironmentSystemPromptForPowerShell(t *testing.T) {
	env := &Environment{
		OS:           "windows",
		Shell:        "powershell",
		ShellVersion: "5.1",
		CWD:          `C:\\work`,
	}

	prompt := env.SystemPrompt(PromptOptions{})

	if !strings.Contains(prompt, "Use PowerShell syntax only") {
		t.Fatalf("prompt missing PowerShell-specific instruction: %q", prompt)
	}
	if !strings.Contains(prompt, "Prefer cmdlets over cmd.exe built-ins") {
		t.Fatalf("prompt missing PowerShell guidance: %q", prompt)
	}
}

func TestEnvironmentSystemPromptIncludesOptionalContext(t *testing.T) {
	env := &Environment{
		OS:        "linux",
		Shell:     "bash",
		CWD:       t.TempDir(),
		GitBranch: "main",
		GitStatus: "M  README.md\n?? notes.txt",
	}

	if err := os.MkdirAll(filepath.Join(env.CWD, "cmd"), 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(env.CWD, "go.mod"), []byte("module example"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	prompt := env.SystemPrompt(PromptOptions{IncludeGit: true, IncludeTree: true})

	if !strings.Contains(prompt, "Git branch: main") {
		t.Fatalf("prompt should contain git branch when enabled: %q", prompt)
	}
	if !strings.Contains(prompt, "Git status -s output:\nM  README.md\n?? notes.txt") {
		t.Fatalf("prompt should contain git status when enabled: %q", prompt)
	}
	if !strings.Contains(prompt, "Workspace hints: go.mod, cmd/") {
		t.Fatalf("prompt should contain workspace hints when enabled: %q", prompt)
	}
}

func TestGetCWD(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Skip("skipping: could not get cwd")
	}
	result := getCWD()
	if result != cwd {
		t.Errorf("getCWD: got %q, want %q", result, cwd)
	}
}

func TestGetWorkspaceHints(t *testing.T) {
	tmpDir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(tmpDir, "cmd"), 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "third_party"), 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "internal"), 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module example"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	hints := getWorkspaceHints(tmpDir)

	if !strings.Contains(hints, "go.mod") {
		t.Fatalf("hints should include go.mod, got %q", hints)
	}
	if !strings.Contains(hints, "cmd/") {
		t.Fatalf("hints should include cmd/, got %q", hints)
	}
	if !strings.Contains(hints, "internal/") {
		t.Fatalf("hints should include internal/, got %q", hints)
	}
	if strings.Contains(hints, "third_party/") {
		t.Fatalf("hints should not include unlisted directories, got %q", hints)
	}
}

func TestGetWorkspaceHintsLimitsOutput(t *testing.T) {
	tmpDir := t.TempDir()

	for _, name := range workspaceHintOrder {
		fullPath := filepath.Join(tmpDir, strings.TrimSuffix(name, "/"))
		if strings.HasSuffix(name, "/") {
			if err := os.MkdirAll(fullPath, 0755); err != nil {
				t.Fatalf("MkdirAll failed: %v", err)
			}
			continue
		}
		if err := os.WriteFile(fullPath, []byte("x"), 0644); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}
	}

	hints := getWorkspaceHints(tmpDir)
	parts := strings.Split(hints, ", ")

	if len(parts) != maxWorkspaceHints {
		t.Fatalf("hints should be capped at %d, got %d in %q", maxWorkspaceHints, len(parts), hints)
	}
	if parts[0] != workspaceHintOrder[0] {
		t.Fatalf("hints should preserve priority order, got %q", hints)
	}
}

func TestDetect(t *testing.T) {
	env, err := Detect(false)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if env.OS == "" {
		t.Error("OS should not be empty")
	}
	if env.Shell == "" {
		t.Error("Shell should not be empty")
	}
	if env.CWD == "" {
		t.Error("CWD should not be empty")
	}
}
