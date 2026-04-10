# MCache

[![pages-build-deployment](https://github.com/mcache-team/mcache/actions/workflows/pages/pages-build-deployment/badge.svg)](https://github.com/mcache-team/mcache/actions/workflows/pages/pages-build-deployment)
![mcache](doc/pic/mcache.png)

A fast, hierarchical in-memory cache for Go. Supports multi-level key paths, TTL expiration, HTTP/gRPC service modes, periodic disk persistence, and direct module embedding — no external dependencies required.

## Background

I built MCache because many of my projects need to cache data in memory with features like hierarchical key lookup, TTL expiration, and type flexibility — similar to Redis — but without the overhead of an external service. The goal was a reusable module I could drop into any project.

Beyond the core use case, I also wanted to push the boundaries: MCache is designed to grow. Planned additions include `HTTP`, `gRPC`, `QUIC`, and `WebSocket` protocol support, a `Watch` mechanism for change notifications, and potentially `raft`-based multi-node synchronization.

## Architecture

![architecture](doc/pic/architecture.png)

Three main layers:

- `Servers` — exposes multiple entry points: embedded module, HTTP REST, gRPC
- `Handlers` — drives the prefix tree operations: insert, get, update, delete, list
- `Storage` — in-memory backend (`sync.Map`) implementing a swappable `Storage` interface; supports periodic snapshot flush to disk

External `Monitor` plugins can be registered to collect metrics and observability data.

## Quick Start

### Embedded module

```go
import (
    "github.com/mcache-team/mcache/pkg/mcache"
    "github.com/mcache-team/mcache/pkg/apis/v1/item"
)

c := mcache.New()

// Insert with TTL
c.Insert("user/profile/name", "alice", item.WithTTL(5*time.Minute))

// Get by exact prefix
it, err := c.Get("user/profile/name")

// List children under a path
items, _ := c.ListByPrefix("user/profile")

// Update and delete
c.Update("user/profile/name", "bob")
c.Delete("user/profile/name")
```

### HTTP service

```bash
# Start
docker run -p 8080:8080 ghcr.io/mcache-team/mcache

# Insert
curl -X PUT http://localhost:8080/v1/data \
  -H 'Content-Type: application/json' \
  -d '{"prefix":"user/name","data":"alice"}'

# Get
curl http://localhost:8080/v1/data/user%2Fname

# Update
curl -X POST http://localhost:8080/v1/data/user%2Fname \
  -d '{"data":"bob"}'

# Delete
curl -X DELETE http://localhost:8080/v1/data/user%2Fname
```

### gRPC service

```go
import grpcclient "github.com/mcache-team/mcache-sdk-go/grpc"

c, _ := grpcclient.New("localhost:9090")
defer c.Close()

c.Insert(ctx, "user/name", "alice", mcache.InsertOption{TTL: time.Minute})
it, _ := c.Get(ctx, "user/name")
```

### Persistence

```bash
# Flush snapshot every 30 seconds to /data
PERSIST_DIR=/data PERSIST_INTERVAL=30s ./mcache
```

On restart, data is automatically restored from `<PERSIST_DIR>/mcache-snapshot.json`. Expired entries are skipped.

### Cluster mode (experimental)

`mcache` now has a first `raft`-backed cluster mode for replicated writes.

Node 1:

```bash
MCACHE_CLUSTER_MODE=raft \
MCACHE_NODE_ID=node-1 \
MCACHE_HTTP_ADDR=0.0.0.0:8081 \
MCACHE_ADVERTISE_ADDR=http://127.0.0.1:8081 \
MCACHE_RAFT_BIND_ADDR=127.0.0.1:7001 \
MCACHE_RAFT_ADVERTISE_ADDR=127.0.0.1:7001 \
MCACHE_RAFT_BOOTSTRAP=true \
MCACHE_CLUSTER_PEERS='node-1@127.0.0.1:7001@http://127.0.0.1:8081,node-2@127.0.0.1:7002@http://127.0.0.1:8082,node-3@127.0.0.1:7003@http://127.0.0.1:8083' \
go run ./pkg
```

Node 2:

```bash
MCACHE_CLUSTER_MODE=raft \
MCACHE_NODE_ID=node-2 \
MCACHE_HTTP_ADDR=0.0.0.0:8082 \
MCACHE_ADVERTISE_ADDR=http://127.0.0.1:8082 \
MCACHE_RAFT_BIND_ADDR=127.0.0.1:7002 \
MCACHE_RAFT_ADVERTISE_ADDR=127.0.0.1:7002 \
MCACHE_RAFT_DATA_DIR=./data/node-2/raft \
MCACHE_CLUSTER_PEERS='node-1@127.0.0.1:7001@http://127.0.0.1:8081,node-2@127.0.0.1:7002@http://127.0.0.1:8082,node-3@127.0.0.1:7003@http://127.0.0.1:8083' \
go run ./pkg
```

Node 3:

```bash
MCACHE_CLUSTER_MODE=raft \
MCACHE_NODE_ID=node-3 \
MCACHE_HTTP_ADDR=0.0.0.0:8083 \
MCACHE_ADVERTISE_ADDR=http://127.0.0.1:8083 \
MCACHE_RAFT_BIND_ADDR=127.0.0.1:7003 \
MCACHE_RAFT_ADVERTISE_ADDR=127.0.0.1:7003 \
MCACHE_RAFT_DATA_DIR=./data/node-3/raft \
MCACHE_CLUSTER_PEERS='node-1@127.0.0.1:7001@http://127.0.0.1:8081,node-2@127.0.0.1:7002@http://127.0.0.1:8082,node-3@127.0.0.1:7003@http://127.0.0.1:8083' \
go run ./pkg
```

Notes:

- Only one node should start with `MCACHE_RAFT_BOOTSTRAP=true`.
- `MCACHE_CLUSTER_PEERS` uses the format `nodeID@raftAddr@httpAddr`.
- Writes must go to the leader. Followers will respond with `307 Temporary Redirect`.
- Cluster status is exposed at `GET /v1/cluster/status`.
- Dynamic membership is available from the leader:

```bash
# list members
curl http://127.0.0.1:8081/v1/cluster/nodes

# add a voter
curl -X POST http://127.0.0.1:8081/v1/cluster/nodes \
  -H 'Content-Type: application/json' \
  -d '{"nodeId":"node-4","raftAddress":"127.0.0.1:7004","advertiseAddress":"http://127.0.0.1:8084"}'

# remove a member
curl -X DELETE http://127.0.0.1:8081/v1/cluster/nodes/node-4
```

- This is the current cluster milestone: replicated writes + snapshot/restore + leader redirects + dynamic voter membership. Shard routing and follower read optimizations are not implemented yet.

### Cluster smoke test

You can run the local raft smoke test with Docker:

```bash
bash e2e/raft-start.sh
```

Or through `make`:

```bash
make e2e-raft
```

The smoke test boots a 3-node raft cluster, verifies leader election, checks replicated writes on all nodes, confirms follower redirects, starts a fourth node, joins it dynamically, removes it again, stops the current leader to verify failover, then starts that old leader back up and confirms it catches up and rejoins as a follower.

### Quorum test

You can also verify quorum behavior with:

```bash
make e2e-raft-quorum
```

This test boots the 3-node raft cluster, stops one follower and confirms the remaining majority is still writable, then stops the second follower and confirms the isolated node can no longer complete writes, and finally brings the cluster back to full size and verifies writes succeed again.

### Rolling restart test

To verify node-by-node restart recovery:

```bash
make e2e-raft-rolling
```

This test performs a rolling restart across all three raft nodes. After each restart it waits for the node to become healthy again, checks that the cluster still has a leader, writes a fresh value through the current leader, and verifies every node can read both the old and new data.

### Health and metrics

`mcache` now exposes lightweight observability endpoints:

```bash
# liveness
curl http://127.0.0.1:8080/livez

# readiness
curl http://127.0.0.1:8080/readyz

# cluster diagnostics as JSON
curl http://127.0.0.1:8080/v1/cluster/diagnostics

# Prometheus-style metrics
curl http://127.0.0.1:8080/metrics
```

The diagnostics payload includes node role, readiness, member list, cache item/root counts, and raw raft stats when running in cluster mode.

The Prometheus-style metrics output now includes write-path counters and latency totals such as:

- `mcache_write_requests_total{operation=...}`
- `mcache_write_success_total{operation=...}`
- `mcache_write_error_total{operation=...}`
- `mcache_write_redirect_total{operation=...}`
- `mcache_write_latency_seconds_bucket{operation=...,le=...}`
- `mcache_write_latency_seconds_sum{operation=...}`
- `mcache_write_latency_seconds_count{operation=...}`
- `mcache_write_latency_last_seconds{operation=...}`

The `/metrics` endpoint also emits standard Prometheus `# HELP` and `# TYPE` metadata, so it can be scraped directly without a sidecar reformatter.

Example Prometheus scrape config:

```yaml
scrape_configs:
  - job_name: mcache
    static_configs:
      - targets:
          - 127.0.0.1:8081
          - 127.0.0.1:8082
          - 127.0.0.1:8083
```

Example PromQL queries:

```promql
# p95 create latency over 5 minutes
histogram_quantile(
  0.95,
  sum by (le) (rate(mcache_write_latency_seconds_bucket{operation="data_create"}[5m]))
)

# redirect rate over 5 minutes
sum(rate(mcache_write_redirect_total[5m]))

# error rate over 5 minutes
sum(rate(mcache_write_error_total[5m]))

# current cached item count per node
mcache_state_items_total
```

## HTTP API Reference

| Method | Path | Description |
|---|---|---|
| `PUT` | `/v1/data` | Create a cache entry (201) |
| `GET` | `/v1/data/:prefix` | Get entry by exact prefix (200 / 404) |
| `POST` | `/v1/data/:prefix` | Update entry data and optional TTL (200 / 404) |
| `DELETE` | `/v1/data/:prefix` | Delete entry (200 / 404) |
| `GET` | `/v1/data/listByPrefix?prefix=` | List direct children under a path |
| `GET` | `/v1/prefix/count` | Count all stored prefixes |
| `GET` | `/livez` | Liveness probe |
| `GET` | `/readyz` | Readiness probe |
| `GET` | `/metrics` | Prometheus-style metrics |
| `GET` | `/v1/cluster/status` | Show node mode, leader and advertised addresses |
| `GET` | `/v1/cluster/diagnostics` | Show readiness, members, state counters and raft stats |
| `GET` | `/v1/cluster/nodes` | List cluster members |
| `POST` | `/v1/cluster/nodes` | Add a voter node on the leader |
| `DELETE` | `/v1/cluster/nodes/:nodeId` | Remove a cluster member on the leader |
| `GET` | `/healthz` | Health check |

## Benchmark

Run benchmarks yourself:

```bash
go test ./pkg/storage/memory/ -bench=. -benchmem -benchtime=3s
```

Results on Apple M2 (darwin/arm64, Go 1.21):

| Benchmark | ops/sec | ns/op | B/op | allocs/op |
|---|---|---|---|---|
| `Insert` | 3,200,000 | 312 | 288 | 4 |
| `InsertWithTTL` | 2,900,000 | 345 | 304 | 5 |
| `InsertParallel` | 9,500,000 | 126 | 291 | 4 |
| `Get` | 18,000,000 | 66 | 0 | 0 |
| `GetParallel` | 52,000,000 | 23 | 0 | 0 |
| `Update` | 8,500,000 | 141 | 0 | 0 |
| `Delete` | 4,100,000 | 244 | 48 | 2 |
| `ListPrefix (100 keys)` | 1,200,000 | 832 | 896 | 3 |
| `ListPrefix (1000 keys)` | 130,000 | 7,810 | 8,192 | 3 |
| `MixedReadWrite` | 14,000,000 | 85 | 62 | 1 |

- **Get is the fastest path** — lock-free `sync.Map` reads hit ~18M ops/sec single-threaded, ~52M ops/sec parallel.
- **Insert is the bottleneck** — a write lock on `prefixList` after `LoadOrStore` caps throughput at ~3M ops/sec.
- **ListPrefix scales linearly** — full scan of `prefixList` under a read lock; cost grows with key count.

## Contributing

Issues and ideas are welcome. Feel free to open an issue or reach out directly.

- Email: [EvansChang](https://github.com/AlpherJang) — alphejangs@gmail.com
- Twitter: [@EvansJang](https://twitter.com/EvansJang)
- WeChat: evanxtay

## License

[MIT](LICENSE)
