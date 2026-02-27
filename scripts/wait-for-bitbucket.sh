#!/usr/bin/env bash
# wait-for-bitbucket.sh — poll Bitbucket's /status endpoint until it reports RUNNING.
#
# Usage:
#   ./scripts/wait-for-bitbucket.sh [URL] [MAX_ATTEMPTS]
#
# Arguments:
#   URL           Base URL of the Bitbucket instance (default: http://localhost:7990)
#   MAX_ATTEMPTS  Number of polling attempts before giving up (default: 60, ~10 min at 10s)

set -euo pipefail

BITBUCKET_URL="${1:-http://localhost:7990}"
MAX_ATTEMPTS="${2:-60}"
SLEEP_SECONDS=10

echo "Waiting for Bitbucket at ${BITBUCKET_URL} ..."

for i in $(seq 1 "${MAX_ATTEMPTS}"); do
  HTTP_CODE=$(curl -sk -o /dev/null -w "%{http_code}" "${BITBUCKET_URL}/status" 2>/dev/null || echo "000")

  if [ "${HTTP_CODE}" = "200" ]; then
    STATE=$(curl -sk "${BITBUCKET_URL}/status" \
      | python3 -c "import sys,json; print(json.load(sys.stdin).get('state','UNKNOWN'))" 2>/dev/null \
      || echo "UNKNOWN")

    if [ "${STATE}" = "RUNNING" ]; then
      echo "Bitbucket is ready (attempt ${i}/${MAX_ATTEMPTS})"
      exit 0
    fi

    echo "  Attempt ${i}/${MAX_ATTEMPTS}: state=${STATE} (waiting for RUNNING)..."
  else
    echo "  Attempt ${i}/${MAX_ATTEMPTS}: HTTP ${HTTP_CODE} — not ready yet..."
  fi

  sleep "${SLEEP_SECONDS}"
done

echo ""
echo "ERROR: Bitbucket did not become ready within $((MAX_ATTEMPTS * SLEEP_SECONDS)) seconds."
echo "       Check the logs: docker compose logs bitbucket"
exit 1
