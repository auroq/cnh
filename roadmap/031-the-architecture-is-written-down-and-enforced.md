# 31. The architecture is written down and enforced

**Open. Filed 2026-11-09, during Ring 3, by Aria. Nobody's call: it writes down what item 025 built.**

Depends on: 025

Numbered 031 because issue numbers share a sequence with pull requests; the item number equals the issue number.

## What

- `.gitar/review/architecture.md` names the four layers (model, use case, adapters, controllers) and the import direction, one quotable line per rule.
- `tools/archcheck.sh` reads the import graph with `go list` and fails with the offending edge named; the `architecture` CI job runs it and becomes a required check once it is green on `main`.
- `docs/architecture.md` is the one page Aria presents: the hexagon with the ports named, one line per package, one line per port (who declares it, who implements it), one line per binary, and a "where to look when" table.
- A Gitar rule asks for `docs/architecture.md` to change whenever a package is added or an import between layers changes.

## Why

Rules 6 and 7 of the hackathon: the participant must be able to explain the architecture and the decisions.
A page in plain words, a check that keeps the code matching it, and a reviewer that asks for the page to be updated make that explanation true all day, not just at submission.

## Acceptance

- `tools/archcheck.sh` passes on `main` and fails on a branch where `pkg/ask` imports `pkg/esindex`.
- `architecture` is a required check on `main`.
- The README's Architecture section links `docs/architecture.md`.

## Touches

`.gitar/review/`, `.gitar/rules/`, `tools/`, `.github/workflows/ci.yml`, `docs/architecture.md`, `README.md`.
