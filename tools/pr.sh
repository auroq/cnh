#!/usr/bin/env bash
set -euo pipefail

repo="${REPO:-auroq/cnh}"

usage() {
  cat <<EOF
usage: $(basename "$0") (--closes N | --refs N) --title TEXT (--body-file FILE | --body TEXT) [--base BRANCH]
       $(basename "$0") --existing NUMBER

Opens a pull request from the current branch and applies this repo's PR policy:
the issue it is for in the body (Gitar approves only PRs that cite one, ADR 0009),
a "gitar display:verbose" and "gitar auto-apply:off" comment, then GitHub auto-merge (squash).
With --existing, only applies the comment and auto-merge to an already open PR.

  --closes N         the roadmap issue this PR finishes ("Closes #N" opens the body)
  --refs N           the roadmap issue this PR is part of ("Refs #N" opens the body)
  --title TEXT       PR title
  --body-file FILE   PR body from a file (the issue line goes first, the Claude Code footer last)
  --body TEXT        PR body inline (same)
  --base BRANCH      target branch (default main)
  --existing NUMBER  apply comment and auto-merge to PR NUMBER instead of opening one
  --no-auto-merge    skip arming auto-merge
  -h, --help         this text
EOF
}

title="" body="" body_file="" base="main" existing="" auto_merge=1 issue_line=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --closes) issue_line="Closes #$2"; shift 2 ;;
    --refs) issue_line="Refs #$2"; shift 2 ;;
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
  [[ -n "$issue_line" ]] || { echo "--closes N or --refs N is required: every PR cites its roadmap issue" >&2; exit 2; }
  if [[ -n "$body_file" ]]; then
    body="$(cat "$body_file")"
  fi
  [[ -n "$body" ]] || { echo "--body or --body-file is required" >&2; exit 2; }
  head="$(git rev-parse --abbrev-ref HEAD)"
  url="$(gh pr create --repo "$repo" --base "$base" --head "$head" --title "$title" --body "${issue_line}"$'\n\n'"${body}${footer}")"
  existing="${url##*/}"
  echo "$url"
fi

gh pr comment "$existing" --repo "$repo" --body $'gitar display:verbose\ngitar auto-apply:off' >/dev/null
echo "PR #$existing: Gitar set to verbose, auto-apply off"

if [[ "$auto_merge" -eq 1 ]]; then
  gh pr merge "$existing" --repo "$repo" --auto --squash
  echo "PR #$existing: auto-merge (squash) armed"
fi
