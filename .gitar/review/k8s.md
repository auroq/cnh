# Kubernetes manifest review

Everything under `deploy/` is synced to the cluster by Argo CD and audited by Radar.
Flag any workload in a PR that breaks one of these, and say which line:

- Every container sets `resources.requests` and `resources.limits` for both cpu and memory.
- Every container has `readinessProbe` and `livenessProbe`.
- Pod `securityContext` sets `runAsNonRoot: true` and `seccompProfile.type: RuntimeDefault`.
- Every container `securityContext` sets `allowPrivilegeEscalation: false`, `readOnlyRootFilesystem: true` and drops all capabilities.
- `automountServiceAccountToken: false` unless the workload talks to the Kubernetes API, and then the PR says why.
- App workloads are Argo Rollouts `Rollout` resources with at least 2 replicas and a matching `PodDisruptionBudget`.
- Images are pinned to a tag or digest, never `latest`.
- No `hostPath`, `hostNetwork`, `privileged` or `NET_RAW`.
- Secrets are `SealedSecret` resources; a plain `Secret` with data in it is a blocker.
