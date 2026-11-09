#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<EOF
usage: $(basename "$0") --image IMAGE[:TAG] [--rollout NAME] [--namespace NS] [--timeout SECONDS]

Waits until the Rollout runs the given image and reports Healthy, then exits 0.
Exits 1 with the Rollout's status if it reports Degraded or the timeout passes.
Needs only get on rollouts in the namespace (KUBECONFIG may point at the cd-gate token).

  --image IMAGE       the image the Rollout must be running, e.g. ghcr.io/auroq/cnh/widedoor:<sha>
  --rollout NAME      Rollout name (default widedoor)
  --namespace NS      namespace (default wide-door)
  --timeout SECONDS   how long to wait (default 600)
  -h, --help          this text
EOF
}

image="" rollout="widedoor" namespace="wide-door" timeout=600
while [[ $# -gt 0 ]]; do
  case "$1" in
    --image) image="$2"; shift 2 ;;
    --rollout) rollout="$2"; shift 2 ;;
    --namespace) namespace="$2"; shift 2 ;;
    --timeout) timeout="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "unknown argument: $1" >&2; usage >&2; exit 2 ;;
  esac
done
[[ -n "$image" ]] || { echo "--image is required" >&2; exit 2; }

deadline=$(( $(date +%s) + timeout ))
last=""
while true; do
  running=$(kubectl get rollout "$rollout" -n "$namespace" -o jsonpath='{.spec.template.spec.containers[0].image}' 2>/dev/null || true)
  phase=$(kubectl get rollout "$rollout" -n "$namespace" -o jsonpath='{.status.phase}' 2>/dev/null || true)
  message=$(kubectl get rollout "$rollout" -n "$namespace" -o jsonpath='{.status.message}' 2>/dev/null || true)
  state="image=${running} phase=${phase} ${message}"
  if [[ "$state" != "$last" ]]; then echo "$state"; last="$state"; fi
  if [[ "$running" == "$image" && "$phase" == "Healthy" ]]; then
    echo "rollout $rollout is Healthy on $image"
    exit 0
  fi
  if [[ "$running" == "$image" && "$phase" == "Degraded" ]]; then
    echo "rollout $rollout is Degraded on $image: $message" >&2
    exit 1
  fi
  if (( $(date +%s) >= deadline )); then
    echo "timed out after ${timeout}s waiting for $rollout on $image (phase=${phase})" >&2
    exit 1
  fi
  sleep 10
done
