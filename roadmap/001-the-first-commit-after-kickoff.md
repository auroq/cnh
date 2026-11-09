# 1. The first commit after kickoff

**Open. Filed 2026-11-09, from the kickoff plan. Nobody's call: the contest rules require all three files.**

Depends on: nothing

## What

The first commit after the `pre-kickoff` tag adds a README skeleton, `AI_USAGE.md` and `.mcp.json`.
The README gets headings for what the app is, its architecture, how to run it and the tracks it enters, with the architecture section left to fill as the rings land.
`AI_USAGE.md` follows [ADR 0007](../docs/adr/0007-ai-disclosure.md).
`.mcp.json` lists the MCP servers the build uses, with no secrets in it.
The language of the API is chosen here and written into the README.

## Why

The rules require a public repo, a README that explains the architecture and disclosed AI use.
Landing all three first means a submission is valid from the first hour, whatever ring the day ends on.
Every track is judged on this repo, so every track depends on it.

## Acceptance

- `git log pre-kickoff..` shows this commit first.
- The README has an Architecture heading and links `ROADMAP.md`, `docs/adr/` and `AI_USAGE.md`.
- `.mcp.json` parses as JSON and contains no token, key or password.

## Touches

`README.md`, `AI_USAGE.md`, `.mcp.json`.
