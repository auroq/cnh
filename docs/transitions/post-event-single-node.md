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

All times UTC.

- 2026-11-11 02:27: `infra/node` applied; droplet `wide-door-node` at 164.92.84.236.
- 02:29: cloud-init's bootstrap was killed while snapd restarted, then failed waiting on an nginx DaemonSet; MicroK8s 1.36's ingress addon is Traefik. Rerun by hand; fixed in PR #132 (wait on Traefik, retry three times, `ignore_changes = [user_data]`).
- 02:44: `infra/node-platform` applied (12 resources); Argo CD synced `deploy/node`.
- 02:49: `make node-bootstrap`: Kibana token, ELSER ready, first ingest indexed 602 sessions, 77 places and 8 needs.
- 02:52: answers over http on the IP matched the old cluster.
- 02:53: `publish_dns=true` applied; the specific A record took over from the wildcard within about four minutes.
- 03:00 to 03:14: no certificate, because the HTTP-01 solver Ingress had no class: the ClusterIssuer set the old `class` field and Traefik only honours `ingressClassName`. A hand-applied fix was reverted by Argo's self-heal, as it should be. Fixed in PR #133, then the order was restarted; Let's Encrypt issued `widedoor-tls` at 03:14.
- 03:20: acceptance against the node and against the old cluster through port-forwards failed the same four cases, so the move changed no answers. The stale expectations are tracked separately.
- 03:25: state detached so the destroys could not take the site with them: `digitalocean_domain.cnh`, `digitalocean_vpc.cnh` and `digitalocean_project_resources.cnh` from `infra/cluster`, and `github_actions_secret.cd_kubeconfig` from `infra/platform` (it now holds the node's cd-gate kubeconfig, owned by `infra/node-platform`).
- 03:27: the old Argo Application deleted with cascade so ECK's three volumes were released; `infra/platform` destroyed (26 of 27; `metrics-server` hung uninstalling and was dropped from state); `infra/cluster` destroyed.
- 03:30: the DigitalOcean API shows no cluster, no load balancer and no volumes; the domain, the VPC, the node and the `widedoor` record remain.

## Going back to the event shape

`infra/cluster` no longer owns the domain or the VPC, so `make up` would try to create both and fail on the name.
Import them first: `terraform -chdir=infra/cluster import digitalocean_domain.cnh cnh.auroq.com` and `terraform -chdir=infra/cluster import digitalocean_vpc.cnh <vpc id>` (through the Makefile's `$(TF)` wrapper for credentials).
