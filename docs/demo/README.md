# Evidence from the day

Each page here was captured while it happened, not staged afterwards.
The live app was https://widedoor.cnh.auroq.com until it was taken down on 2026-11-13; the links below no longer answer.

| page | what it shows |
| --- | --- |
| [2026-11-09-gitar-merge-gate.md](2026-11-09-gitar-merge-gate.md) | Gitar approving PRs, withholding approval from PR 47, and the merge policy that followed |
| [2026-11-09-rollout-abort.md](2026-11-09-rollout-abort.md) | a real canary aborted by Argo Rollouts at two minutes and diagnosed by Radar |

## Five links per track

### Gitar / Sonar, Never Merge Broken

1. [PR 47](https://github.com/auroq/cnh/pull/47): a deleted test, refused by Gitar and closed unmerged.
2. [PR 42](https://github.com/auroq/cnh/pull/42): documentation, approved by Gitar and merged.
3. [PR 48](https://github.com/auroq/cnh/pull/48): `AI_USAGE.md`, auto-approved by Gitar and merged.
4. [ADR 0013](../adr/0013-gitar-blocks-on-important-findings.md): important findings block the merge since 15:25 MT.
5. [2026-11-09-gitar-merge-gate.md](2026-11-09-gitar-merge-gate.md): the whole afternoon, including the nine PRs that merged with findings open.

### Dash0, Stop Flying Blind

1. [PR 44](https://github.com/auroq/cnh/pull/44): traces, metrics and logs reach Dash0, with the verification comment on the PR.
2. [`deploy/dash0/wide-door.json`](../../deploy/dash0/wide-door.json): the "Wide Door" dashboard as code (item 089).
3. [`pkg/telemetry`](../../pkg/telemetry): the one adapter for all three signals.
4. [`pkg/ask/telemetry.go`](../../pkg/ask/telemetry.go): the instruments, including `widedoor.ask.unverified_served`.
5. [README, Observability](../../README.md#observability): what each signal carries.

### Radar, AI SRE

1. [2026-11-09-rollout-abort.md](2026-11-09-rollout-abort.md): the aborted canary and Radar's correlated change.
2. [Item 012](../../roadmap/012-radar-shows-no-findings.md): the four remaining findings, accepted there and dismissed in Radar by Aria.
3. [PR 29](https://github.com/auroq/cnh/pull/29): the Rollout aborts a canary that stalls for two minutes.
4. [PR 27](https://github.com/auroq/cnh/pull/27): Radar-matching Kibana PDB and topology spread.
5. [ADR 0005](../adr/0005-radar-clean-manifests-via-argo.md): manifests pass Radar from the first commit.

### Elasticsearch, Find the Right Talk

1. [`pkg/esindex/store.go`](../../pkg/esindex/store.go): BM25 and `semantic_text` fused with RRF.
2. [`data/venue/needs.yaml`](../../data/venue/needs.yaml): the need vocabulary, indexed as the `needs` index (item 066).
3. [`pkg/esindex/index.go`](../../pkg/esindex/index.go): the mappings for sessions, places, needs and reports.
4. [Item 072](../../roadmap/072-sessions-only-when-they-answer-the-question.md): sessions only when they answer the question.
5. [ADR 0004](../adr/0004-elasticsearch-as-the-single-store.md): Elasticsearch as the one store.

### Merge Forward, Conferences for Everyone

1. https://widedoor.cnh.auroq.com: ask the venue.
2. https://widedoor.cnh.auroq.com/profile: what Wide Door knows about you, kept on your device.
3. [PR 102](https://github.com/auroq/cnh/pull/102): attendees verify, dispute and annotate places.
4. [ADR 0012](../adr/0012-how-we-describe-distance.md): distance as facts, never a walk time.
5. [`docs/venue-sources.md`](../venue-sources.md): where every venue fact came from and whether it is verified.
