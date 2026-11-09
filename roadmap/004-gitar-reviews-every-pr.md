# 4. Gitar reviews every PR

**Open. Filed 2026-11-09, from the kickoff plan. Nobody's call: the track defines what it wants.**

Depends on: 001

## What

Review rules go in `.gitar/review/` and `.gitar/rules/`, written for this repo's own failure modes:
manifests that would fail Radar's checks, a secret outside a SealedSecret, a missing `Co-Authored-By` trailer on AI-assisted commits, and UI changes that drop an accessible name or a keyboard path.
A CI workflow runs the tests and a manifest check on every PR.
Auto-apply is on, so Gitar can push a fix for a failing CI run to the PR.
From Ring 1 onward every change lands on `main` through a PR that Gitar has reviewed.

## Why

Gitar/Sonar "Never Merge Broken" judges review rules and fixed CI.
The rules double as the guard that keeps Radar and accessibility from regressing during the afternoon rush.

## Acceptance

- One PR shows a Gitar comment raised by a rule from `.gitar/rules/`.
- One PR shows CI going red, a Gitar auto-applied fix, and CI going green.
- `main` has no commit after Ring 0 that did not come through a PR.

## Touches

`.gitar/review/`, `.gitar/rules/`, `.github/workflows/`.
