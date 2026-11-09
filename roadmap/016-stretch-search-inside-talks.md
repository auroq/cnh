# 16. Stretch: search inside talks

**Adjudicated 2026-11-09: won't do. Aria's call: the product became venue access shaped by need and verified by attendees; transcript search is a different product.**

"Which sessions have captions or an interpreter" lives on as a session data field (item 072 and the captions-asl chip of item 081), not as a transcript pipeline.

Depends on: 007, 015

This item is part of Option 2 and is a stretch goal.

## What

`/ask` also searches transcript segments with the same hybrid query, and an answer can be "this was said in room 255 at 14:12, in this session", joined to the schedule by session id.

## Why

Searching what was said, not only the abstract, is the stronger "Find the Right Talk" entry.
It also lets someone who missed a talk, or could not follow it live, find the part they need at their own pace (Merge Forward).

## Acceptance

- A phrase spoken in a captioned test session is found by a paraphrase, with the right session and a timestamp within a minute.

## Touches

`api/`.
