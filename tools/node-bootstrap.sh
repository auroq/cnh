#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
usage: node-bootstrap.sh [--namespace NS] [--port PORT] [--timeout SECONDS]

One-time steps after `make up-node` that need the Elasticsearch API rather than a manifest:
  1. a Kibana service-account token, stored as the Secret wide-door-kb-token
  2. the ELSER download, via tools/es-warm.sh
  3. the first ingest, as a Job created from the ingest CronJob

KUBECONFIG must point at the node (make node-kubeconfig writes .kube/node.config).
The elastic password is read from the wide-door-es-elastic-user Secret into a mode-600
temporary file and handed to curl through secret-run; it is never printed or put in argv.

  --namespace NS     app namespace (default wide-door)
  --port PORT        local port for the Elasticsearch port-forward (default 19202)
  --timeout SECONDS  how long to wait for Elasticsearch and ELSER (default 900)
  -h, --help         this text
EOF
}

namespace="wide-door" port=19202 timeout=900
while [[ $# -gt 0 ]]; do
  case "$1" in
    --namespace) namespace="$2"; shift 2 ;;
    --port) port="$2"; shift 2 ;;
    --timeout) timeout="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "unknown argument: $1" >&2; usage >&2; exit 2 ;;
  esac
done

for tool in kubectl curl secret-run python3; do
  command -v "$tool" >/dev/null || { echo "missing: $tool" >&2; exit 1; }
done
[[ -n "${KUBECONFIG:-}" ]] || { echo "KUBECONFIG is not set; run make node-kubeconfig first" >&2; exit 1; }

here=$(cd "$(dirname "$0")" && pwd)
workdir=$(mktemp -d)
cleanup() {
  [[ -n "${pf_pid:-}" ]] && kill "$pf_pid" 2>/dev/null || true
  rm -rf "$workdir"
}
trap cleanup EXIT

echo "waiting for the Elasticsearch pod in $namespace"
kubectl -n "$namespace" rollout status statefulset/wide-door-es-default --timeout="${timeout}s"

umask 077
kubectl -n "$namespace" get secret wide-door-es-elastic-user -o jsonpath='{.data.elastic}' | base64 -d > "$workdir/es-password"

kubectl -n "$namespace" port-forward svc/wide-door-es-http "${port}:9200" >"$workdir/port-forward.log" 2>&1 &
pf_pid=$!
for _ in $(seq 1 30); do
  curl -fsS -o /dev/null "http://127.0.0.1:${port}/" -u "elastic:x" 2>/dev/null && break
  curl -s -o /dev/null -w '%{http_code}' "http://127.0.0.1:${port}/" 2>/dev/null | grep -q 401 && break
  sleep 1
done

url="http://127.0.0.1:${port}"

if kubectl -n "$namespace" get secret wide-door-kb-token >/dev/null 2>&1; then
  echo "wide-door-kb-token exists, keeping it"
else
  echo "creating a Kibana service-account token"
  secret-run --env ES_PASSWORD="file:$workdir/es-password" -- sh -c '
    set -eu
    curl -fsS -u "elastic:$ES_PASSWORD" -X POST "$0/_security/service/elastic/kibana/credential/token/node-$(date +%s)" \
      | python3 -c "import json,sys; sys.stdout.write(json.load(sys.stdin)[\"token\"][\"value\"])"
  ' "$url" | kubectl -n "$namespace" create secret generic wide-door-kb-token --from-file=token=/dev/stdin
fi

echo "warming ELSER (first run downloads the model)"
secret-run --env ES_PASSWORD="file:$workdir/es-password" -- \
  env ES_URL="$url" ES_USERNAME=elastic "$here/es-warm.sh" --timeout "$timeout"

job="ingest-first-$(date +%s)"
echo "running the first ingest as job/$job"
kubectl -n "$namespace" create job --from=cronjob/ingest "$job"
kubectl -n "$namespace" wait --for=condition=complete "job/$job" --timeout="${timeout}s"
kubectl -n "$namespace" logs "job/$job" | tail -20

echo "bootstrap done"
