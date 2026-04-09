package detect

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

const maxShellDetectionDepth = 8

const maxWorkspaceHints = 8

var workspaceHintOrder = []string{
	"go.mod",
	"package.json",
	"pyproject.toml",
	"Cargo.toml",
	"Makefile",
	"README.md",
	"cmd/",
	"internal/",
	"src/",
	"scripts/",
	"tests/",
	"test/",
	"model/",
	"Dockerfile",
	"docker-compose.yml",
	"compose.yaml",
}

type Environment struct {
	OS           string
	Shell        string
	ShellVersion string
	CWD          string
	GitBranch    string
	GitStatus    string
}

type PromptOptions struct {
	IncludeGit  bool
	IncludeTree bool
}

func Detect(includeGit bool) (*Environment, error) {
	env := &Environment{
		OS:  runtime.GOOS,
		CWD: getCWD(),
	}

	shell, err := detectShell()
	if err != nil {
		return nil, fmt.Errorf("detecting shell: %w", err)
	}
	env.Shell = shell

	shellVersion, err := detectShellVersion(shell)
	if err != nil {
		env.ShellVersion = "unknown"
	} else {
		env.ShellVersion = shellVersion
	}

	if includeGit {
		env.GitBranch, env.GitStatus = detectGit()
	}

	return env, nil
}

func getCWD() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "/"
	}
	return cwd
}

func getWorkspaceHints(cwd string) string {
	entries, err := os.ReadDir(cwd)
	if err != nil {
		return ""
	}

	available := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if entry.IsDir() {
			available[name+"/"] = struct{}{}
			continue
		}
		available[name] = struct{}{}
	}

	hints := make([]string, 0, maxWorkspaceHints)
	for _, candidate := range workspaceHintOrder {
		if _, ok := available[candidate]; !ok {
			continue
		}
		hints = append(hints, candidate)
		if len(hints) >= maxWorkspaceHints {
			break
		}
	}

	return strings.Join(hints, ", ")
}

func detectShell() (string, error) {
	processNames, err := detectParentProcessNames(maxShellDetectionDepth)
	if err == nil {
		if shell := detectShellFromProcessNames(processNames); shell != "" {
			return shell, nil
		}
	}

	return resolveShell(
		runtime.GOOS,
		"",
		os.Getenv("SHELL_SPECIAL"),
		os.Getenv("SHELL"),
		os.Getenv("PSModulePath"),
		os.Getenv("COMSPEC"),
	), nil
}

func resolveShell(goos, parentProcessName, shellSpecial, shellEnv, psModulePath, comspec string) string {
	if shell := normalizeShellName(shellSpecial); shell != "" {
		return shell
	}

	if shell := normalizeShellName(parentProcessName); shell != "" {
		return shell
	}

	if shell := normalizeShellName(shellEnv); shell != "" {
		return shell
	}

	if goos == "windows" {
		if shell := detectWindowsShellFromEnvironment(psModulePath, comspec); shell != "" {
			return shell
		}

		return "powershell"
	}

	return "bash"
}

func detectShellFromProcessNames(processNames []string) string {
	for _, processName := range processNames {
		if shell := normalizeShellName(processName); shell != "" {
			return shell
		}
	}

	return ""
}

func detectParentProcessName() (string, error) {
	pid := os.Getppid()
	if pid <= 0 {
		return "", fmt.Errorf("invalid parent pid %d", pid)
	}

	if runtime.GOOS == "windows" {
		return detectWindowsProcessName(pid)
	}

	cmd := exec.Command("ps", "-p", fmt.Sprintf("%d", pid), "-o", "comm=")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func detectParentProcessNames(maxDepth int) ([]string, error) {
	if maxDepth <= 0 {
		return nil, fmt.Errorf("invalid max depth %d", maxDepth)
	}

	pid := os.Getppid()
	if pid <= 0 {
		return nil, fmt.Errorf("invalid parent pid %d", pid)
	}

	var processNames []string
	for depth := 0; depth < maxDepth && pid > 0; depth++ {
		name, nextPID, err := detectProcessInfo(pid)
		if err != nil {
			if len(processNames) > 0 {
				return processNames, nil
			}
			return nil, err
		}

		processNames = append(processNames, name)
		if nextPID == pid {
			break
		}
		pid = nextPID
	}

	return processNames, nil
}

func detectProcessInfo(pid int) (string, int, error) {
	if runtime.GOOS == "windows" {
		return detectWindowsProcessInfo(pid)
	}

	cmd := exec.Command("ps", "-p", fmt.Sprintf("%d", pid), "-o", "ppid=", "-o", "comm=")
	output, err := cmd.Output()
	if err != nil {
		return "", 0, err
	}

	fields := strings.Fields(strings.TrimSpace(string(output)))
	if len(fields) < 2 {
		return "", 0, fmt.Errorf("unexpected ps output for pid %d", pid)
	}

	parentPID, err := strconv.Atoi(fields[0])
	if err != nil {
		return "", 0, fmt.Errorf("parsing parent pid for %d: %w", pid, err)
	}

	return fields[1], parentPID, nil
}

func detectWindowsProcessInfo(pid int) (string, int, error) {
	cmd := exec.Command(
		"powershell",
		"-NoProfile",
		"-Command",
		fmt.Sprintf(`$p = Get-CimInstance Win32_Process -Filter "ProcessId = %d"; if ($null -eq $p) { exit 1 }; Write-Output ($p.Name + "|" + $p.ParentProcessId)`, pid),
	)
	output, err := cmd.Output()
	if err != nil {
		return "", 0, err
	}

	parts := strings.Split(strings.TrimSpace(string(output)), "|")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("unexpected process output for pid %d", pid)
	}

	parentPID, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return "", 0, fmt.Errorf("parsing parent pid for %d: %w", pid, err)
	}

	return strings.TrimSpace(parts[0]), parentPID, nil
}

