package memory

import (
	"github.com/mcache-team/mcache/pkg/apis/v1/item"
	"github.com/mcache-team/mcache/pkg/apis/v1/storage"
	"github.com/sirupsen/logrus"
	"strings"
	"sync"
	"time"
)

type Memory struct {
	mu         sync.RWMutex
	prefixList []string
	dataMap    *sync.Map
}

var MemoryStoreType = "memory"

var _ storage.Storage = &Memory{}

func NewStorage() storage.Storage {
	return &Memory{
		prefixList: make([]string, 0),
		dataMap:    new(sync.Map),
	}
}
func (m *Memory) GetOne(prefix string) (*item.Item, error) {
	data, has := m.dataMap.Load(prefix)
	if !has {
		return nil, item.NoDataError
	}
	it := data.(*item.Item)
	if !it.ExpireTime.IsZero() && it.ExpireTime.Before(time.Now()) {
		m.Delete(prefix)
		return nil, item.NoDataError
	}
	return it, nil
}

// ListPrefixData
// list data by prefix list
// get data with given prefix list
func (m *Memory) ListPrefixData(prefixList []string) ([]*item.Item, error) {
	result := make([]*item.Item, 0, len(prefixList))
	prefixMap := map[string]struct{}{}
	for _, item := range prefixList {
		prefixMap[item] = struct{}{}
	}
	m.dataMap.Range(func(key, value interface{}) bool {
		if _, has := prefixMap[key.(string)]; has {
			data := value.(*item.Item)
			result = append(result, data)
		}
		return true
	})
	if len(result) == 0 {
		return nil, item.NoDataError
	}
	return result, nil
}

func (m *Memory) CountPrefixData(prefixList []string) int {
	count := len(prefixList)
	for _, item := range prefixList {
		if _, has := m.dataMap.Load(item); !has {
			count--
		}
	}
	return count
}

func (m *Memory) ListPrefix(prePrefix string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]string, 0)
	for _, item := range m.prefixList {
		if strings.HasPrefix(item, prePrefix) {
			result = append(result, item)
		}
	}
	return result, nil
}

func (m *Memory) CountPrefix(prePrefix string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, item := range m.prefixList {
		if strings.HasPrefix(item, prePrefix) {
			count++
		}
	}
	return count
}

func (m *Memory) Insert(prefix string, data interface{}, opt ...item.Option) error {
	now := time.Now()
	cacheItem := &item.Item{
		Prefix:    prefix,
		Data:      data,
		CreatedAt: now,
		UpdatedAt: now,
	}
	for _, op := range opt {
		op(cacheItem)
	}
	if cacheItem.Timeout > 0 {
		cacheItem.ExpireTime = cacheItem.CreatedAt.Add(cacheItem.Timeout)
	}
	if _, loaded := m.dataMap.LoadOrStore(prefix, cacheItem); loaded {
		return item.PrefixExisted
	}
	m.mu.Lock()
	m.prefixList = append(m.prefixList, prefix)
	m.mu.Unlock()
	return nil
}

func (m *Memory) Update(prefix string, data interface{}, opt ...item.Option) error {
	existed, has := m.dataMap.Load(prefix)
	if !has {
		return item.NoDataError
	}
	cacheItem := existed.(*item.Item)
	cacheItem.Data = data
	cacheItem.UpdatedAt = time.Now()
	oldExpireTime := cacheItem.ExpireTime
	for _, op := range opt {
		op(cacheItem)
	}
	if cacheItem.Timeout > 0 {
		cacheItem.ExpireTime = time.Now().Add(cacheItem.Timeout)
	} else {
		cacheItem.ExpireTime = oldExpireTime
	}
	m.dataMap.Store(prefix, cacheItem)
	return nil
}

func (m *Memory) Delete(prefix string) (interface{}, error) {
	data, has := m.dataMap.Load(prefix)
	if !has {
		logrus.Warningf("cache item of prefix %s not existed", prefix)
		return nil, item.PrefixNotExisted
	}
	m.dataMap.Delete(prefix)
	m.mu.Lock()
	for idx, item := range m.prefixList {
		if item == prefix {
			m.prefixList = append(m.prefixList[:idx], m.prefixList[idx+1:]...)
			break
		}
	}
	m.mu.Unlock()
	return data.(*item.Item).Data, nil
}
