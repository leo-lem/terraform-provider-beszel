#!/usr/bin/env bash
# Tears down the Beszel hub started by testacc-up.sh, including its data volume.
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."

docker compose down -v
