#!/usr/bin/env bash
set -euo pipefail

repo="${REPO:-auroq/cnh}"
remote="${REMOTE:-github}"
root="$(cd "$(dirname "$0")/.." && pwd)"
kustomization="${root}/deploy/app/kustomization.yaml"

usage() {
  cat <<EOF
usage: $(basename "$0") [--sha SHA] [--dry-run]

Pins every ghcr.io/auroq/cnh image in deploy/app/kustomization.yaml to the commit SHA
of the latest successful "images" workflow run on main (or the given SHA), on a new
branch deploy/bump-<sha7>, and opens the PR with tools/pr.sh.
Argo CD then sees a real manifest change, and Argo Rollouts runs the canary steps.

  --sha SHA    pin to this 40-character commit SHA instead of the latest build
  --dry-run    show the SHA and the diff, change nothing in git
  -h, --help   this text
EOF
}

sha="" dry_run=0
while [[ $# -gt 0 ]]; do
  case "$1" in
    --sha) sha="$2"; shift 2 ;;
    --dry-run) dry_run=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "unknown argument: $1" >&2; usage >&2; exit 2 ;;
  esac
done

if [[ -z "$sha" ]]; then
  sha="$(gh run list --repo "$repo" --workflow images --branch main --status success --limit 1 --json headSha -q '.[0].headSha')"
fi
[[ "$sha" =~ ^[0-9a-f]{40}$ ]] || { echo "not a commit SHA: '$sha'" >&2; exit 1; }
short="${sha:0:7}"

if grep -q "newTag: ${sha}" "$kustomization"; then
  echo "already pinned to ${short}"
  exit 0
fi

if [[ "$dry_run" -eq 1 ]]; then
  echo "would pin to ${sha}"
  sed -E "s|(newTag: )[0-9a-f]{40}|\1${sha}|" "$kustomization" | diff -u "$kustomization" - || true
  exit 0
fi

[[ -z "$(git -C "$root" status --porcelain)" ]] || { echo "working tree is not clean" >&2; exit 1; }
git -C "$root" fetch -q "$remote" main
git -C "$root" checkout -q -B "deploy/bump-${short}" "${remote}/main"
sed -i -E "s|(newTag: )[0-9a-f]{40}|\1${sha}|" "$kustomization"
git -C "$root" add "$kustomization"
git -C "$root" commit -q -m "deploy: images at ${short}" -m "Pins the widedoor and ingest images to the build of ${sha} on main."
git -C "$root" push -q -u "$remote" "deploy/bump-${short}"
"${root}/tools/pr.sh" --title "deploy: images at ${short}" --body "Pins the widedoor and ingest images to the \`images\` build of ${sha} on \`main\`, so Argo CD syncs a real change and the Rollout runs its canary steps."
