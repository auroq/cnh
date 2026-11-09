# cnh-platform

The cluster and cluster-wide operators for my Cloud Native Hackathon entry (WeMakeDevs, KubeCon NA 2026, Salt Lake City).
This repo was built **before** the hackathon as allowed setup (rules 1 and 2): it is the "venue", not the project.
The project itself lives in its own repo, created after kickoff, and links back here.

AI use: this repo was written with Claude Code (Anthropic) pairing with me.

## What it creates

| Root | State | Creates |
|---|---|---|
| `bootstrap/` | Spaces `cnh-tfstate` (seeded locally, then migrated) | DO project `cnh`, Spaces key (saved to 1Password), versioned state bucket |
| `cluster/` | Spaces | VPC, DOKS in `sfo3` (3–5 × `g-8vcpu-32gb`), DO DNS zone `cnh.auroq.com` |
| `platform/` | Spaces | Envoy Gateway + wildcard TLS (cert-manager, DNS-01), external-dns, ECK operator, Argo CD, Dash0 operator, Radar Cloud agent |

`sandbox/` holds learning-only manifests: the OpenTelemetry demo as an Argo CD app, a single-node ECK Elasticsearch + Kibana, and a local docker compose for Elasticsearch.

## Secrets

Nothing secret is in this repo or on a command line.
Terraform reads tokens from a dedicated 1Password vault (`CNH`) through the `1Password/onepassword` provider.
The service-account token for that vault lives gpg-encrypted in `pass` and reaches processes only through `secret-run`.
Expected vault items: `DigitalOcean API`, `Dash0 API`, `Radar Connect` (API Credential, field `credential`); `cnh-tfstate` is created by `bootstrap/`.

## Use

```bash
make help
make bootstrap-seed      # once
make up                  # cluster, kubeconfig, platform
make use-cnh             # make cnh the default kubectl context for the day
make teardown            # after the event
```

The Makefile pins `KUBECONFIG` to `./.kube/config`, so its targets only ever talk to this cluster.
