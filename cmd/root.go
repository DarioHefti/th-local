package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/DarioHefti/th-local/internal/config"
	"github.com/DarioHefti/th-local/internal/detect"
	"github.com/DarioHefti/th-local/internal/llm"
	"github.com/DarioHefti/th-local/internal/output"
	"github.com/spf13/cobra"
)

var (
	copyToClipboard bool
	includeGitInfo  bool
	includeTreeInfo bool
	logPrompt       bool
	Version         = "dev"
)

var rootCmd = &cobra.Command{
	Use:   "th [query]",
	Short: "Get shell commands from the local LFM3-1.7B model",
	Long: `th (Terminal Help) - Get shell commands using the local LFM3-1.7B Q4_K_M GGUF

Examples:
  th "list all files modified today"
	th -g "show me the files I changed"
	th -gt "how do i run the tests in this repo"
	th -log "how do i create a new directory"
  th "find large files over 100MB"
`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Usage()
		}

		ctx := context.Background()

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		llmClient, err := llm.NewGenerator(cfg)
		if err != nil {
			return fmt.Errorf("creating LLM generator: %w", err)
		}

		env, err := detect.Detect(includeGitInfo)
		if err != nil {
			return fmt.Errorf("detecting environment: %w", err)
		}

		query := args[0]
		systemPrompt := env.SystemPrompt(detect.PromptOptions{
			IncludeGit:  includeGitInfo,
			IncludeTree: includeTreeInfo,
		})

		if logPrompt {
			output.PrintPrompt(llm.BuildPrompt(systemPrompt, query))
		}

		thinking := output.NewThinkingIndicator()
		command, err := llmClient.GetCommand(ctx, systemPrompt, query, thinking.Tick)
		thinking.Clear()
		if err != nil {
			return fmt.Errorf("getting command: %w", err)
		}

		output.PrintCommand(command, copyToClipboard)

		return nil
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("th version %s\n", Version)
	},
}

func Execute() {
	rootCmd.SetArgs(normalizeArgs(os.Args[1:]))
	if err := rootCmd.Execute(); err != nil {
		output.PrintError(err)
		os.Exit(1)
	}
}

func normalizeArgs(args []string) []string {
	normalized := make([]string, 0, len(args))
	for _, arg := range args {
		if arg == "--version" {
			normalized = append(normalized, "version")
			continue
		}
		if arg == "-log" {
			normalized = append(normalized, "--log")
			continue
		}
		normalized = append(normalized, arg)
	}
	return normalized
}

func init() {
	rootCmd.Flags().BoolVar(&copyToClipboard, "c", false, "Copy result to clipboard")
	rootCmd.Flags().BoolVarP(&includeGitInfo, "git", "g", false, "Include git branch and git status -s output in the prompt")
	rootCmd.Flags().BoolVarP(&includeTreeInfo, "tree", "t", false, "Include top-level workspace hints in the prompt")
	rootCmd.Flags().BoolVar(&logPrompt, "log", false, "Print the final prompt sent to the LLM")
	rootCmd.AddCommand(versionCmd)
}
