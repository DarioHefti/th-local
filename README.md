# th - Terminal Help

Get shell commands from the local Gemma model directly in your terminal.

## Installation

The installer downloads the `th` binary and the Gemma IQ2_M GGUF from Hugging Face, then runs the local model through an in-process `llama.cpp` bridge.

### macOS/Linux

```bash
curl -sSL https://raw.githubusercontent.com/DarioHefti/th-local/refs/heads/main/scripts/install.sh | bash
```

### Windows

```powershell
irm https://raw.githubusercontent.com/DarioHefti/th-local/refs/heads/main/scripts/install.ps1 | iex
```

### From Source

```bash
make build
make model
```

`make build` runs the native `llama.cpp` build first and then builds `th`. `make model` downloads the GGUF into the managed per-user model directory.

### Building on WSL

Use a Linux shell inside WSL, not PowerShell, and install the native toolchain first:

```bash
sudo apt update
sudo apt install -y build-essential cmake git pkg-config

cd /mnt/c/Users/heda/Documents/dev/th
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
2. Press Enter to auto-detect the managed `th` model path or provide a custom path to the local Gemma GGUF.

### Local MVP Requirements

The local runtime uses a direct `llama.cpp` bridge. For source builds, build the native libraries with the included script or just use `make build` / `make test`.

- Windows: `powershell -ExecutionPolicy Bypass -File scripts/build-llama.ps1`
- macOS/Linux: `bash ./scripts/build-llama.sh`
- `make build` and `make test` run the native build automatically
- Download the model with `make model`
- Windows manual bootstrap: `powershell -ExecutionPolicy Bypass -File scripts/download-model.ps1`
- macOS/Linux manual bootstrap: `bash ./scripts/download-model.sh`
- Real inference test is opt-in: set `TH_RUN_LOCAL_INFERENCE_TEST=1` and optionally `TH_INFERENCE_MODEL_PATH` before running `go test ./...`

- Override the model path with `TH_MODEL_PATH`
- If no override is set, `th` auto-detects the managed model path under your user cache directory
- Compatibility fallback: `./model/google_gemma-4-E2B-it-IQ2_M.gguf`

Example:

```powershell
th --config
th "list files modified today"
```

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

This will trigger the release workflow which builds binaries for Linux, macOS, and Windows.
