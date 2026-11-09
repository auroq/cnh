---
title: "Manifests stay Radar clean"
description: "Block workload manifests that would fail Radar's checks"
type: check
when: "A PR adds or changes a Deployment, StatefulSet, DaemonSet, Rollout or CronJob under deploy/"
actions: "Review against .gitar/review/k8s.md and request changes listing every missing field with its path, or approve if nothing is missing"
---

# Manifests stay Radar clean

Radar audits the cluster continuously and the project namespace must show zero findings.
Check every workload in the PR against the list in `.gitar/review/k8s.md`.
If anything is missing, request changes and list each missing field by its YAML path.
If an app workload is a plain `Deployment`, ask for it to be an Argo Rollouts `Rollout` with a `PodDisruptionBudget`.
