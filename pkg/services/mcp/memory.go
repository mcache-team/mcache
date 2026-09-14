package mcp

import (
	"context"
	"fmt"

	"github.com/mcache-team/mcache/pkg/apis/v1/item"
	"github.com/mcache-team/mcache/pkg/handlers"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerMemoryTools exposes the agent-memory layer as MCP tools so agents
// can persist and retrieve their working memory, facts and scratchpad values
// across conversations.
func registerMemoryTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "memory_create_session",
		Description: "Create a session for an agent to group working memory and scratchpad entries. A sessionId is generated when omitted.",
	}, memoryCreateSession)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "memory_list_sessions",
		Description: "List the sessions of an agent.",
	}, memoryListSessions)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "memory_delete_session",
		Description: "Delete a session together with all of its working-memory and scratchpad entries.",
	}, memoryDeleteSession)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "memory_append_working",
		Description: "Append an ordered working-memory entry to a session, optionally with a TTL. Use it for conversation context the agent wants to survive later steps or sessions.",
	}, memoryAppendWorking)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "memory_get_working",
		Description: "Read a session's working-memory entries in insertion order.",
	}, memoryGetWorking)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "memory_scratchpad_put",
		Description: "Store a named scratchpad value for a session (intermediate results, notes, tool outputs).",
	}, memoryScratchpadPut)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "memory_scratchpad_get",
		Description: "Read one named scratchpad value of a session.",
	}, memoryScratchpadGet)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "memory_put_fact",
		Description: "Store a long-lived fact for an agent (user preferences, environment details). Facts persist until deleted unless a TTL is given.",
	}, memoryPutFact)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "memory_get_facts",
		Description: "Read one fact by key, or all facts of an agent when key is omitted.",
	}, memoryGetFacts)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "memory_delete_fact",
		Description: "Delete one fact of an agent.",
	}, memoryDeleteFact)
}

type memoryAgentArgs struct {
	AgentID string `json:"agentId" jsonschema:"unique identifier of the agent owning the memory"`
}

type memoryCreateSessionArgs struct {
	AgentID    string `json:"agentId"`
	SessionID  string `json:"sessionId,omitempty" jsonschema:"optional session identifier; generated when omitted"`
	TTLSeconds int64  `json:"ttlSeconds,omitempty" jsonschema:"optional session lifetime in seconds"`
}

type memoryCreateSessionResult struct {
	AgentID   string `json:"agentId"`
	SessionID string `json:"sessionId"`
}

func memoryCreateSession(_ context.Context, _ *mcp.CallToolRequest, args memoryCreateSessionArgs) (*mcp.CallToolResult, *memoryCreateSessionResult, error) {
	sessionID, err := handlers.Memory.CreateSession(args.AgentID, args.SessionID, args.TTLSeconds)
	if err != nil {
		return toolError[*memoryCreateSessionResult](err)
	}
	return nil, &memoryCreateSessionResult{AgentID: args.AgentID, SessionID: sessionID}, nil
}

type memoryListSessionsResult struct {
	Sessions []string `json:"sessions"`
}

func memoryListSessions(_ context.Context, _ *mcp.CallToolRequest, args memoryAgentArgs) (*mcp.CallToolResult, *memoryListSessionsResult, error) {
	sessions, err := handlers.Memory.ListSessions(args.AgentID)
	if err != nil || sessions == nil {
		sessions = []string{}
	}
	return nil, &memoryListSessionsResult{Sessions: sessions}, nil
}

type memoryDeleteSessionArgs struct {
	AgentID   string `json:"agentId"`
	SessionID string `json:"sessionId"`
}

type memoryDeleteSessionResult struct {
	Removed int `json:"removed"`
}

func memoryDeleteSession(_ context.Context, _ *mcp.CallToolRequest, args memoryDeleteSessionArgs) (*mcp.CallToolResult, *memoryDeleteSessionResult, error) {
	removed, err := handlers.Memory.DeleteSession(args.AgentID, args.SessionID)
	if err != nil {
		return toolError[*memoryDeleteSessionResult](err)
	}
	return nil, &memoryDeleteSessionResult{Removed: removed}, nil
}

type memoryAppendWorkingArgs struct {
	AgentID    string `json:"agentId"`
	SessionID  string `json:"sessionId"`
	Content    string `json:"content" jsonschema:"the information to remember"`
	TTLSeconds int64  `json:"ttlSeconds,omitempty" jsonschema:"optional lifetime in seconds; 0 means keep until the session is deleted"`
}

type memoryAppendWorkingResult struct {
	EntryID string `json:"entryId"`
}

func memoryAppendWorking(_ context.Context, _ *mcp.CallToolRequest, args memoryAppendWorkingArgs) (*mcp.CallToolResult, *memoryAppendWorkingResult, error) {
	entryID, err := handlers.Memory.AppendWorkingMemory(args.AgentID, args.SessionID, args.Content, args.TTLSeconds)
	if err != nil {
		return toolError[*memoryAppendWorkingResult](err)
	}
	return nil, &memoryAppendWorkingResult{EntryID: entryID}, nil
}

