# 0003. Work tracking

Status: Accepted, 2026-11-09. Aria: issues first so the work is as visible to the judges as possible; the markdown fills the gaps. Item 003 carries it out.

## Context

The plan is a set of numbered roadmap items with an index in `ROADMAP.md`.
Judges should be able to see the plan and what happened to it.
Gitar reviews a PR with the context that PR links to.
None of the sponsor tools (Gitar, Dash0, Radar, Elasticsearch) offers ticket tracking, so whatever is used is outside the tracks.

## Decision

Use GitHub Issues on `auroq/cnh` as the ticket system, one public issue per roadmap item.
`ROADMAP.md` stays the index and `roadmap/NNN-*.md` keeps the detail; an issue holds a one-line summary and a link to its file.
Every PR cites its issue, so Gitar's review sees what the change is for.

## Alternative

Roadmap files only, with no issues: PRs cite item numbers in their body.
It is less work and has one source of truth, but Gitar sees only a number, and the status of the work is visible only by reading files.

## Consequences

- Each item's status is in two places, its file and its issue; the file is authoritative and the issue is closed when the file says Done.
- Opening sixteen issues costs a few minutes and is scriptable with `gh issue create`.
