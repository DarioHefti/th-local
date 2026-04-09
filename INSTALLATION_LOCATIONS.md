# Installation Locations

This document explains where files are installed when running the installer scripts.

Installers now download both the CLI binary and the Gemma GGUF during setup.

The local Gemma runtime now uses an in-process `llama.cpp` bridge rather than a local HTTP server process.

Current prebuilt release assets are published for Linux amd64, macOS arm64, and Windows amd64. Other platform or architecture combinations should use a source build.

## Linux amd64 / macOS arm64 (`scripts/install.sh`)

| Item | Location |
|------|----------|
| Binary | `~/.local/bin/th` |
| Model | `~/.cache/th/models/google_gemma-4-E2B-it-IQ2_M.gguf` |
| Custom directory | User-specified via `-d` or `--dir` option |
| Custom model directory | User-specified via `-m` or `--model-dir` option |

### Examples

```bash
# Default installation
./scripts/install.sh
# Installs binary to: ~/.local/bin/th
# Installs model to: ~/.cache/th/models/google_gemma-4-E2B-it-IQ2_M.gguf

# Custom directory
./scripts/install.sh -d /usr/local/bin
# Installs to: /usr/local/bin/th

# Custom model directory
./scripts/install.sh -m "$HOME/.local/share/th/models"
# Installs model to: ~/.local/share/th/models/google_gemma-4-E2B-it-IQ2_M.gguf
```

### PATH

After installation, ensure the installation directory is in your PATH:

```bash
# Bash
echo 'export PATH="$PATH:$HOME/.local/bin"' >> ~/.bashrc
source ~/.bashrc

# Zsh
echo 'export PATH="$PATH:$HOME/.local/bin"' >> ~/.zshrc
source ~/.zshrc
```

---

## Windows (`scripts/install.ps1`)

| Item | Location |
|------|----------|
| Binary | `%LOCALAPPDATA%\th\bin\th.exe` |
| Model | `%LOCALAPPDATA%\th\models\google_gemma-4-E2B-it-IQ2_M.gguf` |
| Custom directory | User-specified via `-InstallDir` parameter |
| Custom model directory | User-specified via `-ModelDir` parameter |
| PATH | Installation directory is added to user PATH |

### Examples

```powershell
# Default installation
.\scripts\install.ps1
# Installs binary to: C:\Users\<username>\AppData\Local\th\bin\th.exe
# Installs model to: C:\Users\<username>\AppData\Local\th\models\google_gemma-4-E2B-it-IQ2_M.gguf

# Custom directory
.\scripts\install.ps1 -InstallDir "C:\Program Files\th"
# Installs to: C:\Program Files\th\th.exe

# Custom model directory
.\scripts\install.ps1 -ModelDir "D:\models\th"
# Installs model to: D:\models\th\google_gemma-4-E2B-it-IQ2_M.gguf
```

### PATH

The installer automatically adds the installation directory to your user PATH. You may need to restart your terminal for changes to take effect.

---

## Local Model Discovery

`th` resolves the Gemma GGUF in this order:

1. `local_model_path` from the saved config
2. `TH_MODEL_PATH`
3. Managed model path in the user cache directory
4. `./model/google_gemma-4-E2B-it-IQ2_M.gguf`
5. `model/google_gemma-4-E2B-it-IQ2_M.gguf` next to the executable

For source builds, the direct bridge uses the checked-out `third_party/llama.cpp` tree and the generated static libraries under its platform build directory.

- If you cloned without `--recursive`, run `git submodule update --init --recursive` before building
- Windows: build with `powershell -ExecutionPolicy Bypass -File scripts/build-llama.ps1`
- macOS/Linux: build with `bash ./scripts/build-llama.sh`
- `make build` and `make test` run the native build automatically
- Download the model with `make model` or the platform-specific `scripts/download-model.*` helper
- The real inference test is opt-in via `TH_RUN_LOCAL_INFERENCE_TEST=1`; set `TH_INFERENCE_MODEL_PATH` if you want it to use a non-default model path

## Uninstallation

### Linux / macOS

```bash
rm ~/.local/bin/th
rm ~/.cache/th/models/google_gemma-4-E2B-it-IQ2_M.gguf
# Remove from PATH manually if added
```

### Windows

```powershell
# Remove the binary
Remove-Item "$env:LOCALAPPDATA\th\bin\th.exe" -Recurse -Force

# Remove the model
Remove-Item "$env:LOCALAPPDATA\th\models\google_gemma-4-E2B-it-IQ2_M.gguf" -Force

# Remove from PATH (manual)
# Open System Properties → Environment Variables → User variables → Path
# Remove the entry: %LOCALAPPDATA%\th\bin
```
