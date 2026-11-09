# AI usage

This project was built by one person, Aria (`auroq`), pairing with Claude Code (Anthropic, model Claude Fable 5.1) for the whole day.
From about 13:00 MT several Claude Code sessions ran at once, each in its own git worktree: an integrator (infrastructure, Argo CD, CI and CD, roadmap bookkeeping), Aria's own pairing session (the testing rings, the ingest fix), a telemetry session (item 011, Dash0) and a coordinator that relayed Aria's instructions between them.
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

- Chose the problem, the option to build, the app and team names, the tracking method, the merge policy below, and every decision marked "Aria's call" in a roadmap item.
- Reviewed and approved every Terraform plan before it was applied.
- Created the sponsor accounts, the GitHub repositories and the 1Password vault, installed the Gitar app, and ran the commands the agent is not allowed to run.
- Judged what was good enough to ship at each ring, and what to demo.
- Read the merged diffs and asked questions as PR comments; the answers are public on the pull requests.
- Turned Gitar's auto-apply off after reading what it had pushed, turned its auto-approve on, chose the bump-PR shape for CD over bot commits to `main`, and ran the branch-protection, token and dashboard steps that need a human.
- Wrote the testing rings (`testing/`), the Elasticsearch compose setup and the acceptance tests, and is writing the README architecture section and the demo in her own words.

## Merge policy

After the first commit, every change goes through a pull request.
`main` is protected: a PR merges only when the `docs`, `manifests`, `architecture` and `Gitar` CI checks pass and Gitar has posted an approving review against [`.gitar/config/approve.md`](.gitar/config/approve.md) ([ADR 0009](docs/adr/0009-gitar-approval-is-the-merge-gate.md)).
Until 15:00 MT the `Gitar` status check alone was required, and it reported SUCCESS even when Gitar requested changes (PR 18), so `main` was merging on CI alone; the approving review is what makes Gitar's verdict the gate.
The agent opens the PR and arms GitHub auto-merge, so a PR merges without a human reading it, on the strength of CI and Gitar's approval.
A PR Gitar does not approve stays blocked until its findings are answered or Aria approves it by hand.
Aria chose this on 2026-11-09 so that one person could spend the day building rather than clicking merge, and accepted that Gitar's review and the CI checks are the gate.
Gitar's auto-apply may also push fixes to a PR branch; those commits are Gitar's and are visible as such.

## How to see it

- Every commit the AI authored carries a `Co-Authored-By: Claude Fable 5.1` trailer and a `Claude-Session` link.
- PR reviews from Gitar are a second AI reading the first one's work; both are visible on each pull request.
- The git tag `pre-kickoff` marks the end of setup done before the event, so `git log pre-kickoff..` is the hackathon work.

## Other services

- Gitar reviews pull requests and, since 15:00 MT, approves them against [`.gitar/config/approve.md`](.gitar/config/approve.md); its auto-apply is off and it changes code only when asked (sponsor track).
- Dash0 and Radar analyse telemetry and cluster state with their own AI features (sponsor tracks).
- Elasticsearch's `semantic_text` field uses Elastic's ELSER model for semantic search (sponsor track).
