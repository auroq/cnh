# 3. One issue per roadmap item

**Open. Filed 2026-11-09, from the kickoff plan. Nobody's call: ADR 0003 is accepted, GitHub Issues first and the markdown fills the gaps.**

Depends on: 001

## What

ADR 0003 is accepted, so each roadmap item gets one public GitHub issue on `auroq/cnh`.
The issue holds a one-line summary and a link to its `roadmap/` file; the file keeps the detail.
Every PR body cites its issue with `Refs #N` or `Closes #N`.
If the ADR is rejected, this item closes as adjudicated and PRs cite roadmap item numbers instead.

## Why

Gitar reads the PR and what it links, so a linked issue gives its review the intent of the change (Gitar/Sonar "Never Merge Broken").
Judges can see the plan and its progress without cloning the repo.

## Acceptance

- Each open item in `ROADMAP.md` has exactly one issue, and each issue links its item file.
- Every PR merged after this item lands cites an issue.

## Touches

GitHub issues on `auroq/cnh`, PR bodies; no files.
