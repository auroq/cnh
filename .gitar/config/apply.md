Auto-apply may fix bugs, missing edge-case handling and missing tests in Go code, and missing fields in Kubernetes manifests under deploy/ (resources, probes, container securityContext, PodDisruptionBudget) that are not listed as accepted exceptions in .gitar/review/k8s.md.
Auto-apply may fix accessibility issues in templates: missing labels, alt text, lang, title, aria-live.
Every test auto-apply writes follows .gitar/review/testing.md: nested t.Run named "when ..., and ..., it should ...", one assertion per leaf, tables for data-only differences.
Auto-apply makes one change per finding, in its own commit, named for the finding; it does not bundle unrelated fixes.
Do not change files under infra/, .github/workflows/, .gitar/, or any Markdown under docs/ or roadmap/.
Do not change images, tags, hostnames, replicas or pod securityContext in manifests; comment instead.
Do not add probes, flags or code whose only purpose is to satisfy a check.
Do not touch an ECK Elasticsearch or Kibana resource, or a CronJob, beyond resources and labels.
