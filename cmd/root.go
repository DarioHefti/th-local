package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/DarioHefti/th-local/internal/config"
	"github.com/DarioHefti/th-local/internal/detect"
	"github.com/DarioHefti/th-local/internal/llm"
	"github.com/DarioHefti/th-local/internal/output"
	"github.com/spf13/cobra"
)

var (
	copyToClipboard bool
	configFlag      bool
	includeGitInfo  bool
	includeTreeInfo bool
	logPrompt       bool
	Version         = "dev"
)

var rootCmd = &cobra.Command{
	Use:   "th [query]",
	Short: "Get shell commands from the local Gemma model",
	Long: `th (Terminal Help) - Get shell commands using the local Gemma model

Examples:
  th "list all files modified today"
	th -g "show me the files I changed"
	th -gt "how do i run the tests in this repo"
	th -log "how do i create a new directory"
  th "find large files over 100MB"
  th --config  # Re-run setup wizard
`,
	Args: validateRootArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if configFlag {
			return runSetupWizard()
		}

		if len(args) == 0 {
			return cmd.Usage()
		}

		ctx := context.Background()

		cfg, err := config.Load()
		if err != nil {
			if config.IsConfigNotFound(err) {
				if err := runSetupWizard(); err != nil {
					return err
				}
				cfg, err = config.Load()
				if err != nil {
					return fmt.Errorf("loading config after setup: %w", err)
				}
			} else {
				return fmt.Errorf("loading config: %w", err)
			}
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

		command, err := llmClient.GetCommand(ctx, systemPrompt, query)
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
		if arg == "-log" {
			normalized = append(normalized, "--log")
			continue
		}
		normalized = append(normalized, arg)
	}
	return normalized
}

func runSetupWizard() error {
	output.PrintSetupRequired()

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Model: local Gemma 4 IQ2_M")
	fmt.Println("th only uses the local Gemma GGUF through the built-in llama.cpp bridge.")
	fmt.Println()

	autoDetectMessage := "Model path (press Enter to auto-detect the managed th model location): "
	if managedPath, err := config.DefaultManagedModelPath(); err == nil {
		autoDetectMessage = fmt.Sprintf("Model path (press Enter to auto-detect %s): ", managedPath)
	}

	fmt.Print(autoDetectMessage)
	modelPath, _ := reader.ReadString('\n')
	modelPath = strings.TrimSpace(modelPath)

	cfg := &config.Config{
		Provider:       config.ProviderLocal,
		Model:          config.DefaultLocalModel,
		LocalModelPath: modelPath,
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	output.PrintSuccess(fmt.Sprintf("Configuration saved to %s", config.ConfigPath()))

	return nil
}

func validateRootArgs(cmd *cobra.Command, args []string) error {
	cfgMode, err := cmd.Flags().GetBool("config")
	if err != nil {
		return err
	}

	if cfgMode {
		if len(args) > 0 {
			return fmt.Errorf("--config does not accept a query argument")
		}
		return nil
	}

	return cobra.MaximumNArgs(1)(cmd, args)
}

func init() {
	rootCmd.Flags().BoolVar(&copyToClipboard, "c", false, "Copy result to clipboard")
	rootCmd.Flags().BoolVarP(&includeGitInfo, "git", "g", false, "Include git branch and git status -s output in the prompt")
	rootCmd.Flags().BoolVarP(&includeTreeInfo, "tree", "t", false, "Include top-level workspace hints in the prompt")
	rootCmd.Flags().BoolVar(&logPrompt, "log", false, "Print the final prompt sent to the LLM")
	rootCmd.Flags().BoolVar(&configFlag, "config", false, "Run setup wizard")
	rootCmd.AddCommand(versionCmd)
}
