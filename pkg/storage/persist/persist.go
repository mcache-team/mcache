package persist

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/mcache-team/mcache/pkg/apis/v1/item"
	"github.com/sirupsen/logrus"
)

const defaultFile = "mcache-snapshot.json"

// Snapshotter is implemented by the storage backend.
type Snapshotter interface {
	Snapshot() ([]*item.Item, error)
	Restore(items []*item.Item) error
}

// Manager handles periodic flush and startup load.
type Manager struct {
	dir      string
	interval time.Duration
	store    Snapshotter
}

func New(store Snapshotter, dir string, interval time.Duration) *Manager {
	return &Manager{store: store, dir: dir, interval: interval}
}

// Load reads the snapshot file and restores data into storage.
func (m *Manager) Load() error {
	path := filepath.Join(m.dir, defaultFile)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		logrus.Infof("persist: no snapshot found at %s, starting fresh", path)
		return nil
	}
	if err != nil {
		return err
	}
	var items []*item.Item
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}
	if err := m.store.Restore(items); err != nil {
		return err
	}
	logrus.Infof("persist: loaded %d items from %s", len(items), path)
	return nil
}

// flush writes a snapshot to disk.
func (m *Manager) flush() {
	items, err := m.store.Snapshot()
	if err != nil {
		logrus.Errorf("persist: snapshot failed: %v", err)
		return
	}
	data, err := json.Marshal(items)
	if err != nil {
		logrus.Errorf("persist: marshal failed: %v", err)
		return
	}
	if err := os.MkdirAll(m.dir, 0755); err != nil {
		logrus.Errorf("persist: mkdir failed: %v", err)
		return
	}
	path := filepath.Join(m.dir, defaultFile)
	if err := os.WriteFile(path, data, 0644); err != nil {
		logrus.Errorf("persist: write failed: %v", err)
		return
	}
	logrus.Infof("persist: flushed %d items to %s", len(items), path)
}

// Start begins the periodic flush loop (blocking, run in goroutine).
func (m *Manager) Start() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()
	for range ticker.C {
		m.flush()
	}
}
