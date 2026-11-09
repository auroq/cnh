# 5. The app namespace is Argo CD managed and Radar clean

**Done 2026-11-09 (b7fa838, issue #5; placeholder live at https://widedoor.cnh.auroq.com, Argo CD Synced/Healthy). Filed 2026-11-09, from the kickoff plan. Nobody's call: Radar's checks are the spec, see ADR 0005.**

Depends on: 001

## What

An Argo CD Application points at `deploy/` in this repo and syncs the app namespace.
The first thing it syncs is a placeholder container, so the pipeline is proven before there is code to ship.
Every workload is an Argo Rollouts `Rollout` with 2 replicas and a PodDisruptionBudget.
Every pod sets requests and limits, readiness and liveness probes, `runAsNonRoot: true`, `allowPrivilegeEscalation: false`, `readOnlyRootFilesystem: true` and `automountServiceAccountToken: false`.
An HTTPRoute attaches to the existing `public` Gateway under `*.cnh.auroq.com`.

## Why

Radar "AI SRE" judges manifests against exactly these checks, and fixing them later costs more than writing them right once ([ADR 0005](../docs/adr/0005-radar-clean-manifests-via-argo.md)).
Argo CD and Argo Rollouts were installed before kickoff, so this is wiring and not setup.

## Acceptance

- Argo CD shows the Application Synced and Healthy.
- The placeholder answers over HTTPS on its hostname.
- Radar lists no findings for the app namespace.

## Touches

`deploy/`, an Argo CD `Application` manifest; nothing under `infra/`.
