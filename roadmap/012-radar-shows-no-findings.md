# 12. Radar shows no findings

**Done 2026-11-09 (PRs #27, #29, #40; the real aborted canary of 12:37 MT in docs/demo; the four remaining Radar findings accepted here and dismissed in Radar by Aria; issue #12). Filed 2026-11-09, from the kickoff plan. Nobody's call: Radar's checks are the spec.**

Depends on: 005, 009

## What

Once the API, ingest and Elasticsearch are all deployed, check Radar's view of the app namespace and fix every finding through a Gitar-reviewed PR.
Then stage one failure for the demo, such as a Rollout canary whose probe fails, and show Radar finding it and Argo Rollouts aborting it.

## Why

Radar "AI SRE" judges both the clean manifests and the agent's diagnosis of a real failure.

## Acceptance

- Radar lists no findings for the app namespace, or each remaining one has a reason in the README.
- The staged failure is recorded: Radar's diagnosis and the aborted rollout, in screenshots or the demo video.

## Touches

`deploy/`.

## Staging the failure

The Rollout `widedoor` has `progressDeadlineSeconds: 120` and `progressDeadlineAbort: true`, so a canary that never becomes ready is aborted by Argo Rollouts on its own and the stable pods keep serving.
Images are pinned by commit SHA in `deploy/app/kustomization.yaml`, so the staged failure is one ordinary release that goes wrong:

1. `tools/bump-image.sh --sha <a commit SHA the images workflow never built>` opens the bump PR; merge it.
2. Argo CD syncs, the canary pod goes to `ImagePullBackOff`, and Radar's live issues list it under image pull failures with the Rollout as the subject.
3. After two minutes Argo Rollouts aborts the canary; `kubectl argo rollouts get rollout widedoor -n wide-door` shows Degraded with the stable ReplicaSet still at full weight, and the site keeps answering.
4. `tools/bump-image.sh` with no flags pins back to the latest good build, and the Rollout recovers.

Radar's remaining findings on the namespace after items 006 and 012 are the two accepted `readOnlyRootFilesystem` ones on Elasticsearch and Kibana, and two ECK-generated `*-certs-public` Secrets that Radar calls unused because no pod mounts them (HTTP TLS is off inside the cluster and terminated at the Gateway); the operator owns them and they stay.
Aria dismissed all four in Radar with those reasons written in, so Radar's view and this repository say the same thing.
