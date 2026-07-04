package openai_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/mock"

	"srcdex.dev/pkg/embeddings"
	"srcdex.dev/pkg/embeddings/openai"
)

// Compile-time verification that test case types implement the
// TestCase interface.
var _ core.TestCase = embedTestCase{}
var _ core.TestCase = modelsTestCase{}
var _ core.TestCase = configTestCase{}
var _ core.TestCase = nameTestCase{}

// stubService serves the canned response, but only on the path the
// client must derive from the base URL.
func stubService(path string, status int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != path {
				http.NotFound(w, r)
				return
			}
			w.WriteHeader(status)
			_, _ = io.WriteString(w, body)
		}))
}

// embedTestCase exercises Embed against a stub HTTP service.
type embedTestCase struct {
	wantIs error
	name   string
	body   string
	want   [][]float32
	texts  []string
	status int
}

func (tc embedTestCase) Name() string {
	return tc.name
}

func (tc embedTestCase) Test(t *testing.T) {
	t.Helper()

	srv := stubService("/v1/embeddings", tc.status, tc.body)
	defer srv.Close()

	emb := mustNewEmbedder(t, srv.URL)
	vectors, err := emb.Embed(context.Background(), tc.texts)

	if tc.wantIs != nil {
		core.AssertErrorIs(t, err, tc.wantIs, "embed")
		return
	}

	core.AssertMustNoError(t, err, "embed")
	tc.assertVectors(t, vectors)
}

func (tc embedTestCase) assertVectors(t *testing.T, vectors [][]float32) {
	t.Helper()

	if !core.AssertEqual(t, len(tc.want), len(vectors), "vector count") {
		return
	}

	for i, want := range tc.want {
		core.AssertSliceEqual(t, want, vectors[i], "vector %v", i)
	}
}

// newEmbedTestCase declares a row where the service answers 200
// with the given body and Embed succeeds.
func newEmbedTestCase(name string, texts []string, body string,
	want [][]float32) embedTestCase {
	return embedTestCase{
		name:   name,
		body:   body,
		want:   want,
		texts:  texts,
		status: http.StatusOK,
	}
}

// newEmbedTestCaseError declares a row where the service answer
// makes Embed fail with the named error.
func newEmbedTestCaseError(name string, texts []string,
	status int, body string, wantIs error) embedTestCase {
	return embedTestCase{
		wantIs: wantIs,
		name:   name,
		body:   body,
		texts:  texts,
		status: status,
	}
}

func embedTestCases() []embedTestCase {
	return []embedTestCase{
		newEmbedTestCase("single text", core.S("hello"),
			`{"data":[{"embedding":[0.1,0.2],"index":0}]}`,
			[][]float32{{0.1, 0.2}}),
		newEmbedTestCase("out-of-order indices", core.S("a", "b"),
			`{"data":[{"embedding":[1],"index":1},{"embedding":[0.5],"index":0}]}`,
			[][]float32{{0.5}, {1}}),
		newEmbedTestCaseError("server error", core.S("boom"),
			http.StatusInternalServerError,
			`{"error":{"message":"boom"}}`,
			embeddings.ErrServiceFailed),
		newEmbedTestCaseError("vector count mismatch", core.S("a", "b"),
			http.StatusOK,
			`{"data":[{"embedding":[0.1],"index":0}]}`,
			embeddings.ErrBadResponse),
		newEmbedTestCaseError("duplicate index", core.S("a", "b"),
			http.StatusOK,
			`{"data":[{"embedding":[0.1],"index":0},{"embedding":[0.2],"index":0}]}`,
			embeddings.ErrBadResponse),
		newEmbedTestCaseError("malformed response", core.S("a"),
			http.StatusOK,
			`{`,
			embeddings.ErrBadResponse),
	}
}

func TestEmbed(t *testing.T) {
	core.RunTestCases(t, embedTestCases())
}

