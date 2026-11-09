# 75. Distance without assumptions: how Wide Door describes how far something is

**Open. Filed 2026-11-09 by Aria, p1; worker issue-75 in progress, research and ADR 0012 first. Nobody's call: published accessibility practice decides.**

Depends on: 063

Numbered 075 because issue numbers share a sequence with pull requests; the item number equals the issue number.
The full text, discussion and evidence are on [issue #75](https://github.com/auroq/cnh/issues/75); this file is the summary the roadmap keeps.

## What

Decide from published accessibility and wayfinding practice how the app says how far a place or session is, record it in ADR 0012 with sources, then apply it everywhere distance is shown.
The expected shape: the tier in words (same level, one level up or down, other building), the route facts we know (elevator needed, step-free, level changes), and "we do not know" when we do not, never a single walk time that assumes a body and a pace.

## Why

Merge Forward: a wheelchair user, someone with chronic fatigue, a blind attendee with a cane and a parent with a stroller take different routes and times; a wrong estimate is worse than none.

## Acceptance

- ADR 0012 "How we describe distance" with sources.
- `Summarise` and the page follow it; the acceptance rows that quote distance wording pass.

## Touches

`docs/adr/0012-*.md`, `pkg/ask`, `pkg/web`, `testing/acceptance`.
