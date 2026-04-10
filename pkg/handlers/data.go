package handlers

import (
	"github.com/mcache-team/mcache/pkg/apis/v1/item"
	"github.com/mcache-team/mcache/pkg/state"
)

type prefixTree struct {
	stateMachine *state.StateMachine
}

func NewPrefixTree() *prefixTree {
	return &prefixTree{stateMachine: state.DefaultStateMachine}
}

func (p *prefixTree) ListNode(prefix string) (list []*item.Item, err error) {
	return p.stateMachine.ListByPrefix(prefix)
}

func (p *prefixTree) RemoveNode(prefix string) (err error) {
	_, err = p.stateMachine.Delete(prefix)
	return
}

func (p *prefixTree) InsertNode(prefix string, data interface{}) error {
	return p.stateMachine.Insert(prefix, data)
}
