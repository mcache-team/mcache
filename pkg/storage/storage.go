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

func NewStorage() storage.Storage {
	return &Storage{store: &memory.Memory{}, storeType: "memory"}
}

func (s *Storage) GetOne(prefix string) (interface{}, error) {
	return s.store.GetOne(prefix)
}

func (s *Storage) ListPrefixData(prefix string) ([]interface{}, error) {
	return s.store.ListPrefixData(prefix)
}

func (s *Storage) CountPrefixData(prefix string) (int, error) {
	return s.store.CountPrefixData(prefix)
}

func (s *Storage) ListPrefix(prefixStr string) ([]string, error) {
	return s.store.ListPrefix(prefixStr)
}

func (s *Storage) CountPrefix(prefixStr string) (int, error) {
	return s.store.CountPrefix(prefixStr)
}

func (s *Storage) Insert(prefix string, data interface{}, opt ...item.Option) error {
	return s.store.Insert(prefix, data, opt...)
}

func (s *Storage) Update(prefix string, data interface{}, opt ...item.Option) error {
	return s.store.Update(prefix, data, opt...)
}

func (s *Storage) Delete(prefix string) error {
	return s.store.Delete(prefix)
}
