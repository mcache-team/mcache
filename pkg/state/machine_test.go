package state

import "testing"

func TestStateMachineListByPrefix(t *testing.T) {
	sm := NewStateMachine()

	if err := sm.Insert("user/profile/name", "alice"); err != nil {
		t.Fatalf("insert name: %v", err)
	}
	if err := sm.Insert("user/profile/email", "alice@example.com"); err != nil {
		t.Fatalf("insert email: %v", err)
	}
	if err := sm.Insert("user/settings/theme", "dark"); err != nil {
		t.Fatalf("insert theme: %v", err)
	}

	items, err := sm.ListByPrefix("user/profile")
	if err != nil {
		t.Fatalf("list by prefix: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Prefix != "user/profile/email" || items[1].Prefix != "user/profile/name" {
		t.Fatalf("unexpected prefixes: %#v %#v", items[0].Prefix, items[1].Prefix)
	}
}

func TestStateMachineDeletePrunesBranch(t *testing.T) {
	sm := NewStateMachine()

	if err := sm.Insert("user/profile/name", "alice"); err != nil {
		t.Fatalf("insert: %v", err)
	}
	if _, err := sm.Delete("user/profile/name"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := sm.ListByPrefix("user/profile"); err == nil {
		t.Fatal("expected missing prefix after delete")
	}
	if count := sm.CountPrefix("user"); count != 0 {
		t.Fatalf("expected count 0, got %d", count)
	}
}

func TestStateMachineStats(t *testing.T) {
	sm := NewStateMachine()

	if err := sm.Insert("alpha/one", "1"); err != nil {
		t.Fatalf("insert alpha: %v", err)
	}
	if err := sm.Insert("beta/two", "2"); err != nil {
		t.Fatalf("insert beta: %v", err)
	}

	stats := sm.Stats()
	if stats.ItemCount != 2 {
		t.Fatalf("expected 2 items, got %d", stats.ItemCount)
	}
	if stats.RootCount != 2 {
		t.Fatalf("expected 2 roots, got %d", stats.RootCount)
	}
}
