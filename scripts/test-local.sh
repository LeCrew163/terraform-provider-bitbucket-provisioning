#!/usr/bin/env bash
# test-local.sh — end-to-end test of the Terraform provider against a local
#                 Bitbucket Data Center instance running in Docker Compose.
#
# Usage:
#   # Start fresh (build, spin up Docker, run Terraform, tear down):
#   ./scripts/test-local.sh
#
#   # Skip Docker startup (instance already running):
#   SKIP_DOCKER=true ./scripts/test-local.sh
#
#   # Skip Terraform destroy at the end (keep resources for inspection):
#   SKIP_DESTROY=true ./scripts/test-local.sh
#
# Environment variables (all optional — defaults shown):
#   BITBUCKET_BASE_URL  http://localhost:7990
#   BITBUCKET_USERNAME  admin
#   BITBUCKET_PASSWORD  admin
#   SKIP_DOCKER         false
#   SKIP_DESTROY        false

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
TESTS_DIR="${ROOT_DIR}/tests/terraform"
ENV_FILE="${ROOT_DIR}/.env.local"

# ── helpers ────────────────────────────────────────────────────────────────
step() { echo ""; echo "──────────────────────────────────────────────"; echo "  $*"; echo "──────────────────────────────────────────────"; }
die()  { echo "ERROR: $*" >&2; exit 1; }

# ── load .env.local if present ─────────────────────────────────────────────
if [ -f "${ENV_FILE}" ]; then
  echo "Loading environment from ${ENV_FILE} ..."
  # shellcheck source=/dev/null
  set -a; source "${ENV_FILE}"; set +a
fi

# ── check required tools ───────────────────────────────────────────────────
for cmd in docker terraform go curl python3; do
  command -v "${cmd}" &>/dev/null || die "Required command '${cmd}' not found"
done

# ── defaults ───────────────────────────────────────────────────────────────
export BITBUCKET_BASE_URL="${BITBUCKET_BASE_URL:-http://localhost:7990}"
export BITBUCKET_USERNAME="${BITBUCKET_USERNAME:-admin}"
export BITBUCKET_PASSWORD="${BITBUCKET_PASSWORD:-admin}"
SKIP_DOCKER="${SKIP_DOCKER:-false}"
SKIP_DESTROY="${SKIP_DESTROY:-false}"

echo ""
echo "╔══════════════════════════════════════════════════════╗"
echo "║   Terraform Provider Bitbucket DC — Local Tests      ║"
echo "╚══════════════════════════════════════════════════════╝"
echo "  Bitbucket URL : ${BITBUCKET_BASE_URL}"
echo "  Username      : ${BITBUCKET_USERNAME}"
echo "  Skip Docker   : ${SKIP_DOCKER}"
echo "  Skip Destroy  : ${SKIP_DESTROY}"

# ── step 1: docker compose ─────────────────────────────────────────────────
if [ "${SKIP_DOCKER}" != "true" ]; then
  step "1/5  Starting Docker Compose (Bitbucket + PostgreSQL)"
  cd "${ROOT_DIR}"

  if [ -z "${BITBUCKET_LICENSE:-}" ]; then
    echo ""
    echo "  ⚠  BITBUCKET_LICENSE is not set."
    echo "     Bitbucket will start but requires manual setup via the web UI."
    echo "     Get a free developer license at:"
    echo "     https://developer.atlassian.com/platform/marketplace/timebomb-licenses-for-testing-server-apps/"
    echo ""
  fi

  docker compose --env-file "${ENV_FILE:-/dev/null}" up -d
else
  step "1/5  Skipping Docker Compose (SKIP_DOCKER=true)"
fi

# ── step 2: wait for bitbucket ─────────────────────────────────────────────
step "2/5  Waiting for Bitbucket to be ready"
"${SCRIPT_DIR}/wait-for-bitbucket.sh" "${BITBUCKET_BASE_URL}" 60

# ── step 3: build + install provider ──────────────────────────────────────
step "3/5  Building and installing Terraform provider"
cd "${ROOT_DIR}"
make install

# ── step 4: terraform tests ────────────────────────────────────────────────
step "4/5  Running Terraform configuration tests"
cd "${TESTS_DIR}"

# Unset any token that may be set for a different Bitbucket instance so the
# provider doesn't see conflicting auth methods (username+password vs token).
unset BITBUCKET_TOKEN

# Remove the lock file so a freshly rebuilt binary (new checksum) doesn't fail.
rm -f .terraform.lock.hcl
terraform init -reconfigure

echo ""
echo "  → terraform plan"
terraform plan -out=tfplan

echo ""
echo "  → terraform apply"
terraform apply -auto-approve tfplan

echo ""
echo "  → terraform show (current state)"
terraform show

echo ""
echo "  → terraform plan (drift check — expect: No changes)"
if terraform plan -detailed-exitcode 2>&1; then
  echo "  ✓ No drift detected"
else
  EXIT=$?
  if [ "${EXIT}" -eq 2 ]; then
    echo "  ⚠  Drift detected — provider state does not match Bitbucket"
    exit 1
  fi
fi

# ── step 5: destroy ────────────────────────────────────────────────────────
if [ "${SKIP_DESTROY}" != "true" ]; then
  step "5/5  Destroying test resources"
  terraform destroy -auto-approve
  echo "  ✓ Resources destroyed"
else
  step "5/5  Skipping destroy (SKIP_DESTROY=true)"
  echo "  Resources left in place. To clean up manually:"
  echo "    cd ${TESTS_DIR} && terraform destroy -auto-approve"
fi

echo ""
echo "╔══════════════════════════════════════════════════════╗"
echo "║   All tests passed successfully!                     ║"
echo "╚══════════════════════════════════════════════════════╝"
echo ""
echo "To stop Docker containers and remove volumes:"
echo "  docker compose down -v"
