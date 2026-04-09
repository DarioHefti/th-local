.PHONY: native model build build-all clean test install

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/DarioHefti/th-local/cmd.Version=$(VERSION)"

ifeq ($(OS),Windows_NT)
NATIVE_BUILD_COMMAND = powershell -ExecutionPolicy Bypass -File scripts/build-llama.ps1
MODEL_DOWNLOAD_COMMAND = powershell -ExecutionPolicy Bypass -File scripts/download-model.ps1
else
NATIVE_BUILD_COMMAND = bash ./scripts/build-llama.sh
MODEL_DOWNLOAD_COMMAND = bash ./scripts/download-model.sh
endif

native:
	$(NATIVE_BUILD_COMMAND)

model:
	$(MODEL_DOWNLOAD_COMMAND)

build: native
	go build $(LDFLAGS) -o th .

build-all: clean
	@echo "build-all is no longer supported as a local cross-compile target because th now depends on native llama.cpp libraries. Build on each target OS or use CI release jobs."
	@exit 1

clean:
	rm -rf dist/
	rm -f th

test: native
	go test -v ./...

install: build
	cp th /usr/local/bin/th

dev:
	go run .
