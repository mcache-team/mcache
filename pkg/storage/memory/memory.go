package memory

import (
	"github.com/mcache-team/mcache/pkg/apis/v1/item"
	"github.com/mcache-team/mcache/pkg/apis/v1/storage"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

type Memory struct {
	prefixList []string
	dataMap    map[string]*item.Item
}

var MemoryStoreType = "memory"

var _ storage.Storage = &Memory{}

func (m *Memory) GetOne(prefix string) (interface{}, error) {
	if data, has := m.dataMap[prefix]; has {
		return data.Data, nil
	}
	return nil, item.NoDataError
}

// ListPrefixData
// list data by prefix list
// get data with given prefix list
func (m *Memory) ListPrefixData(prefixList []string) ([]interface{}, error) {
	result := make([]interface{}, 0)
	for _, item := range prefixList {
		if data, has := m.dataMap[item]; has {
			result = append(result, data.Data)
		}
	}
	if len(result) == 0 {
		return nil, item.NoDataError
	}
	return result, nil
}

func (m *Memory) CountPrefixData(prefixList []string) int {
	count := len(prefixList)
	for _, item := range prefixList {
		if _, has := m.dataMap[item]; !has {
			count--
		}
	}
	return count
}

func (m *Memory) ListPrefix(prePrefix string) ([]string, error) {
	result := make([]string, 0)
	for _, item := range m.prefixList {
		if strings.HasPrefix(item, prePrefix) {
			result = append(result, item)
		}
	}
	return result, nil
}

func (m *Memory) CountPrefix(prePrefix string) int {
	count := 0
	for _, item := range m.prefixList {
		if strings.HasPrefix(item, prePrefix) {
			count++
		}
	}
	return count
}

func (m *Memory) Insert(prefix string, data interface{}, opt ...item.Option) error {
	if _, has := m.dataMap[prefix]; has {
		return item.PrefixExisted
	}
	cacheItem := &item.Item{
		Prefix:    prefix,
		Data:      data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	for _, op := range opt {
		op(cacheItem)
	}
	m.dataMap[prefix] = cacheItem
	m.prefixList = append(m.prefixList, prefix)
	return nil
}

func (m *Memory) Update(prefix string, data interface{}, opt ...item.Option) error {
	cacheItem, has := m.dataMap[prefix]
	if !has {
		return item.NoDataError
	}
	cacheItem.Data = data
	for _, op := range opt {
		op(cacheItem)
	}
	m.dataMap[prefix] = cacheItem
	return nil
}

func (m *Memory) Delete(prefix string) (interface{}, error) {
	cacheItem, has := m.dataMap[prefix]
	if !has {
		logrus.Warningf("cache item of prefix %s not existed", prefix)
		return nil, item.PrefixNotExisted
	}
	delete(m.dataMap, prefix)
	for idx, item := range m.prefixList {
		if item == prefix {
			m.prefixList = append(m.prefixList[:idx], m.prefixList[idx+1:]...)
		}
	}
	return cacheItem.Data, nil
}
