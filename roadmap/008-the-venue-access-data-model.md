# 8. The venue access data model

**Open. Filed 2026-11-09, from the kickoff plan. Aria's call: which access attributes are seeded as "unverified" from the floor plan.**

Depends on: 006

## What

Three document types in Elasticsearch:

- **place**: a building and level, such as Salt Palace L2 or Hyatt L5.
- **room**: a named room on a level, such as 260A, with the Sessionize room name it matches.
- **feature**: something an attendee needs, such as an elevator, an all-gender restroom, a quiet room, a step-free entrance or a room with captions, with its place, the rooms it is near, a free-text description in `semantic_text`, and a `verified` flag with its source.

Seeded from what the organiser publishes: quiet rooms 260A (all attendees) and 260B (women and non-binary) on Salt Palace L2, the prayer room in Hyatt L5 Zion, all-gender restrooms on Salt Palace L1 by the Solutions Showcase and L2 outside 255 and on Hyatt L1 to L4, and Merge Forward gatherings in Grand Ballroom J.
Elevator positions are not published (the organiser's page says to ask staff) and the venue's OpenStreetMap record has no `wheelchair` tag, so anything read off the floor-plan PDF is seeded as unverified.
"Nearest" is measured by level first and then by a hand-set adjacency between rooms on a level, not by a geometric route.

## Why

This is the gap the research found: the organiser knows where things are and publishes it as static text with no level, route or access data.
Limited mobility, hearing and vision are the Merge Forward "Conferences for Everyone" focus, and each one needs a different feature type.
The `verified` flag keeps the app from telling a wheelchair user a guess as if it were fact.

## Aria's call

Which features to seed as unverified from the floor plan, and which to leave out until someone checks them on the floor.
The default if she does not decide: elevators and step-free entrances seeded as unverified, everything else only from the organiser's page.

## Acceptance

- Every feature document has a `verified` value and a source.
- Every room named in the schedule index resolves to a room document, or is listed as unmatched in the ingest log.

## Touches

`data/venue/`, `ingest/`.
