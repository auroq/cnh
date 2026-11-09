# 0009. Gitar's approving review is the merge gate

Status: Accepted, 2026-11-09. Aria's call, roadmap item 043.

## Context

`main` requires the `Gitar` status check, and `AI_USAGE.md` said that Gitar's review was the gate.
It was not: on PR 18 Gitar requested changes with six findings and the `Gitar` status check still reported SUCCESS, so `main` was merging on CI alone.
Gitar can post a GitHub approving review when the criteria in `.gitar/config/approve.md` hold ("Auto-approve PRs based on code review" in its dashboard, turned on by Aria), and GitHub branch protection can require one approving review.
Gitar reads `approve.md` from the default branch only, so the criteria take effect after they merge.
Gitar's own auto-merge stays off; GitHub's auto-merge, armed by `tools/pr.sh`, does the merging.

## Decision

- `.gitar/config/approve.md` says when Gitar may approve: no blocking finding, no high risk, every applicable rule passed, an issue cited, no test weakened; documentation-only PRs and image bump PRs are approved on their own terms; changes under `infra/`, `.gitar/`, `.github/workflows/` and `tools/pr.sh` need an explanation and an ADR or a plan line in the body.
- `main` requires one approving review besides the status checks `docs`, `manifests`, `Gitar` and `architecture`; `enforce_admins` stays off.
- `tools/pr.sh` requires `--closes N` or `--refs N` and writes it into the body, so every PR meets the citation rule by construction.

## Consequences

- A PR merges only when CI passes and Gitar has approved it; a status check that says SUCCESS no longer suffices.
- A PR Gitar does not approve stays BLOCKED until someone answers its findings (`gitar fix this` or a push) or a person approves it; Aria is an admin and can approve or merge by hand as the fallback.
- Image bump and revert PRs opened by CD are approved on their own terms, so the pipeline in ADR 0008 keeps working without a person.
- The Gitar track gets its evidence: one PR approved and merged by Gitar alone, and one deliberate negative PR (a deleted test, no issue cited) that Gitar withholds approval from, closed unmerged and linked from `docs/demo/`.
