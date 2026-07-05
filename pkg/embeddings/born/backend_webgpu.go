//go:build windows || linux

package born

// cspell:words webgpu

import (
	"github.com/born-ml/born/backend/cpu"
	"github.com/born-ml/born/backend/webgpu"
	"github.com/born-ml/born/tensor"
)

// newBackend selects the compute backend following Born's
// documented fallback idiom: WebGPU when a compute-capable adapter
// is present, the pure-Go CPU backend otherwise.
func newBackend() (tensor.Backend, error) {
	if !webgpu.IsAvailable() {
		return cpu.New(), nil
	}

	gpu, err := webgpu.New()
	if err != nil {
		return nil, err
	}
	return gpu, nil
}
