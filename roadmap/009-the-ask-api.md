# 9. The ask API

**Open. Filed 2026-11-09, from the kickoff plan. Nobody's call: the shape follows from items 007 and 008.**

Depends on: 007, 008

## What

An HTTP API with two kinds of question:

- `GET /nearest?feature=<type>&from=<room>`: the closest features of that type, ordered by level then adjacency, each with its `verified` flag and a one-sentence direction.
- `GET /ask?q=<text>&from=<room>`: free text such as "quiet room near the keynote" or "captioned talk on eBPF this afternoon", answered with a hybrid query, BM25 and `semantic_text` combined with an RRF retriever, across sessions and features.

Answers are plain text first and JSON second, so a screen reader gets a sentence and not a table.
It deploys as a Rollout per item 005.

## Why

Elasticsearch "Find the Right Talk" judges the hybrid query; Merge Forward judges whether the answer helps someone who cannot see the map.

## Acceptance

- `nearest?feature=quiet-room&from=255` returns 260A or 260B first.
- `ask?q=` with a paraphrase of a session abstract returns that session in the top three.
- Every response that includes an unverified feature says so in its text.

## Touches

`api/`, `deploy/api/`.
