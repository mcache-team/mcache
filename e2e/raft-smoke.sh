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
  printf '[raft-e2e] %s\n' "$*"
}

request() {
  local method="$1"
  local url="$2"
  local body="${3:-}"
  if [[ -n "${body}" ]]; then
    curl -sS -X "${method}" "${url}" -H 'Content-Type: application/json' -d "${body}"
  else
    curl -sS -X "${method}" "${url}"
  fi
}

request_code() {
  local method="$1"
  local url="$2"
  local body="${3:-}"
  if [[ -n "${body}" ]]; then
    curl -sS -o /tmp/mcache-e2e-response.txt -w '%{http_code}' -X "${method}" "${url}" -H 'Content-Type: application/json' -d "${body}"
  else
    curl -sS -o /tmp/mcache-e2e-response.txt -w '%{http_code}' -X "${method}" "${url}"
  fi
}

wait_for_http() {
  local url="$1"
  for _ in $(seq 1 60); do
    if curl -fsS "${url}" >/dev/null 2>&1; then
      return 0
    fi
    sleep 2
  done
  echo "timed out waiting for ${url}" >&2
  return 1
}

find_leader() {
  local ports=(18081 18082 18083)
  for _ in $(seq 1 60); do
    for port in "${ports[@]}"; do
      local body
      body="$(curl -fsS "http://127.0.0.1:${port}/v1/cluster/status" 2>/dev/null || true)"
      if [[ "${body}" == *'"isLeader":true'* ]]; then
        echo "http://127.0.0.1:${port}"
        return 0
      fi
    done
    sleep 2
  done
  return 1
}

find_leader_excluding() {
  local excluded_url="$1"
  local ports=(18081 18082 18083)
  for _ in $(seq 1 60); do
    for port in "${ports[@]}"; do
      local candidate="http://127.0.0.1:${port}"
      if [[ "${candidate}" == "${excluded_url}" ]]; then
        continue
      fi
      local body
      body="$(curl -fsS "${candidate}/v1/cluster/status" 2>/dev/null || true)"
      if [[ "${body}" == *'"isLeader":true'* ]]; then
        echo "${candidate}"
        return 0
      fi
    done
    sleep 2
  done
  return 1
}

leader_service_name() {
  case "$1" in
    http://127.0.0.1:18081) echo "mcache1" ;;
    http://127.0.0.1:18082) echo "mcache2" ;;
    http://127.0.0.1:18083) echo "mcache3" ;;
    *)
      echo "unknown leader url: $1" >&2
      exit 1
      ;;
  esac
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
    cat /tmp/mcache-e2e-response.txt >&2 || true
    exit 1
  fi
}

log "starting 3-node raft cluster"
"${COMPOSE[@]}" up -d --build mcache1 mcache2 mcache3

wait_for_http "http://127.0.0.1:18081/healthz"
wait_for_http "http://127.0.0.1:18082/healthz"
wait_for_http "http://127.0.0.1:18083/healthz"

LEADER="$(find_leader)"
log "leader elected at ${LEADER}"

NODES_BODY="$(request GET "${LEADER}/v1/cluster/nodes")"
assert_contains "${NODES_BODY}" '"nodeId":"node-1"'
assert_contains "${NODES_BODY}" '"nodeId":"node-2"'
assert_contains "${NODES_BODY}" '"nodeId":"node-3"'

log "writing replicated value through leader"
CODE="$(request_code PUT "${LEADER}/v1/data" '{"prefix":"cluster/key","data":"value-from-leader"}')"
assert_code "201" "${CODE}"

for port in 18081 18082 18083; do
  wait_for_http "http://127.0.0.1:${port}/healthz"
  BODY="$(request GET "http://127.0.0.1:${port}/v1/data/cluster%2Fkey")"
  assert_contains "${BODY}" '"data":"value-from-leader"'
done

FOLLOWER="http://127.0.0.1:18081"
if [[ "${FOLLOWER}" == "${LEADER}" ]]; then
  FOLLOWER="http://127.0.0.1:18082"
fi

