# Kubernetes manifest review

Everything under `deploy/` is synced to the cluster by Argo CD and audited by Radar.
Flag any workload in a PR that breaks one of these, and say which line:

- Every container sets `resources.requests` and `resources.limits` for both cpu and memory.
- Every long-running container has `readinessProbe` and `livenessProbe`.
- Pod `securityContext` sets `runAsNonRoot: true` and `seccompProfile.type: RuntimeDefault`.
- Every container `securityContext` sets `allowPrivilegeEscalation: false`, `readOnlyRootFilesystem: true` and drops all capabilities.
- `automountServiceAccountToken: false` unless the workload talks to the Kubernetes API, and then the PR says why.
- App workloads are Argo Rollouts `Rollout` resources with at least 2 replicas and a matching `PodDisruptionBudget`.
- Images are pinned to a tag or digest, never `latest`.
- No `hostPath`, `hostNetwork`, `privileged` or `NET_RAW`.
- Secrets are `SealedSecret` resources; a plain `Secret` with data in it is a blocker.

## Accepted exceptions

These are decisions, written down in the roadmap, and not findings.
Do not ask for them to change and do not auto-apply a change to them.

- The ECK `Elasticsearch` and `Kibana` resources run without `readOnlyRootFilesystem`, because both write under their install directory and the operator owns their pod template (roadmap item 006, ADR 0005).
- A `CronJob` or `Job` has no replicas, no `PodDisruptionBudget` and no readiness or liveness probes; a batch pod runs to completion and the probes would be theatre.
- Image tags come from the `images` workflow: `:main` and the SHA of the commit on `main` that built it. A tag that workflow never built does not exist and will not pull.
