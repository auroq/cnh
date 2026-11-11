#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<EOF
usage: $(basename "$0") --image IMAGE[:TAG] [--deployment NAME] [--namespace NS] [--timeout SECONDS]

Waits until the Deployment runs the given image and every replica is updated and available, then exits 0.
Exits 1 if the timeout passes first.
Needs only get and list on deployments in the namespace (KUBECONFIG may point at the cd-gate token).

  --image IMAGE       the image the Deployment must be running, e.g. ghcr.io/auroq/cnh/widedoor:<sha>
  --deployment NAME   Deployment name (default widedoor)
  --namespace NS      namespace (default wide-door)
  --timeout SECONDS   how long to wait (default 600)
  -h, --help          this text
EOF
}

image="" deployment="widedoor" namespace="wide-door" timeout=600
while [[ $# -gt 0 ]]; do
  case "$1" in
    --image) image="$2"; shift 2 ;;
    --deployment) deployment="$2"; shift 2 ;;
    --namespace) namespace="$2"; shift 2 ;;
    --timeout) timeout="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "unknown argument: $1" >&2; usage >&2; exit 2 ;;
  esac
done
[[ -n "$image" ]] || { echo "--image is required" >&2; exit 2; }

jsonpath='{.spec.template.spec.containers[0].image} {.metadata.generation} {.status.observedGeneration} {.spec.replicas} {.status.replicas} {.status.updatedReplicas} {.status.availableReplicas}'

deadline=$(( $(date +%s) + timeout ))
last=""
while true; do
  read -r running generation observed desired replicas updated available < <(
    kubectl get deployment "$deployment" -n "$namespace" -o jsonpath="$jsonpath" 2>/dev/null || true
  )
  running=${running:-} generation=${generation:-0} observed=${observed:-0} desired=${desired:-0}
  replicas=${replicas:-0} updated=${updated:-0} available=${available:-0}
  state="image=${running} generation=${generation}/${observed} replicas=${replicas} updated=${updated} available=${available} desired=${desired}"
  if [[ "$state" != "$last" ]]; then echo "$state"; last="$state"; fi
  if [[ "$running" == "$image" ]] && (( observed >= generation )) && (( replicas > 0 )) \
     && (( updated == replicas )) && (( available == replicas )); then
    echo "deployment $deployment is available on $image"
    exit 0
  fi
  if (( $(date +%s) >= deadline )); then
    echo "timed out after ${timeout}s waiting for $deployment on $image (${state})" >&2
    exit 1
  fi
  sleep 10
done
