package prefixtree

import (
	"github.com/mcache-team/mcache/pkg/apis/v1/item"
	"strings"
	"time"
)

type Prefix string

func (p Prefix) SplitPrefix() []string {
	return strings.Split(string(p), "/")
}

type PrefixNode struct {
	Prefix    Prefix        `json:"prefix"`
	HasData   bool          `json:"hasData"`
	SubPrefix []*PrefixNode `json:"subPrefix"`
	Timeout   time.Time     `json:"timeout"`
}

type PrefixTree interface {
	InsertNode(node *item.Item) error
}
