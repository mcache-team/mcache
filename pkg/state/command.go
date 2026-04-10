package state

import "github.com/mcache-team/mcache/pkg/apis/v1/item"

type CommandType string

const (
	CommandInsert CommandType = "insert"
	CommandUpdate CommandType = "update"
	CommandDelete CommandType = "delete"
)

// Command is the state transition unit that can later be serialized and
// replicated through a consensus layer.
type Command struct {
	Type   CommandType `json:"type"`
	Prefix string      `json:"prefix"`
	Item   *item.Item  `json:"item,omitempty"`
}
