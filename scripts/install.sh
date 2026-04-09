#!/bin/bash
set -e

REPO="DarioHefti/th-local"
BINARY_NAME="th"
MODEL_NAME="google_gemma-4-E2B-it-IQ2_M.gguf"
MODEL_URL="https://huggingface.co/bartowski/google_gemma-4-E2B-it-GGUF/resolve/main/google_gemma-4-E2B-it-IQ2_M.gguf?download=true"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
MODEL_DIR="${MODEL_DIR:-}"
FORCE=false
MAX_RETRIES=3
RETRY_DELAY=2

default_model_dir() {
    case "$(uname -s)" in
        Darwin*) echo "$HOME/Library/Caches/th/models" ;;
        Linux*) echo "${XDG_CACHE_HOME:-$HOME/.cache}/th/models" ;;
        *) echo "${XDG_CACHE_HOME:-$HOME/.cache}/th/models" ;;
    esac
}

legacy_model_dir() {
    case "$(uname -s)" in
        Darwin*) echo "$HOME/.cache/th/models" ;;
        *) echo "" ;;
    esac
}

usage() {
    local default_model_dir_value
    default_model_dir_value="$(default_model_dir)"

    cat <<EOF
Usage: install.sh [OPTIONS]

Install th (Terminal Help) CLI

OPTIONS:
    -d, --dir DIR       Installation directory (default: ~/.local/bin)
    -m, --model-dir DIR Model directory (default: $default_model_dir_value)
    -f, --force         Force reinstall
    -h, --help          Show this help message

EXAMPLES:
    install.sh
    install.sh -d /usr/local/bin
    install.sh -m "$default_model_dir_value"
    install.sh -f
EOF
}

log_info() {
    echo "[INFO] $1"
}

log_error() {
    echo "[ERROR] $1" >&2
}

parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -d|--dir)
                if [[ -z "$2" || "$2" == -* ]]; then
                    log_error "Option $1 requires a directory argument"
                    usage
                    exit 1
                fi
                INSTALL_DIR="$2"
                shift 2
                ;;
            -m|--model-dir)
                if [[ -z "$2" || "$2" == -* ]]; then
                    log_error "Option $1 requires a directory argument"
                    usage
                    exit 1
                fi
                MODEL_DIR="$2"
                shift 2
                ;;
            -f|--force)
                FORCE=true
                shift
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done
}

check_dependencies() {
    local missing=()
    
    if ! command -v curl &> /dev/null; then
        missing+=("curl")
    fi
    
    if [[ ${#missing[@]} -gt 0 ]]; then
        log_error "Missing required dependency: ${missing[*]}"
        log_error "Please install the missing dependency and try again"
        exit 1
    fi
}

get_os() {
    case "$(uname -s)" in
        Linux*)     echo "linux" ;;
        Darwin*)    echo "darwin" ;;
        *)          log_error "Unsupported operating system: $(uname -s)"; exit 1 ;;
    esac
}

get_arch() {
    case "$(uname -m)" in
        x86_64)     echo "amd64" ;;
        aarch64|arm64)  echo "arm64" ;;
        *)          log_error "Unsupported architecture: $(uname -m)"; exit 1 ;;
    esac
}

ensure_supported_release_target() {
    local os="$1"
    local arch="$2"

    case "$os/$arch" in
        linux/amd64|darwin/arm64)
            return 0
            ;;
        *)
            log_error "No prebuilt release is published for $os/$arch"
            log_error "Use the source build instructions in README.md instead"
            exit 1
            ;;
    esac
}

get_latest_version() {
    local version
    version=$(curl -sSL --fail --connect-timeout 10 "https://api.github.com/repos/$REPO/releases/latest" | grep -o '"tag_name": "v[^"]*"' | cut -d'"' -f4 | sed 's/v//')
    
    if [[ -z "$version" ]]; then
        return 1
    fi
    
    if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        log_error "Invalid version format: $version"
        return 1
    fi
    
    echo "$version"
}

ensure_directory() {
    local path="$1"
    local label="$2"

    if [[ ! -d "$path" ]]; then
        mkdir -p "$path" || {
            log_error "Failed to create ${label} directory: $path"
            exit 1
        }
    fi

    if [[ ! -w "$path" ]]; then
        log_error "${label^} directory is not writable: $path"
        exit 1
    fi
}

