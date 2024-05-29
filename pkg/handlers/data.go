package handlers

import (
	v1 "github.com/mcache-team/mcache/pkg/apis/v1/prefix-tree"
	"github.com/mcache-team/mcache/pkg/storage"
)

type prefixTree struct {
	tree map[string]*v1.PrefixNode
}

func NewPrefixTree() *prefixTree {
	return &prefixTree{tree: make(map[string]*v1.PrefixNode)}
}

func (p *prefixTree) InsertNode(prefix string, data []byte) error {
	prefixPath := v1.Prefix(prefix).SplitPrefix()
	node := p.mergeTree(prefixPath)
	if node.Prefix == v1.Prefix(prefixPath[len(prefixPath)-1]) {
		node.HasData = true
	}
	err := storage.StorageClient.Insert(prefix, data)
	return err
}

func (p *prefixTree) mergeTree(treePath []string) *v1.PrefixNode {
	var root *v1.PrefixNode
	if len(treePath) == 0 {
		return nil
	}
	if exist, has := p.tree[treePath[0]]; has {
		root = exist
	} else {
		exist = &v1.PrefixNode{
			Prefix:    v1.Prefix(treePath[0]),
			HasData:   true,
			SubPrefix: make([]*v1.PrefixNode, 0),
		}
		p.tree[treePath[0]] = exist
		root = exist
	}
	for i := 1; i < len(treePath)-1; i++ {
		find := false
		for index, item := range root.SubPrefix {
			if item.Prefix == v1.Prefix(treePath[i]) {
				root = root.SubPrefix[index]
				find = true
			}
		}
		if !find {
			node := &v1.PrefixNode{
				Prefix:    v1.Prefix(treePath[i]),
				HasData:   false,
				SubPrefix: make([]*v1.PrefixNode, 0),
			}
			root.HasData = true
			root.SubPrefix = append(root.SubPrefix, node)
			root = node
		}
	}
	return root
}
