package mcp

import (
	"context"
	"strings"
	"testing"

	"github.com/mcache-team/mcache/pkg/cluster"
	"github.com/mcache-team/mcache/pkg/state"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func newTestClient(t *testing.T) *mcp.ClientSession {
	t.Helper()
	prev := cluster.DefaultNode
	node, err := cluster.NewNode(cluster.Config{Mode: cluster.ModeSingle, NodeID: "test"}, state.DefaultStateMachine)
	if err != nil {
		t.Fatalf("new node: %v", err)
	}
	cluster.DefaultNode = node
	t.Cleanup(func() { cluster.DefaultNode = prev })

	ctx := context.Background()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	server := NewServer()
	go func() {
		_ = server.Run(ctx, serverTransport)
	}()
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0.0.0"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session
}

func callTool(t *testing.T, session *mcp.ClientSession, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		t.Fatalf("call %s: %v", name, err)
	}
	return result
}

func textOf(result *mcp.CallToolResult) string {
	var texts []string
	for _, c := range result.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			texts = append(texts, tc.Text)
		}
	}
	return strings.Join(texts, "\n")
}

func TestMCPListTools(t *testing.T) {
	session := newTestClient(t)
	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	want := map[string]bool{
		"cache_get": false, "cache_set": false, "cache_delete": false,
		"cache_list": false, "cache_list_items": false, "cache_count": false,
		"cluster_status":        false,
		"memory_create_session": false, "memory_list_sessions": false,
		"memory_delete_session": false, "memory_append_working": false,
		"memory_get_working": false, "memory_scratchpad_put": false,
		"memory_scratchpad_get": false, "memory_put_fact": false,
		"memory_get_facts": false, "memory_delete_fact": false,
	}
	for _, tool := range result.Tools {
		if _, ok := want[tool.Name]; ok {
			want[tool.Name] = true
		}
	}
	for name, seen := range want {
		if !seen {
			t.Errorf("tool %q not registered", name)
		}
	}
}

func TestMCPCacheRoundTrip(t *testing.T) {
	session := newTestClient(t)

	callTool(t, session, "cache_set", map[string]any{"prefix": "user/profile/name", "data": "alice", "ttlSeconds": 60})
	if got := textOf(callTool(t, session, "cache_get", map[string]any{"prefix": "user/profile/name"})); !strings.Contains(got, "alice") {
		t.Fatalf("cache_get = %q", got)
	}
	if got := textOf(callTool(t, session, "cache_count", map[string]any{})); !strings.Contains(got, "1") {
		t.Fatalf("cache_count = %q", got)
	}
	if got := textOf(callTool(t, session, "cache_list", map[string]any{})); !strings.Contains(got, "user/profile/name") {
		t.Fatalf("cache_list = %q", got)
	}
	// Upsert flips insert -> update.
	res := callTool(t, session, "cache_set", map[string]any{"prefix": "user/profile/name", "data": "bob"})
	if got := textOf(res); strings.Contains(got, "inserted") && !strings.Contains(got, "updated") {
		t.Fatalf("cache_set upsert = %q", got)
	}
	if got := textOf(callTool(t, session, "cache_get", map[string]any{"prefix": "user/profile/name"})); !strings.Contains(got, "bob") {
		t.Fatalf("cache_get after update = %q", got)
	}
	if got := textOf(callTool(t, session, "cache_delete", map[string]any{"prefix": "user/profile/name"})); strings.Contains(got, "false") {
		t.Fatalf("cache_delete = %q", got)
	}
	if got := textOf(callTool(t, session, "cache_get", map[string]any{"prefix": "user/profile/name"})); !strings.Contains(got, "no item found") {
		t.Fatalf("cache_get after delete = %q", got)
	}
}

func TestMCPMemoryFlow(t *testing.T) {
	session := newTestClient(t)

	res := callTool(t, session, "memory_create_session", map[string]any{"agentId": "bot-1"})
	sessionID := strings.TrimSpace(strings.Trim(textOf(res), `"{}: `+"`"+`"`))
	if sessionID == "" {
		t.Fatalf("memory_create_session = %q", textOf(res))
	}
	// Extract sessionId from structured output instead of guessing format.
	structured := res.StructuredContent
	if m, ok := structured.(map[string]any); ok {
		if id, ok := m["sessionId"].(string); ok && id != "" {
			sessionID = id
		}
	}

	callTool(t, session, "memory_append_working", map[string]any{"agentId": "bot-1", "sessionId": sessionID, "content": "user likes concise answers"})
	if got := textOf(callTool(t, session, "memory_get_working", map[string]any{"agentId": "bot-1", "sessionId": sessionID})); !strings.Contains(got, "concise") {
		t.Fatalf("memory_get_working = %q", got)
	}

	callTool(t, session, "memory_put_fact", map[string]any{"agentId": "bot-1", "key": "lang", "content": "go"})
	if got := textOf(callTool(t, session, "memory_get_facts", map[string]any{"agentId": "bot-1"})); !strings.Contains(got, "go") {
		t.Fatalf("memory_get_facts = %q", got)
	}

	if got := textOf(callTool(t, session, "memory_delete_fact", map[string]any{"agentId": "bot-1", "key": "lang"})); strings.Contains(got, "error") {
		t.Fatalf("memory_delete_fact = %q", got)
	}

	res = callTool(t, session, "memory_delete_session", map[string]any{"agentId": "bot-1", "sessionId": sessionID})
	if got := textOf(res); strings.Contains(got, "error") {
		t.Fatalf("memory_delete_session = %q", got)
	}
	if got := textOf(callTool(t, session, "memory_get_working", map[string]any{"agentId": "bot-1", "sessionId": sessionID})); strings.Contains(got, "concise") {
		t.Fatal("working memory should be gone after session deletion")
	}
}
