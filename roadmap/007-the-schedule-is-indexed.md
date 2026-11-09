# 7. The schedule is indexed

**Open. Filed 2026-11-09, from the kickoff plan. Nobody's call: the source is public.**

Depends on: 006

## What

An ingest job reads the KubeCon NA 2026 schedule from Sessionize's markdown view, `https://sessionize.com/api/v2/uo948i5p/view/markdown`, whose sessions carry Venue, Level and Room columns.
It was fetched on 2026-11-09 and returned 600 sessions and 712 speakers, with one file per session at `/view/markdown/session/{sessionId}`.
It writes one session document per talk: title, abstract, speakers, start, end, venue, level and room, with title and abstract in `semantic_text`.
It runs as a Kubernetes CronJob every 15 minutes, since Sessionize caches for five, and is idempotent by session id.
The CronJob carries the same pod security settings as item 005; replicas and a PDB do not apply to a job.

## Why

Sessions are the "Find the Right Talk" corpus, and their room and level are what join a talk to the venue access data in item 008.

## Acceptance

- After one run the index holds as many sessions as the markdown view lists, 600 on 2026-11-09.
- A second run changes no document count.
- A search for a paraphrase of a known abstract returns that session first.

## Touches

`ingest/`, `deploy/ingest/`.
