package embeddings

import (
	"context"
)

// Embeddings turns text into vectors. One engine serves many
// projects across many workspaces.
type Embeddings interface {
	// Embed returns one vector per input text, in input order.
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}
