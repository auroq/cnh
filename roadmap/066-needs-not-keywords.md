# 66. Needs, not keywords: what people say maps to what the venue offers

**Done 2026-11-09 (PR #84, issue #66). Filed 2026-11-09 by Aria. Nobody's call: the vocabulary is data in `data/venue/needs.yaml`, and the inference builds on the kind table from item 063.**

Depends on: 063

Numbered 066 because issue numbers share a sequence with pull requests; the item number equals the issue number.

## What

A need vocabulary, data-driven, that turns what an attendee says into the kinds of place Wide Door can offer, with an honest note when the place is a stand-in for what was asked.
`data/venue/needs.yaml` has one entry per need with `id`, `name` (how the answer names it back), `phrases` (what people say), `kinds` (the place kinds to offer, in order) and `note` (said when the match is a stand-in).
Seeded needs: mobility-accessible, private-space, medical-private, menstrual, first-aid, emergency-exit, rest and help.
Inference order: a kind word in the question (`ask.KindOf`, item 063) first, then a need phrase, then a semantic match of the question against the `needs` index in Elasticsearch, where each need's `content` is a `semantic_text` field, so paraphrases still land.
The summary names the need it understood ("You asked for somewhere private.") before the place, and ends with the need's note when there is one.
Unknown phrasing still falls through to today's hybrid search over places.

## Why

Merge Forward, "Conferences for Everyone": people describe situations, not taxonomy.
"I need to fix my makeup" and "where can I inject insulin" are the real questions, and the second one deserves a private, dignified answer rather than a keyword miss.
Elasticsearch track: the semantic fallback over the needs index is `semantic_text` doing what BM25 cannot.

## Acceptance

- `pkg/ask` has a need inference step (`NeedOf`) with a table test per seeded phrase, nested when, and, it should, one assertion per leaf.
- The ingest indexes the needs into a `needs` index with `semantic_text` on `content`; `ask.Store` gains `FindNeeds(ctx, question)` and the use case takes the top semantic match only above a score threshold.
- For each seeded need, an acceptance row: a question in that need's words gets a top place whose kind is in the need's kinds, the summary names the need, and the summary carries the note when the need has one.
- Unknown phrasing still falls through to the hybrid place search.
- `docs/architecture.md`'s where-to-look table gains "to change what a phrase means, look in `data/venue/needs.yaml`".

## Touches

`data/venue/needs.yaml`, `pkg/model` (`Need`), `pkg/venue` (the needs file loader), `pkg/config` (`venue.needs_file`, `elasticsearch.needs_index`), `pkg/ask` (`need.go`, the `Store` port, the summary), `pkg/esindex` (needs mapping, `IndexNeeds`, `FindNeeds`), `pkg/ingest` (the `Needs` port and the index step), `pkg/cli` (wiring), `deploy/app/config.yaml`, `testing/helpers`, `testing/integration/ask`, `testing/acceptance/cases.go`, `docs/architecture.md`.
