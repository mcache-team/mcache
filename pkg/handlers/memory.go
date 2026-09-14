package handlers

import (
	"fmt"
	"sort"
	"time"

	"github.com/mcache-team/mcache/pkg/apis/v1/item"
	"github.com/mcache-team/mcache/pkg/cluster"
	"github.com/mcache-team/mcache/pkg/state"
)

// MemoryHandler implements an opinionated hierarchical layout for agent
// memory on top of the mcache state machine:
//
//	memory/agent/{agentID}/sessions/{sessionID}/_session      session marker
//	memory/agent/{agentID}/sessions/{sessionID}/working/{seq} working memory, TTL-backed
//	memory/agent/{agentID}/sessions/{sessionID}/scratch/{key} per-session scratchpad
//	memory/agent/{agentID}/facts/{key}                        long-lived facts
type MemoryHandler struct {
	stateMachine *state.StateMachine
}

func NewMemoryHandler() *MemoryHandler {
	return &MemoryHandler{
		stateMachine: state.DefaultStateMachine,
	}
}

// MemoryContent is the stored payload of every memory entry.
type MemoryContent struct {
	AgentID   string `json:"agentId,omitempty"`
	SessionID string `json:"sessionId,omitempty"`
	Key       string `json:"key,omitempty"`
	Content   string `json:"content"`
	Kind      string `json:"kind"`
}

