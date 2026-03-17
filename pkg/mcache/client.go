package mcache

import (
	"github.com/mcache-team/mcache/pkg/apis/v1/item"
	"github.com/mcache-team/mcache/pkg/handlers"
	"github.com/mcache-team/mcache/pkg/storage"
)

// CacheClient defines the interface for interacting with the in-memory cache.
type CacheClient interface {
	Insert(prefix string, data interface{}, opts ...item.Option) error
	Get(prefix string) (*item.Item, error)
	Update(prefix string, data interface{}, opts ...item.Option) error
	Delete(prefix string) error
	ListByPrefix(prefix string) ([]*item.Item, error)
}

// storageProvider is the subset of storage.storageClient methods we need.
type storageProvider interface {
	GetOne(prefix string) (*item.Item, error)
	Update(prefix string, data interface{}, opt ...item.Option) error
}

// prefixProvider is the subset of handlers.prefixTree methods we need.
type prefixProvider interface {
	InsertNode(prefix string, data interface{}, opts ...item.Option) error
	RemoveNode(prefix string) error
	ListNode(prefix string) ([]*item.Item, error)
}

type cacheClient struct {
	stor       storageProvider
	prefixTree prefixProvider
}

// New returns a CacheClient backed by the global StorageClient and PrefixHandler.
func New() CacheClient {
	return &cacheClient{
		stor:       storage.StorageClient,
		prefixTree: handlers.PrefixHandler,
	}
}

func (c *cacheClient) Insert(prefix string, data interface{}, opts ...item.Option) error {
	return c.prefixTree.InsertNode(prefix, data, opts...)
}

func (c *cacheClient) Get(prefix string) (*item.Item, error) {
	return c.stor.GetOne(prefix)
}

func (c *cacheClient) Update(prefix string, data interface{}, opts ...item.Option) error {
	return c.stor.Update(prefix, data, opts...)
}

func (c *cacheClient) Delete(prefix string) error {
	return c.prefixTree.RemoveNode(prefix)
}

func (c *cacheClient) ListByPrefix(prefix string) ([]*item.Item, error) {
	return c.prefixTree.ListNode(prefix)
}
