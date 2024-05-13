package storage

import (
	"github.com/mcache-team/mcache/pkg/apis/v1/item"
	"github.com/mcache-team/mcache/pkg/apis/v1/storage"
	"github.com/mcache-team/mcache/pkg/storage/memory"
)

var StorageClient = NewStorage()

type storageClient struct {
	store     storage.Storage
	storeType string
}

func (s *storageClient) GetOne(prefix string) (*item.Item, error) {
	return s.store.GetOne(prefix)
}

func (s *storageClient) ListPrefixData(prefix []string) ([]*item.Item, error) {
	return s.store.ListPrefixData(prefix)
}

func (s *storageClient) CountPrefixData(prefixList []string) int {
	return s.store.CountPrefixData(prefixList)
}

func (s *storageClient) ListPrefix(prePrefix string) ([]string, error) {
	return s.store.ListPrefix(prePrefix)
}

func (s *storageClient) CountPrefix(prePrefix string) int {
	return s.store.CountPrefix(prePrefix)
}

func (s *storageClient) Insert(prefix string, data interface{}, opt ...item.Option) error {
	return s.store.Insert(prefix, data, opt...)
}

func (s *storageClient) Update(prefix string, data interface{}, opt ...item.Option) error {
	return s.store.Update(prefix, data, opt...)
}

func (s *storageClient) Delete(prefix string) (interface{}, error) {
	return s.store.Delete(prefix)
}

func NewStorage() *storageClient {
	return &storageClient{store: memory.NewStorage(), storeType: memory.MemoryStoreType}
}
