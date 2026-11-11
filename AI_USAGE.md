# AI usage

This project was built by one person, Aria (`auroq`), pairing with Claude Code (Anthropic, model Claude Fable 5.1) for the whole day.
From about 13:00 MT several Claude Code sessions ran at once, each in its own git worktree: an integrator (infrastructure, Argo CD, CI and CD, roadmap bookkeeping), Aria's own pairing session (the testing rings, the ingest fix), a telemetry session (item 011, Dash0) and a coordinator that relayed Aria's instructions between them.
Later in the afternoon the shape was an integrator session, an advisor and dispatcher session that turned Aria's instructions into issues and handed them out, one worker agent per issue on its own `issue/N-*` branch (20 issue branches: 54, 55, 63, 66, 67, 68, 69, 72, 75, 79, 81, 83, 85, 89, 96, 103, 112, 114, 118, 121), research subagents for web lookups, and Claude in Chrome for pages that refuse a non-browser fetch, such as the Salt Palace floor plan in [`docs/venue-sources.md`](docs/venue-sources.md).
`git log pre-kickoff..` carries 87 distinct `Claude-Session` links.
On 2026-11-10, after the event, the advisor session and one worker built the single-node setup in [ADR 0014](docs/adr/0014-one-node-after-the-event.md) from Aria's four decisions, with every Terraform plan read by her before it was applied.
Aria directed all of them; which session wrote a change is visible from the `Claude-Session` trailer on its commits.
This file is the disclosure the hackathon rules ask for, and [ADR 0007](docs/adr/0007-ai-disclosure.md) is the decision behind it.

## What the AI did

- Wrote most of the code, Kubernetes manifests, Terraform and Markdown in this repository, from instructions given in conversation.
- Ran the tooling: Terraform, kubectl, Helm, gh, the Elasticsearch API, and the sponsors' MCP servers listed in [`.mcp.json`](.mcp.json).
- Did the overnight research behind the project choice, with web search, and wrote it up as a document with ranked options.
- Drafted the roadmap items and ADRs in [`roadmap/`](roadmap/) and [`docs/adr/`](docs/adr/).
- Built the delivery pipeline: CI, the CD workflow that opens image bump pull requests, waits for the Rollout and tests the live site, and the rollback-by-revert ([ADR 0008](docs/adr/0008-ci-cd-separation-and-rollback-by-revert.md)).
- Reviewed Gitar's auto-applied commits, reverted two that reversed written-down decisions, and wrote the review rules and approval criteria Gitar now works from ([`.gitar/`](.gitar/)).

## What the human did

- Chose the problem, the option to build, the app and team names, the tracking method, the merge policy below, and every decision marked "Aria's call" in a roadmap item: 002 (the names), 008 (what is seeded as unverified), 014 to 016 (the stretch is a different product), 032 (the config shape), 036 (CD as its own stage), 043 (the Gitar gate), 050 (whether to split config), 068 (what is verified and how it is worded), 079 (routes are the next project), 081 (the chip list and wording), 085 (the wording of what is learned), 096 (the trust rules) and 112 (chips name needs, not people).
- Filed the issues behind items 025, 031, 032, 036, 043, 050, 054, 055, 057, 063, 066, 067, 068, 069, 072, 075, 079, 081, 083, 085, 089, 096 and 118 (each says "by Aria" on line 3), and set the layout or direction recorded on 103, 114, 118 and 121.
- Reviewed and approved every Terraform plan before it was applied.
- Created the sponsor accounts, the GitHub repositories and the 1Password vault, installed the Gitar app, and ran the commands the agent is not allowed to run.
- Judged what was good enough to ship at each ring, and what to demo.
- Read the merged diffs and asked questions as PR comments; the answers are public on the pull requests.
- Turned Gitar's auto-apply off after reading what it had pushed, turned its auto-approve on, chose the bump-PR shape for CD over bot commits to `main`, and ran the branch-protection, token and dashboard steps that need a human.
- Wrote the testing rings (`testing/`), the Elasticsearch compose setup and the acceptance tests, and is writing the README architecture section and the demo in her own words.

## Merge policy

After the first commit, every change goes through a pull request.
`main` is protected: a PR merges only when the `docs`, `manifests`, `architecture`, `integration` and `Gitar` CI checks pass and Gitar has not requested changes.
Since 15:25 MT Gitar posts a "request changes" review on any finding it rates important or higher, and GitHub refuses the merge until Gitar dismisses it because the finding is fixed or Aria comments `gitar unblock` ([ADR 0013](docs/adr/0013-gitar-blocks-on-important-findings.md)).
Gitar also posts an approving review when [`.gitar/config/approve.md`](.gitar/config/approve.md) holds; that review is advisory and never required, because on a repository with one GitHub author nobody can approve their own PR and Gitar cannot read PR bodies here ([ADR 0009](docs/adr/0009-gitar-approval-is-the-merge-gate.md) tried it between 13:40 and 13:47 MT, [ADR 0010](docs/adr/0010-gitar-approval-is-advisory.md) records why it was withdrawn).
Between 13:47 and 15:25 MT, under the advisory policy, nine PRs merged with Gitar findings still open (73, 76, 78, 80, 82, 86, 88, 90, 93); that is said here rather than hidden.
The agent opens the PR and arms GitHub auto-merge, so a PR merges without a human reading it, on the strength of CI and the absence of a blocking finding.
Aria chose this on 2026-11-09 so that one person could spend the day building rather than clicking merge, and accepted that Gitar's review and the CI checks are the gate.
Gitar's auto-apply may also push fixes to a PR branch; those commits are Gitar's and are visible as such.

## How to see it

- Every commit the AI authored carries a `Co-Authored-By: Claude Fable 5.1` trailer and a `Claude-Session` link.
- PR reviews from Gitar are a second AI reading the first one's work; both are visible on each pull request.
- The git tag `pre-kickoff` marks the end of setup done before the event, so `git log pre-kickoff..` is the hackathon work.

## Other services

- Gitar reviews pull requests, approves them against [`.gitar/config/approve.md`](.gitar/config/approve.md) as an advisory review, and since 15:25 MT blocks the merge on important or critical findings ([ADR 0013](docs/adr/0013-gitar-blocks-on-important-findings.md)); its auto-apply is off and it changes code only when asked (sponsor track).
- Dash0 and Radar analyse telemetry and cluster state with their own AI features (sponsor tracks).
- Elasticsearch's `semantic_text` field uses Elastic's ELSER model for semantic search (sponsor track).
