# 89. Dash0 dashboard as code: the Wide Door panels live in the repo

**Open. Filed 2026-11-09 by Aria (issue #89). Nobody's call: the metrics exist, the dashboard is their reviewed home.**

Depends on: 011

Numbered 089 because issue numbers share a sequence with pull requests; the item number equals the issue number.

## What

The Wide Door dashboard is a file in the repo, reviewed like every other change, and reaches Dash0 without anyone clicking in the UI.
The preferred route was a `PersesDashboard` custom resource synced by the Dash0 operator.
The cluster has no `persesdashboards.perses.dev` CRD (the operator only syncs one if the Perses CRD is installed first, which is a cluster-wide change under `infra/`), so the fallback route is taken: a Perses dashboard in `deploy/dash0/wide-door.json`, pushed by `tools/dash0-dashboard.sh` through `PUT /api/dashboards/wide-door` with the token from the CNH vault via `secret-run`, run from `make dashboard`.
Panels, all filtered on `service.name = widedoor`, dataset `default`:

1. Ask latency: p50 and p95 of `widedoor_ask_duration_milliseconds` over time, split by `has_location`.
2. Answers by kind: `widedoor_ask_answers_total` per minute stacked by `kind`, and a second panel split by `verified`.
3. What volunteers should go verify: running `widedoor_ask_unverified_served_total` by `kind`, with the total as a stat.
4. Ingest: `widedoor_ingest_sessions_indexed_total` and `widedoor_ingest_detail_failures_total` per run, and `widedoor_ingest_duration_milliseconds` per run.
5. Needs we were told about, never who: `widedoor_profile_chip_used_total` by `chip`, empty until item 081 ships.
6. Requests, 5xx and p95 by `http_route` from the `otelhttp` metrics, split by `service_version` so each release is visible.

Each panel has a one-line description a judge can read without us explaining it.

## Why

Dash0 "Stop Flying Blind": the dashboard is a reviewed, versioned artefact, pushed by the same repo as the app, and it tells the accessibility story in metrics (unverified answers as the to-do list, needs counted without identities).

## Acceptance

- The dashboard is visible in Dash0 under the name "Wide Door"; the PR body has the Dash0 id and the API response confirming it exists.
- Metric names and attribute keys match what the app exports, checked against the Dash0 Prometheus API; the PR body lists them.
- `tools/dash0-dashboard.sh --help` works, `make dashboard` pushes, `make dashboard-check` reads back.
- README gains two sentences under Observability about the dashboard and where it lives.

## Touches

`deploy/dash0/wide-door.json`, `tools/dash0-dashboard.sh`, `Makefile`, `README.md`, this file.

## Sequencing

Independent of the other workers; does not touch `pkg/`.
