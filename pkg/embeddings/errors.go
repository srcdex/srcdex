package embeddings

import (
	"errors"

	"darvaza.org/core"
)

// ErrInvalidInput is reported by [Embedder.Embed] when the input
// texts are invalid, such as an empty slice.
var ErrInvalidInput = core.QuietWrap(core.ErrInvalid, "invalid input")

// ErrNoModel is reported by [Engine.New] when no model is named
// and the engine declares no default.
var ErrNoModel = core.QuietWrap(core.ErrInvalid, "no model")

// ErrServiceFailed is reported when the remote service answers a
// request with a failure status; the engine's logger carries the
// service's own diagnosis.
var ErrServiceFailed = errors.New("service request failed")

// ErrBadResponse is reported when the remote service answers with a
// payload that does not follow the protocol.
var ErrBadResponse = errors.New("invalid response")
