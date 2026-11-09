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
- **Store:** one Elasticsearch cluster (ECK) holding venue, room, access-feature and session documents, searched with BM25 plus `semantic_text` ([ADR 0004](docs/adr/0004-elasticsearch-as-the-single-store.md)).
- **Schedule:** the public KubeCon NA 2026 schedule from Sessionize, indexed with room and level.
- **Deploy:** Argo CD syncs [`deploy/`](deploy/) into the `wide-door` namespace; every workload is an Argo Rollouts `Rollout` written to pass Radar's checks ([ADR 0005](docs/adr/0005-radar-clean-manifests-via-argo.md)).
- **Telemetry:** OpenTelemetry traces, metrics and logs to Dash0 through the Dash0 operator.
- **Review:** every PR after the first commit is reviewed by Gitar against [`.gitar/review/`](.gitar/review/); its approval is advisory and the CI checks are the gate ([ADR 0010](docs/adr/0010-gitar-approval-is-advisory.md)).
- **Release:** CI and CD are separate workflows; CD builds the image, opens the bump PR, waits for the canary, tests the live site and reverts by PR on failure ([ADR 0008](docs/adr/0008-ci-cd-separation-and-rollback-by-revert.md)).
- **Types:** `pkg/model` is what the business means, adapter shapes stay private, `pkg/api/v1` is what clients were promised ([ADR 0011](docs/adr/0011-types-live-by-kind.md)).
- **Distance:** an answer says the level change and the route facts the organiser published, and never a walk time, because a single time assumes a pace and a body that the people this app is for do not share ([ADR 0012](docs/adr/0012-how-we-describe-distance.md)).
- **Secrets:** Sealed Secrets only ([ADR 0006](docs/adr/0006-sealed-secrets.md)).

### Observability

The app emits all three OpenTelemetry signals from one adapter, [`pkg/telemetry`](pkg/telemetry), over OTLP gRPC to the Dash0 operator's collector in the cluster; an empty endpoint turns every provider into a noop, which is how tests and local runs stay quiet.
Traces: every request is a span named for its route (`GET /api/ask`), with an `ask.Ask` span under it that carries the question length, where the attendee stands and what came back, and one child span per Elasticsearch call.
Metrics: `widedoor.ask.duration`, `widedoor.ask.answers` by the kind of the top place and whether it was verified, and `widedoor.ask.unverified_served`, which counts how many times we told an attendee something nobody had checked on the floor yet; the ingest job reports `widedoor.ingest.sessions_indexed`, `widedoor.ingest.detail_failures` and `widedoor.ingest.duration`.
Logs: `slog` goes through the `otelslog` bridge, so every request log reaches Dash0 as a log record with the trace and span id it belongs to.
`widedoor.ask.unverified_served` is the accessibility metric: it is the list of what to go and verify next.
The Dash0 dashboard "Wide Door" is a file too: [`deploy/dash0/wide-door.json`](deploy/dash0/wide-door.json) in the Perses format, reviewed like any other change and pushed with `make dashboard` through the Dash0 API by [`tools/dash0-dashboard.sh`](tools/dash0-dashboard.sh), with the token coming from the vault through `secret-run`.
Its panels are the accessibility story in metrics: ask latency, answers by kind and by whether anyone verified the place, "What volunteers should go verify", profile chips as needs without identities, requests by route and release, and the health of the schedule ingest.

## What we store and why

Nothing about a person.
The profile panel "What Wide Door knows about you" holds chips that are needs, not diagnoses: step-free only, avoid crowds and noise, quiet spaces matter, all-gender restrooms, captions or ASL, wide doors and ramps, dietary needs.
A chip names the barrier or the requirement, never the device or the person, so a wheelchair user, a scooter user, a parent with a stroller and a person with a walker state the same need without saying who they are.
The chip id `wheels` is historical, from before the label was renamed in item 112, and stays so that cookies and saved links set earlier keep working.
The chips live on the attendee's device only: in the `prefs` query parameter, in the first-party cookie `wd_prefs` (SameSite=Strict, Secure, expires when the conference ends on Thursday 2026-11-12) and in localStorage.
The server reads them on each request to render the panel and shape the answer, then forgets them: there is no account, no session, no profile in any index, and the request log omits `prefs`.
The only trace is the counter `widedoor.profile.chip_used` with one attribute, `chip`, so Dash0 shows how often each need is stated and never a combination that could identify someone.
"Forget me" clears all three places; "Save my profile" shows the link that carries only the chips, so the same profile can be opened on another device.
Each chip carries a source, `chosen` or `learned` (a trailing `~` in the encoding).
A question that implies a need adds the chip as learned: asking for an elevator adds step-free only, marked "from your question", announced in the live region, with its own remove button and a "Keep" button that makes it a choice.
A chip you remove goes into a second cookie, `wd_prefs_no`, so no question adds it back until you add it yourself; "Forget me" clears that too.
The counter `widedoor.profile.chip_learned`, by `chip`, is the only trace of learning.
The schedule works the same way: "Your schedule" takes a pasted Sched iCal export or a list of titles, matches them to the programme in Elasticsearch (exact title first, then the best fuzzy match with most of the words), and keeps only the matched session ids in the `sched` query parameter, the cookie `wd_sched` (same attributes and expiry) and localStorage.
With a schedule present the page says "Your next session" and "You just left", assumes you are still near the room you just left when no location is given (and says so), and offers the nearest quiet room when three sessions run back to back.
The only trace is the counter `widedoor.schedule.matched`, a count of matched sessions per paste.

## Evidence from the day

- [`docs/demo/2026-11-09-rollout-abort.md`](docs/demo/2026-11-09-rollout-abort.md): a real canary aborted by Argo Rollouts at two minutes, diagnosed by Radar with the change that caused it.
- [`docs/demo/2026-11-09-gitar-merge-gate.md`](docs/demo/2026-11-09-gitar-merge-gate.md): Gitar approving PRs on its own, withholding approval from a PR that deleted a test, and why the required review became advisory.
- The pull requests themselves: every change after Ring 0, bumps and reverts included, with Gitar's review on each.

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
Node 24 for the browser tests; `.nvmrc` pins it.

### Print the door codes

```bash
go run ./cmd/widedoor qr --out qr --sheet
```

This writes one `<place id>.svg` and `<place id>.png` per room, entrance, landmark, quiet room, prayer room, nursing room, first-aid point, info desk and coat check, each encoding `https://widedoor.cnh.auroq.com/?from=<place id>`, and `sheet.html` with one printable page per place.
Open `qr/sheet.html` in a browser and print it; each page has the place name at 48pt, the building and level, the code and the URL in text, so a volunteer can tape it beside the door sign.
Scanning a code opens Wide Door with the location filled in and the cursor in the question field.
Pass `--rooms=false` to skip fetching the schedule and render only the curated places, and `--qr-base-url` to encode another address.

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
