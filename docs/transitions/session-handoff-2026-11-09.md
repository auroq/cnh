# Session hand-off, 2026-11-09 11:45 MT

**Transitional. Delete when the hackathon is submitted (17:30 MT today).**
Written by the integrator Claude session that ran from the night before through Ring 2, for the session that replaces it.
Read it whole before doing anything, then read `ROADMAP.md`, `docs/adr/`, `AI_USAGE.md` and the open PRs.

## The event

- WeMakeDevs **Cloud Native Hackathon**, Monday 2026-11-09, The Shop Workspace, Salt Lake City, an offsite of KubeCon + CloudNativeCon North America (Salt Palace + Hyatt Regency, 8 to 12 November).
- Timeline (Mountain Time): kickoff 10:00, building 11:00 to 17:30, submission 17:30, then judging and demos.
- Participant: Aria, GitHub `auroq`, solo, team name **Curb Cut**. Staff engineer, backend and infra heavy, polyglot, prefers TDD and pairing.
- Rules that bind us: (1) main coding and design only after kickoff; setup and research before are fine. (2) templates and third-party tools allowed; only hackathon-time work is judged. (3) public repo with a README that explains the architecture. (4) one project per team. (5) AI assistants allowed but **must be disclosed**. (6) the participant **must understand the code and be able to explain the architecture and the decisions**. (7) projects generated entirely by AI without meaningful participant contribution, verification or understanding **may be rejected**.
- Tracks, USD 4,000 each: **Gitar / Sonar "Never Merge Broken"** (PR review, rules, auto-apply, fixing CI); **Dash0 "Stop Flying Blind"** (OpenTelemetry traces, metrics, logs); **Radar "AI SRE"** (cluster audit and investigation); **Elasticsearch "Find the Right Talk"** (search over the conference programme); **Merge Forward "Conferences for Everyone"** (accessibility and inclusion; prize an iPad). **Best overall** wins an NVIDIA DGX Spark. We are entering all of them.
- Links: https://www.wemakedevs.org/hackathons/cnh (overview), `/resources`, `/rules`.

## The project

**Wide Door**: ask the venue where the nearest step-free route, elevator, quiet room, all-gender restroom or captioned session is, and get an answer that works with a screen reader, a keyboard and large text, and says whether anyone has verified it.
Built from overnight deep research on accessibility and inclusion at tech conferences (disabled, Deaf, blind, neurodivergent, chronically ill, queer, trans and BIPOC first-person accounts, and the solutions those communities built).
The research doc with the ranked options is https://claude.ai/artifact/MCdgPxoC1Ba751AfpgJF3V and the markdown report is at `~/Documents/cnh-research/reports/Conference accessibility and inclusion options.md` with notes beside it.
Its finding: every group reports the same failure, the information exists but nobody can look it up from where they stand; the organiser publishes room lists as static text with no level, route or access data.

Aria chose **Option 1 "Ask the venue"**; **Option 2 "Find it in the talk"** (timestamped transcript search, live-mic captions as default source) is a stretch for 16:00 to 17:00 only if Ring 4 is done; Option 3 "Change watch" was set aside.

Facts to keep straight:

- KubeCon NA 2026 schedule is public as markdown: `https://sessionize.com/api/v2/uo948i5p/view/markdown` (600 sessions, 712 speakers, `Venue | Level | Room` on each row, per-session pages at `/view/markdown/session/{id}`).
- Organiser's inclusion page lists: quiet rooms Salt Palace L2 260A (all) and 260B (women and non-binary), prayer room Hyatt L5 Zion, a nursing room (ask registration), six all-gender restrooms (SP L1 Solutions Showcase Amber Hive Zone, SP L2 outside 255, Hyatt L1 lobby, L2 by Juniper, L3 by Powder Mountain, L4 by Broadcast Lounge), Mobility City wheelchair rental, reserved seating, Wordly AI captions, pronoun and red/yellow/green stickers, Community Hub with Merge Forward in Grand Ballroom J. Elevator positions are **not** published ("ask staff"); OpenStreetMap has no wheelchair data for the venue. So elevators and the step-free entrance are seeded `verified: false`.
- Utah HB 257 does not restrict standalone restrooms (ACLU reading); don't repeat the LF page's "privately operated" line as fact.
- Wildcard cert and DNS are `*.cnh.auroq.com`, one level only: `kibana.cnh.auroq.com` works, `kibana.widedoor.cnh.auroq.com` would not.

## Decisions already made (do not reopen)

