package cluster

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	hclog "github.com/hashicorp/go-hclog"
	hashiraft "github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
	"github.com/mcache-team/mcache/pkg/apis/v1/item"
	"github.com/mcache-team/mcache/pkg/state"
)

type raftNode struct {
	cfg         Config
	machine     *state.StateMachine
	raft        *hashiraft.Raft
	transport   *hashiraft.NetworkTransport
	logStore    io.Closer
	stableStore io.Closer
	metadata    *metadataStore
}

type raftFSM struct {
	machine  *state.StateMachine
	metadata *metadataStore
}

type raftSnapshot struct {
	CacheSnapshot *state.Snapshot   `json:"cacheSnapshot"`
	AddressBook   map[string]string `json:"addressBook"`
}

type applyResponse struct {
	Value interface{} `json:"value,omitempty"`
	Error string      `json:"error,omitempty"`
}

type raftLogPayload struct {
	Kind             string            `json:"kind"`
	Command          *state.Command    `json:"command,omitempty"`
	MetadataUpsert   *metadataUpsert   `json:"metadataUpsert,omitempty"`
	MetadataDeletion *metadataDeletion `json:"metadataDeletion,omitempty"`
}

type metadataUpsert struct {
	NodeID        string `json:"nodeId"`
	AdvertiseAddr string `json:"advertiseAddress"`
}

type metadataDeletion struct {
	NodeID string `json:"nodeId"`
}

func newRaftNode(cfg Config, machine *state.StateMachine) (Node, error) {
	if err := os.MkdirAll(cfg.RaftDataDir, 0o755); err != nil {
		return nil, fmt.Errorf("create raft data dir: %w", err)
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "mcache-raft",
		Level: hclog.Info,
	})

	advertiseAddr, err := net.ResolveTCPAddr("tcp", cfg.RaftAdvertise)
	if err != nil {
		return nil, fmt.Errorf("resolve raft advertise addr: %w", err)
	}
	transport, err := hashiraft.NewTCPTransportWithLogger(cfg.RaftBindAddr, advertiseAddr, 3, 10*time.Second, logger)
	if err != nil {
		return nil, fmt.Errorf("create raft transport: %w", err)
	}

	logStore, err := raftboltdb.NewBoltStore(filepath.Join(cfg.RaftDataDir, "raft-log.bolt"))
	if err != nil {
		return nil, fmt.Errorf("create raft log store: %w", err)
	}
	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(cfg.RaftDataDir, "raft-stable.bolt"))
	if err != nil {
		_ = logStore.Close()
		_ = transport.Close()
		return nil, fmt.Errorf("create raft stable store: %w", err)
	}
	snapshotStore, err := hashiraft.NewFileSnapshotStoreWithLogger(cfg.RaftDataDir, 3, logger)
	if err != nil {
		_ = stableStore.Close()
		_ = logStore.Close()
		_ = transport.Close()
		return nil, fmt.Errorf("create raft snapshot store: %w", err)
	}

	raftConfig := hashiraft.DefaultConfig()
	raftConfig.LocalID = hashiraft.ServerID(cfg.NodeID)
	raftConfig.SnapshotThreshold = 64
	raftConfig.SnapshotInterval = 30 * time.Second
	raftConfig.LogOutput = io.Discard
	raftConfig.Logger = logger

	addressBook := make(map[string]string, len(cfg.RaftPeers)+1)
	addressBook[cfg.NodeID] = cfg.AdvertiseAddr
	for _, peer := range cfg.RaftPeers {
		addressBook[peer.NodeID] = peer.AdvertiseAddr
	}
	metadata := newMetadataStore(addressBook)

	if err := bootstrapRaftIfNeeded(cfg, raftConfig, logStore, stableStore, snapshotStore, transport); err != nil {
		_ = stableStore.Close()
		_ = logStore.Close()
		_ = transport.Close()
		return nil, err
	}

	raftInstance, err := hashiraft.NewRaft(raftConfig, &raftFSM{machine: machine, metadata: metadata}, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		_ = stableStore.Close()
		_ = logStore.Close()
		_ = transport.Close()
		return nil, fmt.Errorf("start raft node: %w", err)
	}

	return &raftNode{
		cfg:         cfg,
		machine:     machine,
		raft:        raftInstance,
		transport:   transport,
		logStore:    logStore,
		stableStore: stableStore,
		metadata:    metadata,
	}, nil
}

func (n *raftNode) Submit(cmd state.Command) (interface{}, error) {
	return n.applyPayload(raftLogPayload{
		Kind:    "command",
		Command: &cmd,
	})
}

