package prefixtree

import (
	"github.com/mcache-team/mcache/pkg/apis/v1/item"
	"strings"
	"time"
)

type Prefix string

func (p Prefix) SplitPrefix() *PrefixGroup {
	prefixPath := strings.Split(string(p), "/")
	if len(prefixPath) == 0 {
		return &PrefixGroup{}
	}
	if prefixPath[0] == "" {
		prefixPath = prefixPath[1:]
	}

	return &PrefixGroup{
		RootPrefix: prefixPath[0],
		NodePrefix: prefixPath[len(prefixPath)-1],
		PrefixPath: prefixPath[1:],
	}
}

func (p Prefix) String() string {
	return string(p)
}

func (p Prefix) Before(parent Prefix) Prefix {
	return Prefix(strings.Join([]string{parent.String(), p.String()}, "/"))
}

type PrefixGroup struct {
	RootPrefix string
	NodePrefix string
	PrefixPath []string
}

func (pg *PrefixGroup) IsEmpty() bool {
	if pg.RootPrefix == "" {
		return true
	}
	return false
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
