#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<EOF
usage: $(basename "$0") [--timeout SECONDS]

Waits until the Elasticsearch at ES_URL answers and its default ELSER inference endpoint
returns an embedding, so the first semantic_text write does not fail while the model downloads.
The first call on a fresh node starts the download, which takes minutes and a few gigabytes.

Environment:
  ES_URL       base URL (default http://127.0.0.1:19200)
  ES_USERNAME  basic auth user (default elastic)
  ES_PASSWORD  basic auth password

  --timeout SECONDS  how long to wait in total (default 900)
  -h, --help         this text
EOF
}

timeout=900
while [[ $# -gt 0 ]]; do
  case "$1" in
    --timeout) timeout="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "unknown argument: $1" >&2; usage >&2; exit 2 ;;
  esac
done

url="${ES_URL:-http://127.0.0.1:19200}"
user="${ES_USERNAME:-elastic}"
endpoint=".elser-2-elasticsearch"
deadline=$(( $(date +%s) + timeout ))

request() {
  local out="$1"
  shift
  curl -sS -o "$out" -w '%{http_code}' -K <(printf 'user = "%s:%s"\n' "$user" "${ES_PASSWORD:-}") "$@" 2>/dev/null || echo 000
}

echo "waiting for $url"
until [[ "$(request /dev/null "$url/_cluster/health?wait_for_status=yellow&timeout=5s")" == 200 ]]; do
  (( $(date +%s) < deadline )) || { echo "timed out waiting for $url" >&2; exit 1; }
  sleep 3
done

echo "warming inference endpoint $endpoint (first run downloads ELSER)"
body=$(mktemp)
trap 'rm -f "$body"' EXIT
while :; do
  code=$(request "$body" -X POST -H 'Content-Type: application/json' --max-time 120 "$url/_inference/sparse_embedding/$endpoint" -d '{"input":["warm up"]}')
  if [[ "$code" == 200 ]]; then
    echo "inference endpoint $endpoint is ready"
    exit 0
  fi
  (( $(date +%s) < deadline )) || { echo "timed out waiting for $endpoint: HTTP $code: $(head -c 300 "$body")" >&2; exit 1; }
  echo "  not ready yet (HTTP $code), retrying"
  sleep 10
done
