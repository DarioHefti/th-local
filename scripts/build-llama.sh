#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LLAMA_DIR="$ROOT_DIR/third_party/llama.cpp"
BUILD_DIR="${BUILD_DIR:-$LLAMA_DIR/build}"
BUILD_TYPE="${BUILD_TYPE:-Release}"
GENERATOR="${CMAKE_GENERATOR:-Unix Makefiles}"

log_info() {
    echo "[INFO] $1"
}

log_error() {
    echo "[ERROR] $1" >&2
}

require_command() {
    if ! command -v "$1" >/dev/null 2>&1; then
        log_error "Missing required command: $1"
        exit 1
    fi
}

require_command cmake

if [[ "$GENERATOR" == "Unix Makefiles" ]]; then
    require_command make
fi

if [[ ! -d "$LLAMA_DIR" ]]; then
    log_error "llama.cpp checkout not found at $LLAMA_DIR"
    exit 1
fi

log_info "Configuring llama.cpp native libraries in $BUILD_DIR"
cmake -S "$LLAMA_DIR" -B "$BUILD_DIR" -G "$GENERATOR" \
    -DCMAKE_BUILD_TYPE="$BUILD_TYPE" \
    -DBUILD_SHARED_LIBS=OFF \
    -DLLAMA_BUILD_COMMON=OFF \
    -DLLAMA_BUILD_TESTS=OFF \
    -DLLAMA_BUILD_TOOLS=OFF \
    -DLLAMA_BUILD_EXAMPLES=OFF \
    -DLLAMA_BUILD_SERVER=OFF \
    -DGGML_NATIVE=OFF \
    -DGGML_OPENMP=OFF \
    -DGGML_ACCELERATE=OFF \
    -DGGML_METAL=OFF \
    -DGGML_BLAS=OFF

log_info "Building llama.cpp static library"
cmake --build "$BUILD_DIR" --config "$BUILD_TYPE" --target llama

log_info "Native llama.cpp build complete"