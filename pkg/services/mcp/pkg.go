package mcp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mcache-team/mcache/pkg/cluster"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	serverName    = "mcache"
	serverVersion = "0.1.0"
)

// NewServer builds an MCP server that exposes mcache operations and agent
// memory as tools. It is transport-agnostic: callers may serve it over stdio
// or streamable HTTP.
func NewServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    serverName,
		Version: serverVersion,
	}, nil)
	registerCacheTools(server)
	registerMemoryTools(server)
	return server
}

// RunStdio serves the MCP server over stdin/stdout. This blocks until the
// stream is closed.
func RunStdio(ctx context.Context) error {
	return NewServer().Run(ctx, &mcp.StdioTransport{})
}

// HTTPHandler returns an http.Handler implementing the streamable HTTP MCP
// transport, ready to be mounted on the HTTP server.
func HTTPHandler() *mcp.StreamableHTTPHandler {
	return mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return NewServer() }, nil)
}

// toolError packs an error into the tool result so the LLM can see it and
// self-correct, as recommended by the MCP specification.
func toolError[Out any](err error) (*mcp.CallToolResult, Out, error) {
	var zero Out
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
	}, zero, nil
}

// leaderRedirectError renders a NotLeaderError into an actionable message.
func leaderRedirectError[Out any](op string, err error) (*mcp.CallToolResult, Out, error) {
	if redirect, ok := err.(*cluster.NotLeaderError); ok {
		return toolError[Out](fmt.Errorf("this node is not the leader; submit %q again at %s", op, redirect.LeaderAddress))
	}
	var zero Out
	return nil, zero, err
}
