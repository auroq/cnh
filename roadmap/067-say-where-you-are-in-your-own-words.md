# 67. Say where you are in your own words: landmarks, street sides and session names resolve to a place

**Open. Filed 2026-11-09 by Aria (issue #67). Nobody's call: the resolution order is in the issue and the data is seeded unverified like the rest of the floor plan.**

Depends on: 063, 069

Numbered 067 because issue numbers share a sequence with pull requests; the item number equals the issue number.

## What

`Locate` accepts how people describe where they are, not only room names.
A landmark: "where everyone registered", "the sponsor booths", "the community hub", "the coat check".
A street side or entrance: "the 100 S side", "200 South entrance", "West Temple side".
A session: "I just got out of Sign Language Crash Course" resolves through the sessions index to the room that session is in, preferring one that ended in the last half hour.
A room, as before, by exact name first and fuzzy name later.
Places gain `aliases`, `side` and `leads_to` in `data/venue/places.yaml`, and the kinds `landmark` and `entrance`.
When the attendee is outside and the question mentions stairs or mobility, the answer is a step-free entrance and names which side it is on.

## Why

Merge Forward: a blind attendee does not read the room sign; a wheelchair user outside a locked door needs the step-free door on their side; everyone knows "where I registered" and nobody knows "255 EF".
Elasticsearch: resolving a session title to a room is the Find the Right Talk retrieval used the other way round, with `distance_feature` on the end time for recency.

## Acceptance

- `Locate` resolves in this order: place id, exact room name, place alias, session title, fuzzy room name, semantic match on the places `content`; `Here` carries `Side` and `Via` (which rule matched).
- Unit tests in `pkg/ask` prove the order with a fake store; an integration test resolves a session title against a real index.
- Acceptance rows: "where I registered" resolves to Salt Palace; "the 100 S side" with "without stairs" answers the step-free entrance with its side named; "I just got out of Sign Language Crash Course" resolves to Grand Ballroom J, Level 1.
- The page's location help text says "a room, a landmark, or the talk you just left".

## Touches

`pkg/ask` (`Locate`, the outside rule, ranking and summary), `pkg/esindex` (alias, session, fuzzy and semantic lookups; mapping), `pkg/model` and `pkg/api/v1` (`Side`, `Via`, `Aliases`, `LeadsTo`), `pkg/venue`, `data/venue/places.yaml`, `pkg/web`, `testing/`.
