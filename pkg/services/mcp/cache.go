package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mcache-team/mcache/pkg/apis/v1/item"
	"github.com/mcache-team/mcache/pkg/cluster"
	"github.com/mcache-team/mcache/pkg/state"
	"github.com/mcache-team/mcache/pkg/storage"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerCacheTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "cache_get",
		Description: "Get a single cached item by its exact hierarchical key path (e.g. \"user/profile/name\").",
	}, cacheGet)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "cache_set",
		Description: "Insert or update a cached item at a hierarchical key path. Set ttlSeconds>0 to make it expire.",
	}, cacheSet)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "cache_delete",
		Description: "Delete a cached item by its exact key path.",
	}, cacheDelete)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "cache_list",
		Description: "List all key paths under a prefix. Pass an empty prefix to list every key in the store.",
	}, cacheList)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "cache_list_items",
		Description: "List the full items (key paths plus their stored values) under a prefix.",
	}, cacheListItems)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "cache_count",
		Description: "Count the cached items under a prefix.",
	}, cacheCount)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "cluster_status",
		Description: "Show mcache server status: mode, node id, leader and whether the node is ready.",
	}, clusterStatus)
}

type cacheGetArgs struct {
	Prefix string `json:"prefix" jsonschema:"the exact key path of the item, e.g. user/profile/name"`
}

type cacheGetResult struct {
	Prefix     string      `json:"prefix"`
	Data       interface{} `json:"data,omitempty"`
	ExpireTime string      `json:"expireTime,omitempty"`
	CreatedAt  string      `json:"createdAt,omitempty"`
	UpdatedAt  string      `json:"updatedAt,omitempty"`
}

func cacheGet(_ context.Context, _ *mcp.CallToolRequest, args cacheGetArgs) (*mcp.CallToolResult, *cacheGetResult, error) {
	if args.Prefix == "" {
		return toolError[*cacheGetResult](fmt.Errorf("prefix is required"))
	}
	it, err := storage.StorageClient.GetOne(args.Prefix)
	if err != nil {
		return toolError[*cacheGetResult](fmt.Errorf("no item found at prefix %q", args.Prefix))
	}
	return nil, &cacheGetResult{
		Prefix:     it.Prefix,
		Data:       it.Data,
		ExpireTime: formatTime(it.ExpireTime),
		CreatedAt:  formatTime(it.CreatedAt),
		UpdatedAt:  formatTime(it.UpdatedAt),
	}, nil
}

type cacheSetArgs struct {
	Prefix     string      `json:"prefix" jsonschema:"the key path to write, e.g. user/profile/name"`
	Data       interface{} `json:"data" jsonschema:"the JSON value to store"`
	TTLSeconds int64       `json:"ttlSeconds,omitempty" jsonschema:"optional time-to-live in seconds; 0 means never expire"`
	Upsert     *bool       `json:"upsert,omitempty" jsonschema:"true (default) to update when the key exists, false to fail if it already exists"`
}

type cacheSetResult struct {
	Prefix string `json:"prefix"`
	Action string `json:"action" jsonschema:"whether the item was inserted or updated"`
}

