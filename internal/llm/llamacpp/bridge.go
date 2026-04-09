package llamacpp

/*
#cgo CPPFLAGS: -I${SRCDIR}/../../../third_party/llama.cpp/include -I${SRCDIR}/../../../third_party/llama.cpp/ggml/include
#cgo CXXFLAGS: -std=c++17
#cgo windows LDFLAGS: -L${SRCDIR}/../../../third_party/llama.cpp/build-mingw/src -L${SRCDIR}/../../../third_party/llama.cpp/build-mingw/ggml/src -lllama -l:ggml.a -l:ggml-cpu.a -l:ggml-base.a -lstdc++ -lws2_32 -lpsapi -lwinmm -lbcrypt
#cgo linux LDFLAGS: -L${SRCDIR}/../../../third_party/llama.cpp/build/src -L${SRCDIR}/../../../third_party/llama.cpp/build/ggml/src -lllama -l:libggml.a -l:libggml-cpu.a -l:libggml-base.a -lstdc++ -lm -ldl -pthread
#cgo darwin LDFLAGS: -L${SRCDIR}/../../../third_party/llama.cpp/build/src -L${SRCDIR}/../../../third_party/llama.cpp/build/ggml/src -lllama -lggml -lggml-cpu -lggml-base -lc++
#include "bridge.h"
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

type Model struct {
	handle unsafe.Pointer
}

func New(modelPath string, contextSize, threads, gpuLayers int) (*Model, error) {
	path := C.CString(modelPath)
	defer C.free(unsafe.Pointer(path))

	errBuf := make([]C.char, 1024)
	handle := C.th_llama_model_load(path, C.int(contextSize), C.int(threads), C.int(gpuLayers), &errBuf[0], C.int(len(errBuf)))
	if handle == nil {
		return nil, errors.New(C.GoString(&errBuf[0]))
	}

	return &Model{handle: handle}, nil
}

func (m *Model) Predict(prompt string, maxTokens int) (string, error) {
	if m == nil || m.handle == nil {
		return "", fmt.Errorf("llama.cpp model is not initialized")
	}

	promptText := C.CString(prompt)
	defer C.free(unsafe.Pointer(promptText))

	outBuf := make([]C.char, maxInt(maxTokens*64, 8192))
	errBuf := make([]C.char, 1024)

	result := C.th_llama_model_predict(m.handle, promptText, C.int(maxTokens), &outBuf[0], C.int(len(outBuf)), &errBuf[0], C.int(len(errBuf)))
	if result != 0 {
		return "", errors.New(C.GoString(&errBuf[0]))
	}

	return C.GoString(&outBuf[0]), nil
}

func (m *Model) Close() {
	if m == nil || m.handle == nil {
		return
	}

	C.th_llama_model_free(m.handle)
	m.handle = nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
