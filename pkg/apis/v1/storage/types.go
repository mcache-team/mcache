package storage

import "github.com/mcache-team/mcache/pkg/apis/v1/item"

type Storage interface {
	GetOne(prefix string) (*item.Item, error)
	ListPrefixData(prefix []string) ([]*item.Item, error)
	CountPrefixData(prefixList []string) int
	ListPrefix(prePrefix string) ([]string, error)
	CountPrefix(prePrefix string) int
	Insert(prefix string, data []byte, opt ...item.Option) error
	Update(prefix string, data []byte, opt ...item.Option) error
	Delete(prefix string) (interface{}, error)
}
