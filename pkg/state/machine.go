package state

import (
	"encoding/json"
	"io"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mcache-team/mcache/pkg/apis/v1/item"
)

var DefaultStateMachine = NewStateMachine()

type StateMachine struct {
	mu    sync.RWMutex
	items map[string]*item.Item
	roots map[string]*prefixNode
}

type Snapshot struct {
	Items map[string]*item.Item `json:"items"`
}

type Stats struct {
	ItemCount int `json:"itemCount"`
	RootCount int `json:"rootCount"`
}

type prefixNode struct {
	segment  string
	hasData  bool
	children map[string]*prefixNode
}

func NewStateMachine() *StateMachine {
	return &StateMachine{
		items: make(map[string]*item.Item),
		roots: make(map[string]*prefixNode),
	}
}

func (s *StateMachine) Apply(cmd Command) (interface{}, error) {
	switch cmd.Type {
	case CommandInsert:
		if cmd.Item == nil {
			return nil, item.NoDataError
		}
		return nil, s.insertItem(cmd.Prefix, cmd.Item)
	case CommandUpdate:
		if cmd.Item == nil {
			return nil, item.NoDataError
		}
		return nil, s.updateItem(cmd.Prefix, cmd.Item)
	case CommandDelete:
		return s.deleteItem(cmd.Prefix)
	default:
		return nil, item.NoDataError
	}
}

func (s *StateMachine) GetOne(prefix string) (*item.Item, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cacheItem, ok := s.items[normalizePrefix(prefix)]
	if !ok {
		return nil, item.NoDataError
	}
	return cloneItem(cacheItem), nil
}

func (s *StateMachine) ListPrefixData(prefixList []string) ([]*item.Item, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*item.Item, 0, len(prefixList))
	for _, prefix := range prefixList {
		cacheItem, ok := s.items[normalizePrefix(prefix)]
		if !ok {
			continue
		}
		result = append(result, cloneItem(cacheItem))
	}
	if len(result) == 0 {
		return nil, item.NoDataError
	}
	return result, nil
}

func (s *StateMachine) CountPrefixData(prefixList []string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, prefix := range prefixList {
		if _, ok := s.items[normalizePrefix(prefix)]; ok {
			count++
		}
	}
	return count
}

func (s *StateMachine) ListPrefix(prePrefix string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	prefix := normalizePrefix(prePrefix)
	result := make([]string, 0, len(s.items))
	for key := range s.items {
		if prefix == "" || strings.HasPrefix(key, prefix) {
			result = append(result, key)
		}
	}
	sort.Strings(result)
	return result, nil
}

func (s *StateMachine) CountPrefix(prePrefix string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	prefix := normalizePrefix(prePrefix)
	count := 0
	for key := range s.items {
		if prefix == "" || strings.HasPrefix(key, prefix) {
			count++
		}
	}
	return count
}

func (s *StateMachine) Insert(prefix string, data interface{}, opt ...item.Option) error {
	cacheItem := &item.Item{
		Prefix:    normalizePrefix(prefix),
		Data:      data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	for _, op := range opt {
		op(cacheItem)
	}
	_, err := s.Apply(Command{
		Type:   CommandInsert,
		Prefix: cacheItem.Prefix,
		Item:   cacheItem,
	})
	return err
}

func (s *StateMachine) Update(prefix string, data []byte, opt ...item.Option) error {
	cacheItem := &item.Item{
		Prefix:    normalizePrefix(prefix),
		Data:      data,
		UpdatedAt: time.Now(),
	}
	for _, op := range opt {
		op(cacheItem)
	}
	_, err := s.Apply(Command{
		Type:   CommandUpdate,
		Prefix: cacheItem.Prefix,
		Item:   cacheItem,
	})
	return err
}

func (s *StateMachine) Delete(prefix string) (interface{}, error) {
	return s.Apply(Command{
		Type:   CommandDelete,
		Prefix: normalizePrefix(prefix),
	})
}

func (s *StateMachine) ListByPrefix(prefix string) ([]*item.Item, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	parentPath := splitPrefix(prefix)
	if len(parentPath) == 0 {
		return nil, item.NoDataError
	}

	current, ok := s.roots[parentPath[0]]
	if !ok {
		return nil, item.NoDataError
	}

	for _, segment := range parentPath[1:] {
		next, ok := current.children[segment]
		if !ok {
			return nil, item.NoDataError
		}
		current = next
	}

	result := make([]*item.Item, 0, len(current.children))
	basePrefix := strings.Join(parentPath, "/")
	for name, child := range current.children {
		if !child.hasData {
			continue
		}
		fullPrefix := joinPrefix(basePrefix, name)
		cacheItem, ok := s.items[fullPrefix]
		if !ok {
			continue
		}
		result = append(result, cloneItem(cacheItem))
	}
	if len(result) == 0 {
		return nil, item.NoDataError
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Prefix < result[j].Prefix
	})
	return result, nil
}

