#!/usr/bin/env bash
# Create a Radar Cloud cluster connection via API and stash any returned token in 1Password.
# Needs RADAR_PAT and OP_SERVICE_ACCOUNT_TOKEN in env (via secret-run). Prints only redacted output.
set -euo pipefail

if [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
  echo "usage: secret-run --env OP_SERVICE_ACCOUNT_TOKEN=... --env RADAR_PAT=... -- $0 [cluster-name]"
  exit 0
fi

name="${1:-cnh}"
resp="$(printf 'Authorization: Bearer %s\nContent-Type: application/json\n' "$RADAR_PAT" |
  curl -s -w '\n%{http_code}' -H @- -X POST --data "{\"name\":\"$name\"}" https://api.radarhq.io/api/clusters)"
code="${resp##*$'\n'}"
body="${resp%$'\n'*}"

echo "HTTP $code"
jq 'walk(if type == "string" and test("^rh[a-z]_") then (.[0:4] + "<redacted>") else . end)' <<<"$body" 2>/dev/null || echo "(non-JSON body, ${#body} bytes)"

token="$(jq -r '[.. | strings | select(test("^rhc_"))][0] // empty' <<<"$body" 2>/dev/null || true)"
cid="$(jq -r '[.. | strings | select(test("^cl_"))][0] // empty' <<<"$body" 2>/dev/null || true)"

if [[ -n "$token" ]]; then
  jq -n --arg t "$token" --arg cid "$cid" '{
    title: "Radar Cluster Connect CNH",
    category: "API_CREDENTIAL",
    fields: [
      {id: "credential", type: "CONCEALED", label: "credential", value: $t},
      {id: "cluster_id", type: "STRING", label: "cluster_id", value: $cid}
    ]
  }' | op item create --vault CNH --format json - | jq -r '"stored 1Password item: \(.title) (\(.id))"'
  echo "cluster_id=$cid"
else
  echo "no rhc_ token in response"
fi
