package state

import (
	"testing"
	"time"

	"github.com/mcache-team/mcache/pkg/apis/v1/item"
)

func TestExpiredItemNotVisibleInReads(t *testing.T) {
	m := NewStateMachine()
	expire := time.Now().Add(50 * time.Millisecond)
	if err := m.Insert("a/b", "v", item.WithExpireTime(expire)); err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	if _, err := m.GetOne("a/b"); err != nil {
		t.Fatalf("item should be readable before expiry: %v", err)
	}

	time.Sleep(80 * time.Millisecond)

	if _, err := m.GetOne("a/b"); err == nil {
		t.Fatal("expired item should not be returned by GetOne")
	}
	if keys, _ := m.ListPrefix(""); len(keys) != 0 {
		t.Fatalf("expired item should not be listed, got %v", keys)
	}
	if c := m.CountPrefix(""); c != 0 {
		t.Fatalf("expired item should not be counted, got %d", c)
	}
	if _, err := m.ListByPrefix("a"); err == nil {
		t.Fatal("expired item should not be listed by tree walk")
	}
}

func TestExpiredItemUpdateable(t *testing.T) {
	m := NewStateMachine()
	if err := m.Insert("a/b", "v1", item.WithExpireTime(time.Now().Add(-time.Second))); err != nil {
		t.Fatalf("insert failed: %v", err)
	}
	// Upsert path: update first.
	if err := m.Update("a/b", []byte("v2")); err != nil {
		t.Fatalf("update on expired item should still succeed: %v", err)
	}
	if err := m.Insert("a/b", "v3"); err != nil {
		t.Fatalf("insert on expired slot should clear it: %v", err)
	}
	if _, err := m.GetOne("a/b"); err != nil {
		t.Fatalf("item should be readable after reinsert: %v", err)
	}
}