- Monorepo `auroq/cnh`, public; `infra/` predates the event and is tagged **`pre-kickoff`** (`git log pre-kickoff..` is the hackathon work). ADR 0001.
- Language: Go, server-rendered HTML, almost no JavaScript. One Elasticsearch cluster is the only store (ADR 0004). Manifests written to pass Radar from the first commit, deployed by Argo CD + Argo Rollouts (ADR 0005). Sealed Secrets only, never plain secrets with data (ADR 0006). AI disclosure in `AI_USAGE.md` with commit trailers (ADR 0007).
- Work tracking: **GitHub Issues, one per roadmap item, issue number == item number** (#1 to #16 == 001 to 016), created by `tools/issues.sh`; `ROADMAP.md` is the index, `roadmap/NNN-*.md` the detail, `docs/adr/` the decisions, `docs/transitions/` temporary state. ADR 0003, accepted "so things are as visible for the judges as possible".
- Merge policy: `main` protected with required checks `docs`, `manifests`, `Gitar`; auto-merge enabled; **the agent opens PRs and arms `gh pr merge --auto --squash`**; nobody reads them before merge, Gitar and CI are the gate. Aria chose this and `AI_USAGE.md` says so.
- **PRs open green.** Red commits may exist inside a PR, the head must pass before the PR is opened.
- Gitar has no repo-wide auto-apply or verbosity switch (docs checked); so every PR gets the label `gitar-managed` (auto-apply on) and a comment `gitar display:verbose`. Write `tools/pr.sh` to do label + comment + auto-merge in one go; it did not exist at hand-off.
- Radar: Elasticsearch and Kibana run **without** `readOnlyRootFilesystem` on purpose; Aria chose to present that finding in the Radar story rather than hide it. Everything else in `wide-door` passes.
- Team name Curb Cut, app name Wide Door; repo stays `auroq/cnh`.

## Collaboration protocol (Aria's requirement, rules 5 to 7)

Aria wants agents autonomous but needs real hands-on work for the whole day and must be able to explain everything.

- **Issue assignee is the owner.** Assigned to Aria: do not touch without asking. Unassigned: yours.
- Keep PRs small enough to read in five minutes. Aria reads merged diffs on GitHub and asks questions as PR comments; **answer them there**, publicly.
- Aria may run her own Claude session in a git worktree for her item; this session is the integrator (infra, Argo, docs, demo plumbing, bookkeeping).
- Her queue as proposed: (a) Gitar's finding on PR 18, a failed detail fetch overwrites a session's description with empty text, fix with a table-driven test in a follow-up PR; (b) item 011 OpenTelemetry to Dash0 in `cmd/widedoor` and the ingest (traces, latency and answers-by-kind and unverified-answers metrics, slog bridge; OTLP to the Dash0 operator collector service in `dash0-system`); (c) item 010 Orca screen-reader and keyboard-only pass on the deployed page, fixes in template and CSS; (d) item 008 floor-plan data into `data/venue/places.yaml` as `verified: false`; (e) item 013 README architecture and demo script in her own words.
- Only Aria can: make GHCR packages public after the first `images` workflow run (github.com/auroq?tab=packages); run `gh repo create`-class commands; install GitHub apps.

## Infrastructure and how to operate it

- DigitalOcean DOKS `cnh` in `sfo3`, 3 to 5 × `g-8vcpu-32gb`, Kubernetes 1.37. Kubeconfig at `.kube/config` in the repo (gitignored); the Makefile exports `KUBECONFIG` to it; `cnh` is also the current kubectx default (switch back with `make use-wasp` after the event).
- Terraform roots `infra/bootstrap` (Spaces state bucket + key), `infra/cluster`, `infra/platform`; S3 backend on DO Spaces. **Always `plan -out=<file>` then `apply <file>`**; the auto-mode classifier blocks `apply -auto-approve` and plan-and-apply in one command. The Makefile `plan-%`/`apply-%` targets wrap credentials; to pass extra flags use `eval "$(make -s -n plan-platform | sed 's| plan$| plan -no-color -out=…|')"`.
- Secrets: 1Password vault **CNH** via a service-account token kept gpg-encrypted in `pass` at `cnh/op-service-account-token`. Pattern: `secret-run --env OP_SERVICE_ACCOUNT_TOKEN=pass:cnh/op-service-account-token -- secret-run --env X=op:CNH/<item>/<field> -- <cmd>`. Use item IDs when a title has `:`. Vault items: `Digital Ocean API CNH`, `Dash0 API Token CNH`, `Radar API CNH` (PAT), `Radar Cluster Connect CNH`, `cnh-tfstate`, `Argo CD admin`, `Sealed Secrets controller key`. Never print secret values; show labels and lengths only.
- Platform (all Helm via Terraform, versions pinned in `infra/platform/variables.tf`): Envoy Gateway + Gateway API (Gateway `public` in ns `gateway`, listeners `*.cnh.auroq.com`, routes allowed from all namespaces), cert-manager DNS-01 on DO, wildcard A record from Terraform (external-dns 0.23 dropped the DO provider), ECK operator (+ data-less `eck-trial-license` Secret, Enterprise trial), Argo CD at https://argocd.cnh.auroq.com (admin password in vault), Argo Rollouts, Dash0 operator (us-west-2, dataset default), Radar Cloud agent (cluster id `WPcuS8CDZUo`, https://app.radarhq.io/checks?scope=WPcuS8CDZUo), metrics-server, Sealed Secrets (`kubeseal` 0.35), Hubble from DOKS Cilium feeding Radar live traffic. Platform charts got resources, 2 replicas and PDBs this morning so Radar's platform findings dropped; remaining platform findings are chart limitations.
- Helm upgrades through the Terraform helm provider take 7 to 15 minutes each on this cluster; run them in the background and don't assume a hang before 20 minutes.
- Radar's public REST API has clusters but no checks endpoint; read the audit in the UI or through the Radar MCP server. Radar MCP: `https://api.radarhq.io/mcp`; Dash0 MCP: `https://api.us-west-2.aws.dash0.com/mcp` (OAuth). Both are in `.mcp.json`, which is why this session was restarted from the repo directory.
- Demo data from the night (OTel demo, sandbox ES) was deleted before kickoff; `infra/sandbox/` stays as reference.

## What is live right now

- Argo CD Application `wide-door` syncs `deploy/app` (kustomize) into ns `wide-door`: placeholder Rollout `web` (http-echo) at https://widedoor.cnh.auroq.com, Elasticsearch `wide-door` (3 nodes, 8Gi, ML on, TLS off inside the cluster, green), Kibana (2) at https://kibana.cnh.auroq.com, PDBs, HTTPRoutes, ingest CronJob (image not yet built). The ES and Kibana manifests were applied by hand from the branch and will be adopted by Argo when PR 18 merges.
- Indexes: `sessions` (600 docs, `content` is `semantic_text`, ELSER works) and `places` (18 curated + 34 rooms derived from the schedule). Populated by `make ingest` from the laptop (port-forward). ES password: secret `wide-door-es-elastic-user` key `elastic` in ns `wide-door`.
- Verified by hand: from 255 BC, "nearest all-gender restroom" answers the one outside 255 on the same level; from Grand Ballroom J, "quiet room" answers 260A one level up; "sign language or captions" finds the Sign Language Crash Course.

## Open PRs and the one puzzle

- **PR 18** `ring-1/elasticsearch` → `main`: ES, Kibana, ingest, Go CI job, Dockerfile, images workflow. Gitar approved after the matrix fix. **Blocked**: the `ci` runs on this branch report failure but list only `go` and `manifests`; the `docs` job never appears, so the required `docs` check never reports. The same workflow passes on `main` (run on d038832). Diagnose with `gh run view <id>` and `gh api repos/auroq/cnh/actions/runs/<id>/jobs`, fix on the branch, and it auto-merges.
- **PR 19** `ring-2/ask` → `ring-1/elasticsearch`: `data/venue/places.yaml`, `pkg/venue`, `pkg/ask`, `cmd/widedoor`, ingest indexes places, image ships `data/`. Retarget to `main` after 18 merges, then arm auto-merge.
- Gitar's open findings on 18 worth acting on: the description-overwrite bug (Aria's task a), and `fail-fast`/matrix (fixed).

