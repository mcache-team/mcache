package item

import "time"

type Item struct {
	Prefix     string        `json:"prefix"`
	Data       []byte        `json:"data"`
	Timeout    time.Duration `json:"timeout,omitempty"`
	ExpireTime time.Time     `json:"expireTime,omitempty"`
	CreatedAt  time.Time     `json:"createdAt"`
	UpdatedAt  time.Time     `json:"UpdatedAt"`
}

type Option func(item *Item)

func (i *Item) String() string {
	return i.Prefix
}
