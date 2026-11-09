#!/usr/bin/env bash
set -euo pipefail

api="${DASH0_API:-https://api.us-west-2.aws.dash0.com}"
dataset="${DASH0_DATASET:-default}"
op_token="${OP_TOKEN_REF:-pass:cnh/op-service-account-token}"
dash0_token="${DASH0_TOKEN_REF:-op:CNH/yfcbogym5s23ghz3q2qp7lilha/credential}"
root="$(cd "$(dirname "$0")/.." && pwd)"

usage() {
  cat <<EOF
usage: $(basename "$0") [--check] [FILE...]

Pushes each Perses dashboard FILE (default: every deploy/dash0/*.json) to Dash0 through
PUT /api/dashboards/<origin>?dataset=$dataset, where <origin> is the file's metadata.name,
so the same file always updates the same dashboard. Prints the id, name and version Dash0 holds.
The token is never on the command line: it reaches curl through secret-run (1Password CNH vault,
service-account token from pass).

  --check     only GET each dashboard by origin and print what Dash0 holds; push nothing
  -h, --help  this text

environment: DASH0_API ($api), DASH0_DATASET ($dataset), OP_TOKEN_REF, DASH0_TOKEN_REF
EOF
}

check=0
files=()
while [[ $# -gt 0 ]]; do
  case "$1" in
    --check) check=1; shift ;;
    -h|--help) usage; exit 0 ;;
    -*) echo "unknown argument: $1" >&2; usage >&2; exit 2 ;;
    *) files+=("$1"); shift ;;
  esac
done
if [[ ${#files[@]} -eq 0 ]]; then
  files=("$root"/deploy/dash0/*.json)
fi

for tool in jq curl secret-run; do
  command -v "$tool" >/dev/null || { echo "$tool is required" >&2; exit 2; }
done

call() {
  local method="$1" path="$2" body="${3:-}"
  local args=(-sS -f -X "$method" -H "Content-Type: application/json")
  if [[ -n "$body" ]]; then
    args+=(--data-binary "@$body")
  fi
  secret-run --env OP_SERVICE_ACCOUNT_TOKEN="$op_token" -- \
    secret-run --env DASH0_AUTH_TOKEN="$dash0_token" -- \
    sh -c 'curl -H "Authorization: Bearer $DASH0_AUTH_TOKEN" "$@"' sh "${args[@]}" "$api$path"
}

summarise() {
  jq -r '"  id: \(.metadata.dash0Extensions.id // "?")  name: \(.spec.display.name // .metadata.name)  version: \(.metadata.version // "?")  updated: \(.metadata.updatedAt // "?")  panels: \(.spec.panels | length)"'
}

status=0
for f in "${files[@]}"; do
  origin="$(jq -r '.metadata.name' "$f")"
  if [[ -z "$origin" || "$origin" == "null" ]]; then
    echo "$f: metadata.name is required, it is the dashboard's origin in Dash0" >&2
    status=1
    continue
  fi
  if [[ "$check" -eq 1 ]]; then
    echo "$f -> GET $api/api/dashboards/$origin?dataset=$dataset"
    if ! call GET "/api/dashboards/$origin?dataset=$dataset" | summarise; then
      echo "  not found in Dash0" >&2
      status=1
    fi
  else
    echo "$f -> PUT $api/api/dashboards/$origin?dataset=$dataset"
    call PUT "/api/dashboards/$origin?dataset=$dataset" "$f" | summarise
  fi
done
exit "$status"
