# A real aborted canary, 2026-11-09 12:37 to 12:40 MT

Evidence for roadmap item 012, captured live rather than staged.
PR 34 changed the Rollout's pod template (the container now runs `widedoor serve` and reads `/etc/widedoor/config.yaml`) while the image was still pinned to a build without subcommands.

## What Argo Rollouts did

Events on `Rollout/widedoor`, namespace `wide-door` (UTC):

| time | reason | message |
| --- | --- | --- |
| 19:37:46 | RolloutUpdated | Rollout updated to revision 4 |
| 19:37:46 | NewReplicaSetCreated | Created ReplicaSet widedoor-84bdb888d8 (revision 4) |
| 19:37:46 | ScalingReplicaSet | Scaled up ReplicaSet widedoor-84bdb888d8 (revision 4) from 0 to 1 |
| 19:39:47 | RolloutAborted | ReplicaSet "widedoor-84bdb888d8" has timed out progressing. |
| 19:39:47 | RolloutAborted | Rollout aborted update to revision 4 |

Status afterwards: `Degraded: RolloutAborted: Rollout aborted update to revision 4`.
The two revision 3 pods kept serving; https://widedoor.cnh.auroq.com answered 200 throughout.
`progressDeadlineSeconds: 120` with `progressDeadlineAbort: true` (PR 29) is what made the abort automatic; the canary step was `setWeight: 50`, so at most one new pod ever existed.

## What Radar saw

Radar's live issues for the namespace went from zero to one:

- severity `warning`, category `rollout_stalled`, subject `Rollout/widedoor`, "Rollout is not healthy", `issue_timing: started_after_resource_was_healthy`.
- `correlated_changes` ranked first, flagged `application_configuration_change: true`, at 19:37:46: `args(widedoor) changed` (to `serve`), six `ES_*` environment variables removed, `WIDEDOOR_ELASTICSEARCH_PASSWORD` added, a `config` volume from ConfigMap `widedoor-config` mounted at `/etc/widedoor`.
- The image bump of 19:19 and the `progressDeadlineAbort` change of 19:17 were listed as earlier, unflagged changes.

That is the diagnosis a person would reach after ten minutes with `kubectl describe`: the manifest moved to a new contract before the image did.

## The fix

`tools/bump-image.sh` opened a PR pinning the image built from the same commit as the manifests; once merged, Argo CD synced it, revision 5 passed the canary, and Radar's issue list went back to zero.
Item 036 makes this automatic: CD builds, opens the bump PR, waits for the canary, tests the live site, and reverts the bump through a PR if the gate fails.

## To stage it again on purpose

See "Staging the failure" in [`roadmap/012-radar-shows-no-findings.md`](../../roadmap/012-radar-shows-no-findings.md): bump to a SHA the `images` workflow never built, watch Radar list `image_pull_failed`, watch Rollouts abort at two minutes, bump back.
