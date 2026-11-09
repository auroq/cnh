# 15. Stretch: live-mic captions

**Adjudicated 2026-11-09: won't do. Aria's call: the product became venue access shaped by need and verified by attendees; transcript search is a different product.**

"Which sessions have captions or an interpreter" lives on as a session data field (item 072 and the captions-asl chip of item 081), not as a transcript pipeline.

Depends on: 014

This item is part of Option 2 and is a stretch goal.

## What

A page per room showing captions in large text, several lines at once, updated through an ARIA live region set to polite, fed by the live-mic source chosen in item 014.
Each finalised caption segment is also indexed with its room, session id and timestamp.
The audio path runs as its own Rollout with the item 005 settings.

## Why

Deaf and hard-of-hearing attendees ask for captions on the stage screen and not only on a phone; a room-anchored page can go on either (Merge Forward).
The indexed segments are the corpus for item 016.

## Acceptance

- Speech in a test room appears on that room's page within a few seconds.
- The segments appear in Elasticsearch with room, session and timestamp.

## Touches

`captions/`, `deploy/captions/`.
