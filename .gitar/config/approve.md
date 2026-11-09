Approve a pull request only when all of these hold.

- The review found no blocking finding and the risk assessment is not high, and every rule under .gitar/rules/ that applies to the changed files passed.
- The title or body cites a roadmap issue as "Closes #N" or "Refs #N".
- No test was skipped, commented out, weakened or deleted unless the same PR replaces it with a test of the same behaviour.
- A manifest under deploy/app/ meets .gitar/review/k8s.md, apart from the accepted exceptions listed there.
- A change under infra/ comes with a "terraform plan:" line in the body stating adds, changes and destroys; a person applies Terraform regardless of this approval.
- A change to .gitar/, .github/workflows/ or tools/pr.sh comes with one paragraph in the body explaining the policy change and citing an ADR under docs/adr/.

Approve without further conditions when one of these holds.

- The PR changes only Markdown under docs/, roadmap/, README.md or AI_USAGE.md and follows docs/README.md: one sentence per line, a status on line 3 of each roadmap item, and the ROADMAP.md open table consistent with the items' statuses.
- The title starts with "deploy: images at" or "deploy: revert images at", the only changed file is deploy/app/kustomization.yaml, and every new tag is a 40-character commit SHA.

When approval is withheld, say in one sentence which line above was not met; a person can answer with "gitar fix this" or approve by hand.
