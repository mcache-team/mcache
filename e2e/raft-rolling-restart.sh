#!/bin/bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="${ROOT_DIR}/e2e/compose.raft.yaml"
KEEP_CLUSTER="${KEEP_CLUSTER:-0}"

if docker compose version >/dev/null 2>&1; then
  COMPOSE=(docker compose -f "${COMPOSE_FILE}")
elif docker-compose version >/dev/null 2>&1; then
  COMPOSE=(docker-compose -f "${COMPOSE_FILE}")
else
  echo "docker compose or docker-compose is required" >&2
  exit 1
fi

cleanup() {
  if [[ "${KEEP_CLUSTER}" == "1" ]]; then
    return
  fi
  "${COMPOSE[@]}" down -v --remove-orphans >/dev/null 2>&1 || true
}
trap cleanup EXIT

log() {
  printf '[raft-rolling] %s\n' "$*"
}

request() {
  local method="$1"
  local url="$2"
  local body="${3:-}"
  if [[ -n "${body}" ]]; then
    curl -sS --max-time 15 -X "${method}" "${url}" -H 'Content-Type: application/json' -d "${body}"
  else
    curl -sS --max-time 15 -X "${method}" "${url}"
  fi
}

request_code() {
  local method="$1"
  local url="$2"
  local body="${3:-}"
  if [[ -n "${body}" ]]; then
    curl -sS --max-time 15 -o /tmp/mcache-rolling-response.txt -w '%{http_code}' -X "${method}" "${url}" -H 'Content-Type: application/json' -d "${body}"
  else
    curl -sS --max-time 15 -o /tmp/mcache-rolling-response.txt -w '%{http_code}' -X "${method}" "${url}"
  fi
}

wait_for_http() {
  local url="$1"
  for _ in $(seq 1 60); do
    if curl -fsS --max-time 5 "${url}" >/dev/null 2>&1; then
      return 0
    fi
    sleep 2
  done
  echo "timed out waiting for ${url}" >&2
  return 1
}

assert_contains() {
  local haystack="$1"
  local needle="$2"
  if [[ "${haystack}" != *"${needle}"* ]]; then
    echo "expected response to contain ${needle}, got: ${haystack}" >&2
    exit 1
  fi
}

assert_code() {
  local expected="$1"
  local actual="$2"
  if [[ "${expected}" != "${actual}" ]]; then
    echo "expected status ${expected}, got ${actual}" >&2
    cat /tmp/mcache-rolling-response.txt >&2 || true
    exit 1
  fi
}

find_leader() {
  local ports=(18081 18082 18083)
  for _ in $(seq 1 60); do
    for port in "${ports[@]}"; do
      local body
      body="$(curl -fsS --max-time 5 "http://127.0.0.1:${port}/v1/cluster/status" 2>/dev/null || true)"
      if [[ "${body}" == *'"isLeader":true'* ]]; then
        echo "http://127.0.0.1:${port}"
        return 0
      fi
    done
    sleep 2
  done
  return 1
}

url_for_service() {
  case "$1" in
    mcache1) echo "http://127.0.0.1:18081" ;;
    mcache2) echo "http://127.0.0.1:18082" ;;
    mcache3) echo "http://127.0.0.1:18083" ;;
    *)
      echo "unknown service: $1" >&2
      exit 1
      ;;
  esac
}

services=(mcache1 mcache2 mcache3)

log "starting 3-node raft cluster"
"${COMPOSE[@]}" up -d --build "${services[@]}"

for service in "${services[@]}"; do
  wait_for_http "$(url_for_service "${service}")/healthz"
done

LEADER="$(find_leader)"
if [[ -z "${LEADER}" ]]; then
  echo "leader was not elected" >&2
  exit 1
fi
log "initial leader elected at ${LEADER}"

CODE="$(request_code PUT "${LEADER}/v1/data" '{"prefix":"cluster/rolling/base","data":"before-restart"}')"
assert_code "201" "${CODE}"

for service in "${services[@]}"; do
  SERVICE_URL="$(url_for_service "${service}")"
  BODY="$(request GET "${SERVICE_URL}/v1/data/cluster%2Frolling%2Fbase")"
  assert_contains "${BODY}" '"data":"before-restart"'
done

step=1
for service in "${services[@]}"; do
  SERVICE_URL="$(url_for_service "${service}")"
  log "restarting ${service}"
  "${COMPOSE[@]}" restart "${service}" >/dev/null
  wait_for_http "${SERVICE_URL}/healthz"

  LEADER="$(find_leader)"
  if [[ -z "${LEADER}" ]]; then
    echo "leader was not elected after restarting ${service}" >&2
    exit 1
  fi
  log "leader after restarting ${service}: ${LEADER}"

  CODE="$(request_code PUT "${LEADER}/v1/data" "{\"prefix\":\"cluster/rolling/step-${step}\",\"data\":\"after-${service}\"}")"
  assert_code "201" "${CODE}"

  for verify_service in "${services[@]}"; do
    VERIFY_URL="$(url_for_service "${verify_service}")"
    wait_for_http "${VERIFY_URL}/healthz"
    BODY="$(request GET "${VERIFY_URL}/v1/data/cluster%2Frolling%2Fbase")"
    assert_contains "${BODY}" '"data":"before-restart"'
    BODY="$(request GET "${VERIFY_URL}/v1/data/cluster%2Frolling%2Fstep-'"${step}"'")"
    assert_contains "${BODY}" "\"data\":\"after-${service}\""
  done

  step=$((step + 1))
done

BODY="$(request GET "${LEADER}/v1/cluster/nodes")"
assert_contains "${BODY}" '"nodeId":"node-1"'
assert_contains "${BODY}" '"nodeId":"node-2"'
assert_contains "${BODY}" '"nodeId":"node-3"'

log "raft rolling restart test passed"