## Next steps for the integrator, in order

1. Unblock PR 18; retarget and arm PR 19.
2. After the `images` workflow runs, ask Aria to make the GHCR packages public; then add `deploy/app/widedoor.yaml` (Rollout for `ghcr.io/auroq/cnh/widedoor`, same hardening as `web.yaml`, Service, HTTPRoute `widedoor.cnh.auroq.com`) and remove the http-echo placeholder; run the ingest CronJob once (`kubectl create job --from=cronjob/ingest ingest-now -n wide-door`) and close #7.
3. `tools/pr.sh`; mark items 008, 009, 010 Done when 19 merges (status line 3, drop from `ROADMAP.md` open table, close issues).
4. Item 012: one deliberate canary workload for the Radar demo, and verify `wide-door` shows only the accepted read-only-filesystem findings.
5. Dash0 dashboard once Aria's OTel lands; Radar and Dash0 MCP queries for the demo.
6. Ring 5 by 17:00: code freeze, README architecture (Aria writes, you fact-check), `AI_USAGE.md` current, delete this file, submit.

## Conventions in this repo

- Markdown: one sentence per line; no em-dashes; status on line 3 of each roadmap item; `ROADMAP.md` open table lists exactly the Open items (CI `docs` job enforces).
- Go: `cmd/<bin>/main.go` + `pkg/`; tests table-driven and named "when …, and …, it should …", one assertion each; no comments unless asked; every script has `--help`.
- Commits end with `Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>` and `Claude-Session: <link>`; PR bodies end with the Claude Code footer. Remote is `github`, never `origin`.
- Nothing installed outside pacman/AUR; nothing secret on a command line or in the transcript.
