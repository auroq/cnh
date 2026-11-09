# cnh

My Cloud Native Hackathon entry (WeMakeDevs, KubeCon NA 2026, Salt Lake City).

## Before vs during the hackathon

Everything up to the git tag **`pre-kickoff`** is setup done before the event, which rules 1 and 2 allow:
the cluster and cluster-wide operators in `infra/`. It is the "venue", not the project.
All project code and design starts after kickoff (10:00 MT, 2026-11-09), in commits after that tag.
`git log pre-kickoff..` shows exactly the hackathon work.

## Infrastructure (`infra/`)

AI use: this repo was written with Claude Code (Anthropic) pairing with me.

## What it creates

| Root | State | Creates |
|---|---|---|
| `infra/bootstrap/` | Spaces `cnh-tfstate` (seeded locally, then migrated) | DO project `cnh`, Spaces key (saved to 1Password), versioned state bucket |
| `infra/cluster/` | Spaces | VPC, DOKS in `sfo3` (3–5 × `g-8vcpu-32gb`), DO DNS zone `cnh.auroq.com` |
| `infra/platform/` | Spaces | Envoy Gateway + wildcard TLS (cert-manager, DNS-01), external-dns, ECK operator, Argo CD, Dash0 operator, Radar Cloud agent |

`infra/sandbox/` holds learning-only manifests: the OpenTelemetry demo as an Argo CD app, a single-node ECK Elasticsearch + Kibana, and a local docker compose for Elasticsearch.

## Secrets

Nothing secret is in this repo or on a command line.
Terraform reads tokens from a dedicated 1Password vault (`CNH`) through the `1Password/onepassword` provider.
The service-account token for that vault lives gpg-encrypted in `pass` and reaches processes only through `secret-run`.
Expected vault items: `Digital Ocean API CNH`, `Dash0 API Token CNH`, `Radar API CNH` (API Credential, field `credential`); `cnh-tfstate` is created by `bootstrap/`.

## Use

```bash
make help
make bootstrap-seed      # once
make up                  # cluster, kubeconfig, platform
make use-cnh             # make cnh the default kubectl context for the day
make teardown            # after the event
```

The Makefile pins `KUBECONFIG` to `./.kube/config`, so its targets only ever talk to this cluster.
