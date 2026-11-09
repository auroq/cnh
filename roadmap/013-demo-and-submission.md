# 13. Demo and submission

**Open. Filed 2026-11-09, from the kickoff plan. Nobody's call: the deadline is 17:30 MT.**

Depends on: 010, 011, 012

## What

Code freezes at 17:00.
The README's architecture section is written from what was built, with one diagram: Sessionize and the venue seed into Elasticsearch, the API and UI on Kubernetes through Argo CD and Rollouts, telemetry to Dash0, Radar watching the cluster, Gitar on every PR.
When it is written, [`docs/transitions/pre-kickoff-platform.md`](../docs/transitions/pre-kickoff-platform.md) is deleted.

The demo script, in order, one minute per track at most:

1. Merge Forward: screen reader on, keyboard only, "nearest quiet room from 255", then a step-free question that returns an unverified elevator and says so.
2. Elasticsearch: a free-text question that only the semantic half of the hybrid query answers.
3. Dash0: the trace of that question.
4. Radar: the clean namespace and the staged rollout failure.
5. Gitar: the rule comment and the auto-applied CI fix.

## Submission checklist

- [x] Repo public, `main` green.
- [x] README explains the architecture and lists the tracks entered.
- [x] `AI_USAGE.md` complete and every AI-assisted commit has a `Co-Authored-By` trailer.
- [ ] Demo video recorded and linked from the README.
- [ ] One project submitted, naming every track.
- [ ] Teardown of the cluster scheduled for after judging, not before.

The first three were checked by the integrator at 16:40 MT; the video, the submission and the teardown are Aria's.

## Touches

`README.md`, `docs/transitions/`.