log "verifying follower write redirect"
CODE="$(request_code PUT "${FOLLOWER}/v1/data" '{"prefix":"cluster/redirect","data":"should-redirect"}')"
assert_code "307" "${CODE}"

log "starting joiner node"
"${COMPOSE[@]}" --profile joiner up -d mcache4
wait_for_http "http://127.0.0.1:18084/healthz"

log "adding node-4 via leader"
CODE="$(request_code POST "${LEADER}/v1/cluster/nodes" '{"nodeId":"node-4","raftAddress":"mcache4:7004","advertiseAddress":"http://127.0.0.1:18084"}')"
assert_code "202" "${CODE}"

for _ in $(seq 1 60); do
  BODY="$(request GET "${LEADER}/v1/cluster/nodes")"
  if [[ "${BODY}" == *'"nodeId":"node-4"'* ]]; then
    break
  fi
  sleep 2
done
assert_contains "${BODY}" '"nodeId":"node-4"'

for _ in $(seq 1 60); do
  BODY="$(curl -fsS "http://127.0.0.1:18084/v1/data/cluster%2Fkey" 2>/dev/null || true)"
  if [[ "${BODY}" == *'"data":"value-from-leader"'* ]]; then
    break
  fi
  sleep 2
done
assert_contains "${BODY}" '"data":"value-from-leader"'

log "removing node-4 via leader"
CODE="$(request_code DELETE "${LEADER}/v1/cluster/nodes/node-4")"
assert_code "202" "${CODE}"

for _ in $(seq 1 60); do
  BODY="$(request GET "${LEADER}/v1/cluster/nodes")"
  if [[ "${BODY}" != *'"nodeId":"node-4"'* ]]; then
    break
  fi
  sleep 2
done
if [[ "${BODY}" == *'"nodeId":"node-4"'* ]]; then
  echo "node-4 still present after removal: ${BODY}" >&2
  exit 1
fi

OLD_LEADER="${LEADER}"
OLD_LEADER_SERVICE="$(leader_service_name "${OLD_LEADER}")"

log "stopping leader ${OLD_LEADER_SERVICE} to verify failover"
"${COMPOSE[@]}" stop "${OLD_LEADER_SERVICE}" >/dev/null

NEW_LEADER="$(find_leader_excluding "${OLD_LEADER}")"
if [[ -z "${NEW_LEADER}" ]]; then
  echo "no new leader elected after stopping ${OLD_LEADER_SERVICE}" >&2
  exit 1
fi
log "new leader elected at ${NEW_LEADER}"

CODE="$(request_code PUT "${NEW_LEADER}/v1/data" '{"prefix":"cluster/failover","data":"after-leader-stop"}')"
assert_code "201" "${CODE}"

for port in 18081 18082 18083; do
  if [[ "http://127.0.0.1:${port}" == "${OLD_LEADER}" ]]; then
    continue
  fi
  BODY="$(request GET "http://127.0.0.1:${port}/v1/data/cluster%2Ffailover")"
  assert_contains "${BODY}" '"data":"after-leader-stop"'
done

log "restarting old leader ${OLD_LEADER_SERVICE} to verify recovery"
"${COMPOSE[@]}" start "${OLD_LEADER_SERVICE}" >/dev/null
wait_for_http "${OLD_LEADER}/healthz"

for _ in $(seq 1 60); do
  BODY="$(curl -fsS "${OLD_LEADER}/v1/data/cluster%2Ffailover" 2>/dev/null || true)"
  if [[ "${BODY}" == *'"data":"after-leader-stop"'* ]]; then
    break
  fi
  sleep 2
done
assert_contains "${BODY}" '"data":"after-leader-stop"'

RECOVERED_STATUS="$(request GET "${OLD_LEADER}/v1/cluster/status")"
assert_contains "${RECOVERED_STATUS}" '"nodeId":"'"${OLD_LEADER_SERVICE/mcache/node-}"'"'

BODY="$(request GET "${NEW_LEADER}/v1/cluster/nodes")"
assert_contains "${BODY}" "\"nodeId\":\"${OLD_LEADER_SERVICE/mcache/node-}\""

log "raft cluster smoke test passed"
