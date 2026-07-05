// Package born implements the in-process embeddings engine on the
// Born deep-learning framework, embedding without a model server.
package born

// cspell:words webgpu

import (
	"context"

	"github.com/born-ml/born/tensor"

	"darvaza.org/core"

	"srcdex.dev/pkg/embeddings"
)

// Engine is the in-process embeddings engine. It owns the Born
// compute backend it selected for its lifetime; Close releases it.
type Engine struct {
	backend tensor.Backend
}

var _ embeddings.Engine = (*Engine)(nil)

// New creates the engine, selecting the compute backend for this
// host: WebGPU when a compute-capable adapter is present, the
// pure-Go CPU backend otherwise.
func New() (*Engine, error) {
	backend, err := newBackend()
	if err != nil {
		return nil, err
	}

	return &Engine{backend: backend}, nil
}

// Name implements [embeddings.Engine], describing the engine by
// the compute backend it selected, e.g. "Born:CPU".
func (eng *Engine) Name() string {
	return "Born:" + eng.BackendName()
}

// BackendName reports the name of the compute backend the engine
// selected, e.g. "CPU" or "WebGPU (<adapter>)".
func (eng *Engine) BackendName() string {
	if eng == nil || eng.backend == nil {
		return ""
	}

	return eng.backend.Name()
}

// Models implements [embeddings.Engine]. Listing the models Born
// can serve is not implemented yet; it reports [core.ErrTODO].
func (*Engine) Models(context.Context) ([]embeddings.Model, error) {
	return nil, core.ErrTODO
}

// New implements [embeddings.Engine]. Binding a model is not
// implemented yet; it reports [core.ErrTODO].
func (*Engine) New(string) (embeddings.Embedder, error) {
	return nil, core.ErrTODO
}

// Close releases the compute backend. The engine is unusable
// afterwards.
func (eng *Engine) Close() error {
	if eng == nil || eng.backend == nil {
		return nil
	}

	if gpu, ok := eng.backend.(interface{ Release() }); ok {
		gpu.Release()
	}
	eng.backend = nil
	return nil
}
