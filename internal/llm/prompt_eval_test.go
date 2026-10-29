//go:build cgo

package llm

import (
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/DarioHefti/th-local/internal/detect"
	"github.com/DarioHefti/th-local/internal/llm/llamacpp"
)

// Screening checks syntax and selected intent, not full correctness.
// Generated commands are parsed with zsh -n, never executed.
var promptEvalCases = []struct{ query, want string }{
	{"search for filename in folder", `find\s+\.\s+.*-i?name\s+['"]?\*filename\*`},
	{"how to find string in folder mac os terminal", `grep\b.*-\w*[rR]\w*.*string.*(\.|folder)`},
	{"find files whose names contain report under the current directory", `find\b.*-i?name\s+['"]?\*report\*`},
	{"find all .go files recursively in the current folder", `find\b.*-name\s+['"]?\*\.go`},
	{"find directories named node_modules under the current folder", `find\b.*(-type d.*-name.*node_modules|-name.*node_modules.*-type d)`},
	{"search file contents recursively for TODO in src", `grep\b.*TODO.*src`},
	{"show line numbers of lines containing error in app.log", `grep\b.*-\w*n\w*.*error.*app\.log`},
	{"list files in the current directory", `ls\b`},
	{"list all files including hidden files with details", `ls\b.*(-\w*a\w*l|-\w*l\w*a|-[aA].*-l|-l.*-[aA])`},
	{"print the current directory", `pwd\b`},
	{"create a directory named project files", `mkdir\b.*['"]project files['"]`},
	{"copy report.txt to backup.txt", `cp\b.*report\.txt.*backup\.txt`},
	{"rename old.txt to new.txt", `mv\b.*old\.txt.*new\.txt`},
	{"remove the empty directory old-cache", `rmdir\b.*old-cache`},
	{"show the first 20 lines of app.log", `head\b.*(-n\s*20|-20).*app\.log`},
	{"show the last 50 lines of app.log", `tail\b.*(-n\s*50|-50).*app\.log`},
	{"count lines in data.csv", `wc\b.*-l.*data\.csv`},
	{"sort names.txt and remove duplicate lines", `sort\b.*(-u.*names\.txt|names\.txt.*\|\s*uniq)`},
	{"show disk usage of the current folder in human-readable form", `du\b.*-\w*h`},
	{"show available disk space in human-readable form", `df\b.*-h`},
	{"find files larger than 100 MB in the current folder", `find\b.*-size\s+\+100[Mm]`},
	{"find files modified in the last 24 hours", `find\b.*(-mtime\s+-1|-mmin\s+-1440)`},
	{"show git status", `git\s+status\b`},
	{"show unstaged git changes", `git\s+diff\b`},
	{"show the last five git commits on one line each", `git\s+log\b.*(--oneline.*(-5|-n\s*5)|(-5|-n\s*5).*--oneline)`},
	{"show the current git branch", `git\s+(branch\s+--show-current|rev-parse\s+--abbrev-ref\s+HEAD)`},
	{"find the process listening on TCP port 8080", `lsof\b.*-iTCP:8080.*-sTCP:LISTEN`},
	{"show running processes containing python", `(ps\b.*\|\s*grep.*python|pgrep\b.*python)`},
	{"fetch only HTTP response headers from https://example.com", `curl\b.*(-\w*I|--head).*https://example\.com`},
	{"create a gzip compressed tar archive backup.tar.gz of the src folder", `tar\b.*(-[czvf]*c[czvf]*z[czvf]*f|-?[czvf]*z[czvf]*c[czvf]*f).*backup\.tar\.gz.*src`},
}

func TestShellPromptEvaluation(t *testing.T) {
	if testing.Short() || os.Getenv("TH_RUN_PROMPT_EVAL") != "1" {
		t.Skip("set TH_RUN_PROMPT_EVAL=1 and TH_INFERENCE_MODEL_PATH to evaluate 30 requests")
	}
	if runtime.GOOS != "darwin" {
		t.Skip("this evaluation targets macOS commands")
	}
	zsh, err := exec.LookPath("zsh")
	if err != nil {
		t.Fatal(err)
	}
	path, err := resolveInferenceTestModelPath()
	if err != nil || path == "" {
		t.Fatalf("model path missing: %v", err)
	}
	engine, err := llamacpp.New(path, 2048, min(4, runtime.NumCPU()), 0)
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	env := &detect.Environment{OS: "darwin", Shell: "zsh"}
	passed := 0
	for i, tc := range promptEvalCases {
		raw, err := engine.Predict(BuildPrompt(env.SystemPrompt(detect.PromptOptions{}), tc.query), localPredictTokens)
		if err != nil {
			t.Fatal(err)
		}
		command := cleanupCommandOutput(stripTurnMarkers(raw))
		parser := exec.Command(zsh, "-f", "-n")
		parser.Stdin = strings.NewReader(command)
		ok := parser.Run() == nil && command != "" && !strings.ContainsAny(command, "\n\r#`") && regexp.MustCompile("^(?:"+tc.want+")").MatchString(command)
		for _, invalid := range []string{"<think>", "</think>", "-printf", "--human-readable", "--max-depth", "find /", "/path/to/", "/proc/"} {
			if strings.Contains(command, invalid) {
				ok = false
			}
		}
		if tc.query == "show unstaged git changes" && (strings.Contains(command, "--staged") || strings.Contains(command, "--cached")) {
			ok = false
		}
		if ok {
			passed++
		}
		t.Logf("%02d pass=%v query=%q output=%q", i+1, ok, tc.query, command)
	}
	t.Logf("SUMMARY: %d/%d checks passed", passed, len(promptEvalCases))
	if passed != len(promptEvalCases) {
		t.Errorf("%d requests failed screening; inspect logged outputs", len(promptEvalCases)-passed)
	}
}
