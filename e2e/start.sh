#!/bin/bash
set -euo pipefail

if docker compose version >/dev/null 2>&1; then
  COMPOSE=(docker compose)
elif docker-compose version >/dev/null 2>&1; then
  COMPOSE=(docker-compose)
else
  echo "docker compose or docker-compose is required" >&2
  exit 1
fi

"${COMPOSE[@]}" version
"${COMPOSE[@]}" up --build --exit-code-from testing
