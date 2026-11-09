# 43. Gitar's approval is the merge gate

**Open. Filed 2026-11-09, during Ring 4, by Aria. Aria's call: auto-approve on in Gitar's dashboard, Gitar's own auto-merge off, a human approval as the fallback.**

Depends on: 004, 036

Numbered 043 because issue numbers share a sequence with pull requests; the item number equals the issue number.

## What

- `.gitar/config/approve.md` states when Gitar may post an approving review: no blocking finding, no high risk, every applicable rule passed, an issue cited, no test weakened; documentation-only and image bump PRs approved on their own terms; policy files need an explanation and an ADR.
- `main` requires one approving review besides the four status checks; `enforce_admins` stays off so Aria can approve or merge by hand.
- `tools/pr.sh` requires `--closes N` or `--refs N`.
- ADR 0009 records it; `AI_USAGE.md` says the gate is Gitar's approval.
- Evidence for the Gitar track: one PR that Gitar approved and auto-merge completed with no human, and one deliberate negative PR (a deleted test, no issue cited) that stays BLOCKED, closed unmerged and linked from `docs/demo/`.

## Why

On PR 18 Gitar requested changes with six findings while the required `Gitar` status check reported SUCCESS; `main` was merging on CI alone and `AI_USAGE.md` overstated the gate.
"Never Merge Broken" is judged on whether the review is the gate, so the verdict has to be the thing GitHub requires.

## Acceptance

- Branch protection on `main` lists `required_approving_review_count: 1` and the four checks.
- The next ordinary PR shows an APPROVED review by `gitar-bot[bot]` and merges by auto-merge alone.
- The negative PR shows no approval, `mergeStateStatus: BLOCKED`, and is closed unmerged.

## Touches

`.gitar/config/approve.md`, `docs/adr/`, `AI_USAGE.md`, `tools/pr.sh`, `docs/demo/`, branch protection on `main`.
