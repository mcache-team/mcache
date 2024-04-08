package prefixtree

import "github.com/mcache-team/mcache/pkg/apis/v1/node"

type PrefixTree interface {
	InsertNode(node *node.Node) error
}
