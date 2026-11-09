# 25. The search port speaks domain terms

**Open. Filed 2026-11-09, during Ring 3, by Aria. Nobody's call: the shape is clean architecture applied to what Ring 2 built.**

Depends on: 009, 010

Numbered 025 because issue numbers share a sequence with pull requests and 017 to 024 were taken by PRs; the item number still equals the issue number.

## What

`pkg/ask` is the use case.
It declares the port it consumes, in the app's own words, and knows nothing about Elasticsearch:

```go
type Store interface {
    LocateRoom(ctx context.Context, asked string) (Here, error)
    FindPlaces(ctx context.Context, question string) ([]Place, error)
    FindSessions(ctx context.Context, question string) ([]Session, error)
}
```

`pkg/esindex` is the one adapter: the hybrid query (BM25 and `semantic_text` fused with RRF), the room lookup, hit parsing, both index mappings, the document shapes and the bulk writer.
`pkg/web` is the controller: handlers, templates and CSS, consuming an `Asker` interface it declares.
Each `main.go` only reads configuration and wires the adapter, the use case and the server.
`pkg/venue` stays the model and loader for the curated places file.

## Why

Before this, `ask.Searcher` was `Search(ctx, index, body) json.RawMessage`: the use case wrote query DSL, parsed hit JSON, and the ingest held the mappings.
A test of `Ask` had to re-implement Elasticsearch's response, so there were none.
With the port in domain terms, `Ask` is tested with a ten-line fake, the adapter is tested against a fake HTTP server, and swapping or adding a backend touches one package.
Aria's OpenTelemetry work (item 011) and the ingest fix for failed detail fetches (item 007) both land on top of this shape.

## Acceptance

- `pkg/ask` imports nothing from `pkg/esindex` and contains no query DSL.
- `go test ./pkg/ask` covers `Ask` and `Locate` with a fake `Store`, in the shape `.gitar/review/testing.md` asks for.
- The hand-verified answers still hold: from 255 BC, "nearest all-gender restroom" answers the one outside 255 on the same level; from Grand Ballroom J, "quiet room" answers 260A one level up.

## Touches

`pkg/ask`, `pkg/esindex`, `pkg/web` (new), `cmd/widedoor`, `cmd/ingest`.
