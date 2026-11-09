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
  It builds and pushes the image tagged with the commit SHA, then pins that SHA in `deploy/app/kustomization.yaml` with a commit pushed straight to `main` whose message carries `[skip ci]`.
  A bot-opened bump pull request would be the cleaner shape, but events created with the workflow token do not start workflows, so such a PR would never get its required checks and never merge; the direct commit is the version that works reliably.
  The bump commit changes only the image tag of a build that CI already passed, so nothing untested is deployed by skipping CI on it.
- **The gate**: CD then waits for the Rollout to run the new image and report Healthy (its canary finished), using a token for a ServiceAccount that can only read rollouts, replicasets and pods in `wide-door`, and runs the acceptance tests against https://widedoor.cnh.auroq.com.
- **Rollback is a revert**: if the gate fails, CD reverts the bump commit on `main`, waits for the Rollout to be Healthy on the previous image, writes the failing output to the job summary, comments on the bump commit, and fails.
  Git stays the source of truth and Argo CD's self-heal does the rollback; nobody runs `kubectl argo rollouts undo` by hand, and nothing fights the operator.

## Consequences

- A change reaches production in one sentence: PR merges when CI and Gitar pass; CD builds, pins, waits, tests; a failure reverts the pin.
- Two repository secrets exist: `CD_GITHUB_TOKEN` (a fine-grained token with contents and pull-requests write on `auroq/cnh`, which lets the workflow push to the protected branch because admins are not enforced) and `CD_KUBECONFIG` (the `cd-gate` ServiceAccount's time-bound token).
  Both values live in the CNH vault; neither is on a command line or in a log.
- `tools/bump-image.sh` stays for a hand-run release or a deliberate bad release (roadmap item 012's staged failure).
- The Rollout's `progressDeadlineAbort` still aborts a canary that stalls, independently of CD; CD's revert is what makes git agree with the cluster afterwards.
