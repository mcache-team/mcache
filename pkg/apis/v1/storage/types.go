package storage

import "github.com/mcache-team/mcache/pkg/apis/v1/item"

type Storage interface {
	GetOne(prefix string) (interface{}, error)
	ListPrefixData(prefix string) ([]interface{}, error)
	CountPrefixData(prefix string) (int, error)
	ListPrefix(prefixStr string) ([]string, error)
	CountPrefix(prefixStr string) (int, error)
	Insert(prefix string, data interface{}, opt ...item.Option) error
	Update(prefix string, data interface{}, opt ...item.Option) error
	Delete(prefix string) error
}
