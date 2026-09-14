package services

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mcache-team/mcache/pkg/cluster"
	"github.com/mcache-team/mcache/pkg/services/mcp"
	"github.com/mcache-team/mcache/pkg/services/rest"
	"github.com/mcache-team/mcache/pkg/state"
	"github.com/sirupsen/logrus"
)

func Start() {
	cfg := cluster.ConfigFromEnv()
	if err := cluster.Bootstrap(cfg, state.DefaultStateMachine); err != nil {
		logrus.Fatalf("init cluster node failed: %v", err)
	}

	if isStdioMCPEnabled() {
		logrus.Printf("MCP stdio server started (node=%s mode=%s)", cfg.NodeID, cfg.Mode)
		if err := mcp.RunStdio(context.Background()); err != nil {
			logrus.Fatalf("MCP stdio server failed: %v", err)
		}
		return
	}

	gin.SetMode(gin.DebugMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	addAPIs(engine)
	logrus.Printf("Server is listening on %s (node=%s mode=%s advertise=%s mcp=%s)", cfg.HTTPAddr, cfg.NodeID, cfg.Mode, cfg.AdvertiseAddr, cfg.HTTPAddr+"/mcp")
	if err := engine.Run(cfg.HTTPAddr); err != nil {
		logrus.Fatalf("Start server failed: %v", err)
	}
}

// isStdioMCPEnabled reports whether the process was launched as a stdio MCP
// server, e.g. by an agent runtime such as Claude or Cursor.
func isStdioMCPEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("MCACHE_MCP_STDIO"))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func addAPIs(engine *gin.Engine) {
	engine.GET("/livez", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	engine.GET("/readyz", func(ctx *gin.Context) {
		if err := cluster.DefaultNode.Ready(); err != nil {
			ctx.JSON(http.StatusServiceUnavailable, gin.H{"status": "not-ready", "error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	engine.GET("/healthz", func(ctx *gin.Context) {
		if err := cluster.DefaultNode.Ready(); err != nil {
			ctx.JSON(http.StatusServiceUnavailable, gin.H{"status": "not-ready", "error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	engine.GET("/metrics", func(ctx *gin.Context) {
		diag, err := cluster.DefaultNode.Diagnostics()
		if err != nil {
			ctx.String(http.StatusInternalServerError, "# diagnostics error: %v\n", err)
			return
		}
		ctx.Data(http.StatusOK, "text/plain; version=0.0.4; charset=utf-8", []byte(cluster.BuildMetrics(diag)))
	})
	rest.Init(engine)
	engine.Any("/mcp", gin.WrapH(mcp.HTTPHandler()))
}
