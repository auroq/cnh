# 75. Distance without assumptions: how Wide Door describes how far something is

**Open. Filed 2026-11-09 by Aria (issue #75). Nobody's call: the decision follows published accessibility and wayfinding practice, recorded in ADR 0012, and the wording follows the decision.**

Depends on: 008, 009, 010

Numbered 075 because issue numbers share a sequence with pull requests; the item number equals the issue number.

## What

Decide, from published accessibility and wayfinding practice, how the app expresses "how far" a place or session is, then say it that way everywhere distance is shown.
A single walk time assumes a pace and a body.
A wheelchair user, someone with chronic fatigue, a blind attendee navigating by cane and a parent with a stroller each take different routes and times, and a wrong estimate is worse than none.
The research is thirty minutes, cited in [ADR 0012](../docs/adr/0012-how-we-describe-distance.md) and in the PR.

## Why

Merge Forward, Conferences for Everyone.
The app exists so that an attendee with limited mobility, hearing or vision can trust what it says.
A guessed "about 5 minutes" is a promise the app cannot keep for the people it is for, and the practice the research found says so.

## Acceptance

- ADR 0012 "How we describe distance", with the sources and the decision.
- Distance wording follows the decision: the tier in words (same level, one level up or down, other building), the route facts we actually know (elevator needed, step-free, number of level changes, connected to the other building), and no single walk time; if a time is ever shown, it is a range tied to a pace the user chose, never a default.
- "We do not know" is said when we do not know, never guessed.
- `pkg/ask` `Summarise` and the templates updated; the per-session `walk_minutes` of item 072 is not shipped and route facts take its place in the API; acceptance rows updated; the e2e axe suite passes.
- A sentence in the README under Architecture on why there is no walk time.

## Touches

`docs/adr/0012-how-we-describe-distance.md`, `data/venue/places.yaml` (route facts where known), `pkg/model`, `pkg/venue`, `pkg/esindex` (the `route` field), `pkg/ask` (`Summarise`), `pkg/api/v1` and `pkg/web` (the `route` field and the template), `testing/acceptance/cases.go`, README.

## Sequencing

Research and the ADR first, then the data, then the wording, in that order of commits, because items 066, 067 and 072 are editing `pkg/ask` at the same time and the `pkg/ask` change here is kept to the distance sentence.
