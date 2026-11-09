# Architecture review

Four layers, and every import points inward.
`tools/archcheck.sh` enforces the import direction in CI; this file is what it enforces and why.
Quote the line that is broken.

## Layers

- **Three homes for a type** (ADR 0011): `pkg/model` is what the business means and imports nothing in the module; an adapter's document, row or file shape stays private to that adapter and is mapped at its boundary; `pkg/api/v1` is what clients were promised, imports nothing, and is imported only by `pkg/web`. A new exported type goes in one of the three, and a reviewer may ask which sentence it answers to.
- **Model**: the types the app talks in. `pkg/venue` (places, rooms, the curated file's shape), `pkg/schedule` (a session), `pkg/config` (every setting, named once, behind per-area interfaces) and the types in `pkg/ask` (`Place`, `Session`, `Here`, `Answer`). Depends on nothing in this module.
- **Use cases**: `pkg/ask` (declares `Store`; `Locate`, `Ask`, ranking, the summary) and `pkg/ingest` (declares `Source`, `Places`, `Indexer`; `Run`). They import only model packages and the standard library.
- **Adapters**: `pkg/esindex` (Elasticsearch: `Store` implements `ask.Store`, `Indexer` implements `ingest.Indexer`), `pkg/sessionize` (implements `ingest.Source`), `venue.File` (implements `ingest.Places`). They implement ports and may import model and use case packages for the port types. They never import `pkg/web`, `pkg/cli`, `pkg/app` or `cmd/*`.
- **Controllers and wiring**: `pkg/web` (HTTP handlers, templates, CSS; imports `pkg/ask` only), `pkg/cli` (the `widedoor` command, `serve`, `ingest`), `pkg/app` (`Main()`), `cmd/widedoor` (`os.Exit(app.Main())`). They import use cases and adapters and only wire. Nothing imports them.

## Rules

- `pkg/ask` never imports `pkg/esindex`, `pkg/sessionize` or `pkg/web`. If `ask` needs something from outside, it declares an interface for it and the adapter implements it.
- Every new outbound dependency (a database, an HTTP API, the clock, a file) gets a port declared by the package that consumes it, named for what it does, not what it is. `Store`, not `ElasticsearchClient`.
- `pkg/web` imports `pkg/ask` only; it declares the `Asker` interface it needs. It does not know what a `Store` is.
- `pkg/cli` and `cmd/*/main.go` are wiring: read configuration, construct adapters, hand them to use cases, start the server or the job. No branching on business data, no query text, no HTML.
- Every setting is a `Key*` constant in `pkg/config` with a default in `Defaults`; a layer takes the `config.Server`, `config.Elasticsearch`, `config.Ingest` or `config.Telemetry` interface it needs, never the whole viper. Secrets are never flags and never logged.
- A package not listed in `tools/archcheck.sh` fails CI until it is placed in a layer there and described in `docs/architecture.md`.
- Tests follow the layers: a use case is tested with a fake port, an adapter against a fake of the real thing (`httptest`), a controller with a fake use case. No test crosses two layers except under `testing/`.

## What to flag

- A query, mapping, URL or SQL string outside an adapter.
- A concrete adapter type in a use case or controller signature where an interface should be.
- Logic in `main.go` beyond construction and error exit.
- A new package without a line in `docs/architecture.md` and `tools/archcheck.sh`.
- An interface declared in the package that implements it instead of the one that uses it.
