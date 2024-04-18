package storage

import "github.com/mcache-team/mcache/pkg/apis/v1/item"

type Storage interface {
	GetOne(prefix string) (interface{}, error)
	ListPrefixData(prefix []string) ([]interface{}, error)
	CountPrefixData(prefixList []string) int
	ListPrefix(prePrefix string) ([]string, error)
	CountPrefix(prePrefix string) int
	Insert(prefix string, data interface{}, opt ...item.Option) error
	Update(prefix string, data interface{}, opt ...item.Option) error
	Delete(prefix string) (interface{}, error)
}