type memoryGetWorkingArgs struct {
	AgentID   string `json:"agentId"`
	SessionID string `json:"sessionId"`
}

type memoryGetWorkingResult struct {
	Entries []memoryEntry `json:"entries"`
}

func memoryGetWorking(_ context.Context, _ *mcp.CallToolRequest, args memoryGetWorkingArgs) (*mcp.CallToolResult, *memoryGetWorkingResult, error) {
	items, err := handlers.Memory.ListWorkingMemory(args.AgentID, args.SessionID)
	if err != nil {
		return nil, &memoryGetWorkingResult{Entries: []memoryEntry{}}, nil
	}
	return nil, &memoryGetWorkingResult{Entries: toEntries(items)}, nil
}

type memoryScratchpadPutArgs struct {
	AgentID    string `json:"agentId"`
	SessionID  string `json:"sessionId"`
	Key        string `json:"key"`
	Content    string `json:"content"`
	TTLSeconds int64  `json:"ttlSeconds,omitempty"`
}

func memoryScratchpadPut(_ context.Context, _ *mcp.CallToolRequest, args memoryScratchpadPutArgs) (*mcp.CallToolResult, *emptyResult, error) {
	if err := handlers.Memory.PutScratchpad(args.AgentID, args.SessionID, args.Key, args.Content, args.TTLSeconds); err != nil {
		return toolError[*emptyResult](err)
	}
	return nil, &emptyResult{OK: true}, nil
}

type memoryScratchpadGetArgs struct {
	AgentID   string `json:"agentId"`
	SessionID string `json:"sessionId"`
	Key       string `json:"key"`
}

type memoryScratchpadGetResult struct {
	Found   bool        `json:"found"`
	Content interface{} `json:"content,omitempty"`
}

func memoryScratchpadGet(_ context.Context, _ *mcp.CallToolRequest, args memoryScratchpadGetArgs) (*mcp.CallToolResult, *memoryScratchpadGetResult, error) {
	it, err := handlers.Memory.GetScratchpad(args.AgentID, args.SessionID, args.Key)
	if err != nil {
		return nil, &memoryScratchpadGetResult{Found: false}, nil
	}
	return nil, &memoryScratchpadGetResult{Found: true, Content: entryValue(it)}, nil
}

type memoryPutFactArgs struct {
	AgentID    string `json:"agentId"`
	Key        string `json:"key" jsonschema:"a short name for the fact, e.g. preferred-language"`
	Content    string `json:"content"`
	TTLSeconds int64  `json:"ttlSeconds,omitempty"`
}

func memoryPutFact(_ context.Context, _ *mcp.CallToolRequest, args memoryPutFactArgs) (*mcp.CallToolResult, *emptyResult, error) {
	if err := handlers.Memory.PutFact(args.AgentID, args.Key, args.Content, args.TTLSeconds); err != nil {
		return toolError[*emptyResult](err)
	}
	return nil, &emptyResult{OK: true}, nil
}

type memoryGetFactsArgs struct {
	AgentID string `json:"agentId"`
	Key     string `json:"key,omitempty" jsonschema:"optional fact key; omitted lists all facts"`
}

type memoryGetFactsResult struct {
	Facts []memoryEntry `json:"facts"`
}

func memoryGetFacts(_ context.Context, _ *mcp.CallToolRequest, args memoryGetFactsArgs) (*mcp.CallToolResult, *memoryGetFactsResult, error) {
	items, err := handlers.Memory.GetFacts(args.AgentID, args.Key)
	if err != nil {
		return nil, &memoryGetFactsResult{Facts: []memoryEntry{}}, nil
	}
	return nil, &memoryGetFactsResult{Facts: toEntries(items)}, nil
}

type memoryDeleteFactArgs struct {
	AgentID string `json:"agentId"`
	Key     string `json:"key"`
}

func memoryDeleteFact(_ context.Context, _ *mcp.CallToolRequest, args memoryDeleteFactArgs) (*mcp.CallToolResult, *emptyResult, error) {
	if err := handlers.Memory.DeleteFact(args.AgentID, args.Key); err != nil {
		if err == item.PrefixNotExisted {
			return toolError[*emptyResult](fmt.Errorf("fact %q not found for agent %q", args.Key, args.AgentID))
		}
		return leaderRedirectError[*emptyResult]("memory_delete_fact", err)
	}
	return nil, &emptyResult{OK: true}, nil
}

type emptyResult struct {
	OK bool `json:"ok"`
}

type memoryEntry struct {
	Key     string      `json:"key"`
	Content string      `json:"content,omitempty"`
	Raw     interface{} `json:"value,omitempty"`
}

func toEntries(items []*item.Item) []memoryEntry {
	entries := make([]memoryEntry, 0, len(items))
	for _, it := range items {
		entries = append(entries, memoryEntry{
			Key:     it.Prefix,
			Content: contentText(it.Data),
			Raw:     entryValue(it),
		})
	}
	return entries
}

func contentText(data interface{}) string {
	if c, ok := data.(handlers.MemoryContent); ok {
		return c.Content
	}
	return ""
}

func entryValue(it *item.Item) interface{} {
	if it == nil {
		return nil
	}
	return it.Data
}
