# 0013. Gitar blocks a merge on important findings; approval stays advisory

Status: Accepted, 2026-11-09, 15:25 MT. Aria's call. Supersedes the blocking half of [ADR 0010](0010-gitar-approval-is-advisory.md); 0010's facts stand.

## Context

ADR 0010 made Gitar's approval advisory because a required approving review cannot be satisfied on a repository with one GitHub author.
That left the other direction open: between 13:47 and 15:25 MT nine pull requests merged with Gitar findings still open (73, 76, 78, 80, 82, 86, 88, 90, 93), because the `Gitar` status check reports SUCCESS whatever the review said.
GitHub can refuse a merge on a "request changes" review without ever needing an approving one, and Gitar can post that review on findings it rates important or higher.

## Decision

- Branch protection on `main` requires a pull request with `required_approving_review_count: 0`; the status checks (`docs`, `manifests`, `Gitar`, `architecture`, `integration`) are unchanged, strict stays off, `enforce_admins` stays off.
- Gitar's dashboard setting "Block merges: Important or higher" is on: on an important or critical finding Gitar posts a GitHub "request changes" review, and GitHub refuses the merge until Gitar dismisses it because the finding is fixed, or Aria comments `gitar unblock` (enabled in the dashboard).
- Approval is unchanged from ADR 0010: advisory, never required, visible on every PR.

## Consequences

- "Never Merge Broken" now has a lock in the direction that matters: a PR with an important finding cannot merge by auto-merge or by hand, and the record of why is Gitar's review on the PR.
- The single-author problem of ADR 0010 does not apply, because no approving review is needed; a human is involved only to say `gitar unblock`, and that comment is itself the audit trail.
- `AI_USAGE.md` and `docs/demo/2026-11-09-gitar-merge-gate.md` say the same thing as this record.
