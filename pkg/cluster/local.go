package cluster

import (
	"fmt"
	"time"

	"github.com/mcache-team/mcache/pkg/state"
)

type localNode struct {
	cfg     Config
	machine *state.StateMachine
}

func newLocalNode(cfg Config, machine *state.StateMachine) Node {
	return &localNode{
		cfg:     cfg,
		machine: machine,
	}
}

func (n *localNode) Submit(cmd state.Command) (interface{}, error) {
	return n.machine.Apply(cmd)
}

func (n *localNode) ListMembers() ([]Member, error) {
	return []Member{{
		NodeID:        n.cfg.NodeID,
		AdvertiseAddr: n.cfg.AdvertiseAddr,
		IsLeader:      true,
		Suffrage:      "Voter",
	}}, nil
}

func (n *localNode) AddMember(req JoinRequest) error {
	return fmt.Errorf("cluster membership changes require raft mode")
}

func (n *localNode) RemoveMember(nodeID string) error {
	return fmt.Errorf("cluster membership changes require raft mode")
}

func (n *localNode) Ready() error {
	return nil
}

func (n *localNode) Diagnostics() (*Diagnostics, error) {
	members, err := n.ListMembers()
	if err != nil {
		return nil, err
	}
	return &Diagnostics{
		Timestamp: time.Now(),
		Ready:     true,
		Status:    n.Status(),
		Members:   members,
		State:     n.machine.Stats(),
		Telemetry: DefaultTelemetry.Snapshot(),
	}, nil
}

func (n *localNode) Status() Status {
	return Status{
		Mode:          string(ModeSingle),
		NodeID:        n.cfg.NodeID,
		Role:          "leader",
		IsLeader:      true,
		LeaderAddress: n.cfg.AdvertiseAddr,
		AdvertiseAddr: n.cfg.AdvertiseAddr,
	}
}
