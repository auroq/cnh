#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<EOF
usage: $(basename "$0") [BASE_URL]

Asks the live site the two hand-verified questions and checks the answers.
Default BASE_URL is https://widedoor.cnh.auroq.com.
Exits 1 and prints the answer it got when one is wrong; this is the post-deploy gate's
fallback until testing/acceptance exists.
EOF
}

case "${1:-}" in
  -h|--help) usage; exit 0 ;;
esac
base="${1:-https://widedoor.cnh.auroq.com}"

expect() {
  local q="$1" from="$2" want="$3"
  local got
  got=$(curl -sS -m 20 --get "$base/api/ask" --data-urlencode "q=$q" --data-urlencode "from=$from" \
    | python3 -I -c 'import sys,json; print(json.load(sys.stdin)["summary"])')
  if [[ "$got" == *"$want"* ]]; then
    echo "ok   \"$q\" from $from -> $got"
  else
    echo "FAIL \"$q\" from $from" >&2
    echo "     want: *$want*" >&2
    echo "     got:  $got" >&2
    return 1
  fi
}

code=$(curl -sS -m 20 -o /dev/null -w '%{http_code}' "$base/healthz")
[[ "$code" == "200" ]] || { echo "FAIL healthz returned $code" >&2; exit 1; }
echo "ok   healthz 200"
fail=0
expect "nearest all-gender restroom" "255 BC" "All-gender restroom outside 255: Salt Palace, Level 2" || fail=1
expect "quiet room" "Grand Ballroom J" "Same building, you are on Level 1 and it is on Level 2" || fail=1
exit $fail
