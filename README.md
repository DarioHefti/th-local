# th - Terminal Help

Get shell commands from LiquidAI's local LFM3-1.7B model directly in your terminal.

## Installation

The installer downloads the `th` binary and LiquidAI's official `LFM3-1.7B-Q4_K_M.gguf` (4-bit quantization) from [Hugging Face](https://huggingface.co/LiquidAI/LFM3-1.7B-GGUF), then runs the local model through an in-process `llama.cpp` bridge. No MLX runtime or separate model server is used.

The vendored runtime is pinned to `llama.cpp` v0.8.0 (b11960), which adds LFM3 support. Inference uses LiquidAI's recommended temperature 0.1, top-k 50, and repetition penalty 1.1. The LFM3 chat template disables thinking for direct shell-command answers. This is the only supported model and prompt format.

### Command-quality evaluation

The system prompt was compared on 30 macOS requests. The selected structured prompt passed 16/30 syntax-and-intent screening checks with a fixed sampling seed. This small quantized model still sometimes produces reasoning, incorrect paths, or unsuitable commands. Always review suggestions before using them; `th` does not execute them.

Run the production evaluation on macOS (generated commands are parsed, never executed):

```bash
TH_RUN_PROMPT_EVAL=1 TH_INFERENCE_MODEL_PATH="$PWD/model/LFM3-1.7B-Q4_K_M.gguf" go test -v ./internal/llm -run TestShellPromptEvaluation -count=1
```

The evaluation intentionally fails if any request fails screening. Normal unit tests do not load the model or run this evaluation.

Prebuilt release binaries are currently published for:

- Linux amd64
- macOS arm64
- Windows amd64

For other platform and architecture combinations, build from source.

### Linux amd64 / macOS arm64

```bash
curl -sSL https://raw.githubusercontent.com/DarioHefti/th-local/refs/heads/main/scripts/install.sh | bash
```

### Windows

```powershell
irm https://raw.githubusercontent.com/DarioHefti/th-local/refs/heads/main/scripts/install.ps1 | iex
```

### From Source

Initialize the vendored `llama.cpp` checkout first if you cloned without `--recursive`:

```bash
git submodule update --init --recursive
```

### From Source on macOS/Linux/WSL

```bash
make build
make model
```

`make build` runs the native `llama.cpp` build first and then builds `th`. `make model` downloads the GGUF into the managed per-user model directory.

### From Source on Windows

Use PowerShell with a MinGW-w64 toolchain available on `PATH`.

```powershell
git submodule update --init --recursive
powershell -ExecutionPolicy Bypass -File scripts/build-llama.ps1
powershell -ExecutionPolicy Bypass -File scripts/download-model.ps1
go build -ldflags "-X github.com/DarioHefti/th-local/cmd.Version=dev" -o th.exe .
```

### Building on WSL

Use a Linux shell inside WSL, not PowerShell, and install the native toolchain first:

```bash
sudo apt update
sudo apt install -y build-essential cmake git pkg-config

cd ~/src/th-local
git submodule update --init --recursive
make build
make model
make test
./th --version
```

Notes:

- Building from `/mnt/c/...` works, but it is slower than building inside the WSL filesystem.
- If builds are slow, move the repo to a Linux path such as `~/src/th`.
- `make build` already runs the Unix `llama.cpp` native build step automatically.

## Usage

```bash
# Get a command for listing all files modified today
th "list all files modified today"

# Add git context
th -g "show me what to commit"

# Add git and tree context
th -gt "how do i run the tests in this repo"

# Print the exact prompt sent to the LLM
th -log "how do i create a new directory"

# Find large files over 100MB and copy to clipboard
th "find large files over 100MB" --c

# Re-run setup wizard
th --config
```

## Setup

1. Run `th --config`.
2. Press Enter to use portable, per-user model discovery, or provide a path to an existing `LFM3-1.7B-Q4_K_M.gguf`.

With no explicit path, `th` uses the user cache or an existing `model/LFM3-1.7B-Q4_K_M.gguf` in the project. If neither exists, it automatically downloads the default GGUF into the user cache on first use. Interrupted or invalid downloads are not installed as model files. Explicit paths must exist; clear a missing override to restore automatic setup.

There is no model or prompt-format selector. Saved configuration defaults contain no machine-specific paths. An existing copy of the same GGUF can be reused without downloading it again.

### Local MVP Requirements

The local runtime uses a direct `llama.cpp` bridge. For source builds, build the native libraries with the included script or just use `make build` / `make test`.

- If you cloned without `--recursive`, run `git submodule update --init --recursive` before building
- Windows: `powershell -ExecutionPolicy Bypass -File scripts/build-llama.ps1`
- macOS/Linux: `bash ./scripts/build-llama.sh`
- `make build` and `make test` run the native build automatically
- Download the model with `make model`
- Windows manual bootstrap: `powershell -ExecutionPolicy Bypass -File scripts/download-model.ps1`
- macOS/Linux manual bootstrap: `bash ./scripts/download-model.sh`
- Real inference test is opt-in: set `TH_RUN_LOCAL_INFERENCE_TEST=1` and optionally `TH_INFERENCE_MODEL_PATH` before running `go test ./...`

- Override the model path with `TH_MODEL_PATH`
- If no override is set, `th` auto-detects the managed model path under your user cache directory
- Compatibility fallback: `./model/LFM3-1.7B-Q4_K_M.gguf`

Example:

```powershell
th --config
th "list files modified today"
```

Managed model path defaults:

- macOS: `~/Library/Caches/th/models/LFM3-1.7B-Q4_K_M.gguf`
- Linux: `~/.cache/th/models/LFM3-1.7B-Q4_K_M.gguf` unless `XDG_CACHE_HOME` is set
- Windows: `%LOCALAPPDATA%\th\models\LFM3-1.7B-Q4_K_M.gguf`

## Options

- `--c` - Copy result to clipboard
- `-g`, `--git` - Include git branch and `git status -s` output in the prompt
- `-t`, `--tree` - Include top-level workspace hints in the prompt
- `-log`, `--log` - Print the exact final prompt sent to the LLM
- `--config` - Run setup wizard

## Privacy

The prompt context stays inside the local `llama.cpp` runtime.

- **Your prompt** - The text query you type (e.g., "list all files modified today")
- **OS type** - Used to generate context-aware shell commands
- **Shell type** - Used to generate shell-specific syntax
- **Git info** - Included only when you pass `-g`
- **Workspace hints** - Included only when you pass `-t`

## Development

### Creating a Release

```bash
git tag v1.0.0
git push --tags
```

This will trigger the release workflow which currently publishes Linux amd64, macOS arm64, and Windows amd64 binaries.
