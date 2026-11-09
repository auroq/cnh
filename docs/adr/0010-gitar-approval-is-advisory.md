# 0010. Gitar's approval is advisory; the status checks are the gate

Status: Accepted, 2026-11-09. Supersedes [ADR 0009](0009-gitar-approval-is-the-merge-gate.md); its blocking half is superseded by [ADR 0013](0013-gitar-blocks-on-important-findings.md) at 15:25 MT, its facts about approval stand. Aria's call, roadmap item 043.

## Context

ADR 0009 made Gitar's approving review a required review on `main` at 13:40 MT.
Within the hour it proved unsatisfiable on this repository, for two reasons that have nothing to do with the code:

- Every pull request is authored by `auroq`, because every session uses Aria's GitHub login, and GitHub refuses an author's approval of their own pull request; so the human fallback named in ADR 0009 did not exist.
- Gitar's approval step cannot read pull request bodies on this account ("the PR body was not provided"), so any criterion about the body, such as citing an issue or explaining a policy change, can never be met; and the pull request that relaxed those criteria was itself withheld.

Four green pull requests, including the Dash0 telemetry, sat blocked with no path through.

## Decision

- `main` requires the status checks `docs`, `manifests`, `Gitar`, `architecture` and `integration`, and no approving review; set by Aria at 13:47 MT.
- Gitar's auto-approve stays on and its approving review stays visible on every pull request, as advisory evidence of what the reviewer concluded, not as a lock.
- `.gitar/config/approve.md` keeps only criteria Gitar can evaluate from the diff, the title and its own findings; `tools/pr.sh` puts `Refs #N` in the title as well as the body, so the issue is visible to Gitar and to people.
- The evidence from the hour the gate was enforced stands: PR 47 (a deleted test, no issue cited) was withheld and closed unmerged; PR 42 and PR 48 were approved by Gitar alone.

## Consequences

- "Never Merge Broken" is answered honestly: Gitar reviews every pull request, blocks nothing by itself, and its verdict is one click away on each; the status checks are what GitHub enforces.
- A single-author repository cannot have a required human review; a second GitHub identity for the agents would make ADR 0009 possible again, and is noted for after the event.
- `AI_USAGE.md` and `docs/demo/2026-11-09-gitar-merge-gate.md` say the same thing as this record.
