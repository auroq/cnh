#!/usr/bin/env bash
set -euo pipefail

module="github.com/auroq/cnh"
root="$(cd "$(dirname "$0")/.." && pwd)"

usage() {
  cat <<EOF
usage: $(basename "$0") [--graph]

Checks that every package imports only what its layer may import.
The layers and the rule are written in .gitar/review/architecture.md and docs/architecture.md;
this script is the mechanical half.
Exit 1 names every offending edge "from -> to".
A package that is not in the layer map fails too: add it to this script and to docs/architecture.md.

  --graph     print every internal import edge and exit
  -h, --help  this text
EOF
}

# core: pkg/model (what the business means), pkg/api/v1 (what clients were promised), pkg/config; import nothing from this module
# model helpers: pkg/venue, pkg/schedule, pkg/agenda; import pkg/model only
# usecase: declare ports; import model packages only
# adapter: implement ports; import model and use case packages, never a controller
# web: the HTTP controller; imports pkg/ask, pkg/reports, pkg/model and pkg/api/v1, and is the only importer of pkg/api/v1
# wiring: pkg/cli, pkg/app, cmd/*; import anything under pkg/; nobody imports them
layer_of() {
  case "$1" in
    pkg/model|pkg/api/v1|pkg/config) echo core ;;
    pkg/venue|pkg/schedule|pkg/agenda) echo model ;;
    pkg/ask|pkg/ingest|pkg/reports) echo usecase ;;
    pkg/esindex|pkg/sessionize|pkg/telemetry|pkg/qr) echo adapter ;;
    pkg/web) echo web ;;
    pkg/cli|pkg/app|cmd/*) echo wiring ;;
    *) echo unknown ;;
  esac
}

is_core()    { [[ "$1" == pkg/model || "$1" == pkg/api/v1 || "$1" == pkg/config ]]; }
is_model()   { [[ "$1" == pkg/venue || "$1" == pkg/schedule || "$1" == pkg/agenda ]]; }
is_usecase() { [[ "$1" == pkg/ask || "$1" == pkg/ingest || "$1" == pkg/reports ]]; }

allowed() {
  local from_layer="$1" to="$2"
  if [[ "$to" == pkg/api/v1 && "$from_layer" != web ]]; then
    return 1
  fi
  case "$from_layer" in
    core) return 1 ;;
    model) [[ "$to" == pkg/model ]] ;;
    usecase) is_core "$to" || is_model "$to" ;;
    adapter) is_core "$to" || is_model "$to" || is_usecase "$to" ;;
    web) [[ "$to" == pkg/ask || "$to" == pkg/reports || "$to" == pkg/model || "$to" == pkg/api/v1 ]] ;;
    wiring) [[ "$to" == pkg/* ]] ;;
    *) return 1 ;;
  esac
}

graph=0
case "${1:-}" in
  --graph) graph=1 ;;
  -h|--help) usage; exit 0 ;;
  "") ;;
  *) echo "unknown argument: $1" >&2; usage >&2; exit 2 ;;
esac

cd "$root"
fail=0
while read -r pkg imports; do
  from="${pkg#"$module"/}"
  layer="$(layer_of "$from")"
  if [[ "$layer" == unknown ]]; then
    echo "$from: not in the layer map; add it to tools/archcheck.sh and docs/architecture.md"
    fail=1
    continue
  fi
  for imp in $imports; do
    [[ "$imp" == "$module"/* ]] || continue
    to="${imp#"$module"/}"
    if [[ "$graph" -eq 1 ]]; then
      echo "$from -> $to"
      continue
    fi
    if ! allowed "$layer" "$to"; then
      echo "$from -> $to: a $layer package may not import $to"
      fail=1
    fi
  done
done < <(go list -f '{{.ImportPath}} {{join .Imports " "}}' ./cmd/... ./pkg/...)

if [[ "$graph" -eq 0 && "$fail" -eq 0 ]]; then
  echo "every import points inward"
fi
exit "$fail"
