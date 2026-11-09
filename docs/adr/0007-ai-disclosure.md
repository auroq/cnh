# 0007. AI use disclosed in AI_USAGE.md and commit trailers

Status: Accepted, 2026-11-09.

## Context

The rules require AI use to be disclosed.
The project is built pairing with Claude Code, and the planning documents were drafted with it too.

## Decision

`AI_USAGE.md` at the repo root says which tools were used, for what (planning, code, tests, documents, review), and what a human decided or checked.
Every commit with AI-written content ends with a `Co-Authored-By:` trailer naming the model.
Gitar's PR reviews are AI as well and are named in `AI_USAGE.md`.

## Consequences

- `git log --grep Co-Authored-By pre-kickoff..` lists the AI-assisted commits.
- A Gitar rule in item 004 flags a commit without the trailer, so the record stays complete.
