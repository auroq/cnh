# Architecture review

Four layers, and every import points inward.
`tools/archcheck.sh` enforces the import direction in CI; this file is what it enforces and why.
Quote the line that is broken.

## Layers

- **Model**: the types the app talks in. `pkg/venue` (places, rooms, the curated file's shape) and the types in `pkg/ask` (`Place`, `Session`, `Here`, `Answer`). Depends on nothing in this module.
- **Use case**: `pkg/ask`. Declares the ports it consumes (`Store`), orchestrates (`Locate`, `Ask`), ranks and summarises. Imports only the model and the standard library.
- **Adapters**: `pkg/esindex` (Elasticsearch: `Store` implements `ask.Store`, `Indexer` writes), `pkg/sessionize` (the schedule source), the YAML loader in `pkg/venue`. They implement ports and may import model packages. They never import `pkg/web` or `cmd/*`.
- **Controllers**: `pkg/web` (HTTP handlers, templates, CSS) and `cmd/*` (`main.go`). They import use cases and adapters and only wire. Nothing imports a controller.

## Rules

- `pkg/ask` never imports `pkg/esindex`, `pkg/sessionize` or `pkg/web`. If `ask` needs something from outside, it declares an interface for it and the adapter implements it.
- Every new outbound dependency (a database, an HTTP API, the clock, a file) gets a port declared by the package that consumes it, named for what it does, not what it is. `Store`, not `ElasticsearchClient`.
- `pkg/web` imports `pkg/ask` only; it declares the `Asker` interface it needs. It does not know what a `Store` is.
- `cmd/*/main.go` is wiring: read configuration, construct adapters, hand them to use cases, start the server or the job. No branching on business data, no query text, no HTML.
- A package not listed in `tools/archcheck.sh` fails CI until it is placed in a layer there and described in `docs/architecture.md`.
- Tests follow the layers: a use case is tested with a fake port, an adapter against a fake of the real thing (`httptest`), a controller with a fake use case. No test crosses two layers except under `testing/`.

## What to flag

- A query, mapping, URL or SQL string outside an adapter.
- A concrete adapter type in a use case or controller signature where an interface should be.
- Logic in `main.go` beyond construction and error exit.
- A new package without a line in `docs/architecture.md` and `tools/archcheck.sh`.
- An interface declared in the package that implements it instead of the one that uses it.
