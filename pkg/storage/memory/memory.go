package memory

import (
	"github.com/mcache-team/mcache/pkg/apis/v1/item"
	"github.com/mcache-team/mcache/pkg/apis/v1/storage"
)

type Memory struct {
	prefixList []string
	dataMap    map[string][]*item.Item
}

var _ storage.Storage = &Memory{}

func (m *Memory) GetOne(prefix string) (interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (m *Memory) ListPrefixData(prefix string) ([]interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (m *Memory) CountPrefixData(prefix string) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *Memory) ListPrefix(prefixStr string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (m *Memory) CountPrefix(prefixStr string) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *Memory) Insert(prefix string, data interface{}, opt ...item.Option) error {
	//TODO implement me
	panic("implement me")
}

func (m *Memory) Update(prefix string, data interface{}, opt ...item.Option) error {
	//TODO implement me
	panic("implement me")
}

func (m *Memory) Delete(prefix string) error {
	//TODO implement me
	panic("implement me")
}
