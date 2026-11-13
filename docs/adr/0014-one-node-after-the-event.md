# ADR 0014: One MicroK8s node after the event

Date: 2026-11-10
Status: Accepted; superseded by the 2026-11-13 takedown, when the node was destroyed
Issue: #130

## Context

The hackathon ended on 2026-11-09 at 17:30 MT.
The live site stays up for the week after, for bug fixes and for anyone who reads the README.
The event cluster was three `g-8vcpu-32gb` nodes with an HA control plane, a load balancer and three block volumes, about $29 a day at list price.
Node CPU sat at 2 percent and Elasticsearch held 28 MB of data across three 8 GiB nodes.
The app is one static Go binary plus one Elasticsearch.

## Decision

Run the app on one DigitalOcean droplet (`s-4vcpu-16gb-amd`, about $2.80 a day) with MicroK8s, and destroy the DOKS cluster with its load balancer and volumes.

Aria's calls:

- MicroK8s from the Ubuntu snap, not k3s.
  Her rule is that anything installed on a system goes through a package manager, and k3s only ships a `curl | sh` installer.
- Deploys stay automated.
  The image bump PR flow is unchanged, Argo CD syncs the node, and the CD gate waits on a Deployment instead of a Rollout before running Playwright and acceptance against the live site.
- Kibana stays at one replica.
- Argo CD stays, with dex, notifications and the ApplicationSet controller off.

## How

- `infra/node` creates the droplet, its firewall, the SSH key, the elastic password (generated and stored in the CNH vault) and, once `publish_dns` is true, the `widedoor` A record.
  Cloud-init installs MicroK8s from the pinned snap channel, adds the public IP to the API server certificate and enables `dns`, `hostpath-storage`, `ingress`, `cert-manager`, `metrics-server` and `helm3`.
- `infra/node-platform` talks to the node through the fetched admin kubeconfig and creates the namespaces, the two Secrets (`wide-door-es-elastic-user`, `dash0-otlp`), Argo CD, the Application for `deploy/node`, and the cd-gate ServiceAccount with the `CD_KUBECONFIG` repository secret.
- `deploy/node` is a kustomize overlay on `deploy/app`.
  It deletes the ECK, Rollouts, Gateway API, Dash0 operator and PodDisruptionBudget resources and adds a single-node Elasticsearch StatefulSet, a Kibana Deployment, a widedoor Deployment, an Ingress and a ClusterIssuer.
  Service and Secret names match what ECK used, so `config.yaml`, `widedoor.yaml` and `ingest.yaml` are reused unchanged.
- Telemetry goes straight from the app to the Dash0 ingress over OTLP/gRPC with the token in `OTEL_EXPORTER_OTLP_HEADERS`.
  `pkg/telemetry` already builds the exporters with `WithEndpointURL` for an `https://` endpoint, and the OpenTelemetry SDK reads headers from the environment, so no Go changed.

## What is dropped for the week

- Radar's cluster audit, the Argo Rollouts canary, ECK, Envoy Gateway, Sealed Secrets, the Dash0 operator and multi-node scheduling.
  The repository, the ADRs and the merged PRs remain the record of how they were used.
- Node redundancy.
  One droplet is a single point of failure; DigitalOcean monitoring is on and the ingest CronJob rebuilds the indices from YAML and Sessionize every hour, so nothing on the node is irreplaceable.

## Consequences

- The Kubernetes API is reachable on port 16443 from anywhere, because the CD gate on GitHub Actions needs it.
  Access is by token only, the cd-gate token can read deployments, replicasets and pods in one namespace, and SSH is key-only.
  Rotate with `make apply-node-platform`.
- The review rule that app workloads are Rollouts with two replicas and a PodDisruptionBudget does not apply under `deploy/node`; `.gitar/review/k8s.md` records the exception.
- The Elasticsearch trial licence on the node runs 30 days from its first start; ELSER stops when it expires.
- Going back to the event shape is `make up` after importing the domain and the VPC into `infra/cluster`, because the cutover detached them from its state so its destroy could not take the site down; the transition doc has the two import commands.