func (n *raftNode) ListMembers() ([]Member, error) {
	future := n.raft.GetConfiguration()
	if err := future.Error(); err != nil {
		return nil, err
	}

	leaderAddr, leaderID := n.raft.LeaderWithID()
	configuration := future.Configuration()
	members := make([]Member, 0, len(configuration.Servers))
	for _, server := range configuration.Servers {
		member := Member{
			NodeID:        string(server.ID),
			RaftAddress:   string(server.Address),
			AdvertiseAddr: n.lookupHTTPAddress(string(server.ID), string(server.Address)),
			Suffrage:      server.Suffrage.String(),
			IsLeader:      server.ID == leaderID || server.Address == leaderAddr,
		}
		members = append(members, member)
	}
	return members, nil
}

func (n *raftNode) AddMember(req JoinRequest) error {
	if req.NodeID == "" || req.RaftAddress == "" || req.AdvertiseAddr == "" {
		return fmt.Errorf("nodeId, raftAddress and advertiseAddress are required")
	}
	if err := n.ensureLeader(); err != nil {
		return err
	}
	if err := n.raft.AddVoter(hashiraft.ServerID(req.NodeID), hashiraft.ServerAddress(req.RaftAddress), 0, n.cfg.ApplyTimeout).Error(); err != nil {
		if err == hashiraft.ErrNotLeader || err == hashiraft.ErrLeadershipLost {
			return &NotLeaderError{LeaderAddress: n.leaderHTTPAddress()}
		}
		return err
	}
	_, err := n.applyPayload(raftLogPayload{
		Kind: "metadata-upsert",
		MetadataUpsert: &metadataUpsert{
			NodeID:        req.NodeID,
			AdvertiseAddr: req.AdvertiseAddr,
		},
	})
	return err
}

func (n *raftNode) RemoveMember(nodeID string) error {
	if nodeID == "" {
		return fmt.Errorf("nodeId is required")
	}
	if err := n.ensureLeader(); err != nil {
		return err
	}
	if err := n.raft.RemoveServer(hashiraft.ServerID(nodeID), 0, n.cfg.ApplyTimeout).Error(); err != nil {
		if err == hashiraft.ErrNotLeader || err == hashiraft.ErrLeadershipLost {
			return &NotLeaderError{LeaderAddress: n.leaderHTTPAddress()}
		}
		return err
	}
	_, err := n.applyPayload(raftLogPayload{
		Kind: "metadata-delete",
		MetadataDeletion: &metadataDeletion{
			NodeID: nodeID,
		},
	})
	return err
}

func (n *raftNode) Status() Status {
	leaderAddr, leaderID := n.raft.LeaderWithID()
	return Status{
		Mode:          string(ModeRaft),
		NodeID:        n.cfg.NodeID,
		Role:          n.raft.State().String(),
		IsLeader:      n.raft.State() == hashiraft.Leader,
		LeaderAddress: n.lookupHTTPAddress(string(leaderID), string(leaderAddr)),
		AdvertiseAddr: n.cfg.AdvertiseAddr,
		RaftAddress:   string(n.transport.LocalAddr()),
	}
}

func (n *raftNode) Ready() error {
	switch n.raft.State() {
	case hashiraft.Leader:
		return nil
	case hashiraft.Follower:
		leaderAddr, _ := n.raft.LeaderWithID()
		if leaderAddr == "" {
			return fmt.Errorf("raft follower has no known leader")
		}
		return nil
	default:
		return fmt.Errorf("raft node is not ready in state %s", n.raft.State().String())
	}
}

func (n *raftNode) Diagnostics() (*Diagnostics, error) {
	members, err := n.ListMembers()
	if err != nil {
		return nil, err
	}
	readyErr := n.Ready()
	return &Diagnostics{
		Timestamp: time.Now(),
		Ready:     readyErr == nil,
		Status:    n.Status(),
		Members:   members,
		State:     n.machine.Stats(),
		Raft:      n.raft.Stats(),
		Telemetry: DefaultTelemetry.Snapshot(),
	}, nil
}

func (n *raftNode) leaderHTTPAddress() string {
	leaderAddr, leaderID := n.raft.LeaderWithID()
	return n.lookupHTTPAddress(string(leaderID), string(leaderAddr))
}

func (n *raftNode) lookupHTTPAddress(leaderID, raftAddr string) string {
	if leaderID != "" {
		if addr, ok := n.metadata.Lookup(leaderID); ok {
			return addr
		}
	}
	if raftAddr == "" {
		return ""
	}
	return "http://" + raftAddr
}

