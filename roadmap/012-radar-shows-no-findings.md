# 12. Radar shows no findings

**Open. Filed 2026-11-09, from the kickoff plan. Nobody's call: Radar's checks are the spec.**

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
