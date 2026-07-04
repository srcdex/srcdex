# srcdex

**srcdex** is a lightweight, self-contained, pure-Go multi-repository code
intelligence engine. It indexes source workspaces for hybrid keyword and
semantic search, and serves results to human developers through an embedded
web UI and to AI agents through the Model Context Protocol (MCP).

> **Status:** early and aspirational. The architecture is still settling, and
> implementation has only begun.

## Architecture

srcdex runs as a single binary (`srcdex serve`) that decouples ingestion from
query serving across five subsystems, each behind a Go interface:

- **Ingestion & Orchestration** — organises a workspace into projects and
  files, applying each project's include/exclude rules and tracking changes
  from Git index metadata and Git hooks, with an opt-in `fsnotify` watch mode
  for live re-indexing.
- **Structural Parsing** — extracts structural units via `gotreesitter`, a
  pure-Go Tree-sitter port.
- **Backend** — the content-agnostic index beneath each project. It keeps a
  dual store — a keyword index (`bleve`, BM25) alongside a vector index
  (`coder/hnsw`, approximate nearest-neighbour search) — and persists within
  its store root.
- **Embeddings Engine** — a single `Engine` interface with two
  implementations: the default runs the pure-Go Born framework in-process (an
  embedded model, or a weights file fetched with `srcdex pull`); the second is
  an OpenAI-compatible HTTP client covering OpenAI and Ollama. One engine
  serves many projects across many workspaces; each workspace binds one model
  to get the `Embedder` it hands down.
- **Access Layer** — an embedded web UI and a Model Context Protocol (MCP)
  server, over `stdio` or Streamable HTTP.

Keyword and vector hits are merged with Reciprocal Rank Fusion (RRF). Each
project stores its indices under a per-project `.srcdex/` directory. For the
authoritative architecture reference and glossary, see [DESIGN.md](DESIGN.md).

## Development

srcdex vendors the shared darvaza.org build system. Common commands:

```bash
make tidy         # format, lint, spell-check, validate
make test         # run tests
make coverage     # tests with coverage
```

For the full build-system reference (targets, tooling, linting, CI, and the
pre-commit workflow), see [darvaza.org/core BUILDING.md][building]. For
testing conventions, see [darvaza.org/core TESTING.md][testing]. Agent-specific
guidance lives in [AGENTS.md](AGENTS.md).

[building]: https://github.com/darvaza-proxy/core/blob/main/BUILDING.md
[testing]: https://github.com/darvaza-proxy/core/blob/main/TESTING.md

## Credits

srcdex builds on these open-source projects:

| Project | Role | Licence |
| --- | --- | --- |
| [`gotreesitter`][gotreesitter] | Pure-Go Tree-sitter runtime; structural parsing and language detection | [MIT][gotreesitter-licence] |
| [`bleve`][bleve] | Full-text keyword index (BM25) | [Apache-2.0][bleve-licence] |
| [`coder/hnsw`][hnsw] | HNSW approximate-nearest-neighbour vector index | [CC0-1.0][hnsw-licence] |
| [`vek`][vek] | SIMD vector maths (AVX2 on amd64) with a pure-Go fallback | [MIT][vek-licence] |
| [Born][born] | In-process pure-Go embeddings framework | [Apache-2.0][born-licence] |

Full licence texts will be bundled with srcdex release artefacts.

[gotreesitter]: https://github.com/odvcencio/gotreesitter
[gotreesitter-licence]: https://github.com/odvcencio/gotreesitter/blob/HEAD/LICENSE
[bleve]: https://github.com/blevesearch/bleve
[bleve-licence]: https://github.com/blevesearch/bleve/blob/HEAD/LICENSE
[hnsw]: https://github.com/coder/hnsw
[hnsw-licence]: https://github.com/coder/hnsw/blob/HEAD/LICENSE
[vek]: https://github.com/viterin/vek
[vek-licence]: https://github.com/viterin/vek/blob/HEAD/LICENSE
[born]: https://github.com/born-ml/born
[born-licence]: https://github.com/born-ml/born/blob/HEAD/LICENSE

## Licence

This project is licensed under the MIT Licence. See [LICENCE.txt](LICENCE.txt)
for details.
