// Package openai implements the [embeddings.Engine] against an
// OpenAI-compatible HTTP service, covering both a locally-run model
// server such as Ollama and hosted endpoints.
package openai

// cspell:words resp Wrapf

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/discard"

	"srcdex.dev/pkg/embeddings"
)

// Paths appended to the base URL; the base URL itself must not
// carry the /v1 suffix.
const (
	embeddingsPath = "/v1/embeddings"
	modelsPath     = "/v1/models"
)

// maxDiagnosisBody caps how much of a failure response is read for
// diagnostics.
const maxDiagnosisBody = 64 << 10

// Config describes an OpenAI-compatible embeddings service.
type Config struct {
	// Logger receives the client's diagnostics; discarded when
	// nil.
	Logger slog.Logger

	// Client overrides the HTTP client when set.
	Client *http.Client

	// BaseURL points at the service root, without the /v1 suffix.
	BaseURL string
	// APIKey is sent as a bearer token when non-empty.
	APIKey string
}

// New creates a [Client] speaking to the OpenAI-compatible service
// the [Config] describes.
func New(cfg Config) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, core.Wrap(core.ErrInvalid, "BaseURL")
	}

	if cfg.Logger == nil {
		cfg.Logger = discard.New()
	}
	if cfg.Client == nil {
		cfg.Client = http.DefaultClient
	}
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")

	return &Client{cfg: cfg}, nil
}

// Client is the [embeddings.Engine] over an OpenAI-compatible HTTP
// service.
type Client struct {
	cfg Config
}

var _ embeddings.Engine = (*Client)(nil)

// Models implements [embeddings.Engine] by asking the service which
// models it serves.
func (c *Client) Models(ctx context.Context) ([]embeddings.Model, error) {
	resp, err := c.do(ctx, http.MethodGet, modelsPath, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	return c.decodeModels(resp)
}

// New implements [embeddings.Engine] by binding the named model.
// Binding is offline — an unknown model surfaces on the first
// Embed. The service declares no default model, so an empty name
// is rejected with [embeddings.ErrNoModel].
func (c *Client) New(model string) (embeddings.Embedder, error) {
	if model == "" {
		return nil, embeddings.ErrNoModel
	}

	return &embedder{client: c, model: model}, nil
}

// Model is a model the service reported serving, linked to the
// [Client] that listed it.
type Model struct {
	client *Client

	// ID identifies the model on the service.
	ID string `json:"id"`
}

var _ embeddings.Model = Model{}

// Name implements [embeddings.Model].
func (m Model) Name() string {
	return m.ID
}

// New implements [embeddings.Model] by binding this model on the
// [Client] that listed it.
func (m Model) New() (embeddings.Embedder, error) {
	if m.client == nil {
		return nil, core.Wrap(core.ErrInvalid, "model has no client")
	}

	return m.client.New(m.ID)
}

type modelsResponse struct {
	Data []Model `json:"data"`
}

func (c *Client) decodeModels(resp *http.Response) ([]embeddings.Model, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, c.serviceFailed(resp)
	}

	var out modelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, core.Wrap(embeddings.ErrBadResponse, err.Error())
	}

	models := core.SliceMap(out.Data,
		func(_ []embeddings.Model, m Model) []embeddings.Model {
			m.client = c
			return []embeddings.Model{m}
		})
	return models, nil
}

// do sends one request to the service, attaching the JSON and
// bearer headers as appropriate.
func (c *Client) do(ctx context.Context, method, path string,
	body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method,
		c.cfg.BaseURL+path, body)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	}

	return c.cfg.Client.Do(req)
}

// serviceFailed logs the service's own diagnosis and reports the
// failure status.
func (c *Client) serviceFailed(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxDiagnosisBody))
	if len(body) > 0 {
		c.cfg.Logger.Error().
			WithField("status", resp.StatusCode).
			Print(string(body))
	}

	return core.Wrap(embeddings.ErrServiceFailed, resp.Status)
}

// embedder is a [Client]'s view bound to one model.
type embedder struct {
	client *Client
	model  string
}

type embedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embedResponse struct {
	Data []embedVector `json:"data"`
}

type embedVector struct {
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

// Embed implements [embeddings.Embedder] by posting the texts as
// one batch.
func (e *embedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, embeddings.ErrInvalidInput
	}

	body, err := json.Marshal(embedRequest{Model: e.model, Input: texts})
	if err != nil {
		return nil, err
	}

	resp, err := e.client.do(ctx, http.MethodPost, embeddingsPath,
		bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	return e.decode(resp, len(texts))
}

func (e *embedder) decode(resp *http.Response, count int) ([][]float32, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, e.client.serviceFailed(resp)
	}

	var out embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, core.Wrap(embeddings.ErrBadResponse, err.Error())
	}

	if len(out.Data) != count {
		return nil, core.Wrapf(embeddings.ErrBadResponse,
			"expected %v vectors, got %v", count, len(out.Data))
	}

	vectors := make([][]float32, count)
	for _, d := range out.Data {
		if d.Index < 0 || d.Index >= count || vectors[d.Index] != nil {
			return nil, core.Wrapf(embeddings.ErrBadResponse,
				"invalid vector index %v", d.Index)
		}
		vectors[d.Index] = d.Embedding
	}

	return vectors, nil
}
