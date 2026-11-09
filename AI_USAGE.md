# AI usage

This project was built by one person, Aria (`auroq`), pairing with Claude Code (Anthropic, model Claude Fable 5.1) for the whole day.
This file is the disclosure the hackathon rules ask for, and [ADR 0007](docs/adr/0007-ai-disclosure.md) is the decision behind it.

## What the AI did

- Wrote most of the code, Kubernetes manifests, Terraform and Markdown in this repository, from instructions given in conversation.
- Ran the tooling: Terraform, kubectl, Helm, gh, the Elasticsearch API, and the sponsors' MCP servers listed in [`.mcp.json`](.mcp.json).
- Did the overnight research behind the project choice, with web search, and wrote it up as a document with ranked options.
- Drafted the roadmap items and ADRs in [`roadmap/`](roadmap/) and [`docs/adr/`](docs/adr/).

## What the human did

- Chose the problem, the option to build, the app and team names, the tracking method, and every decision marked "Aria's call" in a roadmap item.
- Reviewed and approved every Terraform plan before it was applied, and every PR before it merged.
- Created the sponsor accounts, the GitHub repositories and the 1Password vault, and ran the commands the agent is not allowed to run.
- Judged what was good enough to ship at each ring.

## How to see it

- Every commit the AI authored carries a `Co-Authored-By: Claude Fable 5.1` trailer and a `Claude-Session` link.
- PR reviews from Gitar are a second AI reading the first one's work; both are visible on each pull request.
- The git tag `pre-kickoff` marks the end of setup done before the event, so `git log pre-kickoff..` is the hackathon work.

## Other services

- Gitar reviews and fixes pull requests (sponsor track).
- Dash0 and Radar analyse telemetry and cluster state with their own AI features (sponsor tracks).
- Elasticsearch's `semantic_text` field uses Elastic's ELSER model for semantic search (sponsor track).
