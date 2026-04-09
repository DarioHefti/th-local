param(
    [string]$BuildType = "Release",
    [switch]$Clean
)

$ErrorActionPreference = "Stop"

$RootDir = Split-Path -Parent $PSScriptRoot
$LlamaDir = Join-Path $RootDir "third_party\llama.cpp"
$BuildDir = Join-Path $LlamaDir "build-mingw"

function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Cyan
}

function Write-ErrorMsg {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

function Require-Command {
    param([string]$Name)

    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        Write-ErrorMsg "Missing required command: $Name"
        exit 1
    }
}

Require-Command cmake
Require-Command mingw32-make

if (-not (Test-Path $LlamaDir)) {
    Write-ErrorMsg "llama.cpp checkout not found at $LlamaDir"
    exit 1
}

if ($Clean -and (Test-Path $BuildDir)) {
    Write-Info "Removing existing build directory $BuildDir"
    Remove-Item $BuildDir -Recurse -Force
}

if (-not (Test-Path $BuildDir)) {
    New-Item -ItemType Directory -Path $BuildDir | Out-Null
}

Write-Info "Configuring llama.cpp native libraries in $BuildDir"
cmake -S $LlamaDir -B $BuildDir -G "MinGW Makefiles" `
    "-DCMAKE_BUILD_TYPE=$BuildType" `
    -DBUILD_SHARED_LIBS=OFF `
    -DLLAMA_BUILD_COMMON=OFF `
    -DLLAMA_BUILD_TESTS=OFF `
    -DLLAMA_BUILD_TOOLS=OFF `
    -DLLAMA_BUILD_EXAMPLES=OFF `
    -DLLAMA_BUILD_SERVER=OFF `
    -DGGML_NATIVE=OFF `
    -DGGML_OPENMP=OFF `
    -DGGML_ACCELERATE=OFF `
    -DGGML_BLAS=OFF

Write-Info "Building llama.cpp static library"
cmake --build $BuildDir --config $BuildType --target llama

Write-Info "Native llama.cpp build complete"