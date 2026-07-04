# Design Reference: srcdex

`srcdex` is a pure-Go, multi-repository code intelligence engine for hybrid
keyword and semantic search over source workspaces. This document is the
conceptual reference for its architecture and vocabulary.

> **Status:** conceptual. This document describes the intended architecture
> and the words `srcdex` uses for its parts, not shipped code. Implementation
> details — package layout, interface shapes, on-disk formats — are
> deliberately out of scope and still settling.

## Architecture

`srcdex` compiles to a single binary with no runtime dependencies. To serve
hybrid keyword and semantic search from one process, it separates ingestion
from query serving across five subsystems, each behind a Go interface:

```text
External ─────────────────────────────────────────────────────────
  ┌──────────────────────┐
  │ Source repositories  │   (Git)
  └──────────┬───────────┘
             │ read on startup + Git hooks
             ▼
Source context ───────────────────────────────────────────────────
  Ingestion & Orchestration — Git change tracking, project rules
             │ feeds selected source
             ▼
  Structural Parsing — language, metadata, structural units
             │ envelope (source side): wrap each unit with
             │   its repository, file, and scope
   · · · · · │ indexable items · · · · · · · · · · · ·  parser seam
             ▼
Index context ────────────────────────────────────────────────────
  Backend (per project) — content-agnostic; persists within its store root
  ┌───────────────────────────┐           ┌──────────────────────────┐
  │ Keyword Index (bleve)     │   embed   │ Embeddings Engine        │
  │ Vector Index (coder/hnsw) │ ◀──────── │ shared · 1:N             │
  └───────────────┬───────────┘           │ Born / OpenAI-compatible │
                  ▲                       └──────────────────────────┘
                  │ query both indices; fuse with Reciprocal Rank Fusion
Serving context ──┼────────────────────────────────────────────────
  ┌───────────────┴──────┐
  │ Access Layer         │   web UI · MCP server
  └──────────────────────┘
```

The embeddings engine has two implementations: it runs either in-process
behind the default [Born][born] framework or against an external
[OpenAI-compatible][openai-api] HTTP service.

## Composition and Ownership

The runtime composes top-down; each layer hands the one below it exactly what
it needs and no more.

- A **workspace** holds the shared settings and references a single
  embeddings engine; the model it embeds with is one of those settings,
  bound against the engine to yield the workspace's embedder. It composes
  its projects — handing each one its source tree, that embedder, and
  its settings. The engine is shared, not owned: its lifetime sits above
  the workspace, so one engine can serve many projects across many
  workspaces.
- A **project** is the per-project composition root and owns the source
  side: its include/exclude rules decide which files are worth indexing, and
  it feeds the chosen source to the parser. It opens or initialises its
  `.srcdex/` store and creates the backend beneath it.
- The **parser** is the seam between the source side and the backend. Given
  a file's contents and an optional filename, it returns the detected
  language, the file's metadata, and the structural units to index. Taking
  contents rather than a path keeps the parser off the filesystem, so the
  same seam serves source read from the working tree or from Git. Files it
  recognises as binary or generated are skipped rather than indexed.
- The **backend** is content-agnostic. It ingests ready-made indexable items
  rather than files, so nothing ties its store to source code. It is
  constructed with two dependencies handed down through the project — a
  narrowed, writeable store root (per-project) and the embedder it embeds
  with, the workspace's model-bound view of the shared engine (one engine
  serves many backends, 1:N) — and it persists only within that store root.

Because the backend is content-agnostic, the source side does the enveloping:
each structural unit is wrapped with its repository, file, and scope metadata
before ingestion, so the backend receives indexable items ready-made and never
reaches back into source concepts to build them. The envelope's exact wire
format remains an implementation detail.

## Glossary

The canonical terms used throughout `srcdex`:

<!-- markdownlint-disable MD033 -->
| Term | Meaning |
| --- | --- |
| <a id="access-layer"></a>Access Layer | The subsystem that serves query results: an embedded web UI for developers and a [Model Context Protocol][mcp] server for AI agents. |
| <a id="backend"></a>Backend | The content-agnostic index subsystem beneath a project; ingests [indexable items][indexable-item] and persists them within its store root. |
| <a id="dual-store"></a>Dual store | The pair a [backend][backend] keeps: a keyword index ([`bleve`][bleve]) and a vector index ([`coder/hnsw`][hnsw]), fused at query time with [RRF][rrf]. |
| <a id="embedder"></a>Embedder | The [Embeddings Engine][embeddings-engine]'s view bound to one model: what a workspace hands its projects and a [backend][backend] embeds with. |
| <a id="embeddings-engine"></a>Embeddings Engine | The subsystem that turns text into vectors; a shared service referenced by a workspace, not owned by it — one engine serves many projects across many workspaces (1:N). It lists the models it can serve and binds one on request, yielding an [Embedder][embedder]. |
| <a id="envelope"></a>Envelope | A [structural unit][structural-unit] wrapped in its metadata header (repository, file, scope) before embedding. |
| <a id="indexable-item"></a>Indexable item | What the [backend][backend] ingests: a ready-made record carrying no dependency on source code. |
| <a id="ingestion-and-orchestration"></a>Ingestion & Orchestration | The subsystem that drives indexing across a workspace: it tracks source change through Git and applies each project's rules to decide what is fed onward for parsing. |
| <a id="parser"></a>Parser | The source/backend seam-role of the structural-parsing subsystem: given a file's contents and an optional filename, it returns the detected language, metadata, and [structural units][structural-unit]. |
| <a id="project"></a>Project | A repository and its index; the per-project composition root that owns the source side and its `.srcdex/` store. |
| <a id="project-rules"></a>Project rules | The rules a project applies when indexing; its include/exclude criteria for selecting files are one part, layered over sensible defaults it can extend or override. |
| <a id="structural-parsing"></a>Structural Parsing | The subsystem that turns source files into structural units; the [parser][parser] is its source/backend seam-role. |
| <a id="structural-unit"></a>Structural unit | A self-contained piece of code the [parser][parser] extracts — a function, a struct, an interface — with its surrounding context. The source-side output. |
| <a id="workspace"></a>Workspace | A set of projects; holds the shared settings, references the single embeddings engine it uses, and exposes a workspace-wide API. |
<!-- markdownlint-enable MD033 -->

## Appendix

### BM25

BM25 (Best Matching 25) scores how well a document matches a keyword query.
It rewards a term that appears often in a document (term frequency) but with
diminishing returns, so a word repeated ten times does not count ten times as
much; it weights rarer terms more heavily (inverse document frequency), since
a match on an uncommon word says more than a match on a common one; and it
normalises for document length, so a long document is not ranked highly
merely for containing more words. srcdex's keyword index — `bleve` — ranks
its hits this way.

### Reciprocal Rank Fusion

Reciprocal Rank Fusion (RRF) merges several ranked result lists into one
without needing their scores to be comparable. Each item is scored by
summing `1 / (rank + k)` across every list it appears in — `rank` is its
position in that list and `k` is a small constant (commonly 60) — and the
merged list is ordered by that total. Because only positions count, never
the raw scores, keyword hits carrying [BM25][bm25] relevance and vector hits
carrying cosine similarity combine cleanly even though their scores sit on
different scales.

For `srcdex`'s two rankings — a keyword search and a vector search — that
resolves to:

$$
\text{RRF Score} = \frac{1}{\text{Keyword Rank} + k}
                  + \frac{1}{\text{Vector Rank} + k}
$$

A hit that surfaces in only one ranking takes just the single term it
earns; the absent rank contributes nothing.

[mcp]: https://modelcontextprotocol.io
[bleve]: https://github.com/blevesearch/bleve
[hnsw]: https://github.com/coder/hnsw
[born]: https://github.com/born-ml/born
[openai-api]: https://github.com/openai/openai-openapi
[bm25]: #bm25
[rrf]: #reciprocal-rank-fusion
[parser]: #parser
[structural-unit]: #structural-unit
[indexable-item]: #indexable-item
[backend]: #backend
[embedder]: #embedder
[embeddings-engine]: #embeddings-engine