migrate_legacy_model_if_needed() {
    local current_model_dir="$1"
    local legacy_dir="$2"

    if [[ -z "$legacy_dir" || "$current_model_dir" == "$legacy_dir" ]]; then
        return
    fi

    local final_path="$current_model_dir/$MODEL_NAME"
    local legacy_path="$legacy_dir/$MODEL_NAME"

    if [[ -e "$final_path" || ! -f "$legacy_path" ]]; then
        return
    fi

    ensure_directory "$current_model_dir" "model"

    log_info "Migrating model from $legacy_path to $final_path"
    mv "$legacy_path" "$final_path" || {
        log_error "Failed to migrate existing model to $final_path"
        exit 1
    }
}

download_file() {
    local url="$1"
    local output_path="$2"
    local description="$3"
    local attempt=1
    local success=false
    
    while [[ $attempt -le $MAX_RETRIES ]]; do
        log_info "Downloading $description (attempt $attempt of $MAX_RETRIES)..."
        
        if curl -sSL --fail --connect-timeout 30 --retry 2 "$url" -o "$output_path" 2>/dev/null; then
            if [[ -s "$output_path" ]]; then
                success=true
                break
            else
                log_error "Downloaded file is empty"
            fi
        fi
        
        log_info "Retrying in ${RETRY_DELAY}s..."
        sleep "$RETRY_DELAY"
        ((attempt++))
    done
    
    if [[ "$success" != "true" ]]; then
        rm -f "$output_path"
        return 1
    fi

    log_info "Download complete"

    return 0
}

download_binary() {
    local os=$1
    local arch=$2
    local version=$3
    local filename="th-${os}-${arch}"
    local url="https://github.com/$REPO/releases/download/v${version}/${filename}"
    local output_path="$INSTALL_DIR/$BINARY_NAME"

    if ! download_file "$url" "$output_path" "$filename"; then
        log_error "Failed to download binary after $MAX_RETRIES attempts"
        exit 1
    fi

    chmod +x "$output_path"
}

ensure_model() {
    local final_path="$MODEL_DIR/$MODEL_NAME"
    local partial_path="${final_path}.part"

    ensure_directory "$MODEL_DIR" "model"

    if [[ -f "$final_path" ]] && [[ "$FORCE" != "true" ]]; then
        log_info "Model already present at $final_path"
        return 0
    fi

    rm -f "$partial_path"

    if ! download_file "$MODEL_URL" "$partial_path" "$MODEL_NAME"; then
        log_error "Failed to download model after $MAX_RETRIES attempts"
        exit 1
    fi

    mv "$partial_path" "$final_path"
    log_info "Model ready at $final_path"
}

verify_binary() {
    local binary_path="$1"
    
    if [[ ! -f "$binary_path" ]]; then
        log_error "Binary not found at $binary_path"
        return 1
    fi
    
    if [[ ! -s "$binary_path" ]]; then
        log_error "Binary is empty"
        return 1
    fi
    
    if [[ "$binary_path" == *.exe ]]; then
        return 0
    fi
    
    if [[ ! -x "$binary_path" ]]; then
        log_error "Binary is not executable"
        return 1
    fi
    
    return 0
}

main() {
    parse_args "$@"
    check_dependencies

    if [[ -z "$MODEL_DIR" ]]; then
        MODEL_DIR="$(default_model_dir)"
    fi

    migrate_legacy_model_if_needed "$MODEL_DIR" "$(legacy_model_dir)"
    
    local binary_path="$INSTALL_DIR/$BINARY_NAME"
    
    if [[ -f "$binary_path" ]] && [[ "$FORCE" != "true" ]]; then
        log_info "$BINARY_NAME is already installed at $binary_path"
        if ! verify_binary "$binary_path"; then
            log_error "Existing binary verification failed; rerun with -f to replace it"
            exit 1
        fi
    else
        ensure_directory "$INSTALL_DIR" "installation"
    
        local os
        local arch
        local version
    
        os=$(get_os)
        arch=$(get_arch)
        ensure_supported_release_target "$os" "$arch"
    
        log_info "Fetching latest version..."
        version=$(get_latest_version) || {
            log_error "Failed to get latest version"
            log_error "Please check your network connection and try again"
            exit 1
        }
    
        log_info "Installing th v$version for $os/$arch..."
    
        download_binary "$os" "$arch" "$version"
    
        if ! verify_binary "$binary_path"; then
            log_error "Binary verification failed"
            rm -f "$binary_path"
            exit 1
        fi
    fi

    ensure_model
    
    echo ""
    echo "✓ Binary available at $binary_path"
    echo "✓ Model available at $MODEL_DIR/$MODEL_NAME"
    echo ""
    echo "Add to PATH if not already added:"
    echo "  echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ~/.bashrc"
    echo "  source ~/.bashrc"
}

main "$@"
