#!/usr/bin/env bash
set -euo pipefail

repo="${REPO:-auroq/cnh}"

usage() {
  cat <<EOF
usage: $(basename "$0") NUMBER [--timeout SECONDS]

Waits for pull request NUMBER to merge: polls its checks, re-arms auto-merge once they
pass (GitHub sometimes needs the nudge), and prints the merge commit SHA on success.
Exits 1 if a check fails, the PR is closed without merging, or the timeout passes.

  --timeout SECONDS   how long to wait (default 1200)
  -h, --help          this text
EOF
}

number="" timeout=1200
while [[ $# -gt 0 ]]; do
  case "$1" in
    --timeout) timeout="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) number="$1"; shift ;;
  esac
done
[[ "$number" =~ ^[0-9]+$ ]] || { echo "a PR number is required" >&2; usage >&2; exit 2; }

deadline=$(( $(date +%s) + timeout ))
rearmed=0
while true; do
  state=$(gh pr view "$number" --repo "$repo" --json state -q .state)
  case "$state" in
    MERGED)
      gh pr view "$number" --repo "$repo" --json mergeCommit -q .mergeCommit.oid
      exit 0 ;;
    CLOSED)
      echo "PR #$number was closed without merging" >&2
      exit 1 ;;
  esac
  if gh pr checks "$number" --repo "$repo" >/dev/null 2>&1; then
    if [[ "$rearmed" -eq 0 ]]; then
      gh pr merge "$number" --repo "$repo" --disable-auto >/dev/null 2>&1 || true
      gh pr merge "$number" --repo "$repo" --auto --squash >/dev/null 2>&1 || true
      rearmed=1
      echo "PR #$number: checks passed, auto-merge re-armed"
    fi
  else
    if gh pr checks "$number" --repo "$repo" 2>/dev/null | grep -qE '\bfail\b'; then
      echo "PR #$number has a failing check:" >&2
      gh pr checks "$number" --repo "$repo" >&2 || true
      exit 1
    fi
  fi
  if (( $(date +%s) >= deadline )); then
    echo "timed out after ${timeout}s waiting for PR #$number (state $state)" >&2
    exit 1
  fi
  sleep 20
done
