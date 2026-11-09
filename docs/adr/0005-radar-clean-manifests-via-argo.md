# 0005. Manifests pass Radar from the first commit, deployed by Argo

Status: Accepted, 2026-11-09.

## Context

The Radar Cloud agent is connected to the cluster and flags workloads that miss its checks.
Argo CD and Argo Rollouts were installed before kickoff.
Hardening a workload after it works means touching every manifest again in the last hour, when there is no time.

## Decision

Every workload in the app namespace, from the first placeholder, has:
requests and limits, readiness and liveness probes, `runAsNonRoot: true`, `allowPrivilegeEscalation: false`, `readOnlyRootFilesystem: true` and `automountServiceAccountToken: false`.
Long-running workloads are Argo Rollouts `Rollout`s with 2 replicas and a PodDisruptionBudget.
Jobs carry the same pod settings without replicas or a PDB.
Argo CD syncs `deploy/` from `main`; nothing is applied with `kubectl apply` by hand.

## Consequences

- A read-only root filesystem means every writable path is an explicit `emptyDir`.
- Elasticsearch pods are shaped by ECK, so any check they cannot meet is explained in the README instead of fixed.
- The Gitar rules in item 004 check new manifests against this list, so a regression is caught in review.
- Rollouts give the Radar demo a real failure to diagnose: an aborted canary.
