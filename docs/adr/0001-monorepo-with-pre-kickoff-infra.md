# 0001. One repo, with infra built before the event

Status: Accepted, 2026-11-09.

## Context

The rules allow setup before kickoff but require the project itself to be built during the event, in one public repo.
The cluster and its operators (Terraform in `infra/`) were built before kickoff so the day could go to the app.
Judges need to tell the two apart without trusting anyone's word.

## Decision

Everything lives in `auroq/cnh`: `infra/`, the app, its manifests and its documents.
The last commit before kickoff is tagged `pre-kickoff`, and `git log pre-kickoff..` is the hackathon's work.
`infra/` is not changed during the event unless an item needs it, and any such change is its own commit that says why.
The app deploys through manifests outside `infra/`, synced by Argo CD, so the app never needs a Terraform apply.

## Consequences

- One link gives judges the whole system, and the tag makes the before and after checkable.
- What `pre-kickoff` contains is described in [the pre-kickoff platform transition doc](../transitions/pre-kickoff-platform.md) until the README absorbs it.
- `infra/` auto-applies nothing on push, so the app and the platform cannot break each other through CI.