func detectWindowsProcessName(pid int) (string, error) {
	cmd := exec.Command("tasklist", "/FO", "CSV", "/NH", "/FI", fmt.Sprintf("PID eq %d", pid))
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	rawOutput := strings.TrimSpace(string(output))
	if rawOutput == "" || strings.HasPrefix(rawOutput, "INFO:") {
		return "", fmt.Errorf("process %d not found", pid)
	}

	reader := csv.NewReader(strings.NewReader(rawOutput))
	record, err := reader.Read()
	if err != nil && err != io.EOF {
		return "", err
	}
	if len(record) == 0 {
		return "", fmt.Errorf("process %d not found", pid)
	}

	return record[0], nil
}

func detectWindowsShellFromEnvironment(psModulePath, comspec string) string {
	if shell := detectPowerShellShell(psModulePath); shell != "" {
		return shell
	}

	return normalizeShellName(comspec)
}

func detectPowerShellShell(modulePath string) string {
	lowerModulePath := strings.ToLower(modulePath)
	switch {
	case strings.Contains(lowerModulePath, `\\powershell\\7`), strings.Contains(lowerModulePath, "/powershell/7"):
		return "pwsh"
	case strings.Contains(lowerModulePath, "windowspowershell"):
		return "powershell"
	default:
		return ""
	}
}

func normalizeShellName(raw string) string {
	trimmed := strings.Trim(strings.TrimSpace(raw), `"'`)
	if trimmed == "" {
		return ""
	}

	base := strings.ToLower(trimmed)
	if idx := strings.LastIndexAny(base, `/\\`); idx >= 0 {
		base = base[idx+1:]
	}
	base = strings.TrimSuffix(base, ".exe")

	switch {
	case strings.Contains(base, "powershell"):
		return "powershell"
	case base == "pwsh" || strings.HasPrefix(base, "pwsh-"):
		return "pwsh"
	case base == "cmd":
		return "cmd"
	case base == "bash":
		return "bash"
	case base == "zsh":
		return "zsh"
	case base == "fish":
		return "fish"
	case base == "sh":
		return "sh"
	case base == "nu" || base == "nushell":
		return "nushell"
	case base == "xonsh":
		return "xonsh"
	case base == "tcsh":
		return "tcsh"
	case base == "csh":
		return "csh"
	case base == "ksh":
		return "ksh"
	case base == "dash":
		return "dash"
	case base == "elvish":
		return "elvish"
	default:
		return ""
	}
}

func detectShellVersion(shell string) (string, error) {
	var cmd *exec.Cmd

	switch shell {
	case "bash":
		cmd = exec.Command("bash", "--version")
	case "sh":
		cmd = exec.Command("sh", "--version")
	case "zsh":
		cmd = exec.Command("zsh", "--version")
	case "fish":
		cmd = exec.Command("fish", "--version")
	case "powershell", "pwsh":
		cmd = exec.Command("pwsh", "-Version")
	case "cmd":
		cmd = exec.Command("cmd", "/c", "ver")
	default:
		return "", fmt.Errorf("unknown shell: %s", shell)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func detectGit() (branch, status string) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err == nil {
		branch = strings.TrimSpace(string(output))
	}

	cmd = exec.Command("git", "status", "-s")
	output, err = cmd.Output()
	if err == nil {
		status = strings.TrimSpace(string(output))
	}

	return branch, status
}

func (e *Environment) SystemPrompt(options PromptOptions) string {
	shellInstructions := shellStyleInstructions(e.OS, e.Shell)

	prompt := fmt.Sprintf(`Return one shell command.
Context:
- OS: %s
	- Shell: %s`, e.OS, e.Shell)

	if options.IncludeGit && e.GitBranch != "" {
		prompt += fmt.Sprintf("\n- Git branch: %s", e.GitBranch)
		if e.GitStatus != "" {
			prompt += fmt.Sprintf("\n- Git status -s output:\n%s", e.GitStatus)
		}
	}

	if options.IncludeTree {
		if workspaceHints := getWorkspaceHints(e.CWD); workspaceHints != "" {
			prompt += fmt.Sprintf("\n- Workspace hints: %s", workspaceHints)
		}
	}

	prompt += fmt.Sprintf(`
Rules:
- Output only the raw command.
- No explanation, comments, markdown, or backticks.
- Use syntax valid for %s only.
%s`, e.Shell, shellInstructions)

	return prompt
}

func shellStyleInstructions(goos, shell string) string {
	switch shell {
	case "powershell", "pwsh":
		return `- Use PowerShell syntax only.
- Prefer cmdlets over cmd.exe built-ins.
- Do not wrap the command in cmd /c, powershell -Command, or pwsh -Command.`
	case "cmd":
		return `- Use cmd.exe syntax only.
- Do not return PowerShell or POSIX shell syntax.`
	case "bash", "sh", "zsh", "fish", "dash", "ksh":
		return fmt.Sprintf(`- Use %s-compatible shell syntax only.
- Do not return Windows cmd.exe or PowerShell syntax.`, shell)
	default:
		return fmt.Sprintf(`- Use syntax that is valid for %s on %s.`, shell, goos)
	}
}
