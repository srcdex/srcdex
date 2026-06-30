# AGENTS.md

This file provides guidance to AI agents when working with code in this
repository. For developers and general project information, please refer to
[README.md](README.md) first.

## Related Documentation

- [README.md](README.md) — Project overview and entry point.
- [DESIGN.md](DESIGN.md) — Authoritative architecture reference, composition
  model, and canonical glossary; all other docs follow its taxonomy.
- [darvaza.org/core BUILDING.md][building] — Shared build-system reference
  for all darvaza.org projects (make targets, tooling, linting, CI,
  pre-commit workflow, troubleshooting). srcdex vendors this build system.
- [darvaza.org/core TESTING.md][testing] — Testing patterns and conventions
  (`TestCase`, factories, `RunTestCases`).

[building]: https://github.com/darvaza-proxy/core/blob/main/BUILDING.md
[testing]: https://github.com/darvaza-proxy/core/blob/main/TESTING.md

## Repository Overview

**srcdex** is a pure-Go multi-repository code intelligence engine with no
runtime dependencies. It indexes source workspaces for hybrid keyword and
semantic search, exposed to humans through an embedded web UI and to AI
agents through the Model Context Protocol (MCP).

The canonical import path is `srcdex.dev` (a vanity import); the repository
is hosted at `github.com/srcdex/srcdex`.

The project is in its early stages: little or no implementation exists yet,
and the architecture is still settling. Treat design decisions as direction,
and avoid inventing details that have not been decided.

## Code Architecture

The architecture, composition model, and canonical vocabulary live in
[DESIGN.md](DESIGN.md) — its *Architecture*, *Composition and Ownership*, and
*Glossary* sections are the authority. The principles below are the invariants
to preserve when working in the code; see DESIGN.md for how the parts fit
together.

### Key Design Principles

- **Pure Go, single binary.** No CGo, no external database daemons; the
  engine compiles to a single binary with no runtime dependencies. The
  default libraries — `gotreesitter`, `bleve`, `coder/hnsw`, `vek`, and Born —
  are chosen for this discipline.
- **One binary, one server.** The web UI and the MCP server are both served by
  `srcdex serve`; there is no separate daemon.

## Prerequisites

- Go 1.24 or later (the module's minimum; see [go.mod](go.mod)).
- `make` available.

See [BUILDING.md → Required Tools][building] for the full toolchain.

## Build & Tooling Reference

srcdex vendors the shared darvaza.org build system. Most-used commands:

```bash
make tidy         # format, lint, spell-check, validate
make test         # run tests (no cache reuse)
make coverage     # tests with coverage
make all          # full build cycle (get, generate, tidy, build)
```

The full reference lives in [BUILDING.md][building]. Key sections:

- Build Targets — primary and per-module targets.
- Code Quality Standards — revive limits (function length, complexity,
  argument counts).
- Documentation Standards — markdownlint, CSpell, and LanguageTool
  conventions (80-character prose lines, British English).
- Pre-commit Checklist.

Run `make tidy` until it passes before committing.

## Testing

Tests follow the conventions in [darvaza.org/core's TESTING.md][testing]:

- Table-driven suites use `core.TestCase` plus `core.RunTestCases` when two
  or more rows share one assertion path; otherwise write plain test
  functions.
- Factory functions decouple semantic argument order from struct field
  order.
- Every table-driven test type carries a `var _ core.TestCase = ...`
  compile-time assertion.

## Important Notes

- Staying pure-Go with no runtime dependencies is load-bearing: prefer the
  standard library and pure-Go packages over CGo or external daemons.
- British English throughout prose, comments, and commit messages.
- The design is still settling. When a detail is undecided, leave it open
  rather than divining it.
