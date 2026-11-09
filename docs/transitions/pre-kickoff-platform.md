# The platform predates the event

Status: Current since 2026-11-09.
Delete when: the README's Architecture section describes the platform below, in roadmap item [013](../../roadmap/013-demo-and-submission.md).

## What is true now

Everything under `infra/` at the git tag `pre-kickoff` was built before kickoff, as the rules allow.
It is the venue, not the project: no app code, manifest or design for the app is in it.
`git log pre-kickoff..` is the hackathon's work; see [ADR 0001](../adr/0001-monorepo-with-pre-kickoff-infra.md).

## What is installed

A DigitalOcean Kubernetes cluster in `sfo3` with a wildcard DNS zone `*.cnh.auroq.com`, built by Terraform in `infra/bootstrap/`, `infra/cluster/` and `infra/platform/`.

| tool | what the app uses it for |
| --- | --- |
| Envoy Gateway and Gateway API | the shared `public` Gateway that app HTTPRoutes attach to |
| cert-manager | the wildcard TLS certificate, by DNS-01 |
| Argo CD | syncs the app's manifests from `deploy/` |
| Argo Rollouts | canary rollouts of the app's workloads |
| ECK operator | runs the app's Elasticsearch, which is created after kickoff |
| Dash0 operator | collects the app's OpenTelemetry signals and sends them to Dash0 |
| Radar Cloud agent | checks the app's workloads and diagnoses failures |
| metrics-server | resource metrics for the cluster |
| Sealed Secrets | decrypts the app's committed SealedSecrets |
| Hubble | network flow visibility, from the Cilium that DOKS ships |

`infra/sandbox/` holds learning-only manifests (an OpenTelemetry demo, a single-node Elasticsearch and Kibana, a local compose file); none of it is part of the app.

## What is temporary about it

The platform is described here and nowhere else, so a reader of the README cannot yet see what the app runs on.
