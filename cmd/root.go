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
			output.PrintPrompt(llm.BuildPromptWithFormat(cfg.PromptFormat, systemPrompt, query))
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

func runSetupWizard() error {
	output.PrintSetupRequired()

	reader := bufio.NewReader(os.Stdin)

	existing, _ := config.Load()

	fmt.Println("th uses a local GGUF model through the built-in llama.cpp bridge.")
	fmt.Println()

	if existing != nil && existing.LocalModelPath != "" {
		fmt.Printf("  Current model path:    %s\n", existing.LocalModelPath)
		fmt.Printf("  Current prompt format: %s\n", existing.PromptFormat)
		fmt.Println()
	}

	autoDetectMessage := "Model path (press Enter to auto-detect the managed th model location): "
	if existing != nil && existing.LocalModelPath != "" {
		autoDetectMessage = fmt.Sprintf("Model path (press Enter to keep %s): ", existing.LocalModelPath)
	} else if managedPath, err := config.DefaultManagedModelPath(); err == nil {
		autoDetectMessage = fmt.Sprintf("Model path (press Enter to auto-detect %s): ", managedPath)
	}

	fmt.Print(autoDetectMessage)
	modelPath, _ := reader.ReadString('\n')
	modelPath = strings.TrimSpace(modelPath)

	if modelPath == "" && existing != nil {
		modelPath = existing.LocalModelPath
	}

	fmt.Println()
	fmt.Println("Prompt format determines how the prompt is structured for your model.")
	fmt.Println("  1) gemma   - Gemma models (default)")
	fmt.Println("  2) chatml  - ChatML models (Qwen, Yi, Mistral-Instruct, etc.)")
	fmt.Println("  3) llama3  - Llama 3 / 3.1 / 3.2 models")
	fmt.Println("  4) raw     - Generic markdown format (works as fallback for most models)")

	currentFormat := config.DefaultPromptFormat
	if existing != nil && existing.PromptFormat != "" {
		currentFormat = existing.PromptFormat
	}
	fmt.Printf("Prompt format (press Enter to keep %s): ", currentFormat)
	formatInput, _ := reader.ReadString('\n')
	formatInput = strings.TrimSpace(formatInput)

	promptFormat := currentFormat
	switch formatInput {
	case "1", "gemma":
		promptFormat = config.PromptFormatGemma
	case "2", "chatml":
		promptFormat = config.PromptFormatChatML
	case "3", "llama3":
		promptFormat = config.PromptFormatLlama3
	case "4", "raw":
		promptFormat = config.PromptFormatRaw
	case "":
		// keep current
	default:
		if config.IsValidPromptFormat(formatInput) {
			promptFormat = formatInput
		} else {
			fmt.Printf("Unknown format %q, using %s\n", formatInput, currentFormat)
		}
	}

	cfg := &config.Config{
		Provider:       config.ProviderLocal,
		Model:          config.DefaultLocalModel,
		LocalModelPath: modelPath,
		PromptFormat:   promptFormat,
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Println()
	output.PrintSuccess(fmt.Sprintf("Configuration saved to %s", config.ConfigPath()))
	fmt.Printf("  Model path:    %s\n", describeModelPath(cfg.LocalModelPath))
	fmt.Printf("  Prompt format: %s\n", cfg.PromptFormat)

	return nil
}

func describeModelPath(path string) string {
	if path == "" {
		return "(auto-detect)"
	}
	return path
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
