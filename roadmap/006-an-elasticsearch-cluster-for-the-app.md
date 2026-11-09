# 6. An Elasticsearch cluster for the app

**Done 2026-11-09 (issue #6; 3-node ECK cluster `wide-door` green, trial license active, `semantic_text` smoke-tested with ELSER, Kibana at https://kibana.cnh.auroq.com). Filed 2026-11-09, from the kickoff plan. Nobody's call: ADR 0004 chose Elasticsearch as the one store.**

Depends on: 005

## What

An ECK `Elasticsearch` resource in the app namespace, synced by Argo CD, separate from the learning-only cluster in `infra/sandbox/`.
It needs an ML-capable node of at least 4 GB for ELSER, and the default `.elser-2-elasticsearch` endpoint behind `semantic_text`.
ELSER and `semantic_text` need a trial or paid license on a self-managed cluster, so the trial license is started here.
ELSER's adaptive allocations scale to zero after 15 idle minutes, so the first query after a lull is slow; set a minimum of one allocation for the demo.
Credentials reach the API as a Secret that ECK generates, never typed anywhere.

## Why

Elasticsearch "Find the Right Talk" judges hybrid search, and every later ring reads from this cluster.

## Acceptance

- A test index with a `semantic_text` field accepts a document and returns it for a paraphrased query.
- Radar lists no findings for the Elasticsearch pods, or each remaining one is explained in the README.

## Touches

`deploy/elasticsearch/`.

## Notes from landing it

The ECK trial license is a data-less Secret in `elastic-system` created by `infra/platform`, which starts an Enterprise trial for every cluster the operator manages.
Elasticsearch and Kibana run without `readOnlyRootFilesystem`, because both write under their install directory; every other check (non-root, no privilege escalation, dropped capabilities, resources, probes, no token automount, 3 and 2 replicas with PDBs) is met.
