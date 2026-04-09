package cluster

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	defaultHTTPAddr = "0.0.0.0:8080"
	defaultNodeID   = "mcache-1"
)

type Mode string

const (
	ModeSingle Mode = "single"
	ModeRaft   Mode = "raft"
)

type Config struct {
	Mode          Mode
	NodeID        string
	HTTPAddr      string
	AdvertiseAddr string
	RaftBindAddr  string
	RaftAdvertise string
	RaftDataDir   string
	RaftBootstrap bool
	RaftPeers     []Peer
	ApplyTimeout  time.Duration
}

func ConfigFromEnv() Config {
	httpAddr := envOrDefault("MCACHE_HTTP_ADDR", defaultHTTPAddr)
	mode := Mode(strings.ToLower(envOrDefault("MCACHE_CLUSTER_MODE", string(ModeSingle))))
	peers, err := parsePeers(envOrDefault("MCACHE_CLUSTER_PEERS", ""))
	if err != nil {
		panic(err)
	}
	return Config{
		Mode:          mode,
		NodeID:        envOrDefault("MCACHE_NODE_ID", defaultNodeID),
		HTTPAddr:      httpAddr,
		AdvertiseAddr: normalizeAdvertiseAddr(envOrDefault("MCACHE_ADVERTISE_ADDR", ""), httpAddr),
		RaftBindAddr:  envOrDefault("MCACHE_RAFT_BIND_ADDR", "127.0.0.1:7000"),
		RaftAdvertise: envOrDefault("MCACHE_RAFT_ADVERTISE_ADDR", envOrDefault("MCACHE_RAFT_BIND_ADDR", "127.0.0.1:7000")),
		RaftDataDir:   envOrDefault("MCACHE_RAFT_DATA_DIR", fmt.Sprintf("./data/%s/raft", envOrDefault("MCACHE_NODE_ID", defaultNodeID))),
		RaftBootstrap: strings.EqualFold(envOrDefault("MCACHE_RAFT_BOOTSTRAP", "false"), "true"),
		RaftPeers:     peers,
		ApplyTimeout:  parseDurationOrDefault(envOrDefault("MCACHE_RAFT_APPLY_TIMEOUT", "5s"), 5*time.Second),
	}
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func normalizeAdvertiseAddr(advertiseAddr, httpAddr string) string {
	if advertiseAddr != "" {
		return advertiseAddr
	}
	host := strings.TrimSpace(httpAddr)
	switch {
	case strings.HasPrefix(host, "0.0.0.0:"):
		host = "127.0.0.1:" + strings.TrimPrefix(host, "0.0.0.0:")
	case strings.HasPrefix(host, ":"):
		host = "127.0.0.1" + host
	}
	return fmt.Sprintf("http://%s", host)
}

func parseDurationOrDefault(value string, fallback time.Duration) time.Duration {
	duration, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		return fallback
	}
	return duration
}
