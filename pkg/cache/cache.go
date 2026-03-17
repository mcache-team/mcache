package cache

import (
	"github.com/mcache-team/mcache/pkg/apis/v1/item"
	"github.com/mcache-team/mcache/pkg/handlers"
	"github.com/mcache-team/mcache/pkg/storage"
)

// CacheClient is the embedded module interface for direct in-process cache access.
// It provides the same CRUD and prefix-query operations as the HTTP service,
// without starting any HTTP listener.
type CacheClient interface {
	Insert(prefix string, data interface{}, opts ...item.Option) error
	Get(prefix string) (*item.Item, error)
	Update(prefix string, data interface{}, opts ...item.Option) error
	Delete(prefix string) error
	ListByPrefix(prefix string) ([]*item.Item, error)
}

type cacheClient struct{}

// NewCacheClient returns a CacheClient that reuses the shared PrefixHandler
// and StorageClient, keeping behaviour consistent with the HTTP service.
func NewCacheClient() CacheClient {
	return &cacheClient{}
}

// Insert stores data under prefix in both the PrefixTree and Storage.
// Any item.Option values (e.g. WithTTL) are forwarded to the storage layer.
func (c *cacheClient) Insert(prefix string, data interface{}, opts ...item.Option) error {
	return handlers.PrefixHandler.InsertNode(prefix, data, opts...)
}

// Get retrieves the item stored at the exact prefix.
func (c *cacheClient) Get(prefix string) (*item.Item, error) {
	return storage.StorageClient.GetOne(prefix)
}

// Update replaces the data at an existing prefix and updates the timestamp.
// If opts contains a new TTL it will overwrite the existing expireTime;
// otherwise the original expireTime is preserved.
func (c *cacheClient) Update(prefix string, data interface{}, opts ...item.Option) error {
	return storage.StorageClient.Update(prefix, data, opts...)
}

// Delete removes the item from Storage and marks the PrefixTree node as having no data.
func (c *cacheClient) Delete(prefix string) error {
	return handlers.PrefixHandler.RemoveNode(prefix)
}

// ListByPrefix returns all direct child items under the given prefix path.
func (c *cacheClient) ListByPrefix(prefix string) ([]*item.Item, error) {
	return handlers.PrefixHandler.ListNode(prefix)
}
