# 0011. Types live by kind: domain model, adapter shapes, and a versioned API contract

Status: Accepted, 2026-11-09. Aria's call, roadmap item 057.

## Context

Until this decision `Place` existed twice (`venue.Place` with YAML tags, `ask.Place` with JSON tags), `Session` existed twice (`schedule.Session`, `ask.Session`), and the JSON that `/api/ask` returns was whatever `json.Marshal` made of the use case's own types.
Rule 6 of the hackathon asks the participant to explain where a type lives and why; "it depends which copy" is not an answer, and a field rename inside the use case would have changed the public API by accident.

## Decision

Three homes, each with one sentence of justification, enforced by `tools/archcheck.sh`:

- **`pkg/model`**: what the business means. `Place`, `Room`, `Session`, `Here`, `Answer`. Imports nothing in this module. Every port (`ask.Store`, `ingest.Source`, `ingest.Places`, `ingest.Indexer`, `web.Asker`) speaks these types. `pkg/ask`, `pkg/venue` and `pkg/schedule` keep their old names as aliases of the model types, so call sites read as before and there is one definition.
- **Adapter shapes**: how this adapter stores or reads it. The YAML file shape stays private to `pkg/venue` (`placeYAML`), the Elasticsearch documents and the `semantic_text` content stay private to `pkg/esindex`, the markdown rows to `pkg/sessionize`. Each adapter maps to and from the model at its boundary.
- **`pkg/api/v1`**: what clients were promised. `AskResponse`, `Here`, `Place`, `Session`, `Error`, field names frozen. Imports nothing in this module, and only `pkg/web` imports it; `pkg/web` maps `model.Answer` onto it. A wire change is a `v2`, not an edit.

## Consequences

- `/api/ask` is byte-for-byte what it was: `pkg/web/contract_test.go` compares the encoding of a full and an empty answer with a response recorded before the move, and the acceptance tests of item 054 read the same fields from the live site.
- `pkg/model` and `pkg/api/v1` may import nothing from this module; `pkg/api/v1` may be imported only by `pkg/web`; adapters import the model and never a use case's internals.
- The type aliases are a door left open: when a use case needs a field the API must not show, the model grows and `toV1` stays still.
- `docs/architecture.md` gains "Where a type lives", so the three answers are on the page Aria presents.
