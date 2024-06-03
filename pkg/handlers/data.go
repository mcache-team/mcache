package handlers

import (
	"errors"
	"github.com/mcache-team/mcache/pkg/apis/v1/item"
	v1 "github.com/mcache-team/mcache/pkg/apis/v1/prefix-tree"
	"github.com/mcache-team/mcache/pkg/storage"
	"github.com/sirupsen/logrus"
)

type prefixTree struct {
	tree map[string]*v1.PrefixNode
}

func NewPrefixTree() *prefixTree {
	return &prefixTree{tree: make(map[string]*v1.PrefixNode)}
}

func (p *prefixTree) ListNode(prefix string) (list []*item.Item, err error) {
	prefixGroup := v1.Prefix(prefix).SplitPrefix()
	logrus.Infof("prefix parsed to PrefixGroup, RootPrefix : %s, NodePrefix: %s", prefixGroup.RootPrefix, prefixGroup.NodePrefix)
	if prefixGroup.IsEmpty() {
		err = errors.New("prefix is empty")
		return
	}
	var prefixNode *v1.PrefixNode
	if prefixNode, err = p.searchNode(prefixGroup); err != nil {
		return
	}
	list = make([]*item.Item, 0, len(prefixNode.SubPrefix))
	for _, prefixItem := range prefixNode.SubPrefix {
		var data *item.Item
		data, err = storage.StorageClient.GetOne(prefixItem.Prefix.Before(v1.Prefix(prefix)).String())
		if err != nil {
			return
		}
		list = append(list, data)
	}
	return
}

func (p *prefixTree) searchNode(prefixGroup *v1.PrefixGroup) (node *v1.PrefixNode, err error) {
	root, has := p.tree[prefixGroup.RootPrefix]
	if !has {
		err = errors.New("root prefix not existed")
		return
	}
	for _, prefix := range prefixGroup.PrefixPath {
		find := false
		for index, item := range root.SubPrefix {
			if item.Prefix == v1.Prefix(prefix) {
				root = root.SubPrefix[index]
				find = true
				break
			}
		}
		if !find {
			err = errors.New("prefix path not full match")
			return
		}
	}
	node = root
	return
}

func (p *prefixTree) RemoveNode(prefix string) (err error) {
	prefixGroup := v1.Prefix(prefix).SplitPrefix()
	var node *v1.PrefixNode
	node, err = p.searchNode(prefixGroup)
	if err != nil {
		return
	}
	node.HasData = false
	_, err = storage.StorageClient.Delete(prefix)
	return
}

func (p *prefixTree) InsertNode(prefix string, data interface{}) error {
	prefixGroup := v1.Prefix(prefix).SplitPrefix()
	node := p.mergeTree(prefixGroup)
	logrus.Infof("node prefix : %s", prefixGroup.NodePrefix)
	if node.Prefix == v1.Prefix(prefixGroup.NodePrefix) {
		node.HasData = true
	}
	err := storage.StorageClient.Insert(prefix, data)
	return err
}

func (p *prefixTree) mergeTree(prefixGroup *v1.PrefixGroup) *v1.PrefixNode {
	var root *v1.PrefixNode
	if prefixGroup.IsEmpty() {
		return root
	}
	if exist, has := p.tree[prefixGroup.RootPrefix]; has {
		root = exist
	} else {
		exist = &v1.PrefixNode{
			Prefix:    v1.Prefix(prefixGroup.RootPrefix),
			HasData:   true,
			SubPrefix: make([]*v1.PrefixNode, 0),
		}
		p.tree[prefixGroup.RootPrefix] = exist
		root = exist
	}
	for _, prefix := range prefixGroup.PrefixPath {
		find := false
		for index, item := range root.SubPrefix {
			if item.Prefix == v1.Prefix(prefix) {
				logrus.Infof("prefix %s found", prefix)
				root = root.SubPrefix[index]
				find = true
				break
			}
		}
		if !find {
			logrus.Infof("Prefix %s not found, init it", prefix)
			node := &v1.PrefixNode{
				Prefix:    v1.Prefix(prefix),
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
