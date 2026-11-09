# Installation Locations

This document explains where files are installed when running the installer scripts.

Installers download the CLI binary. The CLI downloads LiquidAI's LFM3-1.7B Q4_K_M (4-bit) GGUF automatically on first use.

The local LFM3 runtime uses an in-process `llama.cpp` v0.8.0 bridge rather than MLX or a local HTTP server process.

Current prebuilt release assets are published for Linux amd64, macOS arm64, and Windows amd64. Other platform or architecture combinations should use a source build.

## Linux amd64 / macOS arm64 (`scripts/install.sh`)

| Item | Location |
|------|----------|
| Binary | `~/.local/bin/th` |
| Model (Linux) | `~/.cache/th/models/LFM3-1.7B-Q4_K_M.gguf` |
| Model (macOS) | `~/Library/Caches/th/models/LFM3-1.7B-Q4_K_M.gguf` |
| Custom directory | User-specified via `-d` or `--dir` option |

### Examples

```bash
# Default installation
./scripts/install.sh
# Installs binary to: ~/.local/bin/th

# Custom directory
./scripts/install.sh -d /usr/local/bin
# Installs to: /usr/local/bin/th
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
| Model | `%LOCALAPPDATA%\th\models\LFM3-1.7B-Q4_K_M.gguf` |
| Custom directory | User-specified via `-InstallDir` parameter |
| PATH | Installation directory is added to user PATH |

### Examples

```powershell
# Default installation
.\scripts\install.ps1
# Installs binary to: C:\Users\<username>\AppData\Local\th\bin\th.exe

# Custom directory
.\scripts\install.ps1 -InstallDir "C:\Program Files\th"
# Installs to: C:\Program Files\th\th.exe
```

### PATH

The installer automatically adds the installation directory to your user PATH. You may need to restart your terminal for changes to take effect.

---

## Local Model Discovery

`th` uses the LFM3 GGUF in the managed user-cache path, or the legacy `~/.cache/th/models/LFM3-1.7B-Q4_K_M.gguf` location from older installs. If neither exists, `th` downloads it to the managed path on first use.

For source builds, the direct bridge uses the checked-out `third_party/llama.cpp` tree and the generated static libraries under its platform build directory.

- If you cloned without `--recursive`, run `git submodule update --init --recursive` before building
- Windows: build with `powershell -ExecutionPolicy Bypass -File scripts/build-llama.ps1`
- macOS/Linux: build with `bash ./scripts/build-llama.sh`
- `make build` and `make test` run the native build automatically
- The real inference test is opt-in via `TH_RUN_LOCAL_INFERENCE_TEST=1`; set `TH_INFERENCE_MODEL_PATH` if you want it to use a non-default model path

## Uninstallation

### Linux / macOS

```bash
rm ~/.local/bin/th
rm ~/.cache/th/models/LFM3-1.7B-Q4_K_M.gguf
rm ~/Library/Caches/th/models/LFM3-1.7B-Q4_K_M.gguf
# Remove from PATH manually if added
```

### Windows

```powershell
# Remove the binary
Remove-Item "$env:LOCALAPPDATA\th\bin\th.exe" -Recurse -Force

# Remove the model
Remove-Item "$env:LOCALAPPDATA\th\models\LFM3-1.7B-Q4_K_M.gguf" -Force

# Remove from PATH (manual)
# Open System Properties → Environment Variables → User variables → Path
# Remove the entry: %LOCALAPPDATA%\th\bin
```
