# Transition: from the event cluster to one node

Decision: [ADR 0014](../adr/0014-one-node-after-the-event.md).
Issue: #130.

## Before

- DOKS cluster `cnh`, 3 × `g-8vcpu-32gb`, HA control plane, one load balancer, three 20 GB volumes.
- Argo CD syncs `deploy/app`; Argo Rollouts canary; ECK Elasticsearch with three nodes and two Kibana; Envoy Gateway; Radar; Dash0 operator.
- Month-to-date usage on 2026-11-10 at 06:09 UTC: $47.63 after 21 hours.

## After

- Droplet `wide-door-node`, `s-4vcpu-16gb-amd`, sfo3, MicroK8s from the `1.36/stable` snap channel.
- Argo CD syncs `deploy/node`; one Elasticsearch, one Kibana, one widedoor pod with a rolling update; nginx ingress with a Let's Encrypt certificate.
- Telemetry straight to Dash0.

## Runbook

Each Terraform step is `plan -out` first, Aria reads the plan, then `apply` of that plan file.

1. `make planfile-node` then `make applyfile-node`.
   Creates the droplet, firewall, SSH key and the 1Password item `Wide Door node elastic`.
   `publish_dns` is false, so the live site is untouched.
2. `make node-kubeconfig`.
   Waits for cloud-init to finish (about four minutes), writes `.kube/node.config`, prints the node.
3. `make planfile-node-platform` then `make applyfile-node-platform`.
   Namespaces, Secrets, Argo CD, the Application, the cd-gate and the `CD_KUBECONFIG` repository secret.
4. `make node-bootstrap`.
   Kibana token, ELSER download, first ingest.
5. Check over http before touching DNS:
   `curl --resolve widedoor.cnh.auroq.com:80:<ipv4> http://widedoor.cnh.auroq.com/healthz`
   and one question through `/api/ask`.
6. `make planfile-node TF_ARGS='-var publish_dns=true'` then `make applyfile-node`.
   The specific record overrides the wildcard at once; cert-manager issues the certificate within a few minutes.
7. Run the manual `acceptance` workflow and check a trace in Dash0 with the droplet's `host.name`.
8. `infra/cluster`: plan and apply the `destroy_all_associated_resources` change, then `make destroy-platform` and `make destroy-cluster`.
   Confirm with the DigitalOcean API that no cluster, load balancer or volume remains.

## Day to day

- Kubeconfig: `KUBECONFIG=.kube/node.config kubectl -n wide-door get pods`.
- Deploy: merge to `main`; `cd.yml` builds, opens the bump PR, Argo syncs, the gate waits on the Deployment and tests the live site.
- Kibana: `kubectl -n wide-door port-forward svc/wide-door-kb-http 5601` and sign in as `elastic` with the password from the vault.
- Argo: `kubectl -n argocd port-forward svc/argo-cd-argocd-server 8080:80`; the admin password is in the `argocd-initial-admin-secret` Secret.
- SSH: `make node-ssh`.

## Log

Filled in during the cutover.