func cacheSet(_ context.Context, _ *mcp.CallToolRequest, args cacheSetArgs) (*mcp.CallToolResult, *cacheSetResult, error) {
	if args.Prefix == "" {
		return toolError[*cacheSetResult](fmt.Errorf("prefix is required"))
	}
	if args.Data == nil {
		return toolError[*cacheSetResult](fmt.Errorf("data is required"))
	}
	upsert := args.Upsert == nil || *args.Upsert

	now := time.Now()
	timeout, expire := applyTTL(args.TTLSeconds)
	cmd := state.Command{
		Prefix: args.Prefix,
		Item: &item.Item{
			Prefix:     args.Prefix,
			Data:       args.Data,
			Timeout:    timeout,
			ExpireTime: expire,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}
	if upsert {
		cmd.Type = state.CommandUpdate
		if _, err := cluster.DefaultNode.Submit(cmd); err != nil {
			if err != item.NoDataError {
				return leaderRedirectError[*cacheSetResult]("cache_set", err)
			}
			cmd.Type = state.CommandInsert
			if _, err := cluster.DefaultNode.Submit(cmd); err != nil {
				return leaderRedirectError[*cacheSetResult]("cache_set", err)
			}
			return nil, &cacheSetResult{Prefix: args.Prefix, Action: "insert"}, nil
		}
		return nil, &cacheSetResult{Prefix: args.Prefix, Action: "update"}, nil
	}
	cmd.Type = state.CommandInsert
	if _, err := cluster.DefaultNode.Submit(cmd); err != nil {
		if err == item.PrefixExisted {
			return toolError[*cacheSetResult](fmt.Errorf("prefix %q already exists", args.Prefix))
		}
		return leaderRedirectError[*cacheSetResult]("cache_set", err)
	}
	return nil, &cacheSetResult{Prefix: args.Prefix, Action: "insert"}, nil
}

type cacheDeleteArgs struct {
	Prefix string `json:"prefix" jsonschema:"the exact key path of the item to delete"`
}

type cacheDeleteResult struct {
	Prefix  string      `json:"prefix"`
	Deleted bool        `json:"deleted"`
	Data    interface{} `json:"data,omitempty"`
}

func cacheDelete(_ context.Context, _ *mcp.CallToolRequest, args cacheDeleteArgs) (*mcp.CallToolResult, *cacheDeleteResult, error) {
	if args.Prefix == "" {
		return toolError[*cacheDeleteResult](fmt.Errorf("prefix is required"))
	}
	data, err := cluster.DefaultNode.Submit(state.Command{
		Type:   state.CommandDelete,
		Prefix: args.Prefix,
	})
	if err != nil {
		if err == item.PrefixNotExisted {
			return nil, &cacheDeleteResult{Prefix: args.Prefix, Deleted: false}, nil
		}
		return leaderRedirectError[*cacheDeleteResult]("cache_delete", err)
	}
	return nil, &cacheDeleteResult{Prefix: args.Prefix, Deleted: true, Data: deletedValue(data)}, nil
}

type cacheListArgs struct {
	Prefix string `json:"prefix,omitempty" jsonschema:"optional key path prefix; empty lists all keys"`
}

type cacheListResult struct {
	Keys []string `json:"keys"`
}

func cacheList(_ context.Context, _ *mcp.CallToolRequest, args cacheListArgs) (*mcp.CallToolResult, *cacheListResult, error) {
	keys, err := storage.StorageClient.ListPrefix(args.Prefix)
	if err != nil {
		return nil, &cacheListResult{Keys: []string{}}, nil
	}
	if keys == nil {
		keys = []string{}
	}
	return nil, &cacheListResult{Keys: keys}, nil
}

type cacheListItemsArgs struct {
	Prefix string `json:"prefix,omitempty" jsonschema:"optional key path prefix; empty lists all items"`
}

type cacheListItemsResult struct {
	Items []cacheGetResult `json:"items"`
}

func cacheListItems(_ context.Context, _ *mcp.CallToolRequest, args cacheListItemsArgs) (*mcp.CallToolResult, *cacheListItemsResult, error) {
	items, err := storage.StorageClient.ListPrefixData([]string{args.Prefix})
	if err != nil {
		return nil, &cacheListItemsResult{Items: []cacheGetResult{}}, nil
	}
	result := make([]cacheGetResult, 0, len(items))
	for _, it := range items {
		result = append(result, cacheGetResult{
			Prefix:     it.Prefix,
			Data:       it.Data,
			ExpireTime: formatTime(it.ExpireTime),
			CreatedAt:  formatTime(it.CreatedAt),
			UpdatedAt:  formatTime(it.UpdatedAt),
		})
	}
	return nil, &cacheListItemsResult{Items: result}, nil
}

type cacheCountArgs struct {
	Prefix string `json:"prefix,omitempty" jsonschema:"optional key path prefix; empty counts all items"`
}

type cacheCountResult struct {
	Count int `json:"count"`
}

func cacheCount(_ context.Context, _ *mcp.CallToolRequest, args cacheCountArgs) (*mcp.CallToolResult, *cacheCountResult, error) {
	return nil, &cacheCountResult{Count: storage.StorageClient.CountPrefix(args.Prefix)}, nil
}

type clusterStatusArgs struct{}

type clusterStatusResult struct {
	Status cluster.Status `json:"status"`
	Ready  bool           `json:"ready"`
}

func clusterStatus(_ context.Context, _ *mcp.CallToolRequest, _ clusterStatusArgs) (*mcp.CallToolResult, *clusterStatusResult, error) {
	ready := cluster.DefaultNode.Ready() == nil
	return nil, &clusterStatusResult{Status: cluster.DefaultNode.Status(), Ready: ready}, nil
}

func applyTTL(seconds int64) (time.Duration, time.Time) {
	if seconds <= 0 {
		return 0, time.Time{}
	}
	timeout := time.Duration(seconds) * time.Second
	return timeout, time.Now().Add(timeout)
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func deletedValue(data interface{}) interface{} {
	if data == nil {
		return nil
	}
	// Delete may return raw bytes; decode so MCP clients get JSON values.
	if raw, ok := data.([]byte); ok {
		var value interface{}
		if json.Unmarshal(raw, &value) == nil {
			return value
		}
		return string(raw)
	}
	return data
}
