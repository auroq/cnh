# 11. Telemetry reaches Dash0

**Open. Filed 2026-11-09, from the kickoff plan. Nobody's call: the operator is already installed.**

Depends on: 009

## What

The API and the ingest job emit OpenTelemetry traces, metrics and logs, through the Dash0 operator's instrumentation of the app namespace or the SDK if auto-instrumentation misses Elasticsearch calls.
Spans cover the HTTP request, the Elasticsearch query and the ingest run.
Custom metrics: questions asked by feature type, answers that contained an unverified feature, and questions with no answer.
Logs are structured and carry the trace id.

## Why

Dash0 "Stop Flying Blind" judges this directly.
"Questions with no answer" is also the list of what to verify on the floor next.

## Acceptance

- In Dash0, one `/ask` request shows a trace with the HTTP span and its Elasticsearch child span.
- The three custom metrics are visible in Dash0.
- A log line from the API links to its trace.

## Touches

`api/`, `ingest/`, a `Dash0Monitoring` resource in `deploy/`.
