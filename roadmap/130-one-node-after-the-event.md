# 130. One MicroK8s node after the event

**Open. Filed 2026-11-10 by Aria (issue #130). Aria's call: MicroK8s from snap, deploys stay automated, Kibana stays, Argo CD stays.**

## What

Move the live site from the three-node DOKS cluster to one MicroK8s droplet for the week after the event, then destroy the cluster, its load balancer and its volumes.

## Why

The cluster costs about $29 a day at list for an app that uses 2 percent of one node.
One droplet is about $2.80 a day and keeps what matters for bug fixes: GitOps, rolling deploys, the CD gate, Dash0 traces and Kibana.

## Acceptance

- `deploy/node` renders with kustomize and uses `deploy/app` as its base, so the image bump PRs keep working.
- `infra/node` and `infra/node-platform` bring the droplet, MicroK8s, Argo CD, the Secrets and the cd-gate up with `make up-node`.
- https://widedoor.cnh.auroq.com answers from the droplet with a Let's Encrypt certificate and a trace arrives in Dash0.
- The CD gate waits on the Deployment instead of the Rollout; a push after the cutover passes bump, sync, Playwright and acceptance.
- The DOKS cluster, load balancer and volumes are gone and the DigitalOcean API confirms it.
- [ADR 0014](../docs/adr/0014-one-node-after-the-event.md) and [the transition doc](../docs/transitions/post-event-single-node.md) record the decision and the cutover.

## Touches

`deploy/node/`, `deploy/argocd/wide-door-node.yaml`, `infra/node/`, `infra/node-platform/`, `tools/wait-deployment.sh`, `tools/node-bootstrap.sh`, `Makefile`, `.github/workflows/ci.yml`, `.github/workflows/cd.yml`, `.gitar/review/k8s.md`, `README.md`, `docs/architecture.md`.
