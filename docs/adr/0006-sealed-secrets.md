# 0006. Secrets as SealedSecrets only

Status: Accepted, 2026-11-09.

## Context

The repo is public, and the app needs at least Elasticsearch credentials and possibly an ASR API key.
The Sealed Secrets controller was installed before kickoff.
Platform secrets in `infra/` already come from 1Password through Terraform and never appear in the repo or on a command line.

## Decision

A secret the app needs is either generated in the cluster (as ECK does for Elasticsearch) or committed as a `SealedSecret` encrypted to the cluster's key.
No plain `Secret` manifest is committed and no secret is created with `kubectl create secret`.
Values reach `kubeseal` through `secret-run` from 1Password, never typed or pasted.

## Consequences

- Argo CD can sync everything from git, secrets included.
- A SealedSecret only decrypts on this cluster, so a rebuilt cluster means resealing; acceptable for a one-day project.
- A Gitar rule flags any committed `kind: Secret`.