// TestServiceDiagnosis verifies the service's own diagnosis reaches
// the configured logger when a request fails.
func TestServiceDiagnosis(t *testing.T) {
	const diagnosis = `{"error":{"message":"model not found"}}`

	srv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = io.WriteString(w, diagnosis)
		}))
	defer srv.Close()

	log := mock.NewLogger()
	client, err := openai.New(openai.Config{
		Logger:  log,
		BaseURL: srv.URL,
	})
	core.AssertMustNoError(t, err, "New")

	emb, err := client.New("test-model")
	core.AssertMustNoError(t, err, "New")

	_, err = emb.Embed(context.Background(), core.S("hello"))
	core.AssertErrorIs(t, err, embeddings.ErrServiceFailed, "embed")

	msgs := log.GetMessages()
	core.AssertMustEqual(t, 1, len(msgs), "messages")
	core.AssertEqual(t, diagnosis, msgs[0].Message, "diagnosis")
	core.AssertEqual(t, slog.Error, msgs[0].Level, "level")
	status := core.AssertMustTypeIs[int](t, msgs[0].Fields["status"],
		"status field")
	core.AssertEqual(t, http.StatusNotFound, status, "status")
}

func TestEmbedNoTexts(t *testing.T) {
	emb := mustNewEmbedder(t, "http://localhost:1")
	_, err := emb.Embed(context.Background(), nil)
	core.AssertErrorIs(t, err, embeddings.ErrInvalidInput, "embed")
}

// modelsTestCase exercises Models against a stub HTTP service.
type modelsTestCase struct {
	wantIs error
	name   string
	body   string
	want   []string
	status int
}

func (tc modelsTestCase) Name() string {
	return tc.name
}

func (tc modelsTestCase) Test(t *testing.T) {
	t.Helper()

	srv := stubService("/v1/models", tc.status, tc.body)
	defer srv.Close()

	client := mustNewClient(t, srv.URL)
	models, err := client.Models(context.Background())

	if tc.wantIs != nil {
		core.AssertErrorIs(t, err, tc.wantIs, "models")
		return
	}

	core.AssertMustNoError(t, err, "models")
	names := core.SliceMap(models,
		func(_ []string, m embeddings.Model) []string {
			return core.S(m.Name())
		})
	core.AssertSliceEqual(t, tc.want, names, "model names")
}

// newModelsTestCase declares a row where the service answers 200
// with the given body and Models succeeds.
func newModelsTestCase(name, body string, want []string) modelsTestCase {
	return modelsTestCase{
		name:   name,
		body:   body,
		want:   want,
		status: http.StatusOK,
	}
}

// newModelsTestCaseError declares a row where the service answer
// makes Models fail with the named error.
func newModelsTestCaseError(name string, status int, body string,
	wantIs error) modelsTestCase {
	return modelsTestCase{
		wantIs: wantIs,
		name:   name,
		body:   body,
		status: status,
	}
}

func modelsTestCases() []modelsTestCase {
	return []modelsTestCase{
		newModelsTestCase("two models",
			`{"object":"list","data":[{"id":"alpha"},{"id":"beta"}]}`,
			core.S("alpha", "beta")),
		newModelsTestCase("empty list",
			`{"object":"list","data":[]}`,
			nil),
		newModelsTestCaseError("server error",
			http.StatusInternalServerError,
			`{"error":{"message":"boom"}}`,
			embeddings.ErrServiceFailed),
		newModelsTestCaseError("malformed response",
			http.StatusOK,
			`{`,
			embeddings.ErrBadResponse),
	}
}

func TestModels(t *testing.T) {
	core.RunTestCases(t, modelsTestCases())
}

func TestNewNoModel(t *testing.T) {
	client := mustNewClient(t, "http://localhost:1")
	emb, err := client.New("")
	core.AssertErrorIs(t, err, embeddings.ErrNoModel, "New")
	core.AssertNil(t, emb, "embedder")
}