func (f *raftFSM) Apply(log *hashiraft.Log) interface{} {
	payload := raftLogPayload{}
	if err := json.Unmarshal(log.Data, &payload); err != nil {
		return &applyResponse{Error: err.Error()}
	}

	switch payload.Kind {
	case "command":
		if payload.Command == nil {
			return &applyResponse{Error: "raft command payload is empty"}
		}
		value, err := f.machine.Apply(*payload.Command)
		if err != nil {
			return &applyResponse{Error: err.Error()}
		}
		return &applyResponse{Value: value}
	case "metadata-upsert":
		if payload.MetadataUpsert == nil {
			return &applyResponse{Error: "metadata upsert payload is empty"}
		}
		f.metadata.Upsert(payload.MetadataUpsert.NodeID, payload.MetadataUpsert.AdvertiseAddr)
		return &applyResponse{}
	case "metadata-delete":
		if payload.MetadataDeletion == nil {
			return &applyResponse{Error: "metadata delete payload is empty"}
		}
		f.metadata.Delete(payload.MetadataDeletion.NodeID)
		return &applyResponse{}
	default:
		return &applyResponse{Error: fmt.Sprintf("unsupported raft payload kind %q", payload.Kind)}
	}
}

func (f *raftFSM) Snapshot() (hashiraft.FSMSnapshot, error) {
	return &raftSnapshot{
		CacheSnapshot: f.machine.Snapshot(),
		AddressBook:   f.metadata.Snapshot(),
	}, nil
}

func (f *raftFSM) Restore(reader io.ReadCloser) error {
	defer reader.Close()
	snapshot := &raftSnapshot{}
	if err := json.NewDecoder(reader).Decode(snapshot); err != nil {
		return err
	}
	if snapshot.CacheSnapshot == nil {
		snapshot.CacheSnapshot = &state.Snapshot{Items: map[string]*item.Item{}}
	}
	f.machine.Restore(snapshot.CacheSnapshot)
	if snapshot.AddressBook == nil {
		snapshot.AddressBook = map[string]string{}
	}
	f.metadata.Restore(snapshot.AddressBook)
	return nil
}

func (s *raftSnapshot) Persist(sink hashiraft.SnapshotSink) error {
	if err := json.NewEncoder(sink).Encode(s); err != nil {
		_ = sink.Cancel()
		return err
	}
	return sink.Close()
}

func (s *raftSnapshot) Release() {}

func bootstrapRaftIfNeeded(
	cfg Config,
	raftConfig *hashiraft.Config,
	logStore hashiraft.LogStore,
	stableStore hashiraft.StableStore,
	snapshotStore hashiraft.SnapshotStore,
	transport hashiraft.Transport,
) error {
	if !cfg.RaftBootstrap {
		return nil
	}

	hasState, err := hashiraft.HasExistingState(logStore, stableStore, snapshotStore)
	if err != nil {
		return fmt.Errorf("check existing raft state: %w", err)
	}
	if hasState {
		return nil
	}

	servers := make([]hashiraft.Server, 0, len(cfg.RaftPeers)+1)
	seen := map[string]struct{}{}
	appendPeer := func(peer Peer) {
		if _, ok := seen[peer.NodeID]; ok {
			return
		}
		seen[peer.NodeID] = struct{}{}
		servers = append(servers, hashiraft.Server{
			ID:       hashiraft.ServerID(peer.NodeID),
			Address:  hashiraft.ServerAddress(peer.RaftAddress),
			Suffrage: hashiraft.Voter,
		})
	}

	appendPeer(Peer{
		NodeID:        cfg.NodeID,
		RaftAddress:   cfg.RaftAdvertise,
		AdvertiseAddr: cfg.AdvertiseAddr,
	})
	for _, peer := range cfg.RaftPeers {
		appendPeer(peer)
	}

	if len(servers) == 0 {
		return fmt.Errorf("raft bootstrap requires at least one voter")
	}

	if err := hashiraft.BootstrapCluster(raftConfig, logStore, stableStore, snapshotStore, transport, hashiraft.Configuration{
		Servers: servers,
	}); err != nil && err != hashiraft.ErrCantBootstrap {
		return fmt.Errorf("bootstrap raft cluster: %w", err)
	}
	return nil
}

func commandError(message string) error {
	switch message {
	case item.NoDataError.Error():
		return item.NoDataError
	case item.PrefixExisted.Error():
		return item.PrefixExisted
	case item.PrefixNotExisted.Error():
		return item.PrefixNotExisted
	default:
		return fmt.Errorf(message)
	}
}

func (n *raftNode) applyPayload(payload raftLogPayload) (interface{}, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	future := n.raft.Apply(raw, n.cfg.ApplyTimeout)
	if err := future.Error(); err != nil {
		if err == hashiraft.ErrNotLeader || err == hashiraft.ErrLeadershipLost {
			return nil, &NotLeaderError{LeaderAddress: n.leaderHTTPAddress()}
		}
		return nil, err
	}

	response := future.Response()
	switch typed := response.(type) {
	case nil:
		return nil, nil
	case *applyResponse:
		if typed.Error != "" {
			return nil, commandError(typed.Error)
		}
		return typed.Value, nil
	case error:
		return nil, typed
	default:
		return typed, nil
	}
}

func (n *raftNode) ensureLeader() error {
	if n.raft.State() == hashiraft.Leader {
		return nil
	}
	return &NotLeaderError{LeaderAddress: n.leaderHTTPAddress()}
}
