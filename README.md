# Wide Door

Ask the venue where the nearest step-free route, elevator, quiet room, all-gender restroom or captioned session is, and get an answer that works with a screen reader, a keyboard and large text.

Team **Curb Cut** (one person, [auroq](https://github.com/auroq)) at the WeMakeDevs [Cloud Native Hackathon](https://www.wemakedevs.org/hackathons/cnh), Salt Lake City, 2026-11-09, alongside KubeCon + CloudNativeCon North America.
The name comes from the curb-cut effect: access built for disabled people ends up used by everyone.

- Plan and status: [`ROADMAP.md`](ROADMAP.md), with one file per item in [`roadmap/`](roadmap/) and one GitHub issue per item.
- Decisions: [`docs/adr/`](docs/adr/).
- AI disclosure: [`AI_USAGE.md`](AI_USAGE.md).

## Before and during the hackathon

Everything up to the git tag [`pre-kickoff`](https://github.com/auroq/cnh/releases/tag/pre-kickoff) is setup done before the event, which rules 1 and 2 allow: the cluster and the cluster-wide operators in [`infra/`](infra/).
It is the venue, not the project.
Project code and design start after kickoff (10:00 MT, 2026-11-09) in the commits after that tag, so `git log pre-kickoff..` is exactly the hackathon work.
[`docs/transitions/pre-kickoff-platform.md`](docs/transitions/pre-kickoff-platform.md) lists what was installed before.

## Architecture

[`docs/architecture.md`](docs/architecture.md) is the one-page tour: the layers, the ports, what each package and binary is for, and where to look when you want to change something.
The shape is fixed by the ADRs:

- **Language:** Go for the API and the ingest, with server-rendered HTML and almost no JavaScript, because plain HTML is the most accessible UI there is.
- **Store:** one Elasticsearch cluster (ECK) holding venue, room, access-feature and session documents, searched with BM25 plus `semantic_text` ([ADR 0004](docs/adr/0004-elasticsearch-is-the-store.md)).
- **Schedule:** the public KubeCon NA 2026 schedule from Sessionize, indexed with room and level.
- **Deploy:** Argo CD syncs [`deploy/`](deploy/) into the `wide-door` namespace; every workload is an Argo Rollouts `Rollout` written to pass Radar's checks ([ADR 0005](docs/adr/0005-radar-clean-manifests-via-argo.md)).
- **Telemetry:** OpenTelemetry traces, metrics and logs to Dash0 through the Dash0 operator.
- **Review:** every PR after the first commit is reviewed by Gitar against [`.gitar/review/`](.gitar/review/).
- **Secrets:** Sealed Secrets only ([ADR 0006](docs/adr/0006-sealed-secrets.md)).

### Observability

The app emits all three OpenTelemetry signals from one adapter, [`pkg/telemetry`](pkg/telemetry), over OTLP gRPC to the Dash0 operator's collector in the cluster; an empty endpoint turns every provider into a noop, which is how tests and local runs stay quiet.
Traces: every request is a span named for its route (`GET /api/ask`), with an `ask.Ask` span under it that carries the question length, where the attendee stands and what came back, and one child span per Elasticsearch call.
Metrics: `widedoor.ask.duration`, `widedoor.ask.answers` by the kind of the top place and whether it was verified, and `widedoor.ask.unverified_served`, which counts how many times we told an attendee something nobody had checked on the floor yet; the ingest job reports `widedoor.ingest.sessions_indexed`, `widedoor.ingest.detail_failures` and `widedoor.ingest.duration`.
Logs: `slog` goes through the `otelslog` bridge, so every request log reaches Dash0 as a log record with the trace and span id it belongs to.
`widedoor.ask.unverified_served` is the accessibility metric: it is the list of what to go and verify next.

## Tracks

| track | what Wide Door does for it |
| --- | --- |
| Merge Forward, Conferences for Everyone | the whole point: venue access answers for attendees with limited mobility, hearing or vision |
| Elasticsearch, Find the Right Talk | hybrid search over sessions and venue facts, including captioned and accessible sessions |
| Dash0, Stop Flying Blind | traces, metrics and logs from the API and the ingest, plus custom accessibility metrics |
| Radar, AI SRE | a namespace with zero findings from the first commit, and a canary to investigate |
| Gitar / Sonar, Never Merge Broken | review rules for manifests, Go and accessibility, with auto-apply fixing what it can |

## Run it

```bash
make help
```

Infrastructure targets are described in [`infra/`](infra/); the app targets are added as the rings land.
The Makefile pins `KUBECONFIG` to `./.kube/config`, so its targets only ever talk to this cluster.

## Infrastructure (`infra/`)

| Root | State | Creates |
| --- | --- | --- |
| `infra/bootstrap/` | Spaces `cnh-tfstate` (seeded locally, then migrated) | DO project `cnh`, Spaces key (saved to 1Password), versioned state bucket |
| `infra/cluster/` | Spaces | VPC, DOKS in `sfo3` (3 to 5 × `g-8vcpu-32gb`), DO DNS zone `cnh.auroq.com` |
| `infra/platform/` | Spaces | Envoy Gateway + wildcard TLS (cert-manager, DNS-01) with a wildcard DNS record, ECK operator, Argo CD, Argo Rollouts, Dash0 operator, Radar Cloud agent, metrics-server, Sealed Secrets |

`infra/sandbox/` holds learning-only manifests used before the event and removed from the cluster before kickoff.

### Secrets

Nothing secret is in this repo or on a command line.
Terraform reads tokens from a dedicated 1Password vault (`CNH`) through the `1Password/onepassword` provider.
The service-account token for that vault lives gpg-encrypted in `pass` and reaches processes only through `secret-run`.
In the cluster, application secrets are `SealedSecret` resources; the controller's key is backed up in the same vault.
