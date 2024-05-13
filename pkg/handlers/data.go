package handlers

import (
	"github.com/mcache-team/mcache/pkg/apis/v1/item"
	prefixtree "github.com/mcache-team/mcache/pkg/apis/v1/prefix-tree"
)

type prefixHandler struct {
}

func (p *prefixHandler) InsertNode(node *item.Item) error {
	//TODO implement me
	panic("implement me")
}

var _ prefixtree.PrefixTree = &prefixHandler{}
