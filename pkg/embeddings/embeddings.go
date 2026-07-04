package embeddings

import (
	"context"
)

// Engine turns text into vectors. An engine is a shared service:
// its lifetime sits above workspaces, and one engine serves many
// projects across many workspaces. Consumers depend on this
// interface alone, and implementations are safe for concurrent use.
type Engine interface {
	// Models lists the models the engine can serve.
	Models(ctx context.Context) ([]Model, error)

	// New binds a model by name, yielding the view a workspace
	// hands down to its projects. The empty string binds the
	// engine's default model; an engine without one reports
	// [ErrNoModel]. Binding is offline — an unknown model
	// surfaces on the first [Embedder.Embed] call.
	New(model string) (Embedder, error)
}

// Embedder is an [Engine]'s view bound to one model: what a
// workspace hands its projects and a backend embeds with.
// Implementations are safe for concurrent use.
type Embedder interface {
	// Embed returns one vector per input text, in input order.
	// An empty texts slice is rejected with [ErrInvalidInput].
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}

// Model identifies a model an [Engine] can serve. Each engine
// implements it over its own data, keeping the model linked to the
// engine that listed it.
type Model interface {
	// Name is the name the engine serves the model under and
	// [Engine.New] accepts; it may be a mutable tag, not a
	// version-precise identity.
	Name() string

	// New binds this model on its own engine, equivalent to
	// [Engine.New] with the model's name.
	New() (Embedder, error)
}
