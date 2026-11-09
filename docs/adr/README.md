# Architecture decision records

One file per decision, numbered from 0001 and never renumbered.
A record is accepted or proposed; a decision that changes gets a new record that supersedes the old one, and the old one stays.

| record | decision | status |
| --- | --- | --- |
| [0001](0001-monorepo-with-pre-kickoff-infra.md) | one repo, with `infra/` built before the event and tagged `pre-kickoff` | Accepted |
| [0002](0002-build-ask-the-venue.md) | build Option 1, "Ask the venue"; Option 2 is a stretch | Accepted |
| [0003](0003-work-tracking.md) | GitHub Issues as tickets, `ROADMAP.md` as the index, `roadmap/` as the detail | Accepted |
| [0004](0004-elasticsearch-as-the-single-store.md) | Elasticsearch as the one store, with hybrid search | Accepted |
| [0005](0005-radar-clean-manifests-via-argo.md) | manifests pass Radar from the first commit; deploy with Argo CD and Argo Rollouts | Accepted |
| [0006](0006-sealed-secrets.md) | secrets as SealedSecrets only | Accepted |
| [0007](0007-ai-disclosure.md) | AI use disclosed in `AI_USAGE.md` and in commit trailers | Accepted |
| [0008](0008-ci-cd-separation-and-rollback-by-revert.md) | CI and CD are separate stages; a post-deploy gate tests the live site; rollback is a revert PR | Accepted |
| [0009](0009-gitar-approval-is-the-merge-gate.md) | Gitar's approving review, against `.gitar/config/approve.md`, is what lets a PR merge | Accepted |
