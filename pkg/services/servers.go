package services

import (
	"github.com/gin-gonic/gin"
	"github.com/mcache-team/mcache/pkg/cluster"
	"github.com/mcache-team/mcache/pkg/services/rest"
	"github.com/mcache-team/mcache/pkg/state"
	"github.com/sirupsen/logrus"
	"net/http"
)

func Start() {
	gin.SetMode(gin.DebugMode)

	cfg := cluster.ConfigFromEnv()
	if err := cluster.Bootstrap(cfg, state.DefaultStateMachine); err != nil {
		logrus.Fatalf("init cluster node failed: %v", err)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	addAPIs(engine)
	logrus.Printf("Server is listening on %s (node=%s mode=%s advertise=%s)", cfg.HTTPAddr, cfg.NodeID, cfg.Mode, cfg.AdvertiseAddr)
	if err := engine.Run(cfg.HTTPAddr); err != nil {
		logrus.Fatalf("Start server failed: %v", err)
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
}
