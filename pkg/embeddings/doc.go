// Package embeddings turns text into vectors. It defines the
// [Engine] interface every engine implements and the [Embedder],
// the engine's view bound to one model, with two engines behind
// them: the in-process Born framework and an external
// OpenAI-compatible HTTP service.
package embeddings
