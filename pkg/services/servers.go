package services

import (
	"github.com/gin-gonic/gin"
	"github.com/mcache-team/mcache/pkg/services/rest"
	"github.com/sirupsen/logrus"
	"net/http"
)

func Start() {
	gin.SetMode(gin.DebugMode)

	engine := gin.New()
	engine.Use(gin.Recovery())
	addAPIs(engine)
	addr := "0.0.0.0:8080"
	logrus.Printf("Server is listening on %s", addr)
	if err := engine.Run(addr); err != nil {
		logrus.Fatalf("Start server failed: %v", err)
	}
}

func addAPIs(engine *gin.Engine) {
	engine.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{})
	})
	rest.Init(engine)
}
