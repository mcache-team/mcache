package prefixtree

import (
	"github.com/mcache-team/mcache/pkg/apis/v1/item"
)

type PrefixTree interface {
	InsertNode(node *item.Item) error
}
