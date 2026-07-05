//go:build !windows && !linux

package born

// cspell:words webgpu

import (
	"github.com/born-ml/born/backend/cpu"
	"github.com/born-ml/born/tensor"
)

// newBackend selects the compute backend. Born's public webgpu
// backend only builds on Windows and Linux, so every other platform
// selects the pure-Go CPU backend outright.
func newBackend() (tensor.Backend, error) {
	return cpu.New(), nil
}
