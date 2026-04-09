#!/bin/bash
set -e

MODEL_NAME="google_gemma-4-E2B-it-IQ2_M.gguf"
MODEL_URL="https://huggingface.co/bartowski/google_gemma-4-E2B-it-GGUF/resolve/main/google_gemma-4-E2B-it-IQ2_M.gguf?download=true"
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
Usage: download-model.sh [OPTIONS]

Download the managed Gemma GGUF for th.

OPTIONS:
	-d, --dir DIR       Model directory (default: $default_model_dir_value)
	-f, --force         Force re-download
	-h, --help          Show this help message
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
	if ! command -v curl &> /dev/null; then
		log_error "Missing required dependency: curl"
		exit 1
	fi
}

ensure_directory() {
	if [[ ! -d "$MODEL_DIR" ]]; then
		mkdir -p "$MODEL_DIR" || {
			log_error "Failed to create model directory: $MODEL_DIR"
			exit 1
		}
	fi

	if [[ ! -w "$MODEL_DIR" ]]; then
		log_error "Model directory is not writable: $MODEL_DIR"
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

	ensure_directory

	log_info "Migrating model from $legacy_path to $final_path"
	mv "$legacy_path" "$final_path" || {
		log_error "Failed to migrate existing model to $final_path"
		exit 1
	}
}

download_model() {
	local final_path="$MODEL_DIR/$MODEL_NAME"
	local partial_path="${final_path}.part"
	local attempt=1

	if [[ -f "$final_path" ]] && [[ "$FORCE" != "true" ]]; then
		log_info "Model already present at $final_path"
		return 0
	fi

	rm -f "$partial_path"

	while [[ $attempt -le $MAX_RETRIES ]]; do
		log_info "Downloading $MODEL_NAME (attempt $attempt of $MAX_RETRIES)..."

		if curl -sSL --fail --connect-timeout 30 --retry 2 "$MODEL_URL" -o "$partial_path" 2>/dev/null; then
			if [[ -s "$partial_path" ]]; then
				mv "$partial_path" "$final_path"
				log_info "Model ready at $final_path"
				return 0
			fi

			log_error "Downloaded model is empty"
		fi

		rm -f "$partial_path"

		if [[ $attempt -lt $MAX_RETRIES ]]; then
			log_info "Retrying in ${RETRY_DELAY}s..."
			sleep "$RETRY_DELAY"
		fi

		((attempt++))
	done

	log_error "Failed to download model after $MAX_RETRIES attempts"
	exit 1
}

main() {
	parse_args "$@"
	check_dependencies
	if [[ -z "$MODEL_DIR" ]]; then
		MODEL_DIR="$(default_model_dir)"
	fi
	migrate_legacy_model_if_needed "$MODEL_DIR" "$(legacy_model_dir)"
	ensure_directory
	download_model

	echo ""
	echo "✓ Model available at $MODEL_DIR/$MODEL_NAME"
}

main "$@"