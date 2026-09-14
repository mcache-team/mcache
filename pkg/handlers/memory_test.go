package handlers

import (
	"testing"
	"time"

	"github.com/mcache-team/mcache/pkg/cluster"
	"github.com/mcache-team/mcache/pkg/state"
)

func withLocalNode(t *testing.T) {
	t.Helper()
	prev := cluster.DefaultNode
	node, err := cluster.NewNode(cluster.Config{Mode: cluster.ModeSingle, NodeID: "test"}, state.DefaultStateMachine)
	if err != nil {
		t.Fatalf("new node: %v", err)
	}
	cluster.DefaultNode = node
	t.Cleanup(func() { cluster.DefaultNode = prev })
}

func TestSessionLifecycle(t *testing.T) {
	withLocalNode(t)
	h := NewMemoryHandler()
	if _, err := h.CreateSession("", "", 0); err == nil {
		t.Fatal("agentId is required")
	}
	sessionID, err := h.CreateSession("agent-1", "", 0)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	sessions, err := h.ListSessions("agent-1")
	if err != nil || len(sessions) != 1 || sessions[0] != sessionID {
		t.Fatalf("list sessions = %v, %v", sessions, err)
	}

	if _, err := h.AppendWorkingMemory("agent-1", "missing", "x", 0); err == nil {
		t.Fatal("appending to a missing session should fail")
	}
	if _, err := h.AppendWorkingMemory("agent-1", sessionID, "first", 0); err != nil {
		t.Fatalf("append working: %v", err)
	}
	if _, err := h.AppendWorkingMemory("agent-1", sessionID, "second", 0); err != nil {
		t.Fatalf("append working: %v", err)
	}
	items, err := h.ListWorkingMemory("agent-1", sessionID)
	if err != nil || len(items) != 2 {
		t.Fatalf("list working = %d items, %v", len(items), err)
	}

	if err := h.PutScratchpad("agent-1", sessionID, "result", "42", 0); err != nil {
		t.Fatalf("put scratchpad: %v", err)
	}
	if err := h.PutScratchpad("agent-1", sessionID, "result", "43", 0); err != nil {
		t.Fatalf("overwrite scratchpad: %v", err)
	}
	it, err := h.GetScratchpad("agent-1", sessionID, "result")
	if err != nil {
		t.Fatalf("get scratchpad: %v", err)
	}
	if content, ok := it.Data.(MemoryContent); !ok || content.Content != "43" {
		t.Fatalf("scratchpad value = %#v", it.Data)
	}

	removed, err := h.DeleteSession("agent-1", sessionID)
	if err != nil || removed != 4 {
		t.Fatalf("delete session removed=%d err=%v (want 4: marker+2working+scratch)", removed, err)
	}
	if sessions, _ := h.ListSessions("agent-1"); len(sessions) != 0 {
		t.Fatalf("sessions after delete: %v", sessions)
	}
}

func TestFacts(t *testing.T) {
	withLocalNode(t)
	h := NewMemoryHandler()
	if err := h.PutFact("agent-2", "lang", "go", 0); err != nil {
		t.Fatalf("put fact: %v", err)
	}
	if err := h.PutFact("agent-2", "lang", "golang", 0); err != nil {
		t.Fatalf("overwrite fact: %v", err)
	}
	if err := h.PutFact("agent-2", "team", "orca", 0); err != nil {
		t.Fatalf("put fact: %v", err)
	}
	facts, err := h.GetFacts("agent-2", "")
	if err != nil || len(facts) != 2 {
		t.Fatalf("list facts = %d, %v", len(facts), err)
	}
	one, err := h.GetFacts("agent-2", "lang")
	if err != nil || len(one) != 1 {
		t.Fatalf("get one fact = %d, %v", len(one), err)
	}
	if err := h.DeleteFact("agent-2", "lang"); err != nil {
		t.Fatalf("delete fact: %v", err)
	}
	if err := h.DeleteFact("agent-2", "lang"); err == nil {
		t.Fatal("deleting a missing fact should fail")
	}
}

func TestWorkingMemoryTTL(t *testing.T) {
	withLocalNode(t)
	h := NewMemoryHandler()
	sessionID, _ := h.CreateSession("agent-3", "", 0)
	if _, err := h.AppendWorkingMemory("agent-3", sessionID, "ephemeral", 1); err != nil {
		t.Fatalf("append working: %v", err)
	}
	time.Sleep(1100 * time.Millisecond)
	items, err := h.ListWorkingMemory("agent-3", sessionID)
	if err == nil && len(items) != 0 {
		t.Fatalf("expired working memory should disappear, got %d", len(items))
	}
}

func TestDefaultHandlerWiredToStateMachine(t *testing.T) {
	if Memory == nil {
		t.Fatal("default Memory handler must be initialized")
	}
	if Memory.stateMachine != state.DefaultStateMachine {
		t.Fatal("default handler should drive the default state machine")
	}
}
