# Gitar's approval as the merge gate, 2026-11-09 afternoon

Evidence for roadmap item 043 and the Gitar/Sonar "Never Merge Broken" track.
From 15:00 MT, `main` requires one approving review besides the status checks, and Gitar posts that review only when [`.gitar/config/approve.md`](../../.gitar/config/approve.md) holds ([ADR 0009](../adr/0009-gitar-approval-is-the-merge-gate.md)).

## Why it was needed

On PR 18 Gitar requested changes with six findings while the required `Gitar` status check reported SUCCESS.
A status check says "I reviewed", not "I agree"; the approving review is the verdict GitHub can require.

## Approved and merged by Gitar alone

- PR 42 (roadmap bookkeeping, documentation only): Gitar `APPROVED`, auto-merge completed.
- PR 48 (`AI_USAGE.md` brought current, documentation only): Gitar commented, then `APPROVED` ("Gitar has auto-approved this PR"), and GitHub's auto-merge merged it at 20:13:40 UTC with no human review.

## Withheld

- PR 47 (deliberate: `pkg/esindex/index_test.go` deleted with no replacement, no roadmap issue cited): Gitar requested changes ("1 finding") and reported "Auto-approval: Withheld: unresolved findings remain"; GitHub showed `REVIEW_REQUIRED` and `BLOCKED` with every status check green.
  Closed unmerged.

## What a person does when Gitar says no

Answer the findings (`gitar fix this`, or a push), cite the issue, or, as the fallback, approve by hand; Aria is an admin and `enforce_admins` is off.
