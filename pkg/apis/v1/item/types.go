package item

import "time"

type Item struct {
	Prefix     string
	Data       interface{}
	Timeout    time.Duration
	ExpireTime time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Option func(item *Item)

func (i *Item) String() string {
	return i.Prefix
}
