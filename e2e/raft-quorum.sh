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
  printf '[raft-quorum] %s\n' "$*"
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
    curl -sS --max-time 15 -o /tmp/mcache-quorum-response.txt -w '%{http_code}' -X "${method}" "${url}" -H 'Content-Type: application/json' -d "${body}"
  else
    curl -sS --max-time 15 -o /tmp/mcache-quorum-response.txt -w '%{http_code}' -X "${method}" "${url}"
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
    cat /tmp/mcache-quorum-response.txt >&2 || true
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

service_name() {
  case "$1" in
    http://127.0.0.1:18081) echo "mcache1" ;;
    http://127.0.0.1:18082) echo "mcache2" ;;
    http://127.0.0.1:18083) echo "mcache3" ;;
    *)
      echo "unknown service url: $1" >&2
      exit 1
      ;;
  esac
}

all_urls=(http://127.0.0.1:18081 http://127.0.0.1:18082 http://127.0.0.1:18083)

log "starting 3-node raft cluster"
"${COMPOSE[@]}" up -d --build mcache1 mcache2 mcache3

for url in "${all_urls[@]}"; do
  wait_for_http "${url}/healthz"
done

LEADER="$(find_leader)"
if [[ -z "${LEADER}" ]]; then
  echo "leader was not elected" >&2
  exit 1
fi
log "leader elected at ${LEADER}"

FOLLOWERS=()
for url in "${all_urls[@]}"; do
  if [[ "${url}" != "${LEADER}" ]]; then
    FOLLOWERS+=("${url}")
  fi
done

FIRST_FOLLOWER="${FOLLOWERS[0]}"
SECOND_FOLLOWER="${FOLLOWERS[1]}"
FIRST_FOLLOWER_SERVICE="$(service_name "${FIRST_FOLLOWER}")"
SECOND_FOLLOWER_SERVICE="$(service_name "${SECOND_FOLLOWER}")"

log "writing baseline value"
CODE="$(request_code PUT "${LEADER}/v1/data" '{"prefix":"cluster/quorum/base","data":"baseline"}')"
assert_code "201" "${CODE}"

for url in "${all_urls[@]}"; do
  BODY="$(request GET "${url}/v1/data/cluster%2Fquorum%2Fbase")"
  assert_contains "${BODY}" '"data":"baseline"'
done

log "stopping one follower ${FIRST_FOLLOWER_SERVICE}; majority should still be writable"
"${COMPOSE[@]}" stop "${FIRST_FOLLOWER_SERVICE}" >/dev/null
sleep 5

CODE="$(request_code PUT "${LEADER}/v1/data" '{"prefix":"cluster/quorum/majority","data":"two-of-three"}')"
assert_code "201" "${CODE}"

BODY="$(request GET "${LEADER}/v1/data/cluster%2Fquorum%2Fmajority")"
assert_contains "${BODY}" '"data":"two-of-three"'
BODY="$(request GET "${SECOND_FOLLOWER}/v1/data/cluster%2Fquorum%2Fmajority")"
assert_contains "${BODY}" '"data":"two-of-three"'

log "stopping second follower ${SECOND_FOLLOWER_SERVICE}; single node should lose quorum"
"${COMPOSE[@]}" stop "${SECOND_FOLLOWER_SERVICE}" >/dev/null

WRITE_FAILED=0
for _ in $(seq 1 20); do
  CODE="$(request_code PUT "${LEADER}/v1/data" '{"prefix":"cluster/quorum/no-majority","data":"should-fail"}' || true)"
  if [[ "${CODE}" != "201" ]]; then
    WRITE_FAILED=1
    break
  fi
  sleep 3
done

if [[ "${WRITE_FAILED}" != "1" ]]; then
  echo "isolated node unexpectedly kept accepting quorum writes" >&2
  exit 1
fi

log "restoring quorum by starting stopped followers"
"${COMPOSE[@]}" start "${FIRST_FOLLOWER_SERVICE}" "${SECOND_FOLLOWER_SERVICE}" >/dev/null
wait_for_http "${FIRST_FOLLOWER}/healthz"
wait_for_http "${SECOND_FOLLOWER}/healthz"

NEW_LEADER="$(find_leader)"
if [[ -z "${NEW_LEADER}" ]]; then
  echo "no leader elected after quorum restore" >&2
  exit 1
fi
log "leader after recovery is ${NEW_LEADER}"

CODE="$(request_code PUT "${NEW_LEADER}/v1/data" '{"prefix":"cluster/quorum/recovered","data":"quorum-restored"}')"
assert_code "201" "${CODE}"

for url in "${all_urls[@]}"; do
  BODY="$(request GET "${url}/v1/data/cluster%2Fquorum%2Frecovered")"
  assert_contains "${BODY}" '"data":"quorum-restored"'
done

log "raft quorum test passed"
