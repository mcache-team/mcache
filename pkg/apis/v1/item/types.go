package item

import "time"

type Item struct {
	Prefix     string        `json:"prefix"`
	Data       interface{}   `json:"data"`
	Timeout    time.Duration `json:"timeout,omitempty"`
	ExpireTime time.Time     `json:"expireTime,omitempty"`
	CreatedAt  time.Time     `json:"createdAt"`
	UpdatedAt  time.Time     `json:"UpdatedAt"`
}

type Option func(item *Item)

func (i *Item) String() string {
	return i.Prefix
}

// WithTTL returns an Option that sets the Timeout field on an Item.
// The actual ExpireTime is computed in the storage layer after CreatedAt is set.
func WithTTL(d time.Duration) Option {
	return func(i *Item) {
		i.Timeout = d
	}
}
