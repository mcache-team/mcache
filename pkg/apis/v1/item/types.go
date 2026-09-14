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

// WithTTL makes the item expire after the given duration.
func WithTTL(d time.Duration) Option {
	return func(i *Item) {
		i.Timeout = d
		i.ExpireTime = time.Now().Add(d)
	}
}

// WithExpireTime makes the item expire at the given absolute time.
func WithExpireTime(t time.Time) Option {
	return func(i *Item) {
		i.ExpireTime = t
	}
}

func (i *Item) String() string {
	return i.Prefix
}
