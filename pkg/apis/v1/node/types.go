package node

import "time"

// Node tree node data
// the endpoint of one route
type Node struct {
	Prefix    string
	Data      interface{}
	Timeout   time.Duration
	CreatedAt time.Time
	UpdateAt  time.Time
}
