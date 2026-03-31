# MCache

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

## HTTP API Reference

| Method | Path | Description |
|---|---|---|
| `PUT` | `/v1/data` | Create a cache entry (201) |
| `GET` | `/v1/data/:prefix` | Get entry by exact prefix (200 / 404) |
| `POST` | `/v1/data/:prefix` | Update entry data and optional TTL (200 / 404) |
| `DELETE` | `/v1/data/:prefix` | Delete entry (200 / 404) |
| `GET` | `/v1/data/listByPrefix?prefix=` | List direct children under a path |
| `GET` | `/v1/prefix/count` | Count all stored prefixes |
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
