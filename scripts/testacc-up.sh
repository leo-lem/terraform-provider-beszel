#!/usr/bin/env bash
# Starts a Beszel hub via docker-compose.yml, provisions a superuser and a
# regular test user, and prints the env vars acceptance tests need.
#
# Usage:
#   eval "$(./scripts/testacc-up.sh)"
#   TF_ACC=1 go test ./... -v
#
# Requires: docker compose, curl, python3.
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."

ENDPOINT="http://localhost:8090"
SUPERUSER_EMAIL="admin@example.com"
SUPERUSER_PASSWORD="acceptance-test-password-1"
TEST_USER_EMAIL="testacc-owner@example.com"
TEST_USER_PASSWORD="acceptance-test-password-1"

log() { echo "$*" >&2; }

log "Starting Beszel hub..."
docker compose up -d --wait beszel

log "Provisioning superuser..."
docker compose exec -T beszel /beszel superuser upsert "$SUPERUSER_EMAIL" "$SUPERUSER_PASSWORD" >&2

SUPERUSER_TOKEN=$(curl -sf -X POST "$ENDPOINT/api/collections/_superusers/auth-with-password" \
  -H "Content-Type: application/json" \
  -d "{\"identity\":\"$SUPERUSER_EMAIL\",\"password\":\"$SUPERUSER_PASSWORD\"}" \
  | python3 -c 'import sys, json; print(json.load(sys.stdin)["token"])')

log "Looking up test user..."
EXISTING=$(curl -sf "$ENDPOINT/api/collections/users/records?filter=email%3D%22$TEST_USER_EMAIL%22" \
  -H "Authorization: $SUPERUSER_TOKEN")
TEST_USER_ID=$(echo "$EXISTING" | python3 -c 'import sys, json; items = json.load(sys.stdin)["items"]; print(items[0]["id"] if items else "")')

if [ -z "$TEST_USER_ID" ]; then
  log "Creating test user..."
  TEST_USER_ID=$(curl -sf -X POST "$ENDPOINT/api/collections/users/records" \
    -H "Authorization: $SUPERUSER_TOKEN" -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_USER_EMAIL\",\"password\":\"$TEST_USER_PASSWORD\",\"passwordConfirm\":\"$TEST_USER_PASSWORD\",\"role\":\"user\",\"verified\":true}" \
    | python3 -c 'import sys, json; print(json.load(sys.stdin)["id"])')
fi

log "Ready."
cat <<EOF
export BESZEL_ENDPOINT="$ENDPOINT"
export BESZEL_EMAIL="$SUPERUSER_EMAIL"
export BESZEL_PASSWORD="$SUPERUSER_PASSWORD"
export BESZEL_TEST_USER_ID="$TEST_USER_ID"
EOF
