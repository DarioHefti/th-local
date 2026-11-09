# th - Terminal Help

Turn plain English into a shell command, using LiquidAI's LFM3-1.7B running locally. Your prompts stay on your machine.

## Install

macOS arm64 / Linux amd64:

```bash
curl -sSL https://raw.githubusercontent.com/DarioHefti/th-local/refs/heads/main/scripts/install.sh | bash
```

Windows amd64:

```powershell
irm https://raw.githubusercontent.com/DarioHefti/th-local/refs/heads/main/scripts/install.ps1 | iex
```

There is nothing to configure. The first run downloads the model (about 1.1 GB) into your user cache.

## Usage

```bash
th "list all files modified today"
th -g "show me what to commit"        # include git branch and status
th -gt "how do i run the tests"       # also include workspace hints
th "find large files over 100MB" --c  # copy the command to the clipboard
th -log "create a new directory"      # print the exact prompt
```

The model's reasoning streams to stderr while you wait, and stdout gets only the command. `th` never runs anything, so review suggestions before using them.

macOS runs on the GPU through Metal. Linux and Windows run on the CPU.

## Build from source

Requires Go, CMake, and a C++ compiler.

```bash
git clone --recursive https://github.com/DarioHefti/th-local
cd th-local
make build
```

On Windows, with MinGW-w64 on `PATH`:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/build-llama.ps1
go build -o th.exe .
```

## Development

Screen command quality on macOS (commands are parsed, never executed):

```bash
TH_RUN_PROMPT_EVAL=1 go test -v ./internal/llm -run TestShellPromptEvaluation -count=1
```

It currently passes 25 of 30 checks, and reports failure unless all pass.

Publish a release by pushing a version tag:

```bash
git tag v1.2.3 && git push --tags
```

See [INSTALLATION_LOCATIONS.md](INSTALLATION_LOCATIONS.md) for file locations and uninstalling.