func (s *StateMachine) Snapshot() *Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make(map[string]*item.Item, len(s.items))
	for key, value := range s.items {
		items[key] = cloneItem(value)
	}
	return &Snapshot{Items: items}
}

func (s *StateMachine) Stats() Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return Stats{
		ItemCount: len(s.items),
		RootCount: len(s.roots),
	}
}

func (s *StateMachine) Restore(snapshot *Snapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items = make(map[string]*item.Item, len(snapshot.Items))
	s.roots = make(map[string]*prefixNode)
	for key, value := range snapshot.Items {
		normalized := normalizePrefix(key)
		s.items[normalized] = cloneItem(value)
		s.ensurePath(normalized).hasData = true
	}
}

func (s *StateMachine) RestoreFromReader(reader io.Reader) error {
	snapshot := &Snapshot{}
	if err := json.NewDecoder(reader).Decode(snapshot); err != nil {
		return err
	}
	if snapshot.Items == nil {
		snapshot.Items = map[string]*item.Item{}
	}
	s.Restore(snapshot)
	return nil
}

func (s *StateMachine) insertItem(prefix string, newItem *item.Item) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	prefix = normalizePrefix(prefix)
	if _, ok := s.items[prefix]; ok {
		return item.PrefixExisted
	}
	s.items[prefix] = cloneItem(newItem)
	s.ensurePath(prefix).hasData = true
	return nil
}

func (s *StateMachine) updateItem(prefix string, updated *item.Item) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	prefix = normalizePrefix(prefix)
	current, ok := s.items[prefix]
	if !ok {
		return item.NoDataError
	}
	current.Data = updated.Data
	current.UpdatedAt = updated.UpdatedAt
	if updated.Timeout != 0 {
		current.Timeout = updated.Timeout
	}
	if !updated.ExpireTime.IsZero() {
		current.ExpireTime = updated.ExpireTime
	}
	return nil
}

func (s *StateMachine) deleteItem(prefix string) (interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	prefix = normalizePrefix(prefix)
	current, ok := s.items[prefix]
	if !ok {
		return nil, item.PrefixNotExisted
	}
	delete(s.items, prefix)
	s.prunePath(splitPrefix(prefix))
	return current.Data, nil
}

func (s *StateMachine) ensurePath(prefix string) *prefixNode {
	path := splitPrefix(prefix)
	root, ok := s.roots[path[0]]
	if !ok {
		root = &prefixNode{
			segment:  path[0],
			children: make(map[string]*prefixNode),
		}
		s.roots[path[0]] = root
	}
	current := root
	for _, segment := range path[1:] {
		next, ok := current.children[segment]
		if !ok {
			next = &prefixNode{
				segment:  segment,
				children: make(map[string]*prefixNode),
			}
			current.children[segment] = next
		}
		current = next
	}
	return current
}

func (s *StateMachine) prunePath(path []string) {
	if len(path) == 0 {
		return
	}
	root, ok := s.roots[path[0]]
	if !ok {
		return
	}
	if len(path) == 1 {
		root.hasData = false
		if len(root.children) == 0 {
			delete(s.roots, path[0])
		}
		return
	}

	stack := make([]*prefixNode, 0, len(path))
	stack = append(stack, root)
	current := root
	for _, segment := range path[1:] {
		next, ok := current.children[segment]
		if !ok {
			return
		}
		stack = append(stack, next)
		current = next
	}

	stack[len(stack)-1].hasData = false
	for index := len(stack) - 1; index >= 1; index-- {
		node := stack[index]
		parent := stack[index-1]
		if node.hasData || len(node.children) > 0 {
			break
		}
		delete(parent.children, node.segment)
	}

	if !root.hasData && len(root.children) == 0 {
		delete(s.roots, path[0])
	}
}

func splitPrefix(prefix string) []string {
	normalized := normalizePrefix(prefix)
	if normalized == "" {
		return nil
	}
	return strings.Split(normalized, "/")
}

func normalizePrefix(prefix string) string {
	return strings.Trim(strings.TrimSpace(prefix), "/")
}

func joinPrefix(parent, child string) string {
	if parent == "" {
		return child
	}
	return parent + "/" + child
}

func cloneItem(src *item.Item) *item.Item {
	if src == nil {
		return nil
	}
	dst := *src
	return &dst
}
