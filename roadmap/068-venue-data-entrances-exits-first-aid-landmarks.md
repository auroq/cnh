# 68. Venue data: entrances by street side, exits and what is outside them, first aid, landmarks and aliases

**Open. Filed 2026-11-09 by Aria, p1. Aria's call: what is verified, what is a guess, and how it is worded; this is the human-authored part.**

Depends on: 008

Numbered 068 because issue numbers share a sequence with pull requests; the item number equals the issue number.
The full text, discussion and evidence are on [issue #68](https://github.com/auroq/cnh/issues/68); this file is the summary the roadmap keeps.

## What

`data/venue/places.yaml` grows: every public entrance with `side` and whether it is step-free, emergency exits with `leads_to`, what is outside the main doors (street, TRAX stop, skybridge), first aid points, landmarks with `aliases`, and aliases on existing places.
Every new fact carries `verified` and `source`; most entrances are `verified: false` from the floor plan until someone checks on the floor.

## Why

Merge Forward: items 066 and 067 are only as good as the facts, and judging what is verified is the human's job.

## Acceptance

- `venue.Load` accepts the file and every entry has `verified` and `source`.
- Items 066 and 067 find their examples in it.

## Touches

`data/venue/places.yaml`, `pkg/venue`.
