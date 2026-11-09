# 0008. CI and CD are separate stages, with a post-deploy gate and rollback by revert

Status: Accepted, 2026-11-09. Aria's call, roadmap item 036.

## Context

Until this decision a push to `main` built an image, and a person ran `tools/bump-image.sh` to pin it; nothing checked the live site afterwards, and a bad release would have been fixed by hand.
Argo CD syncs `deploy/app` with `selfHeal: true`, so any rollback done against the cluster is undone by the next sync.
The judged question for Gitar/Sonar and Radar is what happens when a change is wrong.

## Decision

CI and CD are separate GitHub Actions workflows.

- **CI** (`ci.yml`) runs on every pull request and on every push to `main`: format, vet and unit tests, the manifest check, the docs check, the architecture check, and integration tests against an Elasticsearch service container.
  Its jobs are the required checks on `main`.
- **CD** (`cd.yml`) runs when CI on `main` succeeds, never on a pull request.
  It builds and pushes the image tagged with the commit SHA, then opens a pull request that pins that SHA in `deploy/app/kustomization.yaml` (`tools/bump-image.sh`), arms auto-merge, and waits for it to merge.
  The PR is opened with a fine-grained personal token, not the workflow token, because events created with the workflow token do not start workflows and such a PR would never get its required checks.
  Every change to `main` therefore still arrives through a pull request that CI and Gitar have passed, which is what `AI_USAGE.md` and roadmap item 004 promise.
  A bump commit pushed straight to `main` was considered and rejected for breaking that promise.
- **The gate**: CD then waits for the Rollout to run the new image and report Healthy (its canary finished), using a token for a ServiceAccount that can only read rollouts, replicasets and pods in `wide-door`, and runs the acceptance tests against https://widedoor.cnh.auroq.com.
- **Rollback is a revert**: if the gate fails, CD opens a pull request that reverts the bump, auto-merges it, waits for the Rollout to be Healthy on the previous image, writes the failing output to the job summary, comments on the bump PR, and fails.
  Git stays the source of truth and Argo CD's self-heal does the rollback; nobody runs `kubectl argo rollouts undo` by hand, and nothing fights the operator.
  The revert takes as long as CI and Gitar take on the revert PR, a few minutes; the Rollout's `progressDeadlineAbort` keeps the previous pods serving in the meantime, so the wait costs availability nothing.

## Consequences

- A change reaches production in one sentence: a PR merges when CI and Gitar pass; CD builds the image, opens and merges the bump PR, waits for the canary, tests the live site; a failure opens and merges the revert PR.
- Two repository secrets exist: `CD_GITHUB_TOKEN` (a fine-grained token with contents, pull-requests and secrets write on `auroq/cnh`, created by Aria) and `CD_KUBECONFIG` (the `cd-gate` ServiceAccount's time-bound token, created by `infra/platform`).
  Both values live in the CNH vault; neither is on a command line or in a log.
- A commit whose subject starts with `deploy: images at` or `deploy: revert images at` runs only the gate, never a build, so a bump cannot trigger another bump.
- `tools/bump-image.sh` stays for a hand-run release or a deliberate bad release (roadmap item 012's staged failure).
- The Rollout's `progressDeadlineAbort` still aborts a canary that stalls, independently of CD; CD's revert is what makes git agree with the cluster afterwards.
