# 36. CI and CD are separate stages, nothing deploys untested, a bad deploy rolls itself back

**Done 2026-11-09 (PRs #40, #46, #53, #60, Terraform cd-gate; first full release d2b00ca at 13:58 MT: build, bump PR, canary, gate green; issue #36). Filed 2026-11-09, during Ring 3, by Aria. Aria's call: bump by bot commit or by auto-merged PR, and the GitHub secret step is hers.**

Depends on: 032

Numbered 036 because issue numbers share a sequence with pull requests; the item number equals the issue number.
Sequenced after item 032 because both touch the workflows; do not interleave.

## What

1. **CI**, on every pull request and push to `main`: gofmt, vet and unit tests; manifests; docs; architecture; and integration tests against an Elasticsearch service container (`docker.elastic.co/elasticsearch/elasticsearch:9.5.6`, security on, trial licence, ML off) for the bulk, index and store tests. Tests that need ELSER or `semantic_text` carry a build tag or `t.Skip` when the licence or ML is absent, so CI is deterministic. Builds on the `testing/` layout from Aria's pairing session (`testing/helpers`, `testing/integration`, `testing/acceptance`, `compose.yaml`); does not duplicate it.
2. **CD**, on push to `main` after CI passes, in its own workflow: build and push the image, then bump the image SHA in `deploy/app/kustomization.yaml`, either by a bot commit straight to `main` that CI skips, or by opening the bump PR and auto-merging it as `tools/bump-image.sh` does today. The PR says which and why.
3. **Post-deploy gate**: after the bump lands, wait for Argo CD to report `wide-door` Synced and Healthy and for the Rollout's canary to finish, through a read-and-rollout-scoped ServiceAccount token stored as a GitHub Actions secret (never the admin kubeconfig; the token also goes in the CNH vault), then run `testing/acceptance` against https://widedoor.cnh.auroq.com.
4. **Rollback on failure**: revert the bump commit so git stays the source of truth and Argo's self-heal pulls the cluster back, verify Healthy, fail the workflow with the test output in the job summary, and comment on the bump commit or PR.
5. **ADR**: CI/CD separation, the post-deploy gate, rollback-by-revert. `docs/architecture.md` gets the pipeline diagram.
6. Required checks on `main` stay the CI jobs; CD and the gate run after merge and report on the commit status.

## Why

Gitar/Sonar "Never Merge Broken" and Radar "AI SRE" both judge what happens when a change is wrong.
Separating CI from CD makes "how does a change reach production" a one-paragraph answer, and rollback-by-revert keeps git true, so Argo CD never fights a hand-made rollback.

## Acceptance

- A PR with a failing integration test cannot merge.
- A push to `main` ends with the new SHA live, Argo Synced and Healthy, and acceptance green, with no hand on the wheel.
- A push whose acceptance fails ends with the previous SHA live, a revert commit on `main`, a red CD run with the test output, and a comment on the bump.

## Touches

`.github/workflows/`, `testing/`, `deploy/` or `infra/platform` (ServiceAccount and RBAC), `docs/adr/`, `docs/architecture.md`, `tools/`.