// CreateSession registers a new session for an agent. If sessionID is empty a
// time-based id is generated.
func (m *MemoryHandler) CreateSession(agentID, sessionID string, ttlSeconds int64) (string, error) {
	if agentID == "" {
		return "", fmt.Errorf("agentId is required")
	}
	if sessionID == "" {
		sessionID = fmt.Sprintf("s%016x", time.Now().UnixNano())
	}
	timeout, expire := m.ttl(ttlSeconds)
	_, err := m.submit(state.Command{
		Type:   state.CommandInsert,
		Prefix: m.sessionKey(agentID, sessionID),
		Item: &item.Item{
			Prefix:     m.sessionKey(agentID, sessionID),
			Data:       MemoryContent{AgentID: agentID, SessionID: sessionID, Content: "session", Kind: "session"},
			Timeout:    timeout,
			ExpireTime: expire,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	})
	if err == item.PrefixExisted {
		return sessionID, nil
	}
	return sessionID, err
}

// AppendWorkingMemory appends an ordered entry to a session's working memory
// and returns the generated entry id.
func (m *MemoryHandler) AppendWorkingMemory(agentID, sessionID, content string, ttlSeconds int64) (string, error) {
	if agentID == "" || sessionID == "" {
		return "", fmt.Errorf("agentId and sessionId are required")
	}
	if content == "" {
		return "", fmt.Errorf("content is required")
	}
	if _, err := m.stateMachine.GetOne(m.sessionKey(agentID, sessionID)); err != nil {
		return "", fmt.Errorf("session %q not found for agent %q", sessionID, agentID)
	}
	entryID := fmt.Sprintf("w%020d", time.Now().UnixNano())
	key := m.workingKey(agentID, sessionID, entryID)
	timeout, expire := m.ttl(ttlSeconds)
	_, err := m.submit(state.Command{
		Type:   state.CommandInsert,
		Prefix: key,
		Item: &item.Item{
			Prefix:     key,
			Data:       MemoryContent{AgentID: agentID, SessionID: sessionID, Content: content, Kind: "working"},
			Timeout:    timeout,
			ExpireTime: expire,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	})
	if err == item.PrefixExisted {
		// Clock collision; bump until unique.
		entryID = fmt.Sprintf("w%020d%d", time.Now().UnixNano(), time.Now().UnixMicro()%1000)
		key = m.workingKey(agentID, sessionID, entryID)
		_, err := m.submit(state.Command{
			Type:   state.CommandInsert,
			Prefix: key,
			Item: &item.Item{
				Prefix:     key,
				Data:       MemoryContent{AgentID: agentID, SessionID: sessionID, Content: content, Kind: "working"},
				Timeout:    timeout,
				ExpireTime: expire,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
		})
		return entryID, err
	}
	return entryID, err
}

// ListWorkingMemory returns the surviving working-memory entries of a session
// in insertion order.
func (m *MemoryHandler) ListWorkingMemory(agentID, sessionID string) ([]*item.Item, error) {
	return m.stateMachine.ListByPrefix(m.workingPrefix(agentID, sessionID))
}

// PutScratchpad stores a named scratchpad value for a session.
func (m *MemoryHandler) PutScratchpad(agentID, sessionID, key, content string, ttlSeconds int64) error {
	if agentID == "" || sessionID == "" || key == "" {
		return fmt.Errorf("agentId, sessionId and key are required")
	}
	if _, err := m.stateMachine.GetOne(m.sessionKey(agentID, sessionID)); err != nil {
		return fmt.Errorf("session %q not found for agent %q", sessionID, agentID)
	}
	if _, err := m.stateMachine.GetOne(m.scratchKey(agentID, sessionID, key)); err == nil {
		_, err := m.submit(state.Command{
			Type:   state.CommandUpdate,
			Prefix: m.scratchKey(agentID, sessionID, key),
			Item:   m.entry(MemoryContent{AgentID: agentID, SessionID: sessionID, Key: key, Content: content, Kind: "scratch"}, ttlSeconds, time.Now()),
		})
		return err
	}
	timeout, expire := m.ttl(ttlSeconds)
	_, err := m.submit(state.Command{
		Type:   state.CommandInsert,
		Prefix: m.scratchKey(agentID, sessionID, key),
		Item: &item.Item{
			Prefix:     m.scratchKey(agentID, sessionID, key),
			Data:       MemoryContent{AgentID: agentID, SessionID: sessionID, Key: key, Content: content, Kind: "scratch"},
			Timeout:    timeout,
			ExpireTime: expire,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	})
	return err
}

// GetScratchpad reads one scratchpad value.
func (m *MemoryHandler) GetScratchpad(agentID, sessionID, key string) (*item.Item, error) {
	return m.stateMachine.GetOne(m.scratchKey(agentID, sessionID, key))
}

// ListScratchpad lists all scratchpad values of a session.
func (m *MemoryHandler) ListScratchpad(agentID, sessionID string) ([]*item.Item, error) {
	return m.stateMachine.ListByPrefix(m.scratchPrefix(agentID, sessionID))
}

// PutFact stores a long-lived fact for an agent.
func (m *MemoryHandler) PutFact(agentID, key, content string, ttlSeconds int64) error {
	if agentID == "" || key == "" {
		return fmt.Errorf("agentId and key are required")
	}
	if _, err := m.stateMachine.GetOne(m.factKey(agentID, key)); err == nil {
		_, err := m.submit(state.Command{
			Type:   state.CommandUpdate,
			Prefix: m.factKey(agentID, key),
			Item:   m.entry(MemoryContent{AgentID: agentID, Key: key, Content: content, Kind: "fact"}, ttlSeconds, time.Now()),
		})
		return err
	}
	timeout, expire := m.ttl(ttlSeconds)
	_, err := m.submit(state.Command{
		Type:   state.CommandInsert,
		Prefix: m.factKey(agentID, key),
		Item: &item.Item{
			Prefix:     m.factKey(agentID, key),
			Data:       MemoryContent{AgentID: agentID, Key: key, Content: content, Kind: "fact"},
			Timeout:    timeout,
			ExpireTime: expire,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	})
	return err
}

// GetFacts reads one fact (when key is set) or lists all facts of an agent.
func (m *MemoryHandler) GetFacts(agentID, key string) ([]*item.Item, error) {
	if key != "" {
		return m.stateMachine.ListPrefixData([]string{m.factKey(agentID, key)})
	}
	return m.stateMachine.ListByPrefix(m.factPrefix(agentID))
}

// DeleteFact removes a fact.
func (m *MemoryHandler) DeleteFact(agentID, key string) error {
	_, err := m.submit(state.Command{
		Type:   state.CommandDelete,
		Prefix: m.factKey(agentID, key),
	})
	return err
}

// ListSessions lists the surviving sessions of an agent.
func (m *MemoryHandler) ListSessions(agentID string) ([]string, error) {
	keys, err := m.stateMachine.ListPrefix(m.sessionsPrefix(agentID))
	if err != nil {
		return nil, err
	}
	sessions := make([]string, 0, len(keys))
	for _, key := range keys {
		if ok, _ := m.isSessionMarker(key); ok {
			sessions = append(sessions, sessionOf(key))
		}
	}
	sort.Strings(sessions)
	return sessions, nil
}

// DeleteSession removes a session together with its working memory and
// scratchpad entries.
func (m *MemoryHandler) DeleteSession(agentID, sessionID string) (int, error) {
	prefixes := []string{
		m.sessionKey(agentID, sessionID),
		m.workingPrefix(agentID, sessionID),
		m.scratchPrefix(agentID, sessionID),
	}
	removed := 0
	for _, prefix := range prefixes {
		keys, err := m.stateMachine.ListPrefix(prefix)
		if err != nil {
			continue
		}
		for _, key := range keys {
			if _, err := m.submit(state.Command{Type: state.CommandDelete, Prefix: key}); err == nil {
				removed++
			}
		}
	}
	if removed == 0 {
		return 0, fmt.Errorf("session %q not found for agent %q", sessionID, agentID)
	}
	return removed, nil
}

func (m *MemoryHandler) sessionKey(agentID, sessionID string) string {
	return fmt.Sprintf("memory/agent/%s/sessions/%s/_session", agentID, sessionID)
}

func (m *MemoryHandler) sessionsPrefix(agentID string) string {
	return fmt.Sprintf("memory/agent/%s/sessions/", agentID)
}

func (m *MemoryHandler) workingPrefix(agentID, sessionID string) string {
	return fmt.Sprintf("memory/agent/%s/sessions/%s/working/", agentID, sessionID)
}

func (m *MemoryHandler) workingKey(agentID, sessionID, entryID string) string {
	return fmt.Sprintf("memory/agent/%s/sessions/%s/working/%s", agentID, sessionID, entryID)
}

func (m *MemoryHandler) scratchPrefix(agentID, sessionID string) string {
	return fmt.Sprintf("memory/agent/%s/sessions/%s/scratch/", agentID, sessionID)
}

func (m *MemoryHandler) scratchKey(agentID, sessionID, key string) string {
	return fmt.Sprintf("memory/agent/%s/sessions/%s/scratch/%s", agentID, sessionID, key)
}

func (m *MemoryHandler) factPrefix(agentID string) string {
	return fmt.Sprintf("memory/agent/%s/facts/", agentID)
}

func (m *MemoryHandler) factKey(agentID, key string) string {
	return fmt.Sprintf("memory/agent/%s/facts/%s", agentID, key)
}

func sessionOf(key string) string {
	parts := splitKey(key)
	if len(parts) >= 5 {
		return parts[4]
	}
	return ""
}

func (m *MemoryHandler) isSessionMarker(key string) (bool, error) {
	parts := splitKey(key)
	if len(parts) == 6 && parts[5] == "_session" {
		_, err := m.stateMachine.GetOne(key)
		return err == nil, err
	}
	return false, nil
}

func splitKey(key string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(key); i++ {
		if key[i] == '/' {
			parts = append(parts, key[start:i])
			start = i + 1
		}
	}
	return append(parts, key[start:])
}

func (m *MemoryHandler) entry(content MemoryContent, ttlSeconds int64, now time.Time) *item.Item {
	timeout, expire := m.ttl(ttlSeconds)
	return &item.Item{
		Data:       content,
		Timeout:    timeout,
		ExpireTime: expire,
		UpdatedAt:  now,
	}
}

func (m *MemoryHandler) ttl(ttlSeconds int64) (time.Duration, time.Time) {
	if ttlSeconds <= 0 {
		return 0, time.Time{}
	}
	timeout := time.Duration(ttlSeconds) * time.Second
	return timeout, time.Now().Add(timeout)
}

func (m *MemoryHandler) submit(cmd state.Command) (interface{}, error) {
	return cluster.DefaultNode.Submit(cmd)
}
