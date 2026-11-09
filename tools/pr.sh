#!/usr/bin/env bash
set -euo pipefail

repo="${REPO:-auroq/cnh}"

usage() {
  cat <<EOF
usage: $(basename "$0") --title TEXT (--body-file FILE | --body TEXT) [--base BRANCH]
       $(basename "$0") --existing NUMBER

Opens a pull request from the current branch and applies this repo's PR policy:
a "gitar display:verbose" and "gitar auto-apply:off" comment, then GitHub auto-merge (squash).
With --existing, only applies the policy to an already open PR.

  --title TEXT       PR title
  --body-file FILE   PR body from a file (the Claude Code footer is appended)
  --body TEXT        PR body inline (the footer is appended)
  --base BRANCH      target branch (default main)
  --existing NUMBER  apply comment and auto-merge to PR NUMBER instead of opening one
  --no-auto-merge    skip arming auto-merge
  -h, --help         this text
EOF
}

title="" body="" body_file="" base="main" existing="" auto_merge=1
while [[ $# -gt 0 ]]; do
  case "$1" in
    --title) title="$2"; shift 2 ;;
    --body) body="$2"; shift 2 ;;
    --body-file) body_file="$2"; shift 2 ;;
    --base) base="$2"; shift 2 ;;
    --existing) existing="$2"; shift 2 ;;
    --no-auto-merge) auto_merge=0; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "unknown argument: $1" >&2; usage >&2; exit 2 ;;
  esac
done

footer=$'\n\n🤖 Generated with [Claude Code](https://claude.com/claude-code)'
if [[ -n "${CLAUDE_SESSION_URL:-}" ]]; then
  footer+=$'\n\n'"${CLAUDE_SESSION_URL}"
fi

if [[ -z "$existing" ]]; then
  [[ -n "$title" ]] || { echo "--title is required" >&2; exit 2; }
  if [[ -n "$body_file" ]]; then
    body="$(cat "$body_file")"
  fi
  [[ -n "$body" ]] || { echo "--body or --body-file is required" >&2; exit 2; }
  head="$(git rev-parse --abbrev-ref HEAD)"
  url="$(gh pr create --repo "$repo" --base "$base" --head "$head" --title "$title" --body "${body}${footer}")"
  existing="${url##*/}"
  echo "$url"
fi

gh pr comment "$existing" --repo "$repo" --body $'gitar display:verbose\ngitar auto-apply:off' >/dev/null
echo "PR #$existing: Gitar set to verbose, auto-apply off"

if [[ "$auto_merge" -eq 1 ]]; then
  gh pr merge "$existing" --repo "$repo" --auto --squash
  echo "PR #$existing: auto-merge (squash) armed"
fi
