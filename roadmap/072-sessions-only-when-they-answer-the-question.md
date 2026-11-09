# 72. Sessions appear only when they answer the question, and say when and how far they are

**Done 2026-11-09 (PR #99, issue #72). Filed by Aria, p2. The time window is applied in the use case after the relevance search, and the distance line carries tier words only until item 075 decides the wording.**

Depends on: 063, 075

Numbered 072 because issue numbers share a sequence with pull requests; the item number equals the issue number.
The full text, discussion and evidence are on [issue #72](https://github.com/auroq/cnh/issues/72); this file is the summary the roadmap keeps.

## What

Sessions are listed only when the question is about sessions or names a session feature; the heading says "Sessions matching your question"; each row states its time relative to now ("Starts in 25 minutes", "Happening now, ends 2:50 pm", "Ended at 2:00 pm") and its distance from the attendee in the same words as places.
`/api/ask` adds `sessions_shown_because`, and per session `starts_in_minutes`, `state` and `distance`.

## Why

Merge Forward: a person asking for a quiet room must not be handed five unrelated talks; a Deaf attendee asking about sign language needs to know the crash course is in 25 minutes, one level up.
Elasticsearch: the schedule retrieval used deliberately, time-filtered and explained.

## Acceptance

- A quiet-room question lists no sessions; a sign-language question lists the crash course with its time and distance.
- The new JSON fields appear in `pkg/api/v1` as a `v2` or as additions, never as a change to existing field names.

## Touches

`pkg/ask`, `pkg/web`, `pkg/api/v1`, `testing/acceptance`.
