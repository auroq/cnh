#!/usr/bin/env bash
# Mirror roadmap/NNN-*.md items to GitHub issues on auroq/cnh (ADR 0003).
# Idempotent: an item whose title already exists as an issue is skipped.
set -euo pipefail

repo="${REPO:-auroq/cnh}"
root="$(cd "$(dirname "$0")/.." && pwd)"

usage() {
  cat <<EOF
usage: $(basename "$0") [--dry-run] [NNN ...]

Creates one GitHub issue per roadmap item (or only the given item numbers),
labelled with its ring and tracks, body taken from the item file.
Items whose status line starts with Done are created and closed.
Needs: gh (authenticated), labels created by '$(basename "$0") --labels'.

  --labels    create or update the ring/track labels and exit
  --dry-run   print what would be created
  -h, --help  this text
EOF
}

labels() {
  while IFS='|' read -r name color desc; do
    gh label create "$name" -R "$repo" --color "$color" --description "$desc" --force >/dev/null
    echo "label $name"
  done <<'EOF'
ring-0|0e8a16|11:00-12:00 first commit, pipeline proven
ring-1|0e8a16|12:00-13:00 Elasticsearch up, schedule indexed
ring-2|0e8a16|13:00-14:00 venue data and the ask API
ring-3|0e8a16|14:00-15:00 screen-reader-first UI
ring-4|0e8a16|15:00-16:00 Dash0, Radar, Gitar proven
ring-5|0e8a16|17:00-17:30 demo and submission
stretch|fbca04|16:00-17:00 only if ring 4 is done
track:gitar|1d76db|Gitar / Sonar: Never Merge Broken
track:dash0|1d76db|Dash0: Stop Flying Blind
track:radar|1d76db|Radar: AI SRE
track:elastic|1d76db|Elasticsearch: Find the Right Talk
track:merge-forward|1d76db|Merge Forward: Conferences for Everyone
EOF
}

# item number -> labels
item_labels() {
  case "$1" in
    001) echo "ring-0" ;;
    002) echo "ring-0,track:merge-forward" ;;
    003) echo "ring-0" ;;
    004) echo "ring-0,track:gitar" ;;
    005) echo "ring-0,track:radar" ;;
    006) echo "ring-1,track:elastic" ;;
    007) echo "ring-1,track:elastic" ;;
    008) echo "ring-2,track:elastic,track:merge-forward,accessibility" ;;
    009) echo "ring-2,track:elastic" ;;
    010) echo "ring-3,track:merge-forward,accessibility" ;;
    011) echo "ring-4,track:dash0" ;;
    012) echo "ring-4,track:radar" ;;
    013) echo "ring-5" ;;
    014) echo "stretch,track:merge-forward,accessibility" ;;
    015) echo "stretch,track:merge-forward,accessibility" ;;
    016) echo "stretch,track:elastic" ;;
    *) echo "" ;;
  esac
}

dry=0
case "${1:-}" in
  -h|--help) usage; exit 0 ;;
  --labels) labels; exit 0 ;;
  --dry-run) dry=1; shift ;;
esac

existing="$(gh issue list -R "$repo" --state all --limit 200 --json title --jq '.[].title')"

for f in "$root"/roadmap/[0-9][0-9][0-9]-*.md; do
  n="$(basename "$f" | cut -d- -f1)"
  if [[ $# -gt 0 ]] && ! printf '%s\n' "$@" | grep -qx "$n"; then continue; fi
  title="$(sed -n 1p "$f" | sed -E 's/^# [0-9]+\. //')"
  issue_title="$n. $title"
  if grep -qxF "$issue_title" <<<"$existing"; then
    echo "skip  $issue_title (exists)"
    continue
  fi
  status="$(sed -n 3p "$f")"
  body="$(printf '%s\n\nFull item: https://github.com/%s/blob/main/roadmap/%s\n\n%s' \
    "$status" "$repo" "$(basename "$f")" "$(tail -n +4 "$f")")"
  lbl="$(item_labels "$n")"
  if [[ $dry -eq 1 ]]; then
    echo "would create: $issue_title [$lbl]"
    continue
  fi
  url="$(gh issue create -R "$repo" --title "$issue_title" --body "$body" ${lbl:+--label "$lbl"})"
  echo "created $url  $issue_title"
  if [[ "$status" == *"**Done"* ]]; then
    gh issue close "$url" -R "$repo" --comment "Decided; see the item file." >/dev/null
    echo "closed  $url"
  fi
done