// TestModelNew verifies a listed model binds on the client that
// listed it, while a detached one is rejected.
func TestModelNew(t *testing.T) {
	srv := stubService("/v1/models", http.StatusOK,
		`{"object":"list","data":[{"id":"alpha"}]}`)
	defer srv.Close()

	client := mustNewClient(t, srv.URL)
	models, err := client.Models(context.Background())
	core.AssertMustNoError(t, err, "models")
	core.AssertMustEqual(t, 1, len(models), "model count")

	emb, err := models[0].New()
	core.AssertNoError(t, err, "New")
	core.AssertNotNil(t, emb, "embedder")
}

func TestModelNewDetached(t *testing.T) {
	emb, err := openai.Model{ID: "alpha"}.New()
	core.AssertErrorIs(t, err, core.ErrInvalid, "New")
	core.AssertNil(t, emb, "embedder")
}

// configTestCase exercises the constructor's validation.
type configTestCase struct {
	name    string
	baseURL string
	wantErr bool
}

func (tc configTestCase) Name() string {
	return tc.name
}

func (tc configTestCase) Test(t *testing.T) {
	t.Helper()

	client, err := openai.New(openai.Config{
		BaseURL: tc.baseURL,
	})

	if tc.wantErr {
		core.AssertError(t, err, "New")
		core.AssertNil(t, client, "client")
		return
	}

	core.AssertNoError(t, err, "New")
	core.AssertNotNil(t, client, "client")
}

// newConfigTestCase declares a row where the config is valid and
// the constructor succeeds.
func newConfigTestCase(name, baseURL string) configTestCase {
	return configTestCase{
		name:    name,
		baseURL: baseURL,
	}
}

// newConfigTestCaseError declares a row where the config is
// rejected.
func newConfigTestCaseError(name, baseURL string) configTestCase {
	return configTestCase{
		name:    name,
		baseURL: baseURL,
		wantErr: true,
	}
}

func TestNew(t *testing.T) {
	testCases := []configTestCase{
		newConfigTestCase("valid", "http://localhost:11434"),
		newConfigTestCaseError("missing base URL", ""),
		newConfigTestCaseError("missing scheme", "localhost:11434"),
		newConfigTestCaseError("unsupported scheme", "ftp://example.org"),
		newConfigTestCaseError("malformed", "http://[::1"),
	}

	core.RunTestCases(t, testCases)
}

// nameTestCase exercises Name over the constructor's base-URL
// normalisation.
type nameTestCase struct {
	name    string
	baseURL string
	want    string
}

func (tc nameTestCase) Name() string {
	return tc.name
}

func (tc nameTestCase) Test(t *testing.T) {
	t.Helper()

	client := mustNewClient(t, tc.baseURL)
	core.AssertEqual(t, tc.want, client.Name(), "name")
}

func newNameTestCase(name, baseURL, want string) nameTestCase {
	return nameTestCase{
		name:    name,
		baseURL: baseURL,
		want:    want,
	}
}

func TestName(t *testing.T) {
	testCases := []nameTestCase{
		newNameTestCase("plain", "http://localhost:11434",
			"OpenAI:localhost:11434"),
		newNameTestCase("trailing slash", "http://localhost:11434/",
			"OpenAI:localhost:11434"),
		newNameTestCase("proxy path", "https://gw.example.org/ollama/",
			"OpenAI:gw.example.org/ollama"),
		newNameTestCase("default port", "https://api.openai.com",
			"OpenAI:api.openai.com"),
	}

	core.RunTestCases(t, testCases)
}

func TestClose(t *testing.T) {
	client := mustNewClient(t, "http://localhost:1")
	core.AssertNoError(t, client.Close(), "Close")
}

func mustNewClient(t *testing.T, baseURL string) *openai.Client {
	t.Helper()

	client, err := openai.New(openai.Config{
		BaseURL: baseURL,
	})
	core.AssertMustNoError(t, err, "New")
	core.AssertMustNotNil(t, client, "client")
	return client
}

func mustNewEmbedder(t *testing.T, baseURL string) embeddings.Embedder {
	t.Helper()

	emb, err := mustNewClient(t, baseURL).New("test-model")
	core.AssertMustNoError(t, err, "New")
	core.AssertMustNotNil(t, emb, "embedder")
	return emb
}
