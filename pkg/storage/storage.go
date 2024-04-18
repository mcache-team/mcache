package storage

import (
	"github.com/mcache-team/mcache/pkg/apis/v1/item"
	"github.com/mcache-team/mcache/pkg/apis/v1/storage"
	"github.com/mcache-team/mcache/pkg/storage/memory"
)

type Storage struct {
	store     storage.Storage
	storeType string
}

func (s *Storage) GetOne(prefix string) (interface{}, error) {
	return s.store.GetOne(prefix)
}

func (s *Storage) ListPrefixData(prefix []string) ([]interface{}, error) {
	return s.store.ListPrefixData(prefix)
}

func (s *Storage) CountPrefixData(prefixList []string) int {
	return s.store.CountPrefixData(prefixList)
}

func (s *Storage) ListPrefix(prePrefix string) ([]string, error) {
	return s.store.ListPrefix(prePrefix)
}

func (s *Storage) CountPrefix(prePrefix string) int {
	return s.store.CountPrefix(prePrefix)
}

func (s *Storage) Insert(prefix string, data interface{}, opt ...item.Option) error {
	return s.store.Insert(prefix, data, opt...)
}

func (s *Storage) Update(prefix string, data interface{}, opt ...item.Option) error {
	return s.store.Update(prefix, data, opt...)
}

func (s *Storage) Delete(prefix string) (interface{}, error) {
	return s.store.Delete(prefix)
}

func NewStorage() storage.Storage {
	return &Storage{store: &memory.Memory{}, storeType: memory.MemoryStoreType}
}
