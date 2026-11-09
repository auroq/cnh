# 67. Say where you are in your own words: landmarks, street sides and session names resolve to a place

**Open. Filed 2026-11-09 by Aria, p1; worker issue-67 in progress. Nobody's call: the data decides.**

Depends on: 063, 068

Numbered 067 because issue numbers share a sequence with pull requests; the item number equals the issue number.
The full text, discussion and evidence are on [issue #67](https://github.com/auroq/cnh/issues/67); this file is the summary the roadmap keeps.

## What

`Locate` accepts landmarks ("where I registered", "the sponsor booths"), street sides and entrances ("the 200 South entrance", "I'm outside"), and session titles ("I just got out of the Flux keynote", resolved through the sessions index to its room), as well as room names.
Places gain `aliases`, `side` and `leads_to`, and the kinds `landmark`, `entrance` and `emergency-exit`.
When the attendee is outside, the answer offers entrances, step-free first when the question mentions stairs or mobility, and names the side.

## Why

Merge Forward: a blind attendee does not read the room sign; a wheelchair user outside a locked door needs the step-free door on their side.
Elasticsearch: resolving a session title to a room is the schedule retrieval used the other way round.

## Acceptance

- Each example phrase resolves to the expected place in a unit test with a fake store.
- Acceptance rows for a landmark, a street side and a session title pass against the live site.

## Touches

`pkg/ask`, `pkg/esindex/store.go`, `data/venue/places.yaml`, `testing/acceptance`.
