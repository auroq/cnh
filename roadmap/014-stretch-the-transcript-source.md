# 14. Stretch: the transcript source

**Adjudicated 2026-11-09: won't do. Aria's call: the product became venue access shaped by need and verified by attendees; transcript search is a different product.**

"Which sessions have captions or an interpreter" lives on as a session data field (item 072 and the captions-asl chip of item 081), not as a transcript pipeline.

Depends on: 013

This item is part of Option 2, "Find it in the talk", and starts only if Ring 4 is done by 16:00; it has until the 17:00 code freeze.

## What

Choose the source of transcript text for items 015 and 016:

- **Live mic** (the default): a phone or laptop in the room streams audio to an ASR service in the cluster or a hosted one; works today and needs nothing from the organiser.
- **Recordings**: CNCF publishes session recordings on YouTube under CC-BY with ASR captions, but the 2026 talks are not recorded yet, so this means 2025 talks.
- **The organiser's captions**: Wordly captions every room, but no caption text feed is exposed to third parties.

And the ASR engine for live mic: CPU-only open source in a pod, such as faster-whisper, or a hosted streaming API.
The transcript store is the same Elasticsearch, so the choice changes ingest and not the rest of the stack.

## Why

Hearing access is in the Merge Forward focus, and transcripts are the corpus for searching inside talks (Elasticsearch).
The source decides whether the demo uses this week's talks or last year's.

## Acceptance

- The choice and the ASR engine are written here and in a new ADR.

## Touches

`docs/adr/`.
