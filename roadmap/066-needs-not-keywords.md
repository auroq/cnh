# 66. Needs, not keywords: what people say maps to what the venue offers

**Open. Filed 2026-11-09 by Aria, p1; worker issue-66 in progress (PR 84). Nobody's call: the vocabulary is data in `data/venue/needs.yaml`.**

Depends on: 063

Numbered 066 because issue numbers share a sequence with pull requests; the item number equals the issue number.
The full text, discussion and evidence are on [issue #66](https://github.com/auroq/cnh/issues/66); this file is the summary the roadmap keeps.

## What

A need vocabulary, data-driven: `data/venue/needs.yaml` with one entry per need (`id`, `phrases` people say, `kinds` of place to offer in order, `note` said when the match is a stand-in).
Seeded with mobility-accessible, private-space, medical-private, menstrual, first-aid, emergency-exit, rest and help.
`ask` infers the need from the question, asks the store for those kinds, and says the note when the exact thing is not on offer.

## Why

Merge Forward: people say "I need to inject insulin" or "somewhere to fix my makeup", not the kind name of a room; an honest stand-in with a note beats a wrong match or nothing.

## Acceptance

- Each seeded phrase resolves to its kinds in a unit test over the YAML.
- Acceptance rows for the examples pass against the live site.

## Touches

`data/venue/needs.yaml`, `pkg/ask`, `testing/acceptance`.
