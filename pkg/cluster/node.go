package cluster

import (
	"fmt"
	"sync"
	"time"

	"github.com/mcache-team/mcache/pkg/state"
)

var (
	DefaultNode Node

	defaultNodeMu sync.RWMutex
)

type Node interface {
	Submit(cmd state.Command) (interface{}, error)
	ListMembers() ([]Member, error)
	AddMember(req JoinRequest) error
	RemoveMember(nodeID string) error
	Ready() error
	Diagnostics() (*Diagnostics, error)
	Status() Status
}

type Status struct {
	Mode          string `json:"mode"`
	NodeID        string `json:"nodeId"`
	Role          string `json:"role,omitempty"`
	IsLeader      bool   `json:"isLeader"`
	LeaderAddress string `json:"leaderAddress"`
	AdvertiseAddr string `json:"advertiseAddress"`
	RaftAddress   string `json:"raftAddress,omitempty"`
}

type Member struct {
	NodeID        string `json:"nodeId"`
	RaftAddress   string `json:"raftAddress,omitempty"`
	AdvertiseAddr string `json:"advertiseAddress,omitempty"`
	Suffrage      string `json:"suffrage,omitempty"`
	IsLeader      bool   `json:"isLeader"`
}

type JoinRequest struct {
	NodeID        string `json:"nodeId"`
	RaftAddress   string `json:"raftAddress"`
	AdvertiseAddr string `json:"advertiseAddress"`
}

type Diagnostics struct {
	Timestamp time.Time         `json:"timestamp"`
	Ready     bool              `json:"ready"`
	Status    Status            `json:"status"`
	Members   []Member          `json:"members,omitempty"`
	State     state.Stats       `json:"state"`
	Raft      map[string]string `json:"raft,omitempty"`
	Telemetry TelemetrySnapshot `json:"telemetry"`
}

func Bootstrap(cfg Config, machine *state.StateMachine) error {
	node, err := NewNode(cfg, machine)
	if err != nil {
		return err
	}
	defaultNodeMu.Lock()
	DefaultNode = node
	defaultNodeMu.Unlock()
	return nil
}

func NewNode(cfg Config, machine *state.StateMachine) (Node, error) {
	switch cfg.Mode {
	case "", ModeSingle:
		return newLocalNode(cfg, machine), nil
	case ModeRaft:
		return newRaftNode(cfg, machine)
	default:
		return nil, fmt.Errorf("unsupported cluster mode %q", cfg.Mode)
	}
}
